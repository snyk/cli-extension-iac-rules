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

// jSONAPI describes a service's implementation of the JSON API specification.
type jSONAPI struct {
	Version string `json:"version"`
}

type resourceDocument struct {
	JSONAPI jSONAPI        `json:"jsonapi"`
	Meta    meta           `json:"meta,omitempty"`
	Data    resourceObject `json:"data"`
}

// resourceObject represents a resource object, as defined in
// https://jsonapi.org/format/#document-resource-objects.
type resourceObject struct {
	ID         string     `json:"id,omitempty"`
	Type       string     `json:"type"`
	Attributes attributes `json:"attributes,omitempty"`
	Meta       meta       `json:"meta,omitempty"`
}

type attributes interface{}

// errorDocument represents a JSON API error document,
type errorDocument struct {
	JSONAPI jSONAPI       `json:"jsonapi"`
	Errors  []errorObject `json:"errors"`
}

// errorObject represents a JSON API error object, as defined in
// https://jsonapi.org/format/#error-objects.
type errorObject struct {
	Status string `json:"status"`
	Detail string `json:"detail"`

	ID     string       `json:"id,omitempty"`
	Code   string       `json:"code,omitempty"`
	Title  string       `json:"title,omitempty"`
	Source *errorSource `json:"source,omitempty"`
	Meta   meta         `json:"meta,omitempty"`
}

// errorSource references the source of the error in the request.
type errorSource struct {
	Pointer   string `json:"pointer"`
	Parameter string `json:"parameter"`
}

// meta represents non-standard meta-information as defined in
// https://jsonapi.org/format/#document-meta.
type meta map[string]interface{}
