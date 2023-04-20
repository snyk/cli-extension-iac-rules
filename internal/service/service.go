package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

const version = "2022-12-21~beta"

type Client struct {
	http *http.Client
	url  string
}

func NewClient(http *http.Client, url string) *Client {
	return &Client{
		http: http,
		url:  url,
	}
}

func (c *Client) CustomRules(ctx context.Context, orgID string) error {
	url := fmt.Sprintf(
		"%s/hidden/orgs/%s/cloud/custom_rules?version=%s",
		c.url,
		orgID,
		version,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return err
	}
	rsp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Status Code: %d\n", rsp.StatusCode)
	return nil
}
