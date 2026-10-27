package rules_test

import (
	"testing"

	"github.com/phazeight/tflint-ruleset-trusty/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestManagedRepoRefRule(t *testing.T) {
	rule := rules.NewManagedRepoRefRule()

	cases := []struct {
		name     string
		content  string
		expected helper.Issues
	}{
		{
			name: "hardcoded repository matching managed repo — warning",
			content: `
module "my_repo" {
  source    = "git::https://github.com/isapp/terraform-modules?ref=modules/molecules/managed-repo/v1.0.0"
  repo_name = "my-repo"
}

resource "github_actions_secret" "token" {
  repository  = "my-repo"
  secret_name = "TOKEN"
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `"my-repo" is managed by module.my_repo — reference it (e.g. split("/", module.my_repo.repository_full_name)[1]) instead of a hardcoded string`,
				},
			},
		},
		{
			name: "hardcoded repository on team repository — warning",
			content: `
module "my_repo" {
  source    = "git::https://github.com/isapp/terraform-modules?ref=modules/molecules/managed-repo/v1.0.0"
  repo_name = "my-repo"
}

resource "github_team_repository" "team" {
  team_id    = "123"
  repository = "my-repo"
}`,
			expected: helper.Issues{
				{
					Rule:    rule,
					Message: `"my-repo" is managed by module.my_repo — reference it (e.g. split("/", module.my_repo.repository_full_name)[1]) instead of a hardcoded string`,
				},
			},
		},
		{
			name: "already references the module — pass",
			content: `
module "my_repo" {
  source    = "git::https://github.com/isapp/terraform-modules?ref=modules/molecules/managed-repo/v1.0.0"
  repo_name = "my-repo"
}

resource "github_actions_secret" "token" {
  repository  = split("/", module.my_repo.repository_full_name)[1]
  secret_name = "TOKEN"
}`,
			expected: helper.Issues{},
		},
		{
			name: "repo not managed in this config — pass",
			content: `
resource "github_actions_secret" "token" {
  repository  = "some-external-repo"
  secret_name = "TOKEN"
}`,
			expected: helper.Issues{},
		},
		{
			name: "module is not a managed-repo module — pass",
			content: `
module "vpc" {
  source    = "git::https://github.com/isapp/terraform-modules?ref=modules/atoms/vpc/v1.0.0"
  repo_name = "my-repo"
}

resource "github_actions_secret" "token" {
  repository  = "my-repo"
  secret_name = "TOKEN"
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
