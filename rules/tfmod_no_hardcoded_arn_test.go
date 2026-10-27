package rules_test

import (
	"testing"

	"github.com/phazeight/tflint-ruleset-trusty/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestNoHardcodedARNRule(t *testing.T) {
	rule := rules.NewNoHardcodedARNRule()

	cases := []struct {
		name     string
		content  string
		expected helper.Issues
	}{
		{
			name: "hardcoded S3 ARN — warning",
			content: `
locals {
  bucket_arn = "arn:aws:s3:::my-bucket"
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `"arn:aws:s3:::my-bucket" is a hardcoded ARN — reference the owning resource's .arn attribute, or a variable/data source instead`,
				},
			},
		},
		{
			name: "hardcoded gov-cloud ARN — warning",
			content: `
locals {
  role_arn = "arn:aws-us-gov:iam::123456789012:role/my-role"
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `"arn:aws-us-gov:iam::123456789012:role/my-role" is a hardcoded ARN — reference the owning resource's .arn attribute, or a variable/data source instead`,
				},
			},
		},
		{
			name: "AWS-managed IAM ARN — pass (baked exception)",
			content: `
locals {
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
}`,
			expected: helper.Issues{},
		},
		{
			name: "reference to resource arn — pass",
			content: `
locals {
  bucket_arn = aws_s3_bucket.this.arn
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
