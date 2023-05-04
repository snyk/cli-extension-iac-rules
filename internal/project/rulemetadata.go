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

package project

// RuleMetadata contains all of the rule metadata fields that are supported for
// custom rules.
type RuleMetadata struct {
	ID           string   `json:"id"`
	Severity     string   `json:"severity"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Product      []string `json:"product"`
	Category     string   `json:"category,omitempty"`
	Labels       []string `json:"labels,omitempty"`
	Platform     []string `json:"platform,omitempty"`
	ServiceGroup string   `json:"service_group,omitempty"`
}
