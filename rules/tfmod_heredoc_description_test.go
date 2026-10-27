package rules_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/phazeight/tflint-ruleset-trusty/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestHeredocDescriptionRule(t *testing.T) {
	rule := rules.NewHeredocDescriptionRule()

	cases := []struct {
		name     string
		content  string
		expected helper.Issues
	}{
		{
			name: "object with 2+ attrs and quoted description triggers error",
			content: `
variable "branch_name_pattern" {
  type = object({
    operator = string
    pattern  = string
    name     = optional(string)
  })
  default     = null
  description = "Restrict branch names. operator: starts_with | ends_with | contains | regex."
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `variable "branch_name_pattern": object type with ≥2 attributes must use <<-DOC heredoc for description`,
					Range: hcl.Range{
						Filename: "inputs.tf",
						Start:    hcl.Pos{Line: 9, Column: 3},
						End:      hcl.Pos{Line: 9, Column: 14},
					},
				},
			},
		},
		{
			name: "list(object) with 2+ attrs and quoted description triggers error",
			content: `
variable "bypass_actors" {
  type = list(object({
    actor_id    = number
    actor_type  = string
    bypass_mode = string
  }))
  default     = []
  description = "Actors that can bypass this ruleset."
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `variable "bypass_actors": object type with ≥2 attributes must use <<-DOC heredoc for description`,
					Range: hcl.Range{
						Filename: "inputs.tf",
						Start:    hcl.Pos{Line: 9, Column: 3},
						End:      hcl.Pos{Line: 9, Column: 14},
					},
				},
			},
		},
		{
			name: "object with 2+ attrs and heredoc description passes",
			content: `
variable "branch_name_pattern" {
  type = object({
    operator = string
    pattern  = string
    name     = optional(string)
  })
  default     = null
  description = <<-DOC
    Restrict branch names.

    operator: starts_with | ends_with | contains | regex.
    pattern: the pattern to match.
    name: optional display name.
  DOC
}`,
			expected: helper.Issues{},
		},
		{
			name: "object with 1 attr and quoted description passes",
			content: `
variable "single_attr" {
  type = object({
    name = string
  })
  default     = null
  description = "Single attribute object."
}`,
			expected: helper.Issues{},
		},
		{
			name: "flat string variable passes",
			content: `
variable "bucket" {
  type        = string
  description = "The name of the bucket."
}`,
			expected: helper.Issues{},
		},
		{
			name: "flat bool variable passes",
			content: `
variable "encryption_enabled" {
  type        = bool
  default     = true
  description = "Enable encryption."
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
