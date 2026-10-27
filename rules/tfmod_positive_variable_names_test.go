package rules_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/phazeight/tflint-ruleset-trusty/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestPositiveVariableNamesRule(t *testing.T) {
	rule := rules.NewPositiveVariableNamesRule()

	cases := []struct {
		name     string
		content  string
		expected helper.Issues
	}{
		{
			name: "disable_ prefix on bool triggers error",
			content: `
variable "disable_encryption" {
  type        = bool
  default     = false
  description = "Disable encryption."
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `variable "disable_encryption": use positive naming (avoid "disable_" prefix)`,
					Range: hcl.Range{
						Filename: "inputs.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 30},
					},
				},
			},
		},
		{
			name: "no_ prefix on bool triggers error",
			content: `
variable "no_logging" {
  type        = bool
  default     = false
  description = "Disable logging."
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `variable "no_logging": use positive naming (avoid "no_" prefix)`,
					Range: hcl.Range{
						Filename: "inputs.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 22},
					},
				},
			},
		},
		{
			name: "not_ prefix on bool triggers error",
			content: `
variable "not_enabled" {
  type        = bool
  default     = false
  description = "Not enabled."
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `variable "not_enabled": use positive naming (avoid "not_" prefix)`,
					Range: hcl.Range{
						Filename: "inputs.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 23},
					},
				},
			},
		},
		{
			name: "prevent_ prefix on bool triggers error",
			content: `
variable "prevent_deletion" {
  type        = bool
  default     = false
  description = "Prevent deletion."
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `variable "prevent_deletion": use positive naming (avoid "prevent_" prefix)`,
					Range: hcl.Range{
						Filename: "inputs.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 28},
					},
				},
			},
		},
		{
			name: "positive bool name passes",
			content: `
variable "encryption_enabled" {
  type        = bool
  default     = true
  description = "Enable encryption."
}`,
			expected: helper.Issues{},
		},
		{
			name: "disable_ prefix on string passes",
			content: `
variable "disable_message" {
  type        = string
  default     = "disabled"
  description = "Disable message."
}`,
			expected: helper.Issues{},
		},
		{
			name: "no type attribute passes",
			content: `
variable "no_type_here" {
  description = "No type."
}`,
			expected: helper.Issues{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"inputs.tf": tc.content})
			if err := rule.Check(runner); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			helper.AssertIssues(t, tc.expected, runner.Issues)
		})
	}
}
