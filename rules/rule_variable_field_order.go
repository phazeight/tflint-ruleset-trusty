package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/phazeight/tflint-ruleset-trusty/project"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// expectedFieldOrder defines the canonical attribute ordering inside a variable block.
var expectedFieldOrder = []string{"type", "default", "description"}

// VariableFieldOrderRule enforces that attributes within a variable block
// appear in the order: type, default, description.
type VariableFieldOrderRule struct {
	tflint.DefaultRule
}

// NewVariableFieldOrderRule creates a new VariableFieldOrderRule.
func NewVariableFieldOrderRule() *VariableFieldOrderRule {
	return &VariableFieldOrderRule{}
}

// Name returns the name of the rule.
func (r *VariableFieldOrderRule) Name() string {
	return project.RuleName("variable_field_order")
}

// Enabled returns whether the rule is enabled by default.
func (r *VariableFieldOrderRule) Enabled() bool {
	return true
}

// Severity returns the severity of the rule.
func (r *VariableFieldOrderRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the reference link for the rule.
func (r *VariableFieldOrderRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// fieldRank returns the position of a known field in the expected order.
// Unknown fields return -1 and are ignored.
func fieldRank(name string) int {
	for i, f := range expectedFieldOrder {
		if f == name {
			return i
		}
	}
	return -1
}

// Check verifies that attributes in each variable block follow the
// canonical order: type, default, description.
func (r *VariableFieldOrderRule) Check(runner tflint.Runner) error {
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

		for _, block := range body.Blocks {
			if block.Type != "variable" {
				continue
			}
			if len(block.Labels) == 0 {
				continue
			}

			varName := block.Labels[0]
			varBody := block.Body

			// Collect known fields in source order.
			type fieldInfo struct {
				name string
				rank int
			}
			var fields []fieldInfo
			for _, attr := range varBody.Attributes {
				rank := fieldRank(attr.Name)
				if rank >= 0 {
					fields = append(fields, fieldInfo{name: attr.Name, rank: rank})
				}
			}

			// Sort by source position (Attributes is a map, so we need to sort).
			// Re-collect in source order using SrcRange.
			type posField struct {
				name  string
				rank  int
				start int
			}
			var posFields []posField
			for _, attr := range varBody.Attributes {
				rank := fieldRank(attr.Name)
				if rank >= 0 {
					posFields = append(posFields, posField{
						name:  attr.Name,
						rank:  rank,
						start: attr.SrcRange.Start.Byte,
					})
				}
			}
			// Sort by byte position.
			for i := 1; i < len(posFields); i++ {
				for j := i; j > 0 && posFields[j].start < posFields[j-1].start; j-- {
					posFields[j], posFields[j-1] = posFields[j-1], posFields[j]
				}
			}

			// Check ordering.
			for i := 1; i < len(posFields); i++ {
				if posFields[i].rank < posFields[i-1].rank {
					return runner.EmitIssue(
						r,
						fmt.Sprintf(
							"variable %q: %q should appear before %q (expected order: type, default, description)",
							varName,
							posFields[i].name,
							posFields[i-1].name,
						),
						block.DefRange(),
					)
				}
			}
		}
	}

	return nil
}
