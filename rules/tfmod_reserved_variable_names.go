// rules/trusty_reserved_variable_names.go
package rules

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// reservedVariableNames are the names Terraform reserves as meta-arguments inside
// `module` blocks. A variable with one of these names is not a style problem — the
// module cannot be initialised at all:
//
//	The variable name "version" is reserved due to its special meaning inside
//	module blocks.
//
// This bites atom modules in particular, because atoms mirror a provider
// resource one-to-one and several AWS resources have a `version` attribute
// (aws_eks_cluster, aws_eks_node_group). Those atoms must rename the variable and
// map it back in main.tf.
var reservedVariableNames = map[string]bool{
	"count":      true,
	"depends_on": true,
	"for_each":   true,
	"lifecycle":  true,
	"providers":  true,
	"source":     true,
	"version":    true,
}

// ReservedVariableNamesRule flags variable blocks whose name Terraform reserves.
type ReservedVariableNamesRule struct {
	tflint.DefaultRule
}

// NewReservedVariableNamesRule creates a new ReservedVariableNamesRule.
func NewReservedVariableNamesRule() *ReservedVariableNamesRule {
	return &ReservedVariableNamesRule{}
}

func (r *ReservedVariableNamesRule) Name() string              { return "trusty_reserved_variable_names" }
func (r *ReservedVariableNamesRule) Enabled() bool             { return true }
func (r *ReservedVariableNamesRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *ReservedVariableNamesRule) Link() string {
	return "https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/rules/trusty_reserved_variable_names.md"
}

func (r *ReservedVariableNamesRule) Check(runner tflint.Runner) error {
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
			if !reservedVariableNames[varName] {
				continue
			}

			if err := runner.EmitIssue(r,
				fmt.Sprintf(
					`variable "%s" uses a name Terraform reserves in module blocks; the module cannot be initialised. Rename it (e.g. "kubernetes_version" for a resource's version attribute) and map it back in main.tf`,
					varName,
				),
				block.DefRange(),
			); err != nil {
				return err
			}
		}
	}
	return nil
}
