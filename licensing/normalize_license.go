package licensing

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// nolint: misspell
	licenseSpellingFixer      = regexp.MustCompile(`(?i:\bLicence\b)`)
	agplLicenseFixer          = regexp.MustCompile(`(?i:Affero GPL)`)
	gnuLicenseFixer           = regexp.MustCompile(`(?i:GNU (?:General Public License|([AL]?)GPL))`)
	licenseVersionFixerRegexp = regexp.MustCompile(
		`^(?i:([AL]?GPL|BSD|AFL|Apache|MIT|MPL|PSF|EPL))` +
			`(?i:\s+license\s+)?` +
			`(?:[- ]?(?:v|(?i:version\s*))?)?` +
			`([0-9.]+)`)
	bsdLicenseFixer = regexp.MustCompile(
		`(?i:BSD-(\d)(?:[- ]clause)?)|(?i:(\d)[- ]clause[- ]BSD)`)
	publicDomainFixer = regexp.MustCompile(
		`(?i:Public[ -]Domain(?:[- ]Dedict?ation)?(?:\s*\([^)]*\))?)`)
	licenseAgreementFixer = regexp.MustCompile(`(?i:(?:\s+|-)Software)?` +
		`(?i:(?:\s+|-)License(?:(?:\s+|-)Agreement)?)$`)
	versionPlusFixer  = regexp.MustCompile(`(?i:GPL-)(\d+)\+`)
	gccExceptionFixer = regexp.MustCompile(`(?i:GPL)-([0-9.]+)(?i:-only|-or-later)?` +
		`(?i:[- ]WITH[- ]GCC[- ])` +
		`(?i:exception|Runtime Library)(?:-[0-9.]+)?`)
	guessVersionRegexp = regexp.MustCompile(
		`^(?i:([AL]?GPL|BSL|Boost|bzip2|AFL|Apache|MPL|PSF|EPL|OFL)|(BSD(?:[- ]like)?))` +
			`(?:-(\d+))?$`)
)

// normalizeLicenseId returns a version of id with common "misspellings" fixed.
func normalizeLicenseId(id string) string {
	id = strings.TrimPrefix(id, "LicenseRef-")
	id = licenseSpellingFixer.ReplaceAllLiteralString(id, "License")
	id = gnuLicenseFixer.ReplaceAllString(id, "${1}GPL")
	id = agplLicenseFixer.ReplaceAllLiteralString(id, "AGPL")
	id = licenseVersionFixerRegexp.ReplaceAllString(id, "$1-$2")
	id = bsdLicenseFixer.ReplaceAllString(id, "BSD-$1$2-Clause")
	id = publicDomainFixer.ReplaceAllLiteralString(id, "Public Domain")
	id = versionPlusFixer.ReplaceAllString(id, "GPL-${1}.0-or-later")
	id = gccExceptionFixer.ReplaceAllString(id, "GPL-$1-with-GCC-exception")
	id = licenseAgreementFixer.ReplaceAllLiteralString(id, "")
	return redirectLicense(assumesVersion(id))
}

// assumesVersion assumes a version when none is specified.
func assumesVersion(id string) string {
	if strings.EqualFold(id, "BSD") {
		return "BSD-3-Clause"
	}
	match := guessVersionRegexp.FindStringSubmatch(id)
	if len(match) == 4 {
		if match[3] != "" {
			if match[2] != "" {
				return fmt.Sprintf("BSD-%s-Clause", match[3])
			}
			return fmt.Sprintf("%s-%s.0", match[1], match[3])
		}
		switch match[1] {
		case "AFL", "AGPL":
			return fmt.Sprintf("%s-3.0", match[1])
		case "LGPL":
			return "LGPL-2.1"
		case "BSL", "Boost":
			return "BSL-1.0"
		case "GPL":
			return "GPL-2.0-or-later"
		case "Apache", "EPL", "MPL", "PSF":
			return fmt.Sprintf("%s-2.0", match[1])
		case "OFL":
			return "OFL-1.1"
		case "bzip2":
			return "bzip2-1.0.6"
		}
		if match[2] != "" {
			return "BSD-3-Clause"
		}
	}
	return id
}

// redirectLicense replaces certain licenses with equivalent ones.
func redirectLicense(id string) string {
	// nolint: misspell
	switch id {
	case "Apache-2.0 WITH LLVM-exception", "Apache Software":
		return "Apache-2.0"
	case "C News-like":
		return "BSD-3-Clause"
	case "MIT/X derivate (http://curl.haxx.se/docs/copyright.html)":
		return "curl"
	case "SIL Open Font License, Version 1.1":
		return "OFL-1.1"
	case "fitsio":
		return "Public Domain"
	case "ISC (ISCL)":
		return "ISC"
	case "Tcl/Tk":
		return "TCL"
	case "zlib/libpng":
		return "zlib-acknowledgement"
	case "Boost-1.0":
		return "BSL-1.0"
	case "Perl Artistic":
		return "Artistic-1.0-Perl"
	case "Ubuntu Font License Version 1.0":
		return "UFL-1.0"
	case "FreeType":
		return "FTL"
	default:
		return id
	}
}
