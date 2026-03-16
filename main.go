package main

import (
	"github.com/phazeight/tflint-ruleset-trusty/custom"
	"github.com/phazeight/tflint-ruleset-trusty/project"
	"github.com/phazeight/tflint-ruleset-trusty/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &custom.RuleSet{
			BuiltinRuleSet: tflint.BuiltinRuleSet{
				Name:    project.Name,
				Version: project.Version,
				Rules:   rules.All(),
			},
		},
	})
}
