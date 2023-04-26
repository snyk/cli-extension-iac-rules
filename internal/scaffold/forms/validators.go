package forms

import (
	"fmt"
	"regexp"
	"strings"
)

var verboseValidationTemplate = `
{{- Bold .Prompt }} {{ .Input -}}
{{- if .ValidationError }} {{ Foreground "1" (Bold "✘") }} {{ .ValidationError }}
{{- else }} {{ Foreground "2" (Bold "✔") }}
{{- end -}}
`

const ruleIDMaxLength = 64

var ruleIDPrefixCharset = regexp.MustCompile(`^[A-Za-z]`)
var ruleIDCharset = regexp.MustCompile(`^[A-Za-z0-9-_]*$`)
var ruleIDReservedPrefixes = []string{"SNYK_", "SNYK-", "FG_R"}

func ruleIDValidator(existingIDs []string, existingDirs []string) func(string) error {
	return func(ruleID string) error {
		if len(ruleID) < 1 {
			// This is a workaround to disallow empty inputs but only show an error
			// message when the user has typed something invalid. The input value
			// that's available in the template includes the blinking cursor, so
			// it's difficult to test whether or not it's empty.
			return fmt.Errorf("")
		}
		if len(ruleID) > ruleIDMaxLength {
			return fmt.Errorf("rule ID exceeds max length of %d characters", ruleIDMaxLength)
		}
		if !ruleIDPrefixCharset.MatchString(ruleID) {
			return fmt.Errorf("rule ID must start with a letter")
		}
		if !ruleIDCharset.MatchString(ruleID) {
			return fmt.Errorf("rule ID must only contain letters, numbers, dashes (-), or underscores (_)")
		}
		for _, p := range ruleIDReservedPrefixes {
			if strings.HasPrefix(ruleID, p) {
				return fmt.Errorf("rule ID has reserved prefix '%s'", p)
			}
		}
		for _, existing := range existingIDs {
			if ruleID == existing {
				return fmt.Errorf("rule with ID %s already exists in this project", existing)
			}
		}
		for _, existing := range existingDirs {
			if ruleID == existing {
				return fmt.Errorf("rule with directory %s already exists in this project", existing)
			}
		}
		return nil
	}
}
