package rules_test

import (
	"testing"

	"github.com/phazeight/tflint-ruleset-trusty/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func runVariableLocationCheck(t *testing.T, files map[string]string) helper.Issues {
	t.Helper()
	runner := helper.TestRunner(t, files)
	rule := rules.NewVariableLocationRule()
	if err := rule.Check(runner); err != nil {
		t.Fatalf("Check: %v", err)
	}
	return runner.Issues
}

func TestVariableLocationRule_Metadata(t *testing.T) {
	rule := rules.NewVariableLocationRule()

	if rule.Name() != "trusty_variable_location" {
		t.Errorf("Name() = %q, want trusty_variable_location", rule.Name())
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

func TestVariableLocationRule_NoIssue_VariablesInInputsTf(t *testing.T) {
	issues := runVariableLocationCheck(t, map[string]string{
		"inputs.tf": `
variable "name" {
  type = string
}

variable "region" {
  type = string
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues when variables are in inputs.tf, got %d", len(issues))
	}
}

func TestVariableLocationRule_Issue_VariableInMainTf(t *testing.T) {
	issues := runVariableLocationCheck(t, map[string]string{
		"main.tf": `
variable "name" {
  type = string
}
`,
	})
	if len(issues) != 1 {
		t.Fatalf("want 1 issue, got %d: %v", len(issues), issues)
	}
	want := `variable "name" should be defined in inputs.tf, not main.tf`
	if issues[0].Message != want {
		t.Errorf("issue message = %q, want %q", issues[0].Message, want)
	}
}

func TestVariableLocationRule_Issue_VariableInVariablesTf(t *testing.T) {
	issues := runVariableLocationCheck(t, map[string]string{
		"variables.tf": `
variable "region" {
  type = string
}
`,
	})
	if len(issues) != 1 {
		t.Fatalf("want 1 issue, got %d: %v", len(issues), issues)
	}
	want := `variable "region" should be defined in inputs.tf, not variables.tf`
	if issues[0].Message != want {
		t.Errorf("issue message = %q, want %q", issues[0].Message, want)
	}
}

func TestVariableLocationRule_NoIssue_NonVariableBlocksElsewhere(t *testing.T) {
	issues := runVariableLocationCheck(t, map[string]string{
		"main.tf": `
resource "aws_s3_bucket" "example" {
  bucket = "my-bucket"
}
`,
		"inputs.tf": `
variable "name" {
  type = string
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for non-variable blocks in other files, got %d", len(issues))
	}
}

func TestVariableLocationRule_MultipleVarsInWrongFile(t *testing.T) {
	issues := runVariableLocationCheck(t, map[string]string{
		"main.tf": `
variable "alpha" {
  type = string
}

variable "beta" {
  type = string
}
`,
	})
	if len(issues) != 2 {
		t.Fatalf("want 2 issues, got %d: %v", len(issues), issues)
	}
}

func TestVariableLocationRule_NoIssue_EmptyFiles(t *testing.T) {
	issues := runVariableLocationCheck(t, map[string]string{})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for empty runner, got %d", len(issues))
	}
}
