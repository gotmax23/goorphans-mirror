package pagure

// TODO: Pagination
// TODO: The rest of the endpoints as needed

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
	"go.gtmx.me/goorphans/common"
)

type pagination struct {
	First   string `json:"first"`
	Last    string `json:"last"`
	Next    string `json:"next"`
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Prev    string `json:"prev"`
}

type groups struct {
	Groups     []string    `json:"groups"`
	Pagination *pagination `json:"pagination"`
}

type Client struct {
	URL    *url.URL
	Client *http.Client
}

func NewClient(url *url.URL, client *http.Client) *Client {
	if client == nil {
		// Use retryablehttp for Pagure by default.
		// The service is kind of flaky.
		client = retryablehttp.NewClient().StandardClient()
	}
	return &Client{url, client}
}

func (c *Client) get(dest any, path *url.URL) error {
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

func (c *Client) GetGroups() ([]string, error) {
	var g groups
	url := c.URL.JoinPath("api/0/groups")
	// TODO: Pagination. Set this to the maximum for now.
	url.RawQuery = "per_page=100"
	err := c.get(&g, url)
	if err != nil {
		return []string{}, err
	}
	return g.Groups, nil
}
