// Package buildutil contains common utility methods for
// generating bazel build files.
package buildutil

import (
	"github.com/bazelbuild/buildtools/build"
)

// Build string expression.
func StrExpr(value string) *build.StringExpr {
	return &build.StringExpr{
		Value: value,
		Token: value,
	}
}

// Build a list of string expressions.
func StrExprList(list ...string) []build.Expr {
	l := make([]build.Expr, 0, len(list))
	for _, s := range list {
		l = append(l, StrExpr(s))
	}
	return l
}

// Build list expression.
func ListExpr(items ...build.Expr) *build.ListExpr {
	return &build.ListExpr{List: items}
}

// Build `attr = <value>` expression.
func Attr(name string, value build.Expr) *build.AssignExpr {
	return &build.AssignExpr{
		LHS: &build.Ident{Name: name},
		Op:  "=",
		RHS: value,
	}
}

// Build `attr = "value"` expression.
func StrAttr(name, value string) *build.AssignExpr {
	return Attr(name, StrExpr(value))
}

func StrListAttr(name string, value ...string) *build.AssignExpr {
	return Attr(name, ListExpr(StrExprList(value...)...))
}

func Ident(expr build.Expr) string {
	switch expr := expr.(type) {
	case *build.Ident:
		return expr.Name
	case *build.TypedIdent:
		return expr.Ident.Name
	}
	return ""
}
