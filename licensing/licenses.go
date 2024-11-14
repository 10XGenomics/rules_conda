// Package licensing contains methods for working with license specifiers.
package licensing

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime/trace"
	"strings"
	"time"

	"github.com/10XGenomics/rules_conda/buildutil"
	"github.com/bazelbuild/buildtools/build"
)

type LicenseInfo struct {
	Licenses []License
	Version  string
	Homepage string
	Source   string
	Url      string
	Manifest string
	// pURL metadata
	Qualifiers PurlQualifiers
	Name       string
	about      aboutJson
}

type License struct {
	CanonicalId string
	Copyright   string
	File        string
	Kinds       []string
}

// PurlQualifiers provides an interface for supplying source-dependent
// qualifier information for packages.
type PurlQualifiers interface {
	Append(buf *strings.Builder)
}

// CondaPackageQualifiers stores qualifiers for conda packages.
//
// See https://github.com/package-url/purl-spec/blob/master/PURL-TYPES.rst#conda
type CondaPackageQualifiers struct {
	Build   string
	Channel string
	Subdir  string
	Type    string
}

func (q *CondaPackageQualifiers) Append(buf *strings.Builder) {
	if q == nil {
		return
	}
	sep := '?'
	if q.Build != "" {
		sep = '&'
		if _, err := buf.WriteString("?build="); err != nil {
			panic(err)
		}
		if _, err := buf.WriteString(q.Build); err != nil {
			panic(err)
		}
	}
	if q.Channel != "" {
		if _, err := buf.WriteRune(sep); err != nil {
			panic(err)
		}
		sep = '&'
		if _, err := buf.WriteString("channel="); err != nil {
			panic(err)
		}
		if _, err := buf.WriteString(q.Channel); err != nil {
			panic(err)
		}
	}
	if q.Subdir != "" {
		if _, err := buf.WriteRune(sep); err != nil {
			panic(err)
		}
		sep = '&'
		if _, err := buf.WriteString("subdir="); err != nil {
			panic(err)
		}
		if _, err := buf.WriteString(q.Subdir); err != nil {
			panic(err)
		}
	}
	if q.Type != "" {
		if _, err := buf.WriteRune(sep); err != nil {
			panic(err)
		}
		if _, err := buf.WriteString("type="); err != nil {
			panic(err)
		}
		if _, err := buf.WriteString(q.Type); err != nil {
			panic(err)
		}
	}
}

type aboutJson struct {
	License       string          `json:"license"`
	Family        string          `json:"license_family"`
	LicenseFiles  json.RawMessage `json:"license_file"`
	LicenseUrl    string          `json:"license_url"`
	Home          string          `json:"home"`
	DevUrl        string          `json:"dev_url"`
	DocUrl        string          `json:"doc_url"`
	DocSrc        string          `json:"doc_source_url"`
	Channels      []string        `json:"channels"`
	useLicenseUrl bool
}

func (a *aboutJson) Channel() string {
	if a == nil {
		return ""
	}
	if len(a.Channels) == 0 {
		return ""
	}
	var best string
	for _, ch := range a.Channels {
		cb := path.Base(ch)
		if cb == "main" {
			return "main"
		}
		if best == "" || cb == "free" || cb == "defaults" {
			best = cb
		}
	}
	return best
}

