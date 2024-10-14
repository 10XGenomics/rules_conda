package main

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/10XGenomics/rules_conda/buildutil"
	"github.com/bazelbuild/buildtools/build"
)

func makeSpecFunc(specs map[string]*PkgSpec, extras []string,
	existing *build.Function) build.Function {
	allSpecs := make(map[string]struct{}, len(specs)+len(extras))
	specList := make([]*PkgSpec, 0, len(specs))
	for _, spec := range specs {
		specList = append(specList, spec)
		allSpecs[spec.Name] = struct{}{}
	}
	for _, pkg := range extras {
		allSpecs[pkg] = struct{}{}
	}
	sort.Slice(specList, func(i, j int) bool {
		return specList[i].DistName < specList[j].DistName
	})

	var condaRepo *build.StringExpr
	if existing != nil {
		if exp := getAttr("name", existing.Params); exp != nil {
			if s, ok := exp.RHS.(*build.StringExpr); ok && s.Value != "" {
				condaRepo = s
			}
		}
	}
	if condaRepo == nil {
		condaRepo = buildutil.StrExpr(buildutil.DefaultCondaRepo)
	}

	docstring := `Create remote repositories to download each conda package.

    Also create the repository rule to generate the complete distribution.  In
    general, other targets should depend on targets in the ` +
		"`@" + condaRepo.Value + "`" + `
    repository, rather than individual package repositories.

    Args:
        name (string): The name of the top level distribution repo.
    `
	body := make([]build.Expr, 1, len(specList)+3)
	body[0] = &build.StringExpr{
		Value:       docstring,
		TripleQuote: true,
	}
	var existingBody []build.Expr
	if existing != nil {
		existingBody = existing.Body
	}
	repoCall, pkgListExpr, aliases := getRepoCall(existingBody)
	if aliases != nil {
		for _, pair := range aliases.List {
			if key, ok := pair.Key.(*build.StringExpr); ok {
				allSpecs[key.Value] = struct{}{}
			}
		}
	}
	for _, spec := range specList {
		var rule *build.CallExpr
		rule, existingBody = spec.searchExistingCall(existingBody, allSpecs)
		updateIdent("conda_repo", "name", rule)
		body = append(body, rule)
	}
	var oldPkgList []build.Expr
	oldPkgList = pkgListExpr.List
	pkgList := make([]build.Expr, 0, len(specList)+len(extras))
	var pyVersion string
	for _, spec := range specList {
		if spec.Name == "python" {
			pyVersion = spec.Version
		}
		var str *build.StringExpr
		str, oldPkgList = updateSortedList(spec.Name, oldPkgList)
		pkgList = append(pkgList, str)
	}
	for _, pkg := range extras {
		if pkg == "" {
			fmt.Fprintln(os.Stderr, "WARNING: empty package name")
			continue
		}
		var str *build.StringExpr
		str, oldPkgList = updateSortedList(pkg, oldPkgList)
		pkgList = append(pkgList, str)
	}
	pkgListExpr.List = pkgList
	if pyVersion != "" && pyVersion[0] >= '2' && pyVersion[0] <= '9' {
		fmt.Fprint(os.Stderr,
			"Setting `py_version` ", pyVersion[:1],
			" (python version ", pyVersion, ")\n")
		updateValue("py_version", &build.LiteralExpr{Token: pyVersion[:1]}, repoCall)
	}
	body = append(body, repoCall)
	body = append(body, &build.CallExpr{
		X: &build.Ident{Name: "native.register_toolchains"},
		List: []build.Expr{
			&build.CallExpr{
				X: &build.DotExpr{
					X:    buildutil.StrExpr("@{}//:python_toolchain"),
					Name: "format",
				},
				List: []build.Expr{&build.Ident{Name: "name"}},
			},
		},
	})
	var existingComments build.Comments
	if existing != nil {
		existingComments = existing.Comments
	}
	return build.Function{
		Params: []build.Expr{
			&build.AssignExpr{
				LHS: &build.Ident{Name: "name"},
				Op:  "=",
				RHS: condaRepo,
			},
		},
		Comments: existingComments,
		Body:     body,
	}
}

