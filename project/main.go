package project

import "fmt"

const (
	// Name is the name of the plugin.
	Name = "trusty"

	// Version is the version of the plugin.
	Version = "0.2.0"
)

// RuleName returns the full rule name with the plugin prefix.
func RuleName(name string) string {
	return fmt.Sprintf("%s_%s", Name, name)
}

// ReferenceLink returns the documentation link for the given rule.
func ReferenceLink(ruleName string) string {
	return fmt.Sprintf(
		"https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/%s.md",
		ruleName,
	)
}
