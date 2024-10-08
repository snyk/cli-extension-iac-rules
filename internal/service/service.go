// © 2023 Snyk Limited All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

const versionLegacyApi = "2023-05-22~experimental"
const version = "2024-09-24~beta"

type Client struct {
	http *http.Client
	url  string
	iacNewEngine bool
}

func NewClient(http *http.Client, url string, iacNewEngine bool) *Client {
	return &Client{
		http: http,
		url:  url,
		iacNewEngine: iacNewEngine,
	}
}

func (c *Client) CreateCustomRules(ctx context.Context, orgID string, targz []byte) (string, error) {
	url := fmt.Sprintf(
		"%s/rest/orgs/%s/cloud/rule_bundles?version=%s",
		c.url,
		orgID,
		versionLegacyApi,
	)

	if (c.iacNewEngine) {
		url = fmt.Sprintf(
			"%s/hidden/orgs/%s/cloud/rule_bundles?version=%s",
			c.url,
			orgID,
			version,
		)
	}

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
		"%s/rest/orgs/%s/cloud/rule_bundles/%s?version=%s",
		c.url,
		orgID,
		customRulesID,
		versionLegacyApi,
	)

	if (c.iacNewEngine) {
		url = fmt.Sprintf(
			"%s/hidden/orgs/%s/cloud/rule_bundles/%s?version=%s",
			c.url,
			orgID,
			customRulesID,
			version,
		)
	}

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

func (c *Client) DeleteCustomRules(
	ctx context.Context,
	orgID string,
	customRulesID string,
) error {
	url := fmt.Sprintf(
		"%s/rest/orgs/%s/cloud/rule_bundles/%s?version=%s",
		c.url,
		orgID,
		customRulesID,
		versionLegacyApi,
	)

	if (c.iacNewEngine) {
		url = fmt.Sprintf(
			"%s/hidden/orgs/%s/cloud/rule_bundles/%s?version=%s",
			c.url,
			orgID,
			customRulesID,
			version,
		)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, http.NoBody)
	if err != nil {
		return err
	}
	rsp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	return parseResponse(rsp, http.StatusNoContent, nil)
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
			// If the error is not encoded as JSON, that is less important a detail to
			// surface to the user than the actual content of the error. Notably, this
			// can occur when cerberus bounces the request, as it returns plain text
			// bodies.
			return fmt.Errorf("response %d: %s", rsp.StatusCode, string(body))
		}
		return fmt.Errorf("%s", errorDocumentToString(errorDoc))
	}
	if expectedDocument != nil {
		return json.Unmarshal(body, expectedDocument)
	}
	return nil
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
