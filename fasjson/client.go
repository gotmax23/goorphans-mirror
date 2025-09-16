package fas2email

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/ubccr/kerby/khttp"
	"go.gtmx.me/goorphans/common"
)

type Client struct {
	Client *http.Client
	URL    *url.URL
}

type userResult struct {
	Result User `json:"result"`
}

type User struct {
	Certificates   []string `json:"certificates"`
	Creation       string   `json:"creation"`
	Emails         []string `json:"emails"`
	GithubUsername string   `json:"github_username"`
	GitlabUsername string   `json:"gitlab_username"`
	Givenname      string   `json:"givenname"`
	Gpgkeyids      []string `json:"gpgkeyids"`
	HumanName      string   `json:"human_name"`
	Ircnicks       []string `json:"ircnicks"`
	IsPrivate      bool     `json:"is_private"`
	Locale         string   `json:"locale"`
	Locked         bool     `json:"locked"`
	Pronouns       []string `json:"pronouns"`
	Rhbzemail      string   `json:"rhbzemail"`
	Rssurl         string   `json:"rssurl"`
	Rssurls        []string `json:"rssurls"`
	Sshpubkeys     []string `json:"sshpubkeys"`
	Surname        string   `json:"surname"`
	Timezone       string   `json:"timezone"`
	URI            string   `json:"uri"`
	Username       string   `json:"username"`
	Website        string   `json:"website"`
	Websites       []string `json:"websites"`
}

type membersResult struct {
	Result []Member `json:"result"`
}

type Member struct {
	URI      string `json:"uri"`
	Username string `json:"username"`
}

// NewClient creates a FASJSON Client using Kerberos authentication
func NewClient() *Client {
	client := &http.Client{Transport: &khttp.Transport{}}
	uri, _ := url.Parse("https://fasjson.fedoraproject.org")
	return &Client{client, uri}
}

func (c *Client) do(dest any, urlparts ...string) error {
	path := c.URL.JoinPath(urlparts...)
	log.Printf("GET %s\n", path.String())
	resp, err := c.Client.Get(path.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := common.CheckStatusCode(resp); err != nil {
		return err
	}
	err = json.NewDecoder(resp.Body).Decode(&dest)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetUser(username string) (*User, error) {
	var result userResult
	username = url.PathEscape(username)
	err := c.do(&result, "v1/users", username)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

func (c *Client) GetMembers(groupname string) ([]Member, error) {
	var result membersResult
	groupname = url.PathEscape(groupname)
	err := c.do(&result, "v1/groups", groupname, "members")
	if err != nil {
		return nil, err
	}
	return result.Result, nil
}
