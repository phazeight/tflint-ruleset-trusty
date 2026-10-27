package rules

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

// ModuleSourceConfig is set per-repo in .tflint.hcl, e.g.:
//
//	rule "trusty_module_source" {
//	  enabled = true
//	  repo    = "github.com/your-org/your-module-repo"
//	}
type ModuleSourceConfig struct {
	Repo string `hclext:"repo,optional"`
}

// ModuleSourceRule enforces source conventions for references to the configured module repo.
type ModuleSourceRule struct {
	tflint.DefaultRule
}

// NewModuleSourceRule creates a new ModuleSourceRule.
func NewModuleSourceRule() *ModuleSourceRule {
	return &ModuleSourceRule{}
}

func (r *ModuleSourceRule) Name() string              { return "trusty_module_source" }
func (r *ModuleSourceRule) Enabled() bool             { return true }
func (r *ModuleSourceRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *ModuleSourceRule) Link() string {
	return "https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/rules/trusty_module_source.md"
}

// Check walks all module blocks and validates source values that reference the configured module repo.
// It does nothing unless `repo` is set in the rule config.
func (r *ModuleSourceRule) Check(runner tflint.Runner) error {
	var config ModuleSourceConfig
	if err := runner.DecodeRuleConfig(r.Name(), &config); err != nil {
		return err
	}
	if config.Repo == "" {
		return nil
	}

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
			if block.Type != "module" || len(block.Labels) < 1 {
				continue
			}

			sourceAttr, exists := block.Body.Attributes["source"]
			if !exists {
				continue
			}

			val, diags := sourceAttr.Expr.Value(nil)
			if diags.HasErrors() || val.Type() != cty.String {
				continue
			}
			source := val.AsString()

			if !strings.Contains(source, config.Repo) {
				continue
			}

			if strings.Contains(source, "git::ssh://") {
				if err := runner.EmitIssue(
					r,
					"module source must use git::https://, not git::ssh://",
					sourceAttr.NameRange,
				); err != nil {
					return err
				}
			}

			// Strip the scheme (e.g. "https://") before checking for //subdir,
			// otherwise the :// in the URL scheme itself triggers a false positive.
			withoutScheme := source
			if idx := strings.Index(source, "://"); idx >= 0 {
				withoutScheme = source[idx+3:]
			}
			if strings.Contains(withoutScheme, "//") {
				if err := runner.EmitIssue(
					r,
					"module source must not use //subdir notation — module repo tags are flat commits",
					sourceAttr.NameRange,
				); err != nil {
					return err
				}
			}

			if !strings.Contains(source, "?ref=") {
				if err := runner.EmitIssue(
					r,
					"module source must be pinned with ?ref=<tag>",
					sourceAttr.NameRange,
				); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
