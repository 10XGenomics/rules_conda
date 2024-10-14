// Tool to generate the BUILD file for a conda_package_repository.
package main

import (
	"flag"
	"log"
	"strings"

	"github.com/10XGenomics/rules_conda/buildutil"
	"github.com/10XGenomics/rules_conda/conda"
	"github.com/10XGenomics/rules_conda/licensing"
)

func maybeSplit(s, sep string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return strings.Split(s, sep)
}

func main() {
	var dir, distName, extraDeps, excludeDeps,
		licenses, licenseFile, ccInclude, url, condaRepo string
	flag.StringVar(&dir, "dir", "",
		"The directory in which to create the BUILD file.")
	flag.StringVar(&licenses, "licenses", "",
		"Use the given `license_kind` instead of the one specified by about.json.")
	flag.StringVar(&licenseFile, "license_file", "",
		"Use the given license file instead of the one specified by about.json.")
	flag.StringVar(&extraDeps, "extra_deps", "",
		"Packages on which this package depends, but aren't in the metadata, "+
			"comma-separated list.")
	flag.StringVar(&excludeDeps, "exclude_deps", "",
		"Packages on which this package claims to depend but should be skipped.")
	flag.StringVar(&distName, "distname", "",
		"The distname of the repo, for commenting.")
	flag.StringVar(&ccInclude, "cc_include_path", "",
		"Paths to add to the c/c++ include path.")
	var channel, pkgType string
	flag.StringVar(&channel, "channel", "",
		"The conda channel from which this package was downloaded.")
	flag.StringVar(&pkgType, "type", "tar.bz2",
		"The type of package (e.g. tar.bz2 or conda)")
	flag.StringVar(&url, "url", "",
		"The URL from which the package was downloaded")
	flag.StringVar(&condaRepo, "conda", buildutil.DefaultCondaRepo,
		"The repo used to refer to dependencies.")
	flag.Parse()
	var pkg conda.Package
	if err := pkg.Load(dir, nil, flag.Args(), true); err != nil {
		log.Fatal("Could not load package metadata:", err)
	}
	if q, ok := pkg.License.Qualifiers.(*licensing.CondaPackageQualifiers); ok &&
		q != nil && channel != "" {
		q.Channel = channel
		q.Type = pkgType
	}
	if len(pkg.Paths.Paths) > 0 {
		if err := pkg.License.CanonicalizeConda(dir, pkg.Index.License,
			strings.Fields(licenses), licenseFile); err != nil {
			log.Fatal("Could not parse license information:", err)
		}
	}
	if condaRepo != "" && condaRepo[0] != '@' {
		condaRepo = "@" + condaRepo
	}
	if err := pkg.MakeTarballBuild(maybeSplit(extraDeps, ","),
		maybeSplit(excludeDeps, ","), condaRepo,
		maybeSplit(ccInclude, ":"),
		distName, url); err != nil {
		log.Fatal("Could not generate BUILD file:", err)
	}
}
