// Tool to copy and, in some cases, translate, files from the
// conda package tarballs into the assembled conda distribution
// directory.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"os"
	"path"
	"strings"

	"github.com/10XGenomics/rules_conda/buildutil"
	"github.com/10XGenomics/rules_conda/conda"
)

func main() {
	var roots, install, dest, prefix, condaRepo string
	flag.StringVar(&roots, "roots", "",
		"The root directories under which conda packages can be found.")
	flag.StringVar(&install, "install", "",
		"The package to install, if any.  Incompatible with require.")
	flag.StringVar(&dest, "dest", "",
		"The destination path relative to the working directory.")
	flag.StringVar(&condaRepo, "conda", buildutil.DefaultCondaRepo,
		"The name of the conda repository.")
	flag.StringVar(&prefix, "prefix", "",
		"An additional prefix to use for a noarch install")
	flag.Parse()

	if install == "" {
		flag.Usage()
		os.Exit(1)
	} else {
		conda.SetCondaRepo(condaRepo)
		if prefix != "" {
			dest = path.Join(dest, prefix)
		}
		installPackage(strings.Split(roots, ","), install, dest,
			fileList(flag.Args()))
	}
}

func fileList(args []string) []string {
	result := make([]string, 0, len(args))
	for _, f := range args {
		if len(f) > 1 && f[0] == '@' {
			result = append(result, listFromFile(f[1:])...)
		} else {
			result = append(result, f)
		}
	}
	return result
}

func listFromFile(fn string) []string {
	f, err := os.Open(fn)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	var result []string
	for scan.Scan() {
		ln := bytes.TrimSpace(scan.Bytes())
		if len(ln) > 0 {
			result = append(result, string(ln))
		}
	}
	return result
}

func installPackage(roots []string, install, dest string, files []string) {
	var pkg conda.Package
	if err := pkg.Load(install, nil, nil, false); err != nil {
		panic(err)
	}
	if err := pkg.Install(roots, dest, files); err != nil {
		panic(err)
	}
}
