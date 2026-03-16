package rules_test

import (
	"testing"

	"github.com/phazeight/tflint-ruleset-trusty/config"
	"github.com/phazeight/tflint-ruleset-trusty/custom"
	"github.com/phazeight/tflint-ruleset-trusty/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// testConfig returns a minimal config with predictable resources for testing.
func testConfig() *config.Config {
	return &config.Config{
		Resources: []*config.Resource{
			{Kind: "aws_s3_bucket", Keys: []string{"bucket"}},
			{Kind: "aws_instance", Keys: []string{"ami", "instance_type"}},
			{Kind: "google_container_cluster", Keys: []string{"project", "location", "name"}},
			{Kind: "aws_empty_keys", Keys: []string{}},
		},
	}
}

func runCheck(t *testing.T, files map[string]string) helper.Issues {
	t.Helper()
	base := helper.TestRunner(t, files)
	runner, err := custom.NewRunner(base, testConfig())
	if err != nil {
		t.Fatalf("NewRunner: %v", err)
	}
	rule := rules.NewKeyAttributesRule()
	if err := rule.Check(runner); err != nil {
		t.Fatalf("Check: %v", err)
	}
	return base.Issues
}

func TestKeyAttributesRule_Metadata(t *testing.T) {
	rule := rules.NewKeyAttributesRule()

	if rule.Name() != "trusty_key_attributes" {
		t.Errorf("Name() = %q, want trusty_key_attributes", rule.Name())
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

func TestKeyAttributesRule_NoIssue_UnknownResource(t *testing.T) {
	issues := runCheck(t, map[string]string{
		"main.tf": `
resource "unknown_resource_type" "x" {
  z = "z"
  a = "a"
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for unknown resource, got %d", len(issues))
	}
}

func TestKeyAttributesRule_NoIssue_KeyFirst(t *testing.T) {
	issues := runCheck(t, map[string]string{
		"main.tf": `
resource "aws_s3_bucket" "example" {
  bucket = "my-bucket"
  tags   = {}
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues when key attribute is first, got %d", len(issues))
	}
}

func TestKeyAttributesRule_NoIssue_AllKeysPresent_CorrectOrder(t *testing.T) {
	issues := runCheck(t, map[string]string{
		"main.tf": `
resource "aws_instance" "web" {
  ami           = "ami-12345678"
  instance_type = "t3.micro"
  tags          = {}
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues when all keys are in correct order, got %d", len(issues))
	}
}

func TestKeyAttributesRule_Issue_KeyAfterNonKey(t *testing.T) {
	issues := runCheck(t, map[string]string{
		"main.tf": `
resource "aws_s3_bucket" "example" {
  tags   = {}
  bucket = "my-bucket"
}
`,
	})
	if len(issues) != 1 {
		t.Fatalf("want 1 issue, got %d: %v", len(issues), issues)
	}
	want := "key-attribute `bucket` should be defined before non-key `tags`"
	if issues[0].Message != want {
		t.Errorf("issue message = %q, want %q", issues[0].Message, want)
	}
}

func TestKeyAttributesRule_Issue_LowerPriorityKeyBeforeHigher(t *testing.T) {
	// instance_type appears before ami — ami has higher priority (listed first)
	issues := runCheck(t, map[string]string{
		"main.tf": `
resource "aws_instance" "web" {
  instance_type = "t3.micro"
  ami           = "ami-12345678"
}
`,
	})
	if len(issues) != 1 {
		t.Fatalf("want 1 issue, got %d: %v", len(issues), issues)
	}
	want := "higher-priority key-attribute `ami` should be defined before `instance_type`"
	if issues[0].Message != want {
		t.Errorf("issue message = %q, want %q", issues[0].Message, want)
	}
}

func TestKeyAttributesRule_NoIssue_ForEachSkipped(t *testing.T) {
	// for_each should be transparent — not treated as a non-key blocker
	issues := runCheck(t, map[string]string{
		"main.tf": `
resource "aws_s3_bucket" "example" {
  for_each = toset(["a", "b"])
  bucket   = each.key
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues when for_each precedes key attribute, got %d", len(issues))
	}
}

func TestKeyAttributesRule_NoIssue_EmptyKeyAttributes(t *testing.T) {
	issues := runCheck(t, map[string]string{
		"main.tf": `
resource "aws_empty_keys" "example" {
  z = "z"
  a = "a"
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for resource with empty key list, got %d", len(issues))
	}
}

func TestKeyAttributesRule_NoIssue_NonResourceBlock(t *testing.T) {
	// provider/terraform/locals blocks are not checked
	issues := runCheck(t, map[string]string{
		"main.tf": `
terraform {
  required_version = ">= 1.0"
}

locals {
  z = "z"
  a = "a"
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues for non-resource blocks, got %d", len(issues))
	}
}

func TestKeyAttributesRule_DataBlock(t *testing.T) {
	// data blocks are checked the same as resource blocks
	issues := runCheck(t, map[string]string{
		"main.tf": `
data "aws_s3_bucket" "example" {
  tags   = {}
  bucket = "my-bucket"
}
`,
	})
	if len(issues) != 1 {
		t.Fatalf("want 1 issue for data block, got %d", len(issues))
	}
}

func TestKeyAttributesRule_NoIssue_KeyAttributeAbsent(t *testing.T) {
	// If the key attribute is not defined at all in the resource, no issue
	issues := runCheck(t, map[string]string{
		"main.tf": `
resource "aws_s3_bucket" "example" {
  tags = {}
}
`,
	})
	if len(issues) != 0 {
		t.Errorf("want 0 issues when key attribute is not present, got %d", len(issues))
	}
}

func TestKeyAttributesRule_MultipleResources(t *testing.T) {
	// Only the misconfigured resource should emit an issue
	issues := runCheck(t, map[string]string{
		"main.tf": `
resource "aws_s3_bucket" "ok" {
  bucket = "good"
  tags   = {}
}

resource "aws_s3_bucket" "bad" {
  tags   = {}
  bucket = "oops"
}
`,
	})
	if len(issues) != 1 {
		t.Fatalf("want exactly 1 issue, got %d", len(issues))
	}
}

func TestKeyAttributesRule_NonCustomRunner_Noop(t *testing.T) {
	// If the rule receives a plain helper.Runner (not *custom.Runner), it
	// should return nil without panicking.
	rule := rules.NewKeyAttributesRule()
	base := helper.TestRunner(t, map[string]string{
		"main.tf": `resource "aws_s3_bucket" "x" { tags = {} }`,
	})
	if err := rule.Check(base); err != nil {
		t.Errorf("Check with non-custom runner returned error: %v", err)
	}
}