func getRepoCall(existing []build.Expr) (*build.CallExpr, *build.ListExpr, *build.DictExpr) {
	var call *build.CallExpr
	for _, expr := range existing {
		if c, ok := expr.(*build.CallExpr); ok {
			if buildutil.Ident(c.X) == "conda_environment_repository" {
				call = c
				break
			}
		}
	}
	if call == nil {
		pkgList := &build.ListExpr{ForceMultiLine: true}
		repoAttrs := []build.Expr{
			buildutil.Attr("name", &build.Ident{Name: "name"}),
			buildutil.Attr("conda_packages", pkgList),
			buildutil.Attr("executable_packages", buildutil.ListExpr(
				buildutil.StrExprList(
					"conda",
					"python",
				)...,
			)),
		}
		return &build.CallExpr{
			X:              &build.Ident{Name: "conda_environment_repository"},
			List:           repoAttrs,
			ForceMultiLine: true,
		}, pkgList, nil
	}
	updateValue("name", &build.Ident{Name: "name"}, call)
	pkgListAttr := getAttr("conda_packages", call.List)
	aliasesExpr := getAttr("aliases", call.List)
	var aliases *build.DictExpr
	if aliasesExpr != nil && aliasesExpr.RHS != nil {
		aliases, _ = aliasesExpr.RHS.(*build.DictExpr)
	}
	if pkgListAttr == nil {
		pkgList := &build.ListExpr{ForceMultiLine: true}
		call.List = append(call.List, buildutil.Attr("conda_packages", pkgList))
		return call, pkgList, aliases
	}
	if pkgList, ok := pkgListAttr.RHS.(*build.ListExpr); ok {
		return call, pkgList, aliases
	}
	pkgList := &build.ListExpr{ForceMultiLine: true}
	pkgListAttr.RHS = pkgList
	return call, pkgList, aliases
}

func loadRule(repo, name string, file *build.File) {
	start := 0
	for i, expr := range file.Stmt {
		switch expr := expr.(type) {
		case *build.StringExpr:
			start = i + 1
		case *build.LoadStmt:
			if expr.Module != nil && expr.Module.Value == repo {
				for _, to := range expr.To {
					if to.Name == name {
						return
					}
				}
				id := &build.Ident{Name: name}
				expr.From = append(expr.From, id)
				expr.To = append(expr.To, id)
			}
			start = i + 1
		}
	}
	file.Stmt = append(
		append(file.Stmt[:start:start], buildutil.LoadExpr(repo, name)),
		file.Stmt[start:]...)
}

func docstring(outName string, file *build.File) {
	docstring := fmt.Sprintf(
		`This file contains the workspace rules for the conda environment.

To use, add
`+"```"+`
    load(%q, "conda_environment")
    conda_environment()
`+"```\nto your `WORKSPACE` file.\n", ":"+path.Base(outName))
	if len(file.Stmt) > 0 {
		if s, ok := file.Stmt[0].(*build.StringExpr); ok {
			if strings.Contains(s.Value, docstring) {
				return
			}
			if strings.TrimSpace(s.Value) != "" {
				s.Value = docstring + "\n" + s.Value
			} else {
				s.Value = docstring
			}
			return
		}
	}
	le := 1 + len(file.Stmt)
	if le < 2 {
		le = 4
	}
	stmt := make([]build.Expr, 1, le)
	stmt[0] = &build.StringExpr{
		Value:       docstring,
		Token:       `"""` + docstring + `"""`,
		TripleQuote: true,
	}
	file.Stmt = append(stmt, file.Stmt...)
}

func writeSpecs(specs map[string]*PkgSpec, extras []string, outName string) error {
	var file *build.File
	if b, err := os.ReadFile(outName); err == nil {
		file, err = build.ParseBzl(outName, b)
		if err != nil {
			fmt.Fprintln(os.Stderr, "WARNING: Existing file found, but could not be parsed.")
		}
	}
	if file == nil {
		file = &build.File{
			Path: path.Base(outName),
			Type: build.TypeBzl,
		}
	}
	docstring(outName, file)
	loadRule("@"+buildutil.BazelRulesConda+
		"//rules:conda_environment.bzl",
		"conda_environment_repository", file)
	loadRule("@"+buildutil.BazelRulesConda+
		"//rules:conda_package_repository.bzl",
		"conda_package_repository",
		file)
	addSpecFunc(specs, extras, file)
	return os.WriteFile(
		outName,
		build.Format(file), 0666)
}

func addSpecFunc(specs map[string]*PkgSpec, extras []string, file *build.File) {
	for _, expr := range file.Stmt {
		if def, ok := expr.(*build.DefStmt); ok && def.Name == "conda_environment" {
			def.Function = makeSpecFunc(specs, extras, &def.Function)
			return
		}
	}
	file.Stmt = append(file.Stmt,
		&build.DefStmt{
			Name:           "conda_environment",
			Function:       makeSpecFunc(specs, extras, nil),
			ForceMultiLine: true,
		})
}

var distNameClean = strings.NewReplacer(".", "_", "-", "_")

func (spec *PkgSpec) repoName() string {
	return "conda_package_" + distNameClean.Replace(spec.Name)
}

func getAttr(name string, list []build.Expr) *build.AssignExpr {
	for _, e := range list {
		if attr, ok := e.(*build.AssignExpr); ok {
			if buildutil.Ident(attr.LHS) == name {
				return attr
			}
		}
	}
	return nil
}

