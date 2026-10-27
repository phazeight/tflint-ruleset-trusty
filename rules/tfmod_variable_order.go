package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// VariableOrderRule enforces that variables in inputs.tf are sorted A-Z within each
// section: required variables (no default) first, then optional variables (with default).
type VariableOrderRule struct {
	tflint.DefaultRule
}

func NewVariableOrderRule() *VariableOrderRule {
	return &VariableOrderRule{}
}

func (r *VariableOrderRule) Name() string              { return "trusty_variable_order" }
func (r *VariableOrderRule) Enabled() bool             { return true }
func (r *VariableOrderRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *VariableOrderRule) Link() string {
	return "https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/rules/trusty_variable_order.md"
}

type varEntry struct {
	name       string
	line       int
	hasDefault bool
	defRange   hcl.Range
}

func (r *VariableOrderRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for filename, file := range files {
		if !strings.HasSuffix(filename, "inputs.tf") {
			continue
		}

		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}

		var vars []varEntry
		for _, block := range body.Blocks {
			if block.Type != "variable" || len(block.Labels) < 1 {
				continue
			}
			_, hasDefault := block.Body.Attributes["default"]
			dr := block.DefRange()
			vars = append(vars, varEntry{
				name:       block.Labels[0],
				line:       dr.Start.Line,
				hasDefault: hasDefault,
				defRange:   dr,
			})
		}

		// body.Blocks is in source order, but sort defensively
		sort.Slice(vars, func(i, j int) bool {
			return vars[i].line < vars[j].line
		})

		// Check: no required variable appears after an optional variable
		seenOptional := false
		for _, v := range vars {
			if v.hasDefault {
				seenOptional = true
			} else if seenOptional {
				if err := runner.EmitIssue(
					r,
					fmt.Sprintf(`required variable "%s" (no default) must appear before all optional variables`, v.name),
					v.defRange,
				); err != nil {
					return err
				}
			}
		}

		// Split into sections and check each is sorted A-Z
		var required, optional []varEntry
		for _, v := range vars {
			if v.hasDefault {
				optional = append(optional, v)
			} else {
				required = append(required, v)
			}
		}

		for i := 1; i < len(required); i++ {
			if required[i].name < required[i-1].name {
				if err := runner.EmitIssue(
					r,
					fmt.Sprintf(`required variable "%s" must be sorted A-Z (appears after "%s")`, required[i].name, required[i-1].name),
					required[i].defRange,
				); err != nil {
					return err
				}
			}
		}

		for i := 1; i < len(optional); i++ {
			if optional[i].name < optional[i-1].name {
				if err := runner.EmitIssue(
					r,
					fmt.Sprintf(`optional variable "%s" must be sorted A-Z (appears after "%s")`, optional[i].name, optional[i-1].name),
					optional[i].defRange,
				); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
