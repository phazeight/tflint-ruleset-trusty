package visit_test

import (
	"errors"
	"testing"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/phazeight/tflint-ruleset-trusty/visit"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func TestFiles_VisitsAllFiles(t *testing.T) {
	runner := helper.TestRunner(t, map[string]string{
		"a.tf": `resource "x" "a" {}`,
		"b.tf": `resource "x" "b" {}`,
	})

	visited := map[string]bool{}
	err := visit.Files(nil, runner, func(body *hclsyntax.Body, bytes []byte) error {
		// identify the file by the label of the first block
		if len(body.Blocks) > 0 {
			visited[body.Blocks[0].Labels[1]] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Files returned error: %v", err)
	}
	if !visited["a"] || !visited["b"] {
		t.Errorf("Files did not visit all files: visited = %v", visited)
	}
}

func TestFiles_PropagatesVisitorError(t *testing.T) {
	runner := helper.TestRunner(t, map[string]string{
		"main.tf": `locals { x = 1 }`,
	})

	sentinel := errors.New("visitor error")
	err := visit.Files(nil, runner, func(_ *hclsyntax.Body, _ []byte) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("Files should propagate visitor error, got: %v", err)
	}
}

func TestFiles_EmptyRunner(t *testing.T) {
	runner := helper.TestRunner(t, map[string]string{})
	calls := 0
	err := visit.Files(nil, runner, func(_ *hclsyntax.Body, _ []byte) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("Files returned error on empty runner: %v", err)
	}
	if calls != 0 {
		t.Errorf("visitor called %d times on empty runner, want 0", calls)
	}
}

func TestBlocks_VisitsAllTopLevelBlocks(t *testing.T) {
	runner := helper.TestRunner(t, map[string]string{
		"main.tf": `
resource "aws_s3_bucket" "a" {}
resource "aws_s3_bucket" "b" {}
locals { x = 1 }
`,
	})

	var blockTypes []string
	err := visit.Blocks(nil, runner, func(b *hclsyntax.Block, _ []byte) error {
		blockTypes = append(blockTypes, b.Type)
		return nil
	})
	if err != nil {
		t.Fatalf("Blocks returned error: %v", err)
	}
	// Two resource blocks + one locals block
	if len(blockTypes) != 3 {
		t.Errorf("visited %d blocks, want 3: %v", len(blockTypes), blockTypes)
	}
}

func TestBlocks_DoesNotDescendIntoNestedBlocks(t *testing.T) {
	runner := helper.TestRunner(t, map[string]string{
		"main.tf": `
resource "aws_instance" "x" {
  lifecycle {
    ignore_changes = []
  }
}
`,
	})

	count := 0
	err := visit.Blocks(nil, runner, func(_ *hclsyntax.Block, _ []byte) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("Blocks returned error: %v", err)
	}
	// Only the top-level resource block, not the nested lifecycle block
	if count != 1 {
		t.Errorf("Blocks visited %d blocks, want 1 (top-level only)", count)
	}
}

func TestBlocks_PropagatesVisitorError(t *testing.T) {
	runner := helper.TestRunner(t, map[string]string{
		"main.tf": `resource "x" "y" {}`,
	})

	sentinel := errors.New("block visitor error")
	err := visit.Blocks(nil, runner, func(_ *hclsyntax.Block, _ []byte) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("Blocks should propagate visitor error, got: %v", err)
	}
}

func TestBlocks_MultipleFiles(t *testing.T) {
	runner := helper.TestRunner(t, map[string]string{
		"a.tf": `resource "x" "a" {}`,
		"b.tf": `resource "x" "b" {}`,
	})

	var labels []string
	err := visit.Blocks(nil, runner, func(b *hclsyntax.Block, _ []byte) error {
		labels = append(labels, b.Labels[1])
		return nil
	})
	if err != nil {
		t.Fatalf("Blocks returned error: %v", err)
	}
	if len(labels) != 2 {
		t.Errorf("visited %d blocks across 2 files, want 2", len(labels))
	}
}

// Ensure visit.Files and visit.Blocks accept a nil rule (the rule param is
// unused — it exists only for future hook compatibility).
func TestFiles_NilRuleAccepted(t *testing.T) {
	runner := helper.TestRunner(t, map[string]string{"main.tf": `locals {}`})
	var nilRule tflint.Rule
	if err := visit.Files(nilRule, runner, func(_ *hclsyntax.Body, _ []byte) error { return nil }); err != nil {
		t.Fatalf("Files with nil rule returned error: %v", err)
	}
}
