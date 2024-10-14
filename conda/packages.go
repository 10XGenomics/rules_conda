// Package conda contains methods for working with conda packages.
package conda

import (
	"os"
	"path"
	"sort"
	"strings"

	"github.com/10XGenomics/rules_conda/licensing"
)

type Package struct {
	Dir   string
	Index indexJson
	Paths condaPathFile

	// All files produced by this package.
	allFiles map[string]struct{}

	includeDirs  []string
	pyStubs      []string
	License      licensing.LicenseInfo
	linkPython   bool
	isExecutable bool
}

// Returns the name of the package.
func (pkg *Package) Name() string {
	return pkg.Index.Name
}

var distNameClean = strings.NewReplacer(".", "_", "-", "_")

func (pkg *Package) RepoName() string {
	return "conda_package_" + distNameClean.Replace(pkg.Name())
}

func (pkg *Package) executable() string {
	bestExe := ""
	searchName := pkg.Index.Entry
	if searchName == "" {
		searchName = pkg.Name()
	}
	for _, stub := range pkg.pyStubs {
		n, _, _ := strings.Cut(stub, "=")
		n = strings.TrimSpace(n)
		if bestExe == "" || n == searchName {
			bestExe = "bin/" + n
		}
		if bestExe != "" {
			return bestExe
		}
	}
	for _, f := range pkg.Paths.Paths {
		if strings.HasPrefix(f.Path, "bin/") {
			if bestExe == "" || path.Base(f.Path) == searchName {
				if info, err := os.Stat(path.Join(pkg.Dir, f.Path)); err == nil &&
					info.Mode()&0100 != 0 {
					bestExe = f.Path
				}
			}
		}
	}
	return bestExe
}

func (pkg *Package) Install(roots []string, dest string, files []string) error {
	translate := make(map[string]*condaFilePath, len(pkg.Paths.Paths))
	for i := range pkg.Paths.Paths {
		p := &pkg.Paths.Paths[i]
		if p.NeedsTranslate() {
			translate[p.Path] = p
		}
	}
	roots = cleanRoots(pkg.Dir, roots)
	for _, f := range files {
		sp := stripRoot(f, roots)
		if p := translate[sp]; p != nil {
			// Need to modify the file as it is being installed.
			if err := p.Install(f, dest); err != nil {
				return err
			}
		} else {
			// Simple case.
			dd, b := path.Split(sp)
			if dest != "" {
				dd = path.Join(dest, dd)
			}
			_ = os.MkdirAll(dd, 0777)
			if err := copyFile(f, path.Join(dd, b)); err != nil {
				return err
			}
		}
	}
	return nil
}

func cleanRoots(dir string, roots []string) []string {
	result := make([]string, 0, len(roots)+1)
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	// Ensure all prefixes end in a slash, and include the repository directory.
	for i := range roots {
		result = append(result, path.Join(roots[i], dir))
	}
	result = append(result, dir)
	// Sort longer prefixes before shorter ones.
	sort.SliceStable(result, func(i, j int) bool {
		return len(result[i]) > len(result[j])
	})
	return result
}

func stripRoot(fn string, roots []string) string {
	for _, root := range roots {
		if after, found := strings.CutPrefix(fn, root); found {
			return after
		}
	}
	return fn
}
