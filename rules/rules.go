package rules

import (
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// All returns all rules in the ruleset.
func All() []tflint.Rule {
	return []tflint.Rule{
		NewKeyAttributesRule(),
		NewVariableLocationRule(),
		NewVariableOrderRule(),
	}
}
