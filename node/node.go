package node

import (
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Node is a drop-in replacement for HCL's nodes.
type Node interface {
	Range() hcl.Range
}

// InspectableNode provides additional information about the node.
type InspectableNode interface {
	Node

	AsAttribute() *hclsyntax.Attribute
	AsBlock() *hclsyntax.Block

	IsAttribute() bool
	IsBlock() bool
	Lines() int
	Kind() Kind
	Name() string
	Type() string
}

// Kind indicates the kind of a node.
type Kind int

const (
	// Attribute node kind.
	Attribute Kind = iota

	// Block node kind.
	Block
)

// OrderedInspectableNodesFrom returns all attributes and blocks from the given
// body, sorted by their source position.
func OrderedInspectableNodesFrom(b *hclsyntax.Body) []InspectableNode {
	res := make(
		[]InspectableNode,
		0,
		len(b.Blocks)+len(b.Attributes),
	)

	for _, a := range b.Attributes {
		res = append(res, WrapAttribute(a))
	}
	for _, b := range b.Blocks {
		res = append(res, WrapBlock(b))
	}

	sort.SliceStable(res, func(l, r int) bool {
		ls := res[l].Range().Start.Byte
		rs := res[r].Range().Start.Byte
		return ls < rs
	})

	return res
}
