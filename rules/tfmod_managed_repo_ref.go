package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

// managedRepoSourceMarker identifies module blocks that provision a managed
// GitHub repository — their source path contains "managed-repo".
const managedRepoSourceMarker = "managed-repo"

// managedRepoConsumerTypes are the GitHub provider resource types whose
// `repository` attribute should reference a managed-repo module rather than
// hardcode the repo name string.
var managedRepoConsumerTypes = map[string]bool{
	"github_actions_secret":          true,
	"github_actions_variable":        true,
	"github_actions_environment":     true,
	"github_team_repository":         true,
	"github_repository_collaborator": true,
}

// ManagedRepoRefRule flags string-literal `repository` values that match a
// repo_name provisioned by a managed-repo module in the same configuration,
// nudging toward referencing the module output instead.
type ManagedRepoRefRule struct {
	tflint.DefaultRule
}

// NewManagedRepoRefRule creates a new ManagedRepoRefRule.
func NewManagedRepoRefRule() *ManagedRepoRefRule {
	return &ManagedRepoRefRule{}
}

func (r *ManagedRepoRefRule) Name() string              { return "trusty_managed_repo_ref" }
func (r *ManagedRepoRefRule) Enabled() bool             { return true }
func (r *ManagedRepoRefRule) Severity() tflint.Severity { return tflint.WARNING }
func (r *ManagedRepoRefRule) Link() string {
	return "https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/rules/trusty_managed_repo_ref.md"
}

// staticAttrString returns the static string value of a block attribute, or
// (\"\", false) if the attribute is absent or its expression is a reference.
func staticAttrString(attrs hclsyntax.Attributes, name string) (string, bool) {
	attr, exists := attrs[name]
	if !exists {
		return "", false
	}
	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() || val.Type() != cty.String {
		return "", false
	}
	return val.AsString(), true
}

// Check runs two passes: collect repo_name literals from managed-repo module
// blocks, then flag consumer resources whose repository literal matches.
func (r *ManagedRepoRefRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	// Pass 1: repo_name literal -> module label.
	managed := map[string]string{}
	for _, file := range files {
		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}
		for _, block := range body.Blocks {
			if block.Type != "module" || len(block.Labels) < 1 {
				continue
			}
			source, ok := staticAttrString(block.Body.Attributes, "source")
			if !ok || !strings.Contains(source, managedRepoSourceMarker) {
				continue
			}
			repoName, ok := staticAttrString(block.Body.Attributes, "repo_name")
			if !ok {
				continue
			}
			managed[repoName] = block.Labels[0]
		}
	}
	if len(managed) == 0 {
		return nil
	}

	// Pass 2: flag consumer resources referencing a managed repo by literal.
	for _, file := range files {
		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}
		for _, block := range body.Blocks {
			if block.Type != "resource" || len(block.Labels) < 1 {
				continue
			}
			if !managedRepoConsumerTypes[block.Labels[0]] {
				continue
			}
			repoAttr, exists := block.Body.Attributes["repository"]
			if !exists {
				continue
			}
			repo, ok := staticStringLiteral(repoAttr.Expr)
			if !ok {
				continue
			}
			label, managedByModule := managed[repo]
			if !managedByModule {
				continue
			}
			msg := fmt.Sprintf(
				`%q is managed by module.%s — reference it (e.g. split("/", module.%s.repository_full_name)[1]) instead of a hardcoded string`,
				repo, label, label,
			)
			if err := runner.EmitIssue(r, msg, repoAttr.Expr.Range()); err != nil {
				return err
			}
		}
	}
	return nil
}
