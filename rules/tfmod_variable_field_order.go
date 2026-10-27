package rules

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// VariableFieldOrderRule enforces that fields within each variable block appear in
// this order: type, default (if present), description.
//
// Uses raw hclsyntax parsing to avoid tflint attempting to evaluate type constraint
// expressions (string, bool, list(string), etc.) as regular HCL expressions.
type VariableFieldOrderRule struct {
	tflint.DefaultRule
}

func NewVariableFieldOrderRule() *VariableFieldOrderRule {
	return &VariableFieldOrderRule{}
}

func (r *VariableFieldOrderRule) Name() string              { return "trusty_variable_field_order" }
func (r *VariableFieldOrderRule) Enabled() bool             { return true }
func (r *VariableFieldOrderRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *VariableFieldOrderRule) Link() string {
	return "https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/rules/trusty_variable_field_order.md"
}

func (r *VariableFieldOrderRule) Check(runner tflint.Runner) error {
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
			defaultAttr := block.Body.Attributes["default"]
			descAttr := block.Body.Attributes["description"]

			// type must come before default
			if typeAttr != nil && defaultAttr != nil {
				if typeAttr.NameRange.Start.Line > defaultAttr.NameRange.Start.Line {
					if err := runner.EmitIssue(r,
						fmt.Sprintf(`in variable "%s": "type" must appear before "default"`, varName),
						typeAttr.NameRange,
					); err != nil {
						return err
					}
				}
			}

			// type must come before description
			if typeAttr != nil && descAttr != nil {
				if typeAttr.NameRange.Start.Line > descAttr.NameRange.Start.Line {
					if err := runner.EmitIssue(r,
						fmt.Sprintf(`in variable "%s": "type" must appear before "description"`, varName),
						typeAttr.NameRange,
					); err != nil {
						return err
					}
				}
			}

			// default must come before description
			if defaultAttr != nil && descAttr != nil {
				if defaultAttr.NameRange.Start.Line > descAttr.NameRange.Start.Line {
					if err := runner.EmitIssue(r,
						fmt.Sprintf(`in variable "%s": "default" must appear before "description"`, varName),
						defaultAttr.NameRange,
					); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
