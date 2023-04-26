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
