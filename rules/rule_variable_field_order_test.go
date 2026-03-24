package rules_test

import (
	"testing"

	"github.com/phazeight/tflint-ruleset-trusty/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func runVariableFieldOrderCheck(t *testing.T, files map[string]string) helper.Issues {
	t.Helper()
	runner := helper.TestRunner(t, files)
	rule := rules.NewVariableFieldOrderRule()
	if err := rule.Check(runner); err != nil {
		t.Fatalf("Check: %v", err)
	}
	return runner.Issues
}

func TestVariableFieldOrderRule_Metadata(t *testing.T) {
	rule := rules.NewVariableFieldOrderRule()

	if rule.Name() != "trusty_variable_field_order" {
		t.Errorf("Name() = %q, want trusty_variable_field_order", rule.Name())
	}
	if !rule.Enabled() {
		t.Error("Enabled() should be true")
	}
	if rule.Severity() != tflint.ERROR {
		t.Errorf("Severity() = %v, want ERROR", rule.Severity())
	}
	if rule.Link() == "" {
		t.Error("Link() should not be empty")
	}
}

func TestVariableFieldOrderRule_NoIssue_CorrectOrder(t *testing.T) {
	issues := runVariableFieldOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "name" {
  type        = string
  description = "Name of the resource"
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for correct order, got %d", len(issues))
	}
}

func TestVariableFieldOrderRule_NoIssue_AllThreeCorrect(t *testing.T) {
	issues := runVariableFieldOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "enabled" {
  type        = bool
  default     = true
  description = "Whether to enable the feature"
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for correct three-field order, got %d", len(issues))
	}
}

func TestVariableFieldOrderRule_Issue_DescriptionBeforeType(t *testing.T) {
	issues := runVariableFieldOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "name" {
  description = "Name of the resource"
  type        = string
}
`,
	})
	if len(issues) != 1 {
		t.Fatalf("want 1 issue, got %d: %v", len(issues), issues)
	}
	want := `variable "name": "type" should appear before "description" (expected order: type, default, description)`
	if issues[0].Message != want {
		t.Errorf("issue message = %q, want %q", issues[0].Message, want)
	}
}

func TestVariableFieldOrderRule_Issue_DefaultBeforeType(t *testing.T) {
	issues := runVariableFieldOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "enabled" {
  default     = true
  type        = bool
  description = "Whether to enable"
}
`,
	})
	if len(issues) != 1 {
		t.Fatalf("want 1 issue, got %d: %v", len(issues), issues)
	}
	want := `variable "enabled": "type" should appear before "default" (expected order: type, default, description)`
	if issues[0].Message != want {
		t.Errorf("issue message = %q, want %q", issues[0].Message, want)
	}
}

func TestVariableFieldOrderRule_Issue_DescriptionBeforeDefault(t *testing.T) {
	issues := runVariableFieldOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "enabled" {
  type        = bool
  description = "Whether to enable"
  default     = true
}
`,
	})
	if len(issues) != 1 {
		t.Fatalf("want 1 issue, got %d: %v", len(issues), issues)
	}
	want := `variable "enabled": "default" should appear before "description" (expected order: type, default, description)`
	if issues[0].Message != want {
		t.Errorf("issue message = %q, want %q", issues[0].Message, want)
	}
}

func TestVariableFieldOrderRule_NoIssue_OnlyType(t *testing.T) {
	issues := runVariableFieldOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "name" {
  type = string
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for single field, got %d", len(issues))
	}
}

func TestVariableFieldOrderRule_NoIssue_UnknownFieldsIgnored(t *testing.T) {
	issues := runVariableFieldOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "name" {
  type        = string
  default     = "foo"
  description = "A name"
  sensitive   = true
  nullable    = false
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues when unknown fields present, got %d", len(issues))
	}
}

func TestVariableFieldOrderRule_IgnoresNonInputsTf(t *testing.T) {
	issues := runVariableFieldOrderCheck(t, map[string]string{
		"main.tf": `
variable "name" {
  description = "Bad order"
  type        = string
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for non-inputs.tf file, got %d", len(issues))
	}
}

func TestVariableFieldOrderRule_MultipleVariables_FirstBad(t *testing.T) {
	issues := runVariableFieldOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "alpha" {
  description = "First"
  type        = string
}

variable "beta" {
  type        = string
  description = "Second"
}
`,
	})
	if len(issues) != 1 {
		t.Fatalf("want 1 issue (first bad variable), got %d: %v", len(issues), issues)
	}
}

func TestVariableFieldOrderRule_EmptyFiles(t *testing.T) {
	issues := runVariableFieldOrderCheck(t, map[string]string{})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for empty runner, got %d", len(issues))
	}
}

func TestVariableFieldOrderRule_RealisticCorrect(t *testing.T) {
	issues := runVariableFieldOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "force_delete" {
  type        = bool
  default     = false
  description = "If true, will delete the repository even if it contains images."
}

variable "image_tag_mutability" {
  type        = string
  default     = "MUTABLE"
  description = "The tag mutability setting for the repository."
}

variable "name" {
  type        = string
  description = "Name of the Repo"
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for realistic correct example, got %d", len(issues))
	}
}
