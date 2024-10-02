package service

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

const bundleId = "bundle-id"
const orgId = "org-id"

func TestCreateCustomRules(t *testing.T) {
	tests := []struct {
		name            string
		response        string
		status          int
		expectedPath    string
		expectedVersion string
		expectedResult  string
		expectedError   error
		iacNewEngine    bool
	}{
		{
			name:            "old API success",
			response:        fmt.Sprintf(`{"data":{"id": "%s"}}`, bundleId),
			status:          http.StatusCreated,
			expectedPath:    fmt.Sprintf(`/rest/orgs/%s/cloud/rule_bundles`, orgId),
			expectedVersion: "2023-05-22~experimental",
			expectedResult:  bundleId,
		},
		{
			name:            "new API success",
			response:        fmt.Sprintf(`{"data":{"id": "%s"}}`, bundleId),
			status:          http.StatusCreated,
			expectedPath:    fmt.Sprintf(`/hidden/orgs/%s/cloud/rule_bundles`, orgId),
			expectedVersion: "2024-09-24~beta",
			expectedResult:  bundleId,
			iacNewEngine:    true,
		},
		{
			name:            "forbidden",
			response:        `{"errors":[{"status":"403","detail":"Forbidden"}]}`,
			status:          http.StatusForbidden,
			expectedPath:    fmt.Sprintf(`/rest/orgs/%s/cloud/rule_bundles`, orgId),
			expectedVersion: "2023-05-22~experimental",
			expectedError:   errors.New("403 : Forbidden"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.expectedPath, r.URL.Path)
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, tt.expectedVersion, r.URL.Query().Get("version"))

				w.WriteHeader(tt.status)
				w.Write([]byte(tt.response))
				w.Header().Add("Content-Type", "application/json")
			}))

			defer server.Close()

			client := NewClient(
				server.Client(),
				server.URL,
				tt.iacNewEngine,
			)

			result, err := client.CreateCustomRules(context.Background(), orgId, make([]byte, 0))
			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestUpdateCustomRules(t *testing.T) {
	tests := []struct {
		name            string
		response        string
		status          int
		expectedPath    string
		expectedVersion string
		expectedError   error
		iacNewEngine    bool
	}{
		{
			name:            "old API success",
			response:        fmt.Sprintf(`{"data":{"id": "%s"}}`, bundleId),
			status:          http.StatusOK,
			expectedPath:    fmt.Sprintf(`/rest/orgs/%s/cloud/rule_bundles/%s`, orgId, bundleId),
			expectedVersion: "2023-05-22~experimental",
		},
		{
			name:            "new API success",
			response:        fmt.Sprintf(`{"data":{"id": "%s"}}`, bundleId),
			status:          http.StatusOK,
			expectedPath:    fmt.Sprintf(`/hidden/orgs/%s/cloud/rule_bundles/%s`, orgId, bundleId),
			expectedVersion: "2024-09-24~beta",
			iacNewEngine:    true,
		},
		{
			name:            "forbidden",
			response:        `{"errors":[{"status":"403","detail":"Forbidden"}]}`,
			status:          http.StatusForbidden,
			expectedPath:    fmt.Sprintf(`/rest/orgs/%s/cloud/rule_bundles/%s`, orgId, bundleId),
			expectedVersion: "2023-05-22~experimental",
			expectedError:   errors.New("403 : Forbidden"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.expectedPath, r.URL.Path)
				require.Equal(t, http.MethodPatch, r.Method)
				require.Equal(t, tt.expectedVersion, r.URL.Query().Get("version"))

				w.WriteHeader(tt.status)
				w.Write([]byte(tt.response))
				w.Header().Add("Content-Type", "application/json")
			}))

			defer server.Close()

			client := NewClient(
				server.Client(),
				server.URL,
				tt.iacNewEngine,
			)

			err := client.UpdateCustomRules(context.Background(), orgId, bundleId, make([]byte, 0))
			require.Equal(t, tt.expectedError, err)
		})
	}
}

func TestDeleteCustomRules(t *testing.T) {
	tests := []struct {
		name            string
		response        string
		status          int
		expectedPath    string
		expectedVersion string
		expectedError   error
		iacNewEngine    bool
	}{
		{
			name:            "old API success",
			response:        fmt.Sprintf(`{"data":{"id": "%s"}}`, bundleId),
			status:          http.StatusNoContent,
			expectedPath:    fmt.Sprintf(`/rest/orgs/%s/cloud/rule_bundles/%s`, orgId, bundleId),
			expectedVersion: "2023-05-22~experimental",
		},
		{
			name:            "new API success",
			response:        fmt.Sprintf(`{"data":{"id": "%s"}}`, bundleId),
			status:          http.StatusNoContent,
			expectedPath:    fmt.Sprintf(`/hidden/orgs/%s/cloud/rule_bundles/%s`, orgId, bundleId),
			expectedVersion: "2024-09-24~beta",
			iacNewEngine:    true,
		},
		{
			name:            "forbidden",
			response:        `{"errors":[{"status":"403","detail":"Forbidden"}]}`,
			status:          http.StatusForbidden,
			expectedPath:    fmt.Sprintf(`/rest/orgs/%s/cloud/rule_bundles/%s`, orgId, bundleId),
			expectedVersion: "2023-05-22~experimental",
			expectedError:   errors.New("403 : Forbidden"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.expectedPath, r.URL.Path)
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(t, tt.expectedVersion, r.URL.Query().Get("version"))

				w.WriteHeader(tt.status)
				w.Write([]byte(tt.response))
				w.Header().Add("Content-Type", "application/json")
			}))

			defer server.Close()

			client := NewClient(
				server.Client(),
				server.URL,
				tt.iacNewEngine,
			)

			err := client.DeleteCustomRules(context.Background(), orgId, bundleId)
			require.Equal(t, tt.expectedError, err)
		})
	}
}
