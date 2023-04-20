package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/snyk/rest-go-libs/v5/jsonapi"
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

func (c *Client) CreateCustomRules(ctx context.Context, orgID string, targz []byte) (string, error) {
	url := fmt.Sprintf(
		"%s/hidden/orgs/%s/cloud/custom_rules?version=%s",
		c.url,
		orgID,
		version,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(targz))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	rsp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}

	expectedStatusCode := http.StatusCreated
	if rsp.StatusCode != expectedStatusCode {
		return "", fmt.Errorf("unexpected status code: %d", rsp.StatusCode)
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}
	var response jsonapi.ResourceDocument
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	return response.Data.ID, nil
}

func (c *Client) UpdateCustomRules(
	ctx context.Context,
	orgID string,
	customRulesID string,
	targz []byte,
) error {
	url := fmt.Sprintf(
		"%s/hidden/orgs/%s/cloud/custom_rules/%s?version=%s",
		c.url,
		orgID,
		customRulesID,
		version,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(targz))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	rsp, err := c.http.Do(req)
	if err != nil {
		return err
	}

	expectedStatusCode := http.StatusOK
	if rsp.StatusCode != expectedStatusCode {
		return fmt.Errorf("unexpected status code: %d", rsp.StatusCode)
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	var response jsonapi.ResourceDocument
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}
	return nil
}
