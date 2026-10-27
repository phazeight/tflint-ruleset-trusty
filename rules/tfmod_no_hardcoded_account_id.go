package rules

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// accountIDPattern matches a bare 12-digit AWS account ID.
var accountIDPattern = regexp.MustCompile(`^[0-9]{12}$`)

// NoHardcodedAccountIDRule flags static string literals that look like a
// hardcoded AWS account ID.
type NoHardcodedAccountIDRule struct {
	tflint.DefaultRule
}

// NewNoHardcodedAccountIDRule creates a new NoHardcodedAccountIDRule.
func NewNoHardcodedAccountIDRule() *NoHardcodedAccountIDRule {
	return &NoHardcodedAccountIDRule{}
}

func (r *NoHardcodedAccountIDRule) Name() string              { return "trusty_no_hardcoded_account_id" }
func (r *NoHardcodedAccountIDRule) Enabled() bool             { return true }
func (r *NoHardcodedAccountIDRule) Severity() tflint.Severity { return tflint.WARNING }
func (r *NoHardcodedAccountIDRule) Link() string {
	return "https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/rules/trusty_no_hardcoded_account_id.md"
}

// Check walks every expression and flags 12-digit string literals.
func (r *NoHardcodedAccountIDRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}

		var visitErr error
		diags := hclsyntax.VisitAll(body, func(node hclsyntax.Node) hcl.Diagnostics {
			if visitErr != nil {
				return nil
			}
			s, ok := staticStringLiteral(node)
			if !ok || !accountIDPattern.MatchString(s) {
				return nil
			}
			msg := fmt.Sprintf(
				`%q looks like a hardcoded AWS account ID — use a variable, data.aws_caller_identity, or an account-id map input instead`,
				s,
			)
			if err := runner.EmitIssue(r, msg, node.Range()); err != nil {
				visitErr = err
			}
			return nil
		})
		if visitErr != nil {
			return visitErr
		}
		if diags.HasErrors() {
			return diags
		}
	}
	return nil
}
