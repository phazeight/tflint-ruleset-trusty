package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

// skipRoots lists traversal roots that are NOT resource references and must not
// be flagged by this rule (they are never count-conditional resources).
var skipRoots = map[string]bool{
	"var":       true,
	"local":     true,
	"path":      true,
	"terraform": true,
	"each":      true,
	"count":     true,
	"self":      true,
}

// skipBlockTypes lists block types whose bodies hold literal resource addresses.
// Terraform requires a concrete address in moved/import/removed, so one() is not
// valid syntax there and the [0] cannot be rewritten.
var skipBlockTypes = map[string]bool{
	"import":  true,
	"moved":   true,
	"removed": true,
}

// UseOneForConditionalRule flags resource_type.resource_name[0].attr traversals
// and suggests using one(resource_type.resource_name[*].attr) instead, which
// returns null safely when count = 0 rather than panicking.
type UseOneForConditionalRule struct {
	tflint.DefaultRule
}

// NewUseOneForConditionalRule creates a new UseOneForConditionalRule.
func NewUseOneForConditionalRule() *UseOneForConditionalRule {
	return &UseOneForConditionalRule{}
}

func (r *UseOneForConditionalRule) Name() string              { return "trusty_use_one_for_conditional" }
func (r *UseOneForConditionalRule) Enabled() bool             { return true }
func (r *UseOneForConditionalRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *UseOneForConditionalRule) Link() string {
	return "https://github.com/phazeight/tflint-ruleset-trusty/blob/main/docs/rules/trusty_use_one_for_conditional.md"
}

// Check walks every file's blocks and emits an issue for each ScopeTraversalExpr
// that contains a literal [0] index on a resource reference. Bodies of
// moved/import/removed blocks are skipped wholesale.
func (r *UseOneForConditionalRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}
		if err := r.checkBody(runner, body); err != nil {
			return err
		}
	}
	return nil
}

// checkBody recurses through nested blocks, skipping those whose addresses must
// stay literal, and scans each attribute expression of the bodies it keeps.
func (r *UseOneForConditionalRule) checkBody(runner tflint.Runner, body *hclsyntax.Body) error {
	// Attributes is a map; sort so issue order is stable across runs.
	names := make([]string, 0, len(body.Attributes))
	for name := range body.Attributes {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if err := r.checkExpr(runner, body.Attributes[name].Expr); err != nil {
			return err
		}
	}

	for _, block := range body.Blocks {
		if skipBlockTypes[block.Type] {
			continue
		}
		if err := r.checkBody(runner, block.Body); err != nil {
			return err
		}
	}
	return nil
}

// checkExpr walks one expression tree — traversals can be nested inside function
// calls, templates, and collection constructors.
func (r *UseOneForConditionalRule) checkExpr(runner tflint.Runner, expression hclsyntax.Expression) error {
	var visitErr error
	diags := hclsyntax.VisitAll(expression, func(node hclsyntax.Node) hcl.Diagnostics {
		if visitErr != nil {
			return nil
		}

		expr, ok := node.(*hclsyntax.ScopeTraversalExpr)
		if !ok {
			return nil
		}

		traversal := expr.Traversal
		if len(traversal) < 2 {
			return nil
		}

		// The first step is always a TraverseRoot — get the root name.
		root, ok := traversal[0].(hcl.TraverseRoot)
		if !ok {
			return nil
		}
		if skipRoots[root.Name] {
			return nil
		}

		// Find the position of the [0] TraverseIndex step.
		zeroIdx := -1
		for i, step := range traversal {
			ti, ok := step.(hcl.TraverseIndex)
			if !ok {
				continue
			}
			if ti.Key.Type() == cty.Number {
				bf := ti.Key.AsBigFloat()
				if bf.IsInt() {
					iv, _ := bf.Int64()
					if iv == 0 {
						zeroIdx = i
						break
					}
				}
			}
		}
		if zeroIdx < 0 {
			return nil
		}

		// The [0] must index the resource itself, not one of its list-typed
		// attributes. Only type.name[0] and module.name[0] (index 2), or
		// data.type.name[0] (index 3), address a count-conditional object.
		// Anything deeper — aws_organizations_organization.root.roots[0].id —
		// is an attribute index and cannot be rewritten with one().
		wantIdx := 2
		if root.Name == "data" {
			wantIdx = 3
		}
		if zeroIdx != wantIdx {
			return nil
		}

		// Build the resource reference prefix (everything before the [0]).
		prefixParts := make([]string, 0, zeroIdx)
		for _, step := range traversal[:zeroIdx] {
			switch s := step.(type) {
			case hcl.TraverseRoot:
				prefixParts = append(prefixParts, s.Name)
			case hcl.TraverseAttr:
				prefixParts = append(prefixParts, s.Name)
			}
		}
		prefix := strings.Join(prefixParts, ".")

		// Build the attribute suffix (everything after the [0]).
		var suffixParts []string
		for _, step := range traversal[zeroIdx+1:] {
			if ta, ok := step.(hcl.TraverseAttr); ok {
				suffixParts = append(suffixParts, ta.Name)
			}
		}
		suffix := strings.Join(suffixParts, ".")

		var msg string
		if suffix != "" {
			msg = fmt.Sprintf(
				`use one() instead of [0] indexing on %q — one(%s[*].%s) returns null safely when count = 0`,
				prefix, prefix, suffix,
			)
		} else {
			msg = fmt.Sprintf(
				`use one() instead of [0] indexing on %q — one(%s[*]) returns null safely when count = 0`,
				prefix, prefix,
			)
		}

		if err := runner.EmitIssue(r, msg, expr.SrcRange); err != nil {
			visitErr = err
		}
		return nil
	})
	if visitErr != nil {
		return visitErr
	}
	if diags.HasErrors() {
		return diags
	}
	return nil
}
