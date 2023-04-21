package forms

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var ErrInvalidRuleID = errors.New("invalid rule ID")

const ruleIDMaxLength = 64
const ruleIDMinLength = 1

var ruleIDPrefixCharset = regexp.MustCompile(`^[A-Za-z]`)
var ruleIDCharset = regexp.MustCompile(`^[A-Za-z0-9-_]*$`)
var ruleIDReservedPrefixes = []string{"SNYK_", "SNYK-", "FG_R"}

func validateRuleID(ruleID string) error {
	if len(ruleID) > ruleIDMaxLength {
		return fmt.Errorf("%w: exceeded max length of %d characters", ErrInvalidRuleID, ruleIDMaxLength)
	}
	if len(ruleID) < ruleIDMinLength {
		return fmt.Errorf("%w: must contain at least %d characters", ErrInvalidRuleID, ruleIDMinLength)
	}
	if !ruleIDPrefixCharset.MatchString(ruleID) {
		return fmt.Errorf("%w: must start with a letter", ErrInvalidRuleID)
	}
	if !ruleIDCharset.MatchString(ruleID) {
		return fmt.Errorf("%w: must only contain letters, numbers, dashes (-), or underscores (_)", ErrInvalidRuleID)
	}
	for _, p := range ruleIDReservedPrefixes {
		if strings.HasPrefix(ruleID, p) {
			return fmt.Errorf("%w: has reserved prefix '%s'", ErrInvalidRuleID, p)
		}
	}
	return nil
}
