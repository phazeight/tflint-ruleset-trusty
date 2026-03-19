package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/phazeight/tflint-ruleset-trusty/project"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// VariableLocationRule enforces that all variable blocks are defined in a
// file named inputs.tf rather than scattered across arbitrary files.
type VariableLocationRule struct {
	tflint.DefaultRule
}

// NewVariableLocationRule creates a new VariableLocationRule.
func NewVariableLocationRule() *VariableLocationRule {
	return &VariableLocationRule{}
}

// Name returns the name of the rule.
func (r *VariableLocationRule) Name() string {
	return project.RuleName("variable_location")
}

// Enabled returns whether the rule is enabled by default.
func (r *VariableLocationRule) Enabled() bool {
	return true
}

// Severity returns the severity of the rule.
func (r *VariableLocationRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the reference link for the rule.
func (r *VariableLocationRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check verifies that all variable blocks live in inputs.tf.
func (r *VariableLocationRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for name, file := range files {
		base := filepath.Base(name)
		if strings.EqualFold(base, "inputs.tf") {
			continue
		}

		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}

		for _, block := range body.Blocks {
			if block.Type != "variable" {
				continue
			}

			varName := ""
			if len(block.Labels) > 0 {
				varName = block.Labels[0]
			}

			if err := runner.EmitIssue(
				r,
				fmt.Sprintf(
					"variable %q should be defined in inputs.tf, not %s",
					varName,
					base,
				),
				block.DefRange(),
			); err != nil {
				return err
			}
		}
	}

	return nil
}
