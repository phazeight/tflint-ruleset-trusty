// rules/trusty_reserved_variable_names_test.go
package rules_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/phazeight/tflint-ruleset-trusty/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestReservedVariableNamesRule(t *testing.T) {
	rule := rules.NewReservedVariableNamesRule()

	cases := []struct {
		name     string
		files    map[string]string
		expected helper.Issues
	}{
		{
			name: "valid — renamed away from the reserved name",
			files: map[string]string{
				"modules/aws/atoms/eks-cluster/inputs.tf": `
variable "kubernetes_version" {
  type = string
}`,
			},
			expected: helper.Issues{},
		},
		{
			name: "invalid — variable named version",
			files: map[string]string{
				"modules/aws/atoms/eks-cluster/inputs.tf": `
variable "version" {
  type = string
}`,
			},
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `variable "version" uses a name Terraform reserves in module blocks; the module cannot be initialised. Rename it (e.g. "kubernetes_version" for a resource's version attribute) and map it back in main.tf`,
					Range: hcl.Range{
						Filename: "modules/aws/atoms/eks-cluster/inputs.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 19},
					},
				},
			},
		},
		{
			name: "invalid — variable named source",
			files: map[string]string{
				"inputs.tf": `
variable "source" {
  type = string
}`,
			},
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `variable "source" uses a name Terraform reserves in module blocks; the module cannot be initialised. Rename it (e.g. "kubernetes_version" for a resource's version attribute) and map it back in main.tf`,
					Range: hcl.Range{
						Filename: "inputs.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			name: "valid — reserved word as a nested object field, not a variable name",
			files: map[string]string{
				"modules/aws/atoms/eks-node-group/inputs.tf": `
variable "launch_template" {
  type = object({
    name    = optional(string, null)
    version = string
  })
}`,
			},
			expected: helper.Issues{},
		},
		{
			name: "valid — reserved word as a resource attribute, not a variable name",
			files: map[string]string{
				"modules/aws/atoms/eks-cluster/main.tf": `
resource "aws_eks_cluster" "this" {
  version = var.kubernetes_version
}`,
			},
			expected: helper.Issues{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner := helper.TestRunner(t, tc.files)
			if err := rule.Check(runner); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			helper.AssertIssues(t, tc.expected, runner.Issues)
		})
	}
}
