package conda

import (
	"bufio"
	"debug/elf"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/10XGenomics/rules_conda/buildutil"
	"github.com/10XGenomics/rules_conda/licensing"
	"github.com/bazelbuild/buildtools/build"
)

func skipFile(path string, excludePatterns []string) (bool, error) {
	for len(path) > 0 && path != "/" && path != "." {
		for _, pattern := range excludePatterns {
			match, err := filepath.Match(pattern, path)
			if err != nil {
				return false, err
			} else if match {
				return true, nil
			}
		}
		path = filepath.Dir(path)
	}
	return false, nil
}

// Load metadata for the package in the given directory.
func (pkg *Package) Load(dir string, extraDeps map[string][]string,
	excludePatterns []string, requirePaths bool) error {
	pkg.Dir = dir
	if err := pkg.Index.Load(dir, extraDeps); err != nil {
		return fmt.Errorf("loading index metadata: %w", err)
	}
	pkg.linkPython, pkg.pyStubs = isLinkPython(dir)
	pkg.isExecutable = pkg.Index.Entry != "" && len(pkg.pyStubs) == 0
	pkg.License.Name = pkg.Index.Name
	pkg.License.Qualifiers = &licensing.CondaPackageQualifiers{
		Build:  pkg.Index.Build,
		Subdir: pkg.Index.Subdir,
	}
	if err := pkg.Paths.Load(dir); err != nil {
		if requirePaths || !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("loading package file manifest: %w", err)
		}
	} else if len(excludePatterns) > 0 {
		filteredPath := make([]condaFilePath, 0, len(pkg.Paths.Paths))
		for _, c := range pkg.Paths.Paths {
			if sk, err := skipFile(c.Path, excludePatterns); err != nil {
				return err
			} else if !sk {
				filteredPath = append(filteredPath, c)
			} else if err := os.Remove(path.Join(dir, c.Path)); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
		pkg.Paths.Paths = filteredPath
	}
	if len(pkg.Paths.Paths) > 0 {
		pkg.allFiles = make(map[string]struct{}, len(pkg.Paths.Paths))
		for _, f := range pkg.Paths.Paths {
			pkg.allFiles[f.Path] = struct{}{}
		}
	}
	return nil
}

func (pkg *Package) depsRule(includeDeps, excludeDeps []string, condaRepo string) *build.CallExpr {
	depsSet := make(map[string]struct{}, len(includeDeps)+len(pkg.Index.Depends))
	for _, dep := range pkg.Index.Depends {
		dep := strings.TrimSpace(dep)
		if i := strings.IndexByte(dep, ' '); i > 0 {
			dep = dep[:i]
		}
		if !strings.HasPrefix(dep, "__") && !strInList(dep, excludeDeps) {
			depsSet[condaRepo+"//:"+dep] = struct{}{}
		}
	}
	for _, dep := range includeDeps {
		dep = strings.TrimSpace(dep)
		if dep != "" {
			depsSet[condaRepo+"//:"+dep] = struct{}{}
		}
	}
	var deps build.ListExpr
	if len(depsSet) > 0 {
		depList := make([]string, 0, len(depsSet))
		for dep := range depsSet {
			depList = append(depList, dep)
		}
		sort.Strings(depList)
		deps.List = buildutil.StrExprList(depList...)
	}
	return &build.CallExpr{
		X: &build.Ident{Name: "conda_deps"},
		List: []build.Expr{
			buildutil.StrAttr("name", "conda_deps"),
			buildutil.Attr("deps", &deps),
			buildutil.Attr("visibility", buildutil.ListExpr(
				buildutil.StrExpr(condaRepo+"//:__pkg__"))),
		},
	}
}

func strInList(dep string, exclude []string) bool {
	for _, v := range exclude {
		if v == dep {
			return true
		}
	}
	return false
}