func (info *LicenseInfo) LoadConda(ctx context.Context, dir string) error {
	defer trace.StartRegion(ctx, "LicenseInfo.Load").End()
	b, err := os.ReadFile(path.Join(dir, "info", "about.json"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &info.about); err != nil {
		return err
	}
	if q, ok := info.Qualifiers.(*CondaPackageQualifiers); ok && q != nil && q.Channel == "" {
		q.Channel = info.about.Channel()
	}
	if len(info.about.LicenseFiles) == 0 && info.about.LicenseUrl != "" {
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, info.about.LicenseUrl, nil)
		if err != nil {
			return fmt.Errorf("retrieving license: %w", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("retrieving license: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			f, err := os.Create(path.Join(dir, "LICENSE"))
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(f, resp.Body); err != nil {
				return fmt.Errorf("writing license file: %w", err)
			}
			if err := f.Close(); err != nil {
				return err
			}
			info.about.useLicenseUrl = true
		}
	}
	return nil
}

func (info *LicenseInfo) purl(version string) string {
	var buf strings.Builder
	if _, err := buf.WriteString("pkg:"); err != nil {
		panic(err)
	}
	if _, err := buf.WriteString(info.Source); err != nil {
		panic(err)
	}
	if err := buf.WriteByte('/'); err != nil {
		panic(err)
	}
	if _, err := buf.WriteString(strings.ReplaceAll(info.Name, "@", "%40")); err != nil {
		panic(err)
	}
	if version != "" {
		if err := buf.WriteByte('@'); err != nil {
			panic(err)
		}
		if _, err := buf.WriteString(version); err != nil {
			panic(err)
		}
	}
	if info.Qualifiers != nil {
		info.Qualifiers.Append(&buf)
	}
	return buf.String()
}

func (info *LicenseInfo) metadataRule(name,
	pkgName, version, url *build.StringExpr) build.Expr {
	c := &build.CallExpr{
		X: &build.Ident{Name: "package_info"},
		List: []build.Expr{
			buildutil.Attr("name", name),
			buildutil.Attr("package_name", pkgName),
		},
	}
	if version != nil {
		c.List = append(c.List, buildutil.Attr("package_version", version))
	}
	if url != nil {
		c.List = append(c.List, buildutil.Attr("package_url", url))
	}
	if info.Source != "" && info.Name != "" {
		var v string
		if version != nil {
			v = version.Value
		}
		c.List = append(c.List, buildutil.StrAttr(
			"purl", info.purl(v)))
	}
	return c
}

func (info *LicenseInfo) Rules(pkgName,
	version, url string) []build.Expr {
	if len(info.Licenses) == 0 {
		return nil
	}
	var versionExp, urlExp *build.StringExpr
	if version != "" {
		versionExp = buildutil.StrExpr(version)
	}
	if url != "" {
		urlExp = buildutil.StrExpr(url)
	}
	// Allocate space for
	// 1. the load statement for the license rule
	// 2. the load statement for the package_info rule.
	// 3. the package declaration
	// N. the license blocks,
	// 4. package metadata
	// 5. conda_files
	// 6. conda_manifest
	// 7. conda_deps
	// 8. a comment block.
	result := make([]build.Expr, 3, 8+len(info.Licenses))
	licenseNames := make([]build.Expr, len(info.Licenses), len(info.Licenses)+1)
	pkgNameExp := buildutil.StrExpr(pkgName)
	// TODO: remove the empty file check, which is currently required due to
	// https://github.com/bazelbuild/rules_license/issues/31
	if len(info.Licenses) == 1 && info.Licenses[0].File != "" {
		licenseNames[0] = buildutil.StrExpr("license")
		result = append(result,
			info.Licenses[0].LicenseRule(pkgNameExp, licenseNames[0],
				versionExp, urlExp))
	} else {
		for i, lic := range info.Licenses {
			// TODO: remove the empty file check, which is currently required due to
			// https://github.com/bazelbuild/rules_license/issues/31
			if lic.File != "" {
				licenseNames[i] = buildutil.StrExpr(lic.NameFromFile(info.Licenses))
				result = append(result,
					lic.LicenseRule(pkgNameExp, licenseNames[i],
						versionExp, urlExp))
			}
		}
	}
	metadataName := buildutil.StrExpr("package_info")
	licenseNames = append(licenseNames, metadataName)
	result = append(result, info.metadataRule(metadataName,
		pkgNameExp, versionExp, urlExp))
	result[0] = buildutil.LoadExpr("@rules_license//rules:license.bzl", "license")
	result[1] = buildutil.LoadExpr("@rules_license//rules:package_info.bzl", "package_info")
	result[2] = &build.CallExpr{
		X: &build.Ident{Name: "package"},
		List: []build.Expr{
			buildutil.Attr("default_applicable_licenses",
				&build.ListExpr{
					List:           licenseNames,
					ForceMultiLine: len(licenseNames) > 1,
				}),
		},
	}
	return result
}

func (lic *License) LicenseRule(pkgName, name build.Expr,
	version, url *build.StringExpr) build.Expr {
	c := &build.CallExpr{
		X: &build.Ident{Name: "license"},
		List: []build.Expr{
			buildutil.Attr("name", name),
			buildutil.Attr("package_name", pkgName),
			buildutil.Attr("license_kinds", &build.ListExpr{
				List:           buildutil.StrExprList(lic.Kinds...),
				ForceMultiLine: len(lic.Kinds) > 0,
			}),
		},
	}
	if version != nil {
		c.List = append(c.List, buildutil.Attr("package_version", version))
	}
	if url != nil {
		c.List = append(c.List, buildutil.Attr("package_url", url))
	}
	if lic.File == "" {
		// TODO: tolerate this.
		// https://github.com/bazelbuild/rules_license/issues/31
		log.Fatal("No license file found")
		// c.List = append(c.List,
		// 	buildutil.Attr("license_text", &build.Ident{Name: "None"}))
	} else {
		c.List = append(c.List, buildutil.StrAttr("license_text", lic.File))
	}
	if lic.Copyright != "" {
		c.List = append(c.List, buildutil.StrAttr("copyright_notice", lic.Copyright))
	}
	return c
}

func (lic *License) NameFromFile(allFiles []License) string {
	fn := path.Base(lic.File)
	if len(allFiles) > 1 {
		// If we have a bunch of license files, each in a unique directory,
		// use the full path rather than the file name.
		names := make(map[string]struct{}, len(allFiles))
		for _, lic := range allFiles {
			names[strings.ReplaceAll(lic.File, "/", "_")] = struct{}{}
		}
		if len(names) == len(allFiles) {
			fn = strings.ReplaceAll(lic.File, "/", "_")
		}
	} else {
		fn = strings.TrimSuffix(fn, ".md")
		fn = strings.TrimSuffix(fn, ".txt")
		fn = strings.TrimSuffix(fn, ".markdown")
	}
	return fn
}

func (about *aboutJson) getId(fallback string) string {
	if about.License == "" || strings.EqualFold(about.License, "unknown") {
		if fallback == "" || strings.EqualFold(fallback, "unknown") {
			return about.Family
		}
		return fallback
	}
	return about.License
}

var (
	andOrRegexp = regexp.MustCompile(
		`(?i:.\s+and\s+.)|(?:\s*[^0-9]\+\s*.)|(?:.,\s*[^vV\s])|(?i:.\s+or\s+.)`)

	orStart  = regexp.MustCompile(`^(?i:or)\s`)
	andStart = regexp.MustCompile(`^(?i:and)\s`)

	// Known-good licenses.
	goodLicenseRegexp = [...]*regexp.Regexp{
		// Best: licenses without a notice requirement.
		regexp.MustCompile(
			`^(?i:0BSD|Unlicense|Public Domain|WTFPL|Zlib)\b`),
		// Second best: commonly known permissive licenses.
		regexp.MustCompile(
			`^(?i:BSD|BSL|AFL|Apache|MIT|Python|PSF|MPL)\b`),
	}
	// Licenses we definitely prefer to avoid.
	badLicenseRegexp = regexp.MustCompile(`^(?:A?GPL|Commercial)`)
)

func licensePriority(id string) int {
	if badLicenseRegexp.FindString(id) != "" {
		return len(goodLicenseRegexp) + 2
	}
	for i, re := range goodLicenseRegexp {
		if re.FindString(id) != "" {
			return i
		}
	}
	return len(goodLicenseRegexp)
}

func licenseListPriority(ids []License) int {
	if len(ids) == 0 {
		return len(goodLicenseRegexp) + 1
	}
	var result int
	for _, lic := range ids {
		if p := licensePriority(lic.CanonicalId); p > result {
			result = p
		}
	}
	return result
}

func pickLicense(first, second []License) []License {
	if len(second) == 0 {
		return first
	} else if len(first) == 0 {
		return second
	}
	p1 := licenseListPriority(first)
	if p1 == 0 && len(second) >= len(first) {
		return first
	}
	p2 := licenseListPriority(second)
	if p1 < p2 {
		return first
	} else if p1 == p2 {
		// Equal priority, return the shorter list.
		if len(second) < len(first) {
			return second
		}
		return first
	}
	return second
}

func splitAndOr(id string) (string, string, bool) {
	if idx := andOrRegexp.FindStringIndex(id); len(idx) == 2 {
		return id[:idx[0]+1], id[idx[1]-1:],
			orStart.MatchString(strings.TrimSpace(id[idx[0]+1:]))
	}
	return id, "", false
}

func findClose(id string) int {
	open := 0
	for i, c := range id {
		if c == '(' {
			open++
		} else if c == ')' {
			open--
		}
		if open == 0 {
			return i
		}
	}
	return -1
}

func licenseOr(first1, first2, second1, second2 []License) ([]License, []License) {
	// [prefix AND] first1 OR first2 OR second1 OR second2 =
	// [prefix AND] first1 OR (first2 OR second1 OR second2)
	return first1, pickLicense(first2, pickLicense(second1, second2))
}

func licenseAnd(first1, first2, second1, second2 []License) ([]License, []License) {
	if len(first2) == 0 {
		// [prefix AND] first1 AND second1 OR second2 =
		// ([prefix AND] first1 AND second1) OR second2
		return mergeAnd(first1, second1), second2
	}
	// [prefix AND] first1 OR first2 AND second1 OR second2 =
	// [prefix AND] first1 OR (first2 AND second1 OR second2)
	return first1, pickLicense(mergeAnd(first2, second1), second2)
}

// Combine two lists, skipping duplicates.
//
// This is basically like `append` except handles weird cases like
// MIT AND (MIT OR Apache-2.0).
func mergeAnd(first, second []License) []License {
	if len(second) == 0 {
		return first
	}
	if len(first) == 0 {
		return second
	}
	if cap(first)-len(first) < len(second) {
		if cap(second)-len(second) >= len(first) {
			// Trade places to avoid needing to reallocate a slice, since we don't
			// actually care about order.
			first, second = second, first
		} else {
			n := make([]License, len(first), cap(first)+cap(second))
			copy(n, first)
			first = n
		}
	}
	// Generally these are short lists.
	// Except in extreem cases, the cost of allocating a map is going to be
	// much more than the cost of using O(N) search.
	in := func(lst []License, item License) bool {
		for _, e := range lst {
			if e.CanonicalId == item.CanonicalId {
				return true
			}
		}
		return false
	}
	for _, item := range second {
		if !in(first, item) {
			first = append(first, item)
		}
	}
	return first
}

// Splits a license clause into component licenses.
//
// This implements collapsing of boolean logical operators as we go, choosing
// one license set over another (for OR clauses) eagerly where possible, meaning
//
//	 A  OR   B AND C  OR   D AND E AND F  OR  G
//	(A) OR  (B AND C  OR   D AND E AND F  OR  G)
//	(A) OR ((B AND C) OR  (D AND E AND F  OR  G))
//	(A) OR ((B AND C) OR ((D AND E AND F) OR (G)))
//	                       [prefer one over other]
//	(A) OR ((B AND C) OR  (D AND E AND F))
//	         [prefer one over other]
//	(A) OR  (B AND C)
//	 [prefer one over other]
//	A
func splitLicense(id string) ([]License, []License, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return []License{{}}, nil, nil
	}
	// If the leading expression is parenthesied, process it separately.
	if id[0] == '(' {
		i := findClose(id)
		if i < 0 {
			return nil, nil, fmt.Errorf("%w in %q",
				unbalancedParenthesis, id)
		}
		head := id[1:i]
		first, firstOr, err := splitLicense(head)
		if err != nil {
			return first, firstOr, fmt.Errorf("parenthetical clause (%q): %w",
				head, err)
		}
		first = pickLicense(first, firstOr)
		tail := strings.TrimSpace(id[i+1:])
		if tail == "" {
			return first, nil, nil
		}
		if orStart.MatchString(tail) {
			// PREFIX AND X OR Y OR Z = PREFIX AND X OR (Y OR Z)
			second, secondOr, err := splitLicense(tail[3:])
			if err != nil {
				return first, nil, fmt.Errorf("or clause %q ((%q) OR %q): %w",
					id, head, tail[3:], err)
			}
			return first, pickLicense(second, secondOr), nil
		} else if andStart.MatchString(tail) {
			// PREFIX AND X AND Y OR Z = (PREFIX AND X AND Y) OR Z
			second, secondOr, err := splitLicense(tail[4:])
			if err != nil {
				return first, nil, fmt.Errorf("and clause %q ((%q) AND %q): %w",
					id, head, tail[4:], err)
			}
			return mergeAnd(first, second), secondOr, nil
		}
	}
	// Find the first AND or OR.
	if fid, sid, isOr := splitAndOr(id); sid != "" {
		first, firstOr, err := splitLicense(fid)
		if err != nil {
			if isOr {
				return first, firstOr, fmt.Errorf("or clause %q (%q OR %q): %w",
					id, fid, sid, err)
			} else {
				return first, firstOr, fmt.Errorf("and clause %q (%q AND %q): %w",
					id, fid, sid, err)
			}
		}
		second, secondOr, err := splitLicense(sid)
		if err != nil {
			if isOr {
				return first, firstOr, fmt.Errorf("or clause %q (%q OR %q): %w",
					id, fid, sid, err)
			} else {
				return first, firstOr, fmt.Errorf("and clause %q (%q AND %q): %w",
					id, fid, sid, err)
			}
		}
		if isOr {
			first, firstOr = licenseOr(first, firstOr, second, secondOr)
			return first, firstOr, nil
		} else {
			first, firstOr = licenseAnd(first, firstOr, second, secondOr)
			return first, firstOr, nil
		}
	}
	if strings.EqualFold("others", id) {
		return nil, nil, nil
	}
	lic, err := canonicalizeLicense(id)
	if err != nil {
		if first, second, _ := strings.Cut(id, "/"); second != "" {
			first, firstOr, err2 := splitLicense(first)
			if err2 != nil {
				return first, firstOr, err
			}
			second, secondOr, err2 := splitLicense(second)
			if err2 != nil {
				return first, firstOr, err
			}
			first, firstOr = licenseOr(first, firstOr, second, secondOr)
			return first, firstOr, nil
		}
		return nil, nil, fmt.Errorf("parsing %q: %w", id, err)
	}
	return []License{lic}, nil, nil
}

func getLicence(ctx context.Context, id string) (License, error) {
	defer trace.StartRegion(ctx, "getLicense").End()
	licenses, licensesOr, err := splitLicense(id)
	licenses = pickLicense(licenses, licensesOr)
	if err != nil {
		return License{CanonicalId: id}, err
	}
	license := licenses[0]
	if len(licenses) > 1 {
		license.CanonicalId = id
		for _, lic := range licenses[1:] {
			license.Kinds = append(license.Kinds, lic.Kinds...)
		}
	}
	return license, nil
}

var (
	missingLicense        = errors.New("license missing")
	unrecognizedLicense   = errors.New("unrecognized license ID")
	unbalancedParenthesis = errors.New("unbalanced parenthesis")
)

func canonicalizeLicense(id string) (License, error) {
	if id == "" || strings.EqualFold("unknown", id) {
		return License{}, missingLicense
	}
	id = normalizeLicenseId(id)
	if cid := getKnown(id); cid != nil {
		return *cid, nil
	}
	// Use %w here rather than just the literal string so that errors.Is works.
	return License{}, fmt.Errorf("%w %s", unrecognizedLicense, id)
}

func licenseFiles(ctx context.Context, files json.RawMessage) ([]string, error) {
	defer trace.StartRegion(ctx, "licenseFiles").End()
	if len(files) < 3 {
		return []string{""}, nil
	}
	if files[0] == '"' {
		var s string
		err := json.Unmarshal(files, &s)
		return []string{path.Join("info", "licenses", s)}, err
	}
	var options []string
	err := json.Unmarshal(files, &options)
	for i, s := range options {
		options[i] = path.Join("info", "licenses", s)
	}
	return options, err
}

var copyrightSearch = regexp.MustCompile(`^\s*((?i:copyright:?\s+[(c©0-9].*?))\s*$`)

func findCopyright(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		matches := copyrightSearch.FindSubmatch(scanner.Bytes())
		if len(matches) == 2 {
			return string(matches[1]), nil
		}
	}

	if err := scanner.Err(); errors.Is(err, io.EOF) {
		return "", nil
	} else {
		return "", err
	}
}

