package rules

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// staticStringLiteral returns the string value of a node if and only if it is a
// static string literal with no interpolation. A quoted HCL string parses as a
// *hclsyntax.TemplateExpr; IsStringLiteral() is true only when the template is a
// single literal part with no `${...}` interpolation — anything interpolated is
// already partly a reference and must be skipped.
//
// Matching only TemplateExpr (not the child LiteralValueExpr parts) also avoids
// double-emitting during hclsyntax.VisitAll, which visits both the template and
// its literal parts.
func staticStringLiteral(node hclsyntax.Node) (string, bool) {
	tmpl, ok := node.(*hclsyntax.TemplateExpr)
	if !ok || !tmpl.IsStringLiteral() {
		return "", false
	}
	val, diags := tmpl.Value(nil)
	if diags.HasErrors() || val.Type() != cty.String {
		return "", false
	}
	return val.AsString(), true
}
