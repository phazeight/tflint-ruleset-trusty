package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

var negativeBoolPrefixes = []string{"disable_", "no_", "not_", "prevent_"}

// PositiveVariableNamesRule enforces that bool variables use positive naming
// (e.g., "encryption_enabled" not "disable_encryption").
type PositiveVariableNamesRule struct {
	tflint.DefaultRule
}

// NewPositiveVariableNamesRule creates a new PositiveVariableNamesRule.
func NewPositiveVariableNamesRule() *PositiveVariableNamesRule {
	return &PositiveVariableNamesRule{}
}

func (r *PositiveVariableNamesRule) Name() string              { return "trusty_positive_variable_names" }
func (r *PositiveVariableNamesRule) Enabled() bool             { return true }
func (r *PositiveVariableNamesRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *PositiveVariableNamesRule) Link() string {
	return "https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/rules/trusty_positive_variable_names.md"
}

func (r *PositiveVariableNamesRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}

		for _, block := range body.Blocks {
			if block.Type != "variable" || len(block.Labels) < 1 {
				continue
			}
			varName := block.Labels[0]

			typeAttr := block.Body.Attributes["type"]
			if typeAttr == nil {
				continue
			}

			exprRange := typeAttr.Expr.Range()
			start := exprRange.Start.Byte
			end := exprRange.End.Byte
			if end > len(file.Bytes) {
				continue
			}
			typeSrc := strings.TrimSpace(string(file.Bytes[start:end]))
			if typeSrc != "bool" {
				continue
			}

			for _, prefix := range negativeBoolPrefixes {
				if strings.HasPrefix(varName, prefix) {
					if err := runner.EmitIssue(r,
						fmt.Sprintf(`variable "%s": use positive naming (avoid "%s" prefix)`, varName, prefix),
						block.DefRange(),
					); err != nil {
						return err
					}
					break
				}
			}
		}
	}
	return nil
}
