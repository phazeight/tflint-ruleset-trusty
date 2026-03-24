package rules_test

import (
	"testing"

	"github.com/phazeight/tflint-ruleset-trusty/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func runVariableOrderCheck(t *testing.T, files map[string]string) helper.Issues {
	t.Helper()
	runner := helper.TestRunner(t, files)
	rule := rules.NewVariableOrderRule()
	if err := rule.Check(runner); err != nil {
		t.Fatalf("Check: %v", err)
	}
	return runner.Issues
}

func TestVariableOrderRule_Metadata(t *testing.T) {
	rule := rules.NewVariableOrderRule()

	if rule.Name() != "trusty_variable_order" {
		t.Errorf("Name() = %q, want trusty_variable_order", rule.Name())
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

func TestVariableOrderRule_NoIssue_AlphabeticalOrder(t *testing.T) {
	issues := runVariableOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "alpha" {
  type = string
}

variable "beta" {
  type = string
}

variable "gamma" {
  type = string
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for alphabetically sorted variables, got %d", len(issues))
	}
}

func TestVariableOrderRule_Issue_OutOfOrder(t *testing.T) {
	issues := runVariableOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "beta" {
  type = string
}

variable "alpha" {
  type = string
}
`,
	})
	if len(issues) != 1 {
		t.Fatalf("want 1 issue, got %d: %v", len(issues), issues)
	}
	want := `variable "alpha" should be defined before "beta" (alphabetical order)`
	if issues[0].Message != want {
		t.Errorf("issue message = %q, want %q", issues[0].Message, want)
	}
}

func TestVariableOrderRule_NoIssue_SingleVariable(t *testing.T) {
	issues := runVariableOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "name" {
  type = string
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for single variable, got %d", len(issues))
	}
}

func TestVariableOrderRule_NoIssue_NoVariables(t *testing.T) {
	issues := runVariableOrderCheck(t, map[string]string{
		"inputs.tf": `
locals {
  x = 1
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues when no variables present, got %d", len(issues))
	}
}

func TestVariableOrderRule_IgnoresNonInputsTf(t *testing.T) {
	// Out-of-order variables in main.tf should NOT trigger this rule.
	// (The location rule handles that concern separately.)
	issues := runVariableOrderCheck(t, map[string]string{
		"main.tf": `
variable "zebra" {
  type = string
}

variable "apple" {
  type = string
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for non-inputs.tf file, got %d", len(issues))
	}
}

func TestVariableOrderRule_CaseInsensitive(t *testing.T) {
	issues := runVariableOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "Alpha" {
  type = string
}

variable "beta" {
  type = string
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for case-insensitive alphabetical order, got %d", len(issues))
	}
}

func TestVariableOrderRule_NonVariableBlocksIgnored(t *testing.T) {
	// Non-variable blocks (like locals) between variables should not
	// affect the ordering check.
	issues := runVariableOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "alpha" {
  type = string
}

locals {
  x = 1
}

variable "beta" {
  type = string
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues when variables are in order with non-variable blocks between, got %d", len(issues))
	}
}

func TestVariableOrderRule_RealisticExample_Sorted(t *testing.T) {
	// Alphabetically sorted version of the Linear issue example
	issues := runVariableOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "force_delete" {
  default     = false
  description = "If true, will delete the repository even if it contains images."
  type        = bool
}

variable "image_tag_mutability" {
  default     = "MUTABLE"
  description = "The tag mutability setting for the repository."
  type        = string
}

variable "name" {
  description = "Name of the Repo"
  type        = string
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for sorted realistic example, got %d", len(issues))
	}
}

func TestVariableOrderRule_RealisticExample_Unsorted(t *testing.T) {
	// Original order from the Linear issue — name comes before force_delete
	issues := runVariableOrderCheck(t, map[string]string{
		"inputs.tf": `
variable "name" {
  description = "Name of the Repo"
  type        = string
}

variable "force_delete" {
  default     = false
  description = "If true, will delete the repository even if it contains images."
  type        = bool
}

variable "image_tag_mutability" {
  default     = "MUTABLE"
  description = "The tag mutability setting for the repository."
  type        = string
}
`,
	})
	if len(issues) != 1 {
		t.Fatalf("want 1 issue for unsorted realistic example, got %d: %v", len(issues), issues)
	}
}

func TestVariableOrderRule_EmptyFiles(t *testing.T) {
	issues := runVariableOrderCheck(t, map[string]string{})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for empty runner, got %d", len(issues))
	}
}