// Generates a BUILD file (with a filegroup rule) for the
// untarred package.
func (pkg *Package) MakeTarballBuild(includeDeps, excludeDeps []string,
	condaRepo string,
	ccInclude []string, distName, url string) error {
	if pkg.Name() == "python" {
		if err := pkg.writePythonVars(condaRepo); err != nil {
			return err
		}
	}
	groups := make([]build.Expr, 1, 6)
	groups[0] = buildutil.LoadExpr(
		"@"+buildutil.BazelRulesConda+"//rules:conda_manifest.bzl",
		"conda_deps",
		"conda_files",
		"conda_manifest")
	if len(pkg.pyStubs) > 0 {
		groups = append(groups, buildutil.LoadExpr(
			"@bazel_skylib//rules:write_file.bzl",
			"write_file"))
	}
	if pkg.linkPython {
		groups = append(groups, buildutil.LoadExpr(
			"@conda_package_python//:vars.bzl",
			"PYTHON_PREFIX"))
	}
	groups = append(groups, pkg.License.Rules(pkg.Name(), pkg.Index.Version, url)...)
	groups = append(groups, pkg.fileGroups(ccInclude, condaRepo)...)
	groups = append(groups, pkg.depsRule(includeDeps, excludeDeps, condaRepo))
	f := build.File{
		Path: "BUILD",
		Type: build.TypeBuild,
		Comments: build.Comments{
			Before: []build.Comment{
				{
					Token: fmt.Sprintf(
						"# Generated BUILD file for %s\n",
						pkg.Name()),
				},
				{
					Token: fmt.Sprintf(
						"# Package dist name: %s\n",
						distName),
				},
			},
		},
		Stmt: groups,
	}
	build.DisableRewrites = []string{"listsort"}
	return os.WriteFile(
		path.Join(pkg.Dir, "BUILD.bazel"),
		build.Format(&f), 0666)
}

var (
	libToolRe = regexp.MustCompile(`^lib/[^/]*\.la$`)
	aLibRe    = regexp.MustCompile(`^lib/[^/]*\.(?:l?a|lo|lib)(?:\.|$)`)
	soLibRe   = regexp.MustCompile(`^lib/[^/]*\.so(?:\..*)?`)
	// Header files are allowed to be at most one level deep.
	hdrRe = regexp.MustCompile(
		`^(.*/)?[^/]*\.(?:c(?:c|pp|xx|\+\+)?|C|h(?:h|pp|xx)?|ipp|inc)$`)
	srcRe = regexp.MustCompile(
		`.*\.(?:c(?:c|pp|xx|\+\+)?|C|h(?:h|pp|xx)?|ipp|inc|asm|[Sso])$`)
	pyRe = regexp.MustCompile(`\.py[co]?$`)
	// Check possible c/c++ header files.
	cHeaderContentRe = regexp.MustCompile(`/[/*]|{|.;$|` +
		`^\s*#\s*(?:include\s*["<]|define\s|pragma\s|if(?:n?def)?\s)`)
)

type fileType int

const (
	header = fileType(iota)
	aLib
	soLib
	libTool
	srcFile
	pyFile
	otherFile
	metadataFile
)

func anyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func typeFromName(file string, ccInclude []string) fileType {
	if strings.HasPrefix(file, "info/") {
		return metadataFile
	} else if libToolRe.MatchString(file) {
		return libTool
	} else if soLibRe.MatchString(file) {
		return soLib
	} else if aLibRe.MatchString(file) {
		return aLib
	} else if hdrPrefix := hdrRe.FindStringSubmatch(file); len(hdrPrefix) > 1 &&
		(len(ccInclude) == 0 && (strings.HasPrefix(hdrPrefix[1], "include/") ||
			strings.Contains(hdrPrefix[1], "/include/")) ||
			anyPrefix(file, ccInclude)) {
		return header
	} else if srcRe.MatchString(file) {
		return srcFile
	} else if pyRe.MatchString(file) {
		return pyFile
	} else if ext := filepath.Ext(file); ext == "" && anyPrefix(file, ccInclude) {
		// Maybe a C/C++ header file without an extension.
		f, err := os.Open(file)
		if err == nil {
			defer f.Close()
			scanner := bufio.NewScanner(f)
			// Limit line length.
			var buf [512]byte
			scanner.Buffer(buf[:], 512)
			i := 0
			for scanner.Scan() {
				line := scanner.Bytes()
				if i == 0 && len(line) > 2 &&
					line[0] == '#' && line[1] == '!' || i > 20 || !utf8.Valid(line) {
					// Files with #! lines aren't header files.
					// We also don't want to search too far into a file looking
					// for something that looks like c++ before we give up.
					// Also files which aren't valid utf-8 are probably binary
					// and should be ignored.
					break
				}
				if cHeaderContentRe.Match(line) {
					return header
				}
				i++
			}
		}
	}
	return otherFile
}

