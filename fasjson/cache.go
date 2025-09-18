package fasjson

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"iter"
	"slices"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schema string

var DefaultTTL = (time.Hour * 24).Seconds()

// EmailCacheClient is a wrapper around FASJSON that caches username and group
// -> email mappings.
// Each function caches its result in the database
type EmailCacheClient struct {
	db         *sql.DB
	Client     *Client
	TTLSeconds float64
}

// Clean entries greater than TTL
func (cache *EmailCacheClient) Clean() error {
	_, err := cache.db.Exec(`
		DELETE FROM fas_user WHERE (cache_time + ?) <= unixepoch('now','subsec');
		DELETE FROM fas_group WHERE (cache_time + ?) <= unixepoch('now','subsec');
	`, cache.TTLSeconds, cache.TTLSeconds)
	if err != nil {
		return err
	}
	return nil
}

func OpenCacheDB(filename string, ttl float64) (*EmailCacheClient, error) {
	db, err := sql.Open("sqlite3", filename+"?_foreign_keys=1")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	cache := EmailCacheClient{db, NewClient(), ttl}
	// cache.Clean()
	return &cache, nil
}

func (cache *EmailCacheClient) queryUserEmail(username string) (string, error) {
	var email string
	err := cache.db.QueryRow(`
		SELECT email FROM fas_user
		WHERE user_name = ? AND (cache_time + ?) > unixepoch('now','subsec')
	`, username, cache.TTLSeconds).
		Scan(&email)
	if err != nil {
		return "", err
	}
	return email, nil
}

func (cache *EmailCacheClient) insertUserEmail(username string, email string) error {
	_, err := cache.db.Exec(`
		INSERT OR REPLACE INTO fas_user (user_name, email, cache_time)
		VALUES (?, ?, unixepoch('now','subsec'));
	`, username, email, cache.TTLSeconds)
	if err != nil {
		return err
	}
	return nil
}

// GetUserEmail gets the email for a user.
func (cache *EmailCacheClient) GetUserEmail(username string) (string, error) {
	result, err := cache.queryUserEmail(username)
	if err == nil {
		return result, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	user, err := cache.Client.GetUser(username)
	if err != nil {
		return "", err
	}
	result = user.Emails[0]
	err = cache.insertUserEmail(username, user.Emails[0])
	return result, err
}

func (cache *EmailCacheClient) queryMembers(groupname string) ([]string, error) {
	results := []string{}
	tsx, err := cache.db.Begin()
	defer tsx.Rollback()
	if err != nil {
		return results, err
	}
	err = tsx.QueryRow(`
		SELECT group_name FROM fas_group
		WHERE group_name = ? AND (cache_time + ?) > unixepoch('now','subsec');
	`, groupname, cache.TTLSeconds).Scan(&groupname)
	if err != nil {
		return results, err
	}
	rows, err := cache.db.Query(`
		SELECT user_name FROM group_member
		WHERE group_name = ?;
	`, groupname)
	if err != nil {
		return results, err
	}
	defer rows.Close()
	var user string
	for rows.Next() {
		err = rows.Scan(&user)
		if err != nil {
			return results, err
		}
		results = append(results, user)
	}
	err = tsx.Commit()
	return results, err
}

func (cache *EmailCacheClient) insertMembers(groupname string, members []string) error {
	tx, err := cache.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete any old group members
	_, err = tx.Exec(`DELETE FROM group_member WHERE group_name = ?;`, groupname)
	if err != nil {
		return err
	}
	// Create group
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO fas_group (group_name, cache_time)
		VALUES (?, unixepoch('now','subsec'));
	`, groupname)
	if err != nil {
		return err
	}

	// Add group members
	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO group_member (group_name, user_name)
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	for _, member := range members {
		_, err = stmt.Exec(groupname, member)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	return err
}

// GetMembers returns a slice of member usernames for a given group.
func (cache *EmailCacheClient) GetMembers(groupname string) ([]string, error) {
	qresult, err := cache.queryMembers(groupname)
	if err == nil || !errors.Is(err, sql.ErrNoRows) {
		return qresult, err
	}

	// Otherwise, request members again
	members, err := cache.Client.GetMembers(groupname)
	if err != nil {
		return []string{}, err
	}
	memberstrings := make([]string, 0, len(members))
	for _, member := range members {
		memberstrings = append(memberstrings, member.Username)
	}
	if err = cache.insertMembers(groupname, memberstrings); err != nil {
		return memberstrings, err
	}
	return memberstrings, nil
}

// GetUserIterEmailsMap returns a map of username->email for multiple usernames.
func (cache *EmailCacheClient) GetUserIterEmailsMap(
	usernames iter.Seq[string],
) (map[string]string, error) {
	result := map[string]string{}
	for username := range usernames {
		email, err := cache.GetUserEmail(username)
		if err != nil {
			return result, err
		}
		result[username] = email
	}
	return result, nil
}

func (cache *EmailCacheClient) GetMemberEmailsMap(
	groupname string,
) (map[string]string, error) {
	// Function not currently used by the CLI code
	members, err := cache.GetMembers(groupname)
	if err != nil {
		return map[string]string{}, fmt.Errorf(
			"failed to get group member list: %v",
			err,
		)
	}
	emails, err := cache.GetUserIterEmailsMap(slices.Values(members))
	if err != nil {
		return emails, fmt.Errorf("failed to get user emails: %v", err)
	}
	return emails, nil
}

// GetAllEmailsMap returns a map of username -> emails.
// Names that start with "@" are treated as group names.
func (cache *EmailCacheClient) GetAllEmailsMap(names []string) (map[string]string, error) {
	// Use a custom set type. We can have users repeated in names and the same
	// user present in mutliple groups.
	usernames := mapset.NewThreadUnsafeSetWithSize[string](len(names))
	for _, name := range names {
		group, found := strings.CutPrefix(name, "@")
		if found {
			members, err := cache.GetMembers(group)
			if err != nil {
				return map[string]string{}, err
			}
			usernames.Append(members...)
		} else {
			usernames.Add(name)
		}
	}
	return cache.GetUserIterEmailsMap(mapset.Elements(usernames))
}
