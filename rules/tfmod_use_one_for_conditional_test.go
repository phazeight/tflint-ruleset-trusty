package rules_test

import (
	"testing"

	"github.com/phazeight/tflint-ruleset-trusty/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestUseOneForConditionalRule(t *testing.T) {
	rule := rules.NewUseOneForConditionalRule()

	cases := []struct {
		name     string
		content  string
		expected helper.Issues
	}{
		{
			name: "aws_instance[0].id in locals — error",
			content: `
locals {
  instance_id = aws_instance.web[0].id
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `use one() instead of [0] indexing on "aws_instance.web" — one(aws_instance.web[*].id) returns null safely when count = 0`,
				},
			},
		},
		{
			name: "data.aws_route53_zone[0].zone_id in resource attribute — error",
			content: `
resource "aws_route53_record" "example" {
  zone_id = data.aws_route53_zone.this[0].zone_id
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `use one() instead of [0] indexing on "data.aws_route53_zone.this" — one(data.aws_route53_zone.this[*].zone_id) returns null safely when count = 0`,
				},
			},
		},
		{
			name: "module.foo[0].attr — error",
			content: `
locals {
  ruleset_id = module.merge_queue_ruleset[0].id
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `use one() instead of [0] indexing on "module.merge_queue_ruleset" — one(module.merge_queue_ruleset[*].id) returns null safely when count = 0`,
				},
			},
		},
		{
			name: "module.foo[0] with no attribute — error",
			content: `
locals {
  ruleset = module.merge_queue_ruleset[0]
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `use one() instead of [0] indexing on "module.merge_queue_ruleset" — one(module.merge_queue_ruleset[*]) returns null safely when count = 0`,
				},
			},
		},
		{
			name: "list-attribute index — pass (index is not on the resource)",
			content: `
locals {
  root_id = aws_organizations_organization.root.roots[0].id
}`,
			expected: helper.Issues{},
		},
		{
			name: "data source list-attribute index — pass",
			content: `
locals {
  cidr = data.aws_vpc.this.cidr_block_associations[0].cidr_block
}`,
			expected: helper.Issues{},
		},
		{
			name: "moved block address — pass (one() invalid in moved)",
			content: `
moved {
  from = module.ecs_task_execution.aws_iam_role.this
  to   = module.ecs_task_execution.module.role[0].aws_iam_role.this
}`,
			expected: helper.Issues{},
		},
		{
			name: "moved block with a count-shaped address — pass",
			content: `
moved {
  from = aws_instance.old[0]
  to   = aws_instance.web[0]
}`,
			expected: helper.Issues{},
		},
		{
			name: "import block address — pass",
			content: `
import {
  to = aws_instance.web[0]
  id = "i-abc123"
}`,
			expected: helper.Issues{},
		},
		{
			name: "removed block address — pass",
			content: `
removed {
  from = aws_instance.web[0]
}`,
			expected: helper.Issues{},
		},
		{
			name: "[0] inside a nested resource block — error",
			content: `
resource "aws_lb_listener" "this" {
  default_action {
    target_group_arn = aws_lb_target_group.this[0].arn
  }
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `use one() instead of [0] indexing on "aws_lb_target_group.this" — one(aws_lb_target_group.this[*].arn) returns null safely when count = 0`,
				},
			},
		},
		{
			name: "one(aws_instance.web[*].id) — pass",
			content: `
locals {
  instance_id = one(aws_instance.web[*].id)
}`,
			expected: helper.Issues{},
		},
		{
			name: "var.subnets[0] — pass (var root skipped)",
			content: `
locals {
  subnet = var.subnets[0]
}`,
			expected: helper.Issues{},
		},
		{
			name: "local.items[0] — pass (local root skipped)",
			content: `
locals {
  item = local.items[0]
}`,
			expected: helper.Issues{},
		},
		{
			name: "aws_instance.web[1].id — pass (index 1, not 0)",
			content: `
locals {
  instance_id = aws_instance.web[1].id
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