// isLinkSafe returns true if it is expected to be safe to use
// ctx.actions.symlink install the file into the final location, rather than
// actually copying it.
//
// Generally these are files where it is unlikely that someone will be resolving
// the symlinks.  That does _not_ include binaries or script files
// (including python sources).
func isLinkSafe(fn string) bool {
	fn = path.Base(fn)
	switch fn {
	case "Makefile", "ChangeLog",
		"LICENSE",
		"CHANGELOG",
		"AUTHORS",
		"INSTALLER",
		"METADATA",
		"README",
		"RECORD",
		"REQUESTED",
		"WHEEL",
		"py.typed":
		return true
	}
	switch filepath.Ext(fn) {
	case ".pc", ".in",
		".md", ".rst", ".txt", ".css", ".html", ".js",
		".pickle", ".json",
		".gif", ".jpg", ".jpeg", ".png", ".ico",
		".pyi",
		".whl", ".zip",
		".orig",
		".o", ".so":
		return true
	}
	return false
}

// isPython38 returns true if the package is for python itself and the version
// is at least 3.8.
func (pkg *Package) isPython38() bool {
	if pkg.Name() != "python" {
		return false
	}
	version_parts := strings.SplitN(pkg.Index.Version, ".", 3)
	if len(version_parts) < 2 {
		return false
	}
	if i, err := strconv.Atoi(version_parts[0]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing python version %s: %v\n", pkg.Index.Version, err)
	} else if i < 3 {
		return false
	}
	if i, err := strconv.Atoi(version_parts[1]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing python version %s: %v\n", pkg.Index.Version, err)
	} else if i < 8 {
		return false
	}
	return true
}

func condaVis(len int, condaRepo string) build.Expr {
	if len == 0 {
		return buildutil.ListExpr(buildutil.StrExpr(condaRepo + "//:__pkg__"))
	} else {
		return buildutil.PublicVis()
	}
}

// parseStubDecl parses a string like `pytest = pytest:console_main` into the
// stub name and content.
func parseStubDecl(decl string) (string, []string, error) {
	name, after, found := strings.Cut(decl, "=")
	if !found {
		return "", nil, fmt.Errorf("invalid stub declaration %q", decl)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil, fmt.Errorf("invalid stub declaration %q", decl)
	}
	module, function, found := strings.Cut(strings.TrimSpace(after), ":")
	if !found {
		return "", nil, fmt.Errorf("invalid stub declaration %q", decl)
	}
	module = strings.TrimSpace(module)
	function = strings.TrimSpace(function)
	if module == "" || function == "" {
		return "", nil, fmt.Errorf("invalid stub declaration %q", decl)
	}
	return name, []string{
		"#!/usr/bin/env python3",
		"",
		"import sys",
		fmt.Sprintf("from %s import %s", module, function),
		"",
		`if __name__ == "__main__":`,
		fmt.Sprintf(`    sys.exit(%s())`, function),
		"",
	}, nil
}

