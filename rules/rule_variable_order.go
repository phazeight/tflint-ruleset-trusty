package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/phazeight/tflint-ruleset-trusty/project"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// VariableOrderRule enforces that variable blocks within inputs.tf are
// sorted alphabetically (A–Z) by variable name.
type VariableOrderRule struct {
	tflint.DefaultRule
}

// NewVariableOrderRule creates a new VariableOrderRule.
func NewVariableOrderRule() *VariableOrderRule {
	return &VariableOrderRule{}
}

// Name returns the name of the rule.
func (r *VariableOrderRule) Name() string {
	return project.RuleName("variable_order")
}

// Enabled returns whether the rule is enabled by default.
func (r *VariableOrderRule) Enabled() bool {
	return true
}

// Severity returns the severity of the rule.
func (r *VariableOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the reference link for the rule.
func (r *VariableOrderRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check verifies that variable blocks in inputs.tf are sorted A–Z by name.
func (r *VariableOrderRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for name, file := range files {
		base := filepath.Base(name)
		if !strings.EqualFold(base, "inputs.tf") {
			continue
		}

		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}

		var prev string
		for _, block := range body.Blocks {
			if block.Type != "variable" {
				continue
			}
			if len(block.Labels) == 0 {
				continue
			}

			current := block.Labels[0]
			if prev != "" && strings.ToLower(current) < strings.ToLower(prev) {
				return runner.EmitIssue(
					r,
					fmt.Sprintf(
						"variable %q should be defined before %q (alphabetical order)",
						current,
						prev,
					),
					block.DefRange(),
				)
			}
			prev = current
		}
	}

	return nil
}