// guessCopyright makes a guess as to the copyright for the package.
//
// This is unfortunately necessary because conda does not include that
// information explicitly in any of the metadata.
func (info *LicenseInfo) guessCopyright(ctx context.Context, dir string) error {
	defer trace.StartRegion(ctx, "guessCopyright").End()
	for i, lic := range info.Licenses {
		if lic.File == "" {
			continue
		}
		result, err := findCopyright(path.Join(dir, lic.File))
		if err != nil {
			return err
		}
		if result != "" {
			info.Licenses[i].Copyright = result
			return nil
		}
	}
	return nil
}

func (info *LicenseInfo) CanonicalizeConda(dir, fallback string,
	licenseOverrides []string, licenceFile string) error {
	ctx, task := trace.NewTask(context.Background(), "CanonicalizeLicense")
	defer task.End()
	if len(licenseOverrides) > 0 || licenceFile == "" {
		// Only attempt to load about.json if we have to.
		if err := info.LoadConda(ctx, dir); err != nil {
			return err
		}
	}
	var license License
	if len(licenseOverrides) > 0 {
		license = License{
			CanonicalId: "license",
			Kinds:       licenseOverrides,
		}
	} else {
		var err error
		license, err = getLicence(ctx, info.about.getId(fallback))
		if err != nil {
			return err
		}
	}
	var files []string
	if licenceFile != "" {
		files = []string{licenceFile}
	} else if info.about.useLicenseUrl {
		files = []string{"LICENSE"}
	} else {
		var err error
		files, err = licenseFiles(ctx, info.about.LicenseFiles)
		if err != nil {
			return err
		}
	}
	info.Licenses = make([]License, 0, len(files))
	for _, file := range files {
		if file != "" {
			license.File = path.Clean(file)
		}
		info.Licenses = append(info.Licenses, license)
	}
	info.Source = "conda"
	if info.about.Home != "" {
		info.Homepage = info.about.Home
	} else if info.about.DevUrl != "" {
		info.Homepage = info.about.DevUrl
	} else if info.about.DocUrl != "" {
		info.Homepage = info.about.DocUrl
	} else if info.about.DocSrc != "" {
		info.Homepage = info.about.DocSrc
	}
	return info.guessCopyright(ctx, dir)
}