func (pkg *Package) fileGroups(ccInclude []string, condaRepo string) []build.Expr {
	files, symlinks, executable := pkg.Paths.filesList(pkg.Dir)

	linkMap := make(map[string]*symlinkEntry)
	for i := range symlinks {
		linkMap[symlinks[i].location] = &symlinks[i]
	}
	py := make([]build.Expr, 0, len(files))
	laLibs := make([]build.Expr, 0, len(files)/2)
	staticLibs := make([]build.Expr, 0, len(files)/2)
	dyLibs := make([]build.Expr, 0, len(files)/2)
	hdrs := make([]build.Expr, 0, len(files))
	var hdrsWithPlaceholder []build.Expr
	runfiles := make([]build.Expr, 0, len(files))
	linkSafeRunfiles := make([]build.Expr, 0, len(files))
	// Exclude libpython from alibs and solibs, because they're not actually
	// required for linking an extension module in python 3.8+,
	// which is the most common use case for depending on the python package
	// in a cc_* target.
	py38 := pkg.isPython38()
	for _, file := range files {
		switch typeFromName(file.Path, ccInclude) {
		case aLib:
			if py38 {
				linkSafeRunfiles = append(linkSafeRunfiles, buildutil.StrExpr(file.Path))
			} else {
				if file.NeedsTranslate() {
					panic(file.Path + " is a static library but has a placeholder.")
				}
				staticLibs = append(staticLibs, buildutil.StrExpr(file.Path))
			}
		case soLib:
			if py38 {
				linkSafeRunfiles = append(linkSafeRunfiles, buildutil.StrExpr(file.Path))
			} else {
				if file.NeedsTranslate() {
					panic(file.Path + " is a dynamic library but has a placeholder.")
				}
				if soName := getSoName(path.Join(pkg.Dir, file.Path)); soName != "" &&
					soName != path.Base(file.Path) {
					// It is not uncommon for a `.so` file to set DT_SONAME to
					// something other than the name of the `.so` file itself.
					// Most commonly, you have a file libfoo.so.1.2 with
					// `libfoo.so.1` as the DT_SONAME, where that path is a
					// symlink to the actual file.
					// Unfortunately this causes problems in bazel, because
					// either
					// 1. You link against both "files", which is redundant,
					//    but worse, in various situations they end up both
					//    getting treated like regular files which ends up
					//    bloating the sizes of release tarballs or remote
					//    execution inputs.
					// 2. Try to link with the symlink, which will be a broken
					//    link, at least in remote execution.
					// 3. Link with the actual file, which works fine at link
					//    time, but at runtime you have a problem because the
					//    file you linked against, which bazel will put in the
					//    runfiles (and rpath) won't actually be the file that
					//    the dynamic linker is looking for.
					//
					// The solution here is that we reverse the symlink, so that
					// we treat `libfoo.so.1` as the actual file and
					// `libfoo.so.1.2` as the symlink.
					soPath := path.Join(path.Dir(file.Path), soName)
					if link := linkMap[soPath]; link != nil {
						delete(linkMap, soPath)
						linkMap[file.Path] = link
						link.location = file.Path
						link.relPath = soName
						dyLibs = append(dyLibs, buildutil.StrExpr(soPath))
						continue
					}
				}
				dyLibs = append(dyLibs, buildutil.StrExpr(file.Path))
			}
		case header:
			if file.NeedsTranslate() {
				hdrsWithPlaceholder = append(hdrsWithPlaceholder, buildutil.StrExpr(file.Path))
			} else {
				hdrs = append(hdrs, buildutil.StrExpr(file.Path))
			}
		case libTool:
			if py38 {
				runfiles = append(runfiles, buildutil.StrExpr(file.Path))
			} else {
				laLibs = append(laLibs, buildutil.StrExpr(file.Path))
			}
		case pyFile:
			py = append(py, buildutil.StrExpr(file.Path))
		case metadataFile:
		case srcFile:
			linkSafeRunfiles = append(linkSafeRunfiles, buildutil.StrExpr(file.Path))
		default:
			if !file.NeedsTranslate() && isLinkSafe(file.Path) {
				linkSafeRunfiles = append(linkSafeRunfiles, buildutil.StrExpr(file.Path))
			} else {
				runfiles = append(runfiles, buildutil.StrExpr(file.Path))
			}
		}
	}
	if len(staticLibs) > 0 {
		// Move a lib for the package name to the end of the list, in case
		// other libs along the way depend on it.
		libName := pkg.Name()
		if i := strings.IndexByte(libName, '-'); i > 0 {
			libName = libName[:i]
		}
		if !strings.HasPrefix(libName, "lib") {
			libName = "lib" + libName
		}
		sort.SliceStable(staticLibs, func(i, j int) bool {
			ni := path.Base(staticLibs[i].(*build.StringExpr).Value)
			if d := strings.LastIndexByte(ni, '.'); d > 0 {
				ni = ni[:d]
			}
			nj := path.Base(staticLibs[j].(*build.StringExpr).Value)
			if d := strings.LastIndexByte(nj, '.'); d > 0 {
				nj = nj[:d]
			}
			return ni != libName && nj == libName
		})
	}
	// Whether to set `target_compatible_with` on the target.
	// It can make error messages a bit more clear in some situations,
	// but isn't essential, so false negatives are ok.
	archSpecific := py38 || len(staticLibs) > 0 || len(laLibs) > 0 || len(dyLibs) > 0
	result := make([]build.Expr, 1, 9)
	repoName := pkg.RepoName()
	fmt.Fprintf(os.Stderr, "Generating BUILD file for %s\n", repoName)
	filegroup := &build.Ident{Name: "filegroup"}
	makeFilegroup := func(name string, srcs []build.Expr) *build.CallExpr {
		return &build.CallExpr{
			X: filegroup,
			List: []build.Expr{
				buildutil.StrAttr("name", name),
				buildutil.Attr("srcs", buildutil.ListExpr(srcs...)),
				buildutil.Attr("visibility", buildutil.PublicVis()),
			},
		}
	}
	filesRule := &build.CallExpr{
		X: &build.Ident{Name: "conda_files"},
		List: []build.Expr{
			buildutil.StrAttr("name", "files"),
			buildutil.Attr("visibility", condaVis(0, condaRepo)),
		},
	}
	result[0] = filesRule
	if len(py) > 0 {
		filesRule.List = append(filesRule.List,
			buildutil.StrListAttr("py_srcs", ":py_srcs"))
		result = append(result, makeFilegroup("py_srcs", py))
	}
	if len(laLibs) > 0 {
		filesRule.List = append(filesRule.List,
			buildutil.StrListAttr("lalibs", ":lalibs"))
		result = append(result, makeFilegroup("lalibs", laLibs))
	}
	if len(staticLibs) > 0 {
		filesRule.List = append(filesRule.List,
			buildutil.StrListAttr("staticlibs", ":staticlibs"))
		result = append(result, makeFilegroup("staticlibs", staticLibs))
	}
	if len(dyLibs) > 0 {
		filesRule.List = append(filesRule.List,
			buildutil.StrListAttr("dylibs", ":dylibs"))
		result = append(result, makeFilegroup("dylibs", dyLibs))
	}
	if len(hdrs) > 0 {
		filesRule.List = append(filesRule.List,
			buildutil.StrListAttr("hdrs", ":hdrs"))
		result = append(result, makeFilegroup("hdrs", hdrs))
	}
	if len(hdrsWithPlaceholder) > 0 {
		filesRule.List = append(filesRule.List,
			buildutil.StrListAttr("hdrs_with_placeholders", ":hdrs_with_placeholders"))
		result = append(result, makeFilegroup("hdrs_with_placeholders", hdrsWithPlaceholder))
	}
	if len(runfiles) > 0 {
		filesRule.List = append(filesRule.List,
			buildutil.StrListAttr("runfiles", ":runfiles"))
		result = append(result, makeFilegroup("runfiles", runfiles))
	}
	if len(linkSafeRunfiles) > 0 {
		filesRule.List = append(filesRule.List,
			buildutil.StrListAttr("link_safe_runfiles", ":link_safe_runfiles"))
		result = append(result, makeFilegroup("link_safe_runfiles", linkSafeRunfiles))
	}
	if len(pkg.pyStubs) > 0 {
		stubList := &build.ListExpr{
			List: make([]build.Expr, 0, len(pkg.pyStubs)),
		}
		stubGroup := &build.CallExpr{
			X: filegroup,
			List: []build.Expr{
				buildutil.StrAttr("name", "py_stubs"),
				buildutil.Attr("srcs", stubList),
			},
		}
		writeFileRule := &build.Ident{Name: "write_file"}
		for _, stub := range pkg.pyStubs {
			name, content, err := parseStubDecl(stub)
			if err != nil {
				panic(err)
			}
			target := ":" + name + "_stub"
			fn := "bin/" + name
			result = append(result, &build.CallExpr{
				X: writeFileRule,
				List: []build.Expr{
					buildutil.StrAttr("name", target[1:]),
					buildutil.StrAttr("out", fn),
					buildutil.Attr("content", buildutil.ListExpr(
						buildutil.StrExprList(content...)...,
					)),
				},
			})
			stubList.List = append(stubList.List, buildutil.StrExpr(target))
		}
		result = append(result, stubGroup)
	}

	if len(ccInclude) > 0 {
		pkg.includeDirs = ccInclude
	}
	if len(hdrs) > 0 || len(hdrsWithPlaceholder) > 0 {
		if len(ccInclude) == 0 {
			includeDirs := make(map[string]struct{}, 1)
			for _, hdrExpr := range hdrs {
				hdr := hdrExpr.(*build.StringExpr).Value
				i := strings.Index(hdr, "/include/")
				if i > 0 {
					includeDirs[hdr[:i+len("/include")]] = struct{}{}
				}
				// special case: python headers, which are in include/python
				// but usually included without "python/"
				if pkg.Name() == "python" {
					dir, file := path.Split(hdr)
					if file == "Python.h" {
						includeDirs[path.Clean(dir)] = struct{}{}
					}
				}
			}
			pkg.includeDirs = make([]string, 1, 1+len(includeDirs))
			pkg.includeDirs[0] = "include"
			for inc := range includeDirs {
				pkg.includeDirs = append(pkg.includeDirs, inc)
			}
			sort.Strings(pkg.includeDirs)
		}
	}
	result = append(result,
		pkg.manifestRule(symlinks, executable, archSpecific),
	)
	return result
}