func (spec *PkgSpec) searchExistingCall(body []build.Expr,
	allSpecs map[string]struct{}) (*build.CallExpr, []build.Expr) {
	rn := spec.repoName()
	for len(body) > 0 {
		if c, ok := body[0].(*build.CallExpr); ok {
			if buildutil.Ident(c.X) == "conda_package_repository" {
				if repo := getAttr("name", c.List); repo != nil {
					if x, ok := repo.RHS.(*build.StringExpr); ok {
						if x.Value == rn {
							return spec.updateRule(c, allSpecs), body[1:]
						} else if x.Value > rn {
							break
						}
					}
				}
			}
		}
		body = body[1:]
	}
	return spec.repoRule(rn, allSpecs), body
}

func (spec *PkgSpec) repoRule(rn string, allSpecs map[string]struct{}) *build.CallExpr {
	c := build.CallExpr{
		X: &build.Ident{Name: "conda_package_repository"},
		List: []build.Expr{
			buildutil.StrAttr("name", rn),
			buildutil.Attr("base_urls", buildutil.ListExpr(
				buildutil.StrExprList(spec.BaseUrl)...)),
			buildutil.StrAttr("dist_name", spec.DistName),
			buildutil.StrAttr("sha256", spec.Sha256),
		},
		ForceMultiLine: true,
	}
	if strings.HasSuffix(spec.Url, ".conda") {
		c.List = append(c.List,
			buildutil.StrAttr("archive_type", "conda"))
	}
	spec.excludeDeps(&c, allSpecs)
	return &c
}

func (spec *PkgSpec) updateUrl(c *build.CallExpr) {
	ensureInList("base_urls", spec.BaseUrl, c)
}

func ensureInList(attr, value string, c *build.CallExpr) {
	if v := getAttr(attr, c.List); v != nil {
		if list, ok := v.RHS.(*build.ListExpr); ok {
			for _, e := range list.List {
				if s, ok := e.(*build.StringExpr); ok && s.Value == value {
					return
				}
			}
		}
		v.RHS = buildutil.ListExpr(buildutil.StrExpr(value))
	} else {
		c.List = append(c.List,
			buildutil.Attr(attr,
				buildutil.ListExpr(buildutil.StrExpr(value))))
	}
}

func removeAttr(c *build.CallExpr, i int) {
	if i == 0 {
		c.List = c.List[1:]
	} else if i == len(c.List)-1 {
		c.List = c.List[:len(c.List)-1]
	} else {
		c.List = append(c.List[:i], c.List[i+1:]...)
	}
}

func unsetStr(name string, c *build.CallExpr) {
	for i, e := range c.List {
		if attr, ok := e.(*build.AssignExpr); ok {
			if buildutil.Ident(attr.LHS) == name {
				removeAttr(c, i)
				return
			}
		}
	}
}

func updateStr(key, value string, c *build.CallExpr) {
	if v := getAttr(key, c.List); v != nil {
		if s, ok := v.RHS.(*build.StringExpr); ok {
			s.Value = value
		} else {
			v.RHS = buildutil.StrExpr(value)
		}
	} else {
		c.List = append(c.List, buildutil.StrAttr(key, value))
	}
}

func updateIdent(key, value string, c *build.CallExpr) {
	if v := getAttr(key, c.List); v != nil {
		if s, ok := v.RHS.(*build.Ident); ok {
			s.Name = value
		} else {
			v.RHS = &build.Ident{Name: value}
		}
	} else {
		c.List = append(c.List, buildutil.Attr(key, &build.Ident{Name: value}))
	}
}

func updateSortedList(value string, existing []build.Expr) (*build.StringExpr, []build.Expr) {
	for len(existing) > 0 {
		if s, ok := existing[0].(*build.StringExpr); ok {
			if s.Value == value {
				return s, existing[1:]
			} else if s.Value > value {
				break
			}
		}
		existing = existing[1:]
	}
	return buildutil.StrExpr(value), existing
}

func updateValue(key string, value build.Expr, c *build.CallExpr) {
	if v := getAttr(key, c.List); v != nil {
		if v.RHS.Comment() != nil {
			*value.Comment() = *v.RHS.Comment()
		}
		v.RHS = value
	} else {
		c.List = append(c.List, buildutil.Attr(key, value))
	}
}

func (spec *PkgSpec) updateRule(c *build.CallExpr, allSpecs map[string]struct{}) *build.CallExpr {
	spec.updateUrl(c)
	updateStr("dist_name", spec.DistName, c)
	updateStr("sha256", spec.Sha256, c)
	if strings.HasSuffix(spec.Url, ".conda") {
		updateStr("archive_type", "conda", c)
	} else if strings.HasSuffix(spec.Url, ".tar.bz2") {
		unsetStr("archive_type", c)
	}
	spec.excludeDeps(c, allSpecs)
	return c
}

func (spec *PkgSpec) excludeDeps(c *build.CallExpr, allSpecs map[string]struct{}) {
	for _, d := range spec.Depends {
		if strings.HasPrefix(d, "__") {
			// virtual package
			continue
		}
		if _, ok := allSpecs[d]; !ok {
			ensureInList("exclude_deps", d, c)
		}
	}
}
