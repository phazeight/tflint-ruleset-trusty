package rules

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// VariableSectionOrderRule enforces that inputs.tf files with both required and optional
// variables have a DEFAULTS separator comment, with all required variables before it and
// all optional variables after it.
type VariableSectionOrderRule struct {
	tflint.DefaultRule
}

func NewVariableSectionOrderRule() *VariableSectionOrderRule {
	return &VariableSectionOrderRule{}
}

func (r *VariableSectionOrderRule) Name() string              { return "trusty_variable_section_order" }
func (r *VariableSectionOrderRule) Enabled() bool             { return true }
func (r *VariableSectionOrderRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *VariableSectionOrderRule) Link() string {
	return "https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/rules/trusty_variable_section_order.md"
}

type sectionVarEntry struct {
	name       string
	line       int
	hasDefault bool
	defRange   hcl.Range
}

// findDefaultsSeparatorLine scans raw file bytes for a line that starts with '#'
// and contains the word "DEFAULTS" (case-sensitive). Returns the 1-based line number,
// or 0 if not found.
func findDefaultsSeparatorLine(src []byte) int {
	scanner := bufio.NewScanner(bytes.NewReader(src))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") && strings.Contains(line, "DEFAULTS") {
			return lineNum
		}
	}
	return 0
}

func (r *VariableSectionOrderRule) Check(runner tflint.Runner) error {
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

		var vars []sectionVarEntry
		for _, block := range body.Blocks {
			if block.Type != "variable" || len(block.Labels) < 1 {
				continue
			}
			_, hasDefault := block.Body.Attributes["default"]
			dr := block.DefRange()
			vars = append(vars, sectionVarEntry{
				name:       block.Labels[0],
				line:       dr.Start.Line,
				hasDefault: hasDefault,
				defRange:   dr,
			})
		}

		// Sort by source line defensively
		sort.Slice(vars, func(i, j int) bool {
			return vars[i].line < vars[j].line
		})

		// Classify
		var required, optional []sectionVarEntry
		for _, v := range vars {
			if v.hasDefault {
				optional = append(optional, v)
			} else {
				required = append(required, v)
			}
		}

		// Only check files that have BOTH required and optional variables
		if len(required) == 0 || len(optional) == 0 {
			continue
		}

		separatorLine := findDefaultsSeparatorLine(file.Bytes)
		if separatorLine == 0 {
			// No separator found — emit error on the first optional variable
			if err := runner.EmitIssue(
				r,
				`inputs.tf has both required and optional variables but is missing the DEFAULTS separator comment (###...DEFAULTS...###)`,
				optional[0].defRange,
			); err != nil {
				return err
			}
			continue
		}

		// Required variables must appear before the separator
		for _, v := range required {
			if v.line > separatorLine {
				if err := runner.EmitIssue(
					r,
					fmt.Sprintf(`required variable "%s" must appear before the DEFAULTS separator`, v.name),
					v.defRange,
				); err != nil {
					return err
				}
			}
		}

		// Optional variables must appear after the separator
		for _, v := range optional {
			if v.line < separatorLine {
				if err := runner.EmitIssue(
					r,
					fmt.Sprintf(`optional variable "%s" (has default) must appear after the DEFAULTS separator`, v.name),
					v.defRange,
				); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
