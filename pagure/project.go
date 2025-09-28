package pagure

import mapset "github.com/deckarep/golang-set/v2"

type Project struct {
	AccessGroups AccessGroups         `json:"access_groups"`
	AccessUsers  AccessUsers          `json:"access_users"`
	CloseStatus  []string             `json:"close_status"`
	CustomKeys   []any                `json:"custom_keys"`
	DateCreated  string               `json:"date_created"`
	DateModified string               `json:"date_modified"`
	Description  string               `json:"description"`
	FullURL      string               `json:"full_url"`
	Fullname     string               `json:"fullname"`
	ID           int                  `json:"id"`
	Milestones   map[string]Milestone `json:"milestones"`
	Name         string               `json:"name"`
	Namespace    string               `json:"namespace"`
	Parent       any                  `json:"parent"`
	Priorities   map[string]string    `json:"priorities"`
	Tags         []string             `json:"tags"`
	URLPath      string               `json:"url_path"`
	User         User                 `json:"user"`
}
type AccessGroups struct {
	Admin        []string `json:"admin"`
	Collaborator []string `json:"collaborator"`
	Commit       []string `json:"commit"`
	Ticket       []string `json:"ticket"`
}
type AccessUsers struct {
	Admin        []string `json:"admin"`
	Collaborator []string `json:"collaborator"`
	Commit       []string `json:"commit"`
	Owner        []string `json:"owner"`
	Ticket       []string `json:"ticket"`
}
type Milestone struct {
	Active bool    `json:"active"`
	Date   *string `json:"date"`
}

type User struct {
	FullURL  string `json:"full_url"`
	Fullname string `json:"fullname"`
	Name     string `json:"name"`
	URLPath  string `json:"url_path"`
}

func (c *Client) GetProject(project string) (*Project, error) {
	var data Project
	err := c.get(&data, c.URL.JoinPath("api/0", project))
	return &data, err
}

type Contributors struct {
	Users  ContributorsRoles `json:"users"`
	Groups ContributorsRoles `json:"groups"`
}

type ContributorsRoles struct {
	Admin         []string                   `json:"admin"`
	Collaborators []ContributorsCollaborator `json:"collaborators"`
	Commit        []string                   `json:"commit"`
	Ticket        []string                   `json:"ticket"`
}

type ContributorsCollaborator struct {
	Branches string `json:"branches"`
	User     string `json:"user"`
}

func (c *Client) GetContributors(project string) (*Contributors, error) {
	var data Contributors
	err := c.get(&data, c.URL.JoinPath("api/0", project, "contributors"))
	return &data, err
}

// GetAllMaints is a helper function to get a list of all project maintainers.
// Groups are @-prefixed.
// EXPERIMENTAL!
func (c *Client) GetAllMaints(project string, includeGroups bool) ([]string, error) {
	data, err := c.GetContributors(project)
	if err != nil {
		return []string{}, err
	}
	s := mapset.NewThreadUnsafeSet[string]()
	s.Append(data.Users.Admin...)
	s.Append(data.Users.Commit...)
	for _, c := range data.Users.Collaborators {
		s.Add(c.User)
	}
	s.Append(data.Users.Ticket...)
	if includeGroups {
		for _, g := range data.Groups.Admin {
			s.Add("@" + g)
		}
		for _, g := range data.Groups.Commit {
			s.Add("@" + g)
		}
		for _, c := range data.Groups.Collaborators {
			s.Add("@" + c.User)
		}
		for _, g := range data.Groups.Ticket {
			s.Add("@" + g)
		}
	}
	return s.ToSlice(), nil
}
