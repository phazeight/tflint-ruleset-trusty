package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// arnPattern matches any AWS ARN partition (aws, aws-cn, aws-us-gov).
var arnPattern = regexp.MustCompile(`^arn:aws[a-z-]*:`)

// awsManagedARNPrefix is a baked exception: AWS-managed IAM policies/roles have
// no owning resource in the config to reference, so they are always allowed.
const awsManagedARNPrefix = "arn:aws:iam::aws:"

// NoHardcodedARNRule flags static string literals that look like a hardcoded
// AWS ARN, except AWS-managed IAM ARNs.
type NoHardcodedARNRule struct {
	tflint.DefaultRule
}

// NewNoHardcodedARNRule creates a new NoHardcodedARNRule.
func NewNoHardcodedARNRule() *NoHardcodedARNRule {
	return &NoHardcodedARNRule{}
}

func (r *NoHardcodedARNRule) Name() string              { return "trusty_no_hardcoded_arn" }
func (r *NoHardcodedARNRule) Enabled() bool             { return true }
func (r *NoHardcodedARNRule) Severity() tflint.Severity { return tflint.WARNING }
func (r *NoHardcodedARNRule) Link() string {
	return "https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/rules/trusty_no_hardcoded_arn.md"
}

// Check walks every expression and flags hardcoded ARN string literals.
func (r *NoHardcodedARNRule) Check(runner tflint.Runner) error {
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
			if !ok || !arnPattern.MatchString(s) {
				return nil
			}
			if strings.HasPrefix(s, awsManagedARNPrefix) {
				return nil
			}
			msg := fmt.Sprintf(
				`%q is a hardcoded ARN — reference the owning resource's .arn attribute, or a variable/data source instead`,
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