// getSoName returns the first DT_SONAME string in the shared object at the
// given path, or an empty string if the file can't be parsed as an ELF shared
// object or if there is no such string.
func getSoName(soPath string) string {
	f, err := elf.Open(soPath)
	if err != nil {
		return ""
	}
	defer f.Close()
	ns, err := f.DynString(elf.DT_SONAME)
	if err != nil {
		return ""
	}
	for _, n := range ns {
		if n != "" {
			return n
		}
	}
	return ""
}

func formatStringDict(m []symlinkEntry) *build.DictExpr {
	result := build.DictExpr{
		ForceMultiLine: len(m) > 0,
	}
	for _, p := range m {
		result.List = append(result.List, &build.KeyValueExpr{
			Key:   buildutil.StrExpr(p.location),
			Value: buildutil.StrExpr(p.relPath),
		})
	}
	return &result
}

func (pkg *Package) manifestRule(symlinks []symlinkEntry,
	executable []string,
	archSpecific bool) *build.CallExpr {
	c := build.CallExpr{
		X: &build.Ident{Name: "conda_manifest"},
		List: []build.Expr{
			buildutil.StrAttr("name", "conda_metadata"),
			buildutil.StrAttr("manifest", pkg.Paths.Manifest[0]),
			buildutil.StrAttr("index", "info/index.json"),
			buildutil.Attr("info_files", buildutil.ListExpr(
				buildutil.StrExprList(append(pkg.Paths.Manifest[1:],
					"info/index.json")...)...)),
		},
	}
	if len(symlinks) > 0 {
		c.List = append(c.List, buildutil.Attr("symlinks", formatStringDict(symlinks)))
	}
	if len(pkg.includeDirs) > 0 {
		c.List = append(c.List,
			buildutil.Attr("includes", buildutil.ListExpr(
				buildutil.StrExprList(pkg.includeDirs...)...)))
	}
	if pkg.linkPython {
		c.List = append(c.List, buildutil.StrAttr("noarch", "python"))
		c.List = append(c.List, buildutil.Attr("python_prefix", &build.Ident{
			Name: "PYTHON_PREFIX",
		}))
	} else if archSpecific {
		constraints := buildutil.ListExpr(buildutil.StrExprList(
			"@platforms//os:linux",
			"@platforms//cpu:x86_64")...)
		c.List = append(c.List, buildutil.Attr("exec_compatible_with", constraints))
		c.List = append(c.List, buildutil.Attr("target_compatible_with", constraints))
	}
	if e := pkg.executable(); e != "" {
		c.List = append(c.List, buildutil.StrAttr("executable", e))
	}
	if len(executable) > 0 {
		c.List = append(c.List, buildutil.Attr("executables", buildutil.ListExpr(
			buildutil.StrExprList(
				executable...,
			)...)))
	}
	if len(pkg.pyStubs) > 0 {
		c.List = append(c.List, buildutil.StrAttr("py_stubs", ":py_stubs"))
	}
	c.List = append(c.List, buildutil.Attr("visibility", buildutil.PublicVis()))
	return &c
}
