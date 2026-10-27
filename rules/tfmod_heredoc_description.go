package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// HeredocDescriptionRule enforces that variable blocks using an object type with
// ≥2 attributes use <<-DOC heredoc syntax for the description. This produces
// cleaner terraform-docs output and readable per-attribute documentation.
type HeredocDescriptionRule struct {
	tflint.DefaultRule
}

// NewHeredocDescriptionRule creates a new HeredocDescriptionRule.
func NewHeredocDescriptionRule() *HeredocDescriptionRule {
	return &HeredocDescriptionRule{}
}

func (r *HeredocDescriptionRule) Name() string              { return "trusty_heredoc_description" }
func (r *HeredocDescriptionRule) Enabled() bool             { return true }
func (r *HeredocDescriptionRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *HeredocDescriptionRule) Link() string {
	return "https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/rules/trusty_heredoc_description.md"
}

func (r *HeredocDescriptionRule) Check(runner tflint.Runner) error {
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

			typeRange := typeAttr.Expr.Range()
			typeStart := typeRange.Start.Byte
			typeEnd := typeRange.End.Byte
			if typeEnd > len(file.Bytes) {
				continue
			}
			typeSrc := string(file.Bytes[typeStart:typeEnd])

			// Only applies to types containing object(...)
			if !strings.Contains(typeSrc, "object(") {
				continue
			}
			// Count attribute assignments; fewer than 2 means a trivial object, skip
			if strings.Count(typeSrc, " = ") < 2 {
				continue
			}

			descAttr := block.Body.Attributes["description"]
			if descAttr == nil {
				continue
			}

			descStart := descAttr.Expr.Range().Start.Byte
			if descStart >= len(file.Bytes) {
				continue
			}
			// Heredoc expressions start with '<' (<<-DOC or <<DOC)
			if file.Bytes[descStart] == '<' {
				continue
			}

			if err := runner.EmitIssue(r,
				fmt.Sprintf(`variable "%s": object type with ≥2 attributes must use <<-DOC heredoc for description`, varName),
				descAttr.NameRange,
			); err != nil {
				return err
			}
		}
	}
	return nil
}
