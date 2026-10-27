package main

import (
	"github.com/phazeight/tflint-ruleset-trusty/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &tflint.BuiltinRuleSet{
			Name:    "trusty",
			Version: "0.5.0",
			Rules: []tflint.Rule{
				rules.NewVariableLocationRule(),
				rules.NewVariableOrderRule(),
				rules.NewVariableFieldOrderRule(),
				rules.NewKeyAttributesRule(),
				rules.NewVariableSectionOrderRule(),
				rules.NewModuleSourceRule(),
				rules.NewPositiveVariableNamesRule(),
				rules.NewReservedVariableNamesRule(),
				rules.NewHeredocDescriptionRule(),
				rules.NewUseOneForConditionalRule(),
				rules.NewManagedRepoRefRule(),
				rules.NewNoHardcodedAccountIDRule(),
				rules.NewNoHardcodedARNRule(),
				rules.NewNoHardcodedResourceIDRule(),
			},
		},
	})
}
