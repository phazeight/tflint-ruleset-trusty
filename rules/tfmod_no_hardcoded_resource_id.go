package rules

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// resourceIDPattern matches AWS resource IDs for a tight set of prefixes.
// Expand the prefix set as false-positive experience allows.
var resourceIDPattern = regexp.MustCompile(`^(vpc|subnet|sg|ami|igw|rtb|acl|eni|nat|eipalloc|pcx)-[0-9a-f]{8,}$`)

// NoHardcodedResourceIDRule flags static string literals that look like a
// hardcoded AWS resource ID.
type NoHardcodedResourceIDRule struct {
	tflint.DefaultRule
}

// NewNoHardcodedResourceIDRule creates a new NoHardcodedResourceIDRule.
func NewNoHardcodedResourceIDRule() *NoHardcodedResourceIDRule {
	return &NoHardcodedResourceIDRule{}
}

func (r *NoHardcodedResourceIDRule) Name() string              { return "trusty_no_hardcoded_resource_id" }
func (r *NoHardcodedResourceIDRule) Enabled() bool             { return true }
func (r *NoHardcodedResourceIDRule) Severity() tflint.Severity { return tflint.WARNING }
func (r *NoHardcodedResourceIDRule) Link() string {
	return "https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/rules/trusty_no_hardcoded_resource_id.md"
}

// Check walks every expression and flags hardcoded resource ID string literals.
func (r *NoHardcodedResourceIDRule) Check(runner tflint.Runner) error {
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
			if !ok || !resourceIDPattern.MatchString(s) {
				return nil
			}
			msg := fmt.Sprintf(
				`%q is a hardcoded resource ID — use a data source or resource reference instead`,
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
