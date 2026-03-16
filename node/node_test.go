package node_test

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/phazeight/tflint-ruleset-trusty/node"
)

func mustParseBody(t *testing.T, src string) *hclsyntax.Body {
	t.Helper()
	file, diags := hclsyntax.ParseConfig([]byte(src), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse error: %s", diags)
	}
	return file.Body.(*hclsyntax.Body)
}

func TestWrapAttribute(t *testing.T) {
	body := mustParseBody(t, `foo = "bar"`)
	attr := body.Attributes["foo"]
	n := node.WrapAttribute(attr)

	if !n.IsAttribute() {
		t.Error("IsAttribute() should be true")
	}
	if n.IsBlock() {
		t.Error("IsBlock() should be false")
	}
	if n.Name() != "foo" {
		t.Errorf("Name() = %q, want %q", n.Name(), "foo")
	}
	if n.Kind() != node.Attribute {
		t.Errorf("Kind() = %v, want Attribute", n.Kind())
	}
	if n.AsAttribute() == nil {
		t.Error("AsAttribute() should not be nil")
	}
	if n.AsBlock() != nil {
		t.Error("AsBlock() should be nil for an attribute")
	}
	if n.Lines() < 1 {
		t.Errorf("Lines() = %d, want >= 1", n.Lines())
	}
}

func TestWrapBlock(t *testing.T) {
	body := mustParseBody(t, `resource "aws_s3_bucket" "example" {}`)
	b := body.Blocks[0]
	n := node.WrapBlock(b)

	if n.IsAttribute() {
		t.Error("IsAttribute() should be false")
	}
	if !n.IsBlock() {
		t.Error("IsBlock() should be true")
	}
	if n.Name() != "resource aws_s3_bucket example" {
		t.Errorf("Name() = %q, want %q", n.Name(), "resource aws_s3_bucket example")
	}
	if n.Kind() != node.Block {
		t.Errorf("Kind() = %v, want Block", n.Kind())
	}
	if n.AsBlock() == nil {
		t.Error("AsBlock() should not be nil")
	}
	if n.AsAttribute() != nil {
		t.Error("AsAttribute() should be nil for a block")
	}
}

func TestWrapBlock_Dynamic(t *testing.T) {
	// dynamic blocks omit the block type from Name()
	body := mustParseBody(t, `
resource "x" "y" {
  dynamic "tag" {
    content {}
  }
}
`)
	inner := body.Blocks[0].Body
	var dynBlock *hclsyntax.Block
	for _, bl := range inner.Blocks {
		if bl.Type == "dynamic" {
			dynBlock = bl
			break
		}
	}
	if dynBlock == nil {
		t.Fatal("could not find dynamic block")
	}
	n := node.WrapBlock(dynBlock)
	if n.Name() != "tag" {
		t.Errorf("dynamic block Name() = %q, want %q", n.Name(), "tag")
	}
}

func TestOrderedInspectableNodesFrom_Order(t *testing.T) {
	// Attributes in a map have no guaranteed iteration order; nodes must come
	// out sorted by source position.
	body := mustParseBody(t, `
z = "last"
a = "first"
`)
	nodes := node.OrderedInspectableNodesFrom(body)
	if len(nodes) != 2 {
		t.Fatalf("want 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Name() != "z" {
		t.Errorf("first node = %q, want %q", nodes[0].Name(), "z")
	}
	if nodes[1].Name() != "a" {
		t.Errorf("second node = %q, want %q", nodes[1].Name(), "a")
	}
}

func TestOrderedInspectableNodesFrom_MixedKinds(t *testing.T) {
	body := mustParseBody(t, `
first_attr = "x"
some_block {}
second_attr = "y"
`)
	nodes := node.OrderedInspectableNodesFrom(body)
	if len(nodes) != 3 {
		t.Fatalf("want 3 nodes, got %d", len(nodes))
	}
	wantNames := []string{"first_attr", "some_block", "second_attr"}
	for i, want := range wantNames {
		if nodes[i].Name() != want {
			t.Errorf("node[%d] = %q, want %q", i, nodes[i].Name(), want)
		}
	}
}

func TestOrderedInspectableNodesFrom_Empty(t *testing.T) {
	body := mustParseBody(t, ``)
	nodes := node.OrderedInspectableNodesFrom(body)
	if len(nodes) != 0 {
		t.Errorf("want 0 nodes, got %d", len(nodes))
	}
}
