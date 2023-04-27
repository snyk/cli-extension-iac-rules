package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	var response resourceDocument
	if err := parseResponse(rsp, http.StatusCreated, &response); err != nil {
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
	var response resourceDocument
	return parseResponse(rsp, http.StatusOK, &response)
}

func parseResponse(rsp *http.Response, expectedStatusCode int, expectedDocument interface{}) error {
	body, err := io.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return err
	}

	if rsp.StatusCode != expectedStatusCode {
		var errorDoc errorDocument
		if err := json.Unmarshal(body, &errorDoc); err != nil {
			return fmt.Errorf("response %d: %s", rsp.StatusCode, err)
		}
		return fmt.Errorf("%s", errorDocumentToString(errorDoc))
	}
	return json.Unmarshal(body, expectedDocument)
}

func errorDocumentToString(err errorDocument) string {
	msgs := []string{}
	if len(err.Errors) == 0 {
		msgs = append(msgs, "unknown error")
	} else {
		for _, obj := range err.Errors {
			msgs = append(msgs, errorObjectToString(obj))
		}
	}
	return strings.Join(msgs, "\n")
}

func errorObjectToString(err errorObject) string {
	return fmt.Sprintf("%s %s: %s", err.Status, err.Title, err.Detail)
}