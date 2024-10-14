package buildutil

import "github.com/bazelbuild/buildtools/build"

// Build expression for `["//visibility:public"]`.
func PublicVis() *build.ListExpr {
	return ListExpr(StrExpr("//visibility:public"))
}

// Build expression for `glob(["l"])`.
func GlobExpr(l ...string) *build.CallExpr {
	return &build.CallExpr{
		X:    &build.Ident{Name: "glob"},
		List: []build.Expr{ListExpr(StrExprList(l...)...)},
	}
}

func LoadExpr(repo string, values ...string) *build.LoadStmt {
	names := make([]*build.Ident, len(values))
	for i, v := range values {
		names[i] = &build.Ident{Name: v}
	}
	load := &build.LoadStmt{
		Module: StrExpr(repo),
		From:   names,
		To:     names,
	}
	build.SortLoadArgs(load)
	return load
}
