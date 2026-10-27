package rules

import (
	"fmt"
	"path/filepath"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// VariableLocationRule enforces that all variable blocks are defined in inputs.tf.
type VariableLocationRule struct {
	tflint.DefaultRule
}

func NewVariableLocationRule() *VariableLocationRule {
	return &VariableLocationRule{}
}

func (r *VariableLocationRule) Name() string              { return "trusty_variable_location" }
func (r *VariableLocationRule) Enabled() bool             { return true }
func (r *VariableLocationRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *VariableLocationRule) Link() string {
	return "https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/rules/trusty_variable_location.md"
}

func (r *VariableLocationRule) Check(runner tflint.Runner) error {
	content, err := runner.GetModuleContent(
		&hclext.BodySchema{
			Blocks: []hclext.BlockSchema{
				{
					Type:       "variable",
					LabelNames: []string{"name"},
					Body:       &hclext.BodySchema{},
				},
			},
		},
		&tflint.GetModuleContentOption{},
	)
	if err != nil {
		return err
	}

	for _, block := range content.Blocks {
		filename := filepath.Base(block.DefRange.Filename)
		if filename == "inputs.tf" {
			continue
		}
		if err := runner.EmitIssue(
			r,
			fmt.Sprintf(`variable "%s" must be defined in inputs.tf, not in %s`, block.Labels[0], filename),
			block.DefRange,
		); err != nil {
			return err
		}
	}
	return nil
}
