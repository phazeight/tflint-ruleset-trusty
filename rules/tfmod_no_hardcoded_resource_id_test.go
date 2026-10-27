package rules_test

import (
	"testing"

	"github.com/phazeight/tflint-ruleset-trusty/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestNoHardcodedResourceIDRule(t *testing.T) {
	rule := rules.NewNoHardcodedResourceIDRule()

	cases := []struct {
		name     string
		content  string
		expected helper.Issues
	}{
		{
			name: "hardcoded VPC ID — warning",
			content: `
locals {
  vpc = "vpc-0a1b2c3d4e5f6a7b8"
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `"vpc-0a1b2c3d4e5f6a7b8" is a hardcoded resource ID — use a data source or resource reference instead`,
				},
			},
		},
		{
			name: "hardcoded subnet ID nested in a list — warning",
			content: `
locals {
  subnets = ["subnet-0a1b2c3d"]
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `"subnet-0a1b2c3d" is a hardcoded resource ID — use a data source or resource reference instead`,
				},
			},
		},
		{
			name: "reference to data source — pass",
			content: `
locals {
  vpc = data.aws_vpc.this.id
}`,
			expected: helper.Issues{},
		},
		{
			name: "unrelated prefixed string — pass",
			content: `
locals {
  name = "vpc-main"
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
