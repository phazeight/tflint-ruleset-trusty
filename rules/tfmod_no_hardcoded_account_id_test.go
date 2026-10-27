package rules_test

import (
	"testing"

	"github.com/phazeight/tflint-ruleset-trusty/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestNoHardcodedAccountIDRule(t *testing.T) {
	rule := rules.NewNoHardcodedAccountIDRule()

	cases := []struct {
		name     string
		content  string
		expected helper.Issues
	}{
		{
			name: "12-digit literal in attribute — warning",
			content: `
locals {
  account = "123456789012"
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `"123456789012" looks like a hardcoded AWS account ID — use a variable, data.aws_caller_identity, or an account-id map input instead`,
				},
			},
		},
		{
			name: "12-digit literal nested in a list — warning",
			content: `
locals {
  accounts = ["123456789012"]
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `"123456789012" looks like a hardcoded AWS account ID — use a variable, data.aws_caller_identity, or an account-id map input instead`,
				},
			},
		},
		{
			name: "reference to caller identity — pass",
			content: `
locals {
  account = data.aws_caller_identity.current.account_id
}`,
			expected: helper.Issues{},
		},
		{
			name: "interpolated string — pass",
			content: `
locals {
  arn = "arn:aws:iam::${var.account_id}:role/x"
}`,
			expected: helper.Issues{},
		},
		{
			name: "not 12 digits — pass",
			content: `
locals {
  short = "12345"
}`,
			expected: helper.Issues{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"main.tf": tc.content})
			if err := rule.Check(runner); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			helper.AssertIssuesWithoutRange(t, tc.expected, runner.Issues)
		})
	}
}
