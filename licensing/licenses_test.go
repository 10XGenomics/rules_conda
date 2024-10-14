package licensing

import (
	"context"
	"errors"
	"reflect"
	"runtime/trace"
	"testing"
)

func Test_getLicence(t *testing.T) {
	// These test cases are a (large) subset of the license IDs in the conda
	// dependencies for cellranger and turing.
	tests := []struct {
		err   error
		name  string
		kinds []string
	}{
		{
			name:  "10X Genomics",
			kinds: nil,
		},
		{
			name:  "MIT OR 10X Genomics",
			kinds: nil,
		},
		{
			name:  "10X Genomics OR MIT",
			kinds: nil,
		},
		{
			name:  "3-clause BSD",
			kinds: []string{"@rules_license//licenses/spdx:BSD-3-Clause"},
		},
		{
			name: "Adobe+GPLv2",
			kinds: []string{
				"@com_github_10XGenomics_rules_conda//licensing/known:Adobe",
				"@rules_license//licenses/spdx:GPL-2.0",
			},
		},
		{
			name:  "Affero GPL",
			kinds: []string{"@rules_license//licenses/spdx:AGPL-3.0"},
		},
		{
			name:  "Apache 2.0",
			kinds: []string{"@rules_license//licenses/spdx:Apache-2.0"},
		},
		{
			name:  "Apache 2.0 or BSD 2-Clause",
			kinds: []string{"@rules_license//licenses/spdx:Apache-2.0"},
		},
		{
			name:  "Apache License 2.0",
			kinds: []string{"@rules_license//licenses/spdx:Apache-2.0"},
		},
		{
			name:  "Apache Software",
			kinds: []string{"@rules_license//licenses/spdx:Apache-2.0"},
		},
		{
			name:  "Apache-2.0",
			kinds: []string{"@rules_license//licenses/spdx:Apache-2.0"},
		},
		{
			name:  "Apache-2.0 and others",
			kinds: []string{"@rules_license//licenses/spdx:Apache-2.0"},
		},
		{
			name:  "Apache-2.0 OR GPL-2.0",
			kinds: []string{"@rules_license//licenses/spdx:Apache-2.0"},
		},
		{
			name:  "Apache-2.0 OR MIT AND GPL-2.0",
			kinds: []string{"@rules_license//licenses/spdx:Apache-2.0"},
		},
		{
			name:  "0BSD OR MIT AND GPL-2.0",
			kinds: []string{"@rules_license//licenses/spdx:0BSD"},
		},
		{
			name: "Apache-2.0 AND BSD-3-Clause AND PSF-2.0 AND MIT",
			kinds: []string{
				"@rules_license//licenses/spdx:Apache-2.0",
				"@rules_license//licenses/spdx:BSD-3-Clause",
				"@rules_license//licenses/spdx:PSF-2.0",
				"@rules_license//licenses/spdx:MIT",
			},
		},
		{
			name: "Apache-2.0 AND PSF-2.0 AND MIT",
			kinds: []string{
				"@rules_license//licenses/spdx:Apache-2.0",
				"@rules_license//licenses/spdx:PSF-2.0",
				"@rules_license//licenses/spdx:MIT",
			},
		},
		{
			name: "Apache-2.0 WITH LLVM-exception",
			kinds: []string{
				"@rules_license//licenses/spdx:Apache-2.0",
			},
		},
		{
			name: "Apache-2.0 or BSD-2-Clause",
			kinds: []string{
				"@rules_license//licenses/spdx:Apache-2.0",
			},
		},
		{
			name:  "BSD",
			kinds: []string{"@rules_license//licenses/spdx:BSD-3-Clause"},
		},
		{
			name:  "BSD 2-Clause",
			kinds: []string{"@rules_license//licenses/spdx:BSD-2-Clause"},
		},
		{
			name:  "BSD Like",
			kinds: []string{"@rules_license//licenses/spdx:BSD-3-Clause"},
		},
		{
			name: "BSD-2-Clause AND PSF-2.0",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-2-Clause",
				"@rules_license//licenses/spdx:PSF-2.0",
			},
		},
		{
			name: "BSD-2-Clause, PSF2",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-2-Clause",
				"@rules_license//licenses/spdx:PSF-2.0",
			},
		},
		{
			name:  "BSD-2-clause",
			kinds: []string{"@rules_license//licenses/spdx:BSD-2-Clause"},
		},
		{
			name:  "BSD-3",
			kinds: []string{"@rules_license//licenses/spdx:BSD-3-Clause"},
		},
		{
			name:  "BSD-3-Clause",
			kinds: []string{"@rules_license//licenses/spdx:BSD-3-Clause"},
		},
		{
			name: "BSD-3-Clause AND PSF-2.0",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
				"@rules_license//licenses/spdx:PSF-2.0",
			},
		},
		{
			name: "(BSD-3-Clause AND PSF-2.0)",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
				"@rules_license//licenses/spdx:PSF-2.0",
			},
		},
		{
			name: "(BSD-3-Clause) AND PSF-2.0",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
				"@rules_license//licenses/spdx:PSF-2.0",
			},
		},
		{
			name: "(BSD-3-Clause)AND PSF-2.0",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
				"@rules_license//licenses/spdx:PSF-2.0",
			},
		},
		{
			name: "BSD-3-Clause and Apache",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
				"@rules_license//licenses/spdx:Apache-2.0",
			},
		},
		{
			name:  "BSD-like",
			kinds: []string{"@rules_license//licenses/spdx:BSD-3-Clause"},
		},
		{
			name:  "Biopython License Agreement",
			kinds: []string{"@com_github_10XGenomics_rules_conda//licensing/known:Biopython"},
		},
		{
			name: "Biopython License Agreement and GPL-3.0",
			kinds: []string{
				"@com_github_10XGenomics_rules_conda//licensing/known:Biopython",
				"@rules_license//licenses/spdx:GPL-3.0",
			},
		},
		{
			name: "Commercial, GPL-2.0, GPL-3.0",
			kinds: []string{
				"@com_github_10XGenomics_rules_conda//licensing/known:Commercial",
				"@rules_license//licenses/spdx:GPL-2.0",
				"@rules_license//licenses/spdx:GPL-3.0",
			},
		},
		{
			name: "Free software (MIT-like)",
			err:  unrecognizedLicense,
		},
		{
			name:  "GD",
			kinds: []string{"@rules_license//licenses/spdx:GD"},
		},
		{
			name:  "GNU GENERAL PUBLIC LICENSE",
			kinds: []string{"@rules_license//licenses/spdx:GPL-2.0-or-later"},
		},
		{
			name:  "GNU GPL 3+ with GCC Runtime Library",
			kinds: []string{"@rules_license//licenses/spdx:GPL-3.0-with-GCC-exception"},
		},
		{
			name: "GPL-2.0-only and LicenseRef-FreeType",
			kinds: []string{
				"@rules_license//licenses/spdx:GPL-2.0-only",
				"@rules_license//licenses/spdx:FTL",
			},
		},
		{
			name:  "HDF5",
			kinds: []string{"@com_github_10XGenomics_rules_conda//licensing/known:HDF5"},
		},
		{
			name:  "LGPL 2.1 or MPL 1.1",
			kinds: []string{"@rules_license//licenses/spdx:MPL-1.1"},
		},
		{
			name:  "LGPL or MIT",
			kinds: []string{"@rules_license//licenses/spdx:MIT"},
		},
		{
			name:  "LicenseRef-Public-Domain AND BSD-3-clause",
			kinds: []string{"@rules_license//licenses/spdx:BSD-3-Clause"},
		},
		{
			name:  "LicenseRef-cuDNN-Software-License-Agreement",
			kinds: []string{"@com_github_10XGenomics_rules_conda//licensing/known:cuDNN"},
		},
		{
			// nolint: misspell
			name:  "MIT/X derivate (http://curl.haxx.se/docs/copyright.html)",
			kinds: []string{"@rules_license//licenses/spdx:curl"},
		},
		{
			name: "Public Domain",
		},
		{
			name:  "Public-Domain AND BSD-3-clause",
			kinds: []string{"@rules_license//licenses/spdx:BSD-3-Clause"},
		},
		{
			name:  "SIL Open Font License, Version 1.1",
			kinds: []string{"@rules_license//licenses/spdx:OFL-1.1"},
		},
		{
			name:  "Tcl/Tk",
			kinds: []string{"@rules_license//licenses/spdx:TCL"},
		},
		{
			// nolint: misspell
			name:  "Ubuntu Font Licence Version 1.0",
			kinds: []string{"@com_github_10XGenomics_rules_conda//licensing/known:UFL-1.0"},
		},
		{
			name: "Unknown",
			err:  missingLicense,
		},
		{
			name:  "zlib",
			kinds: []string{"@rules_license//licenses/spdx:Zlib"},
		},
		{
			name:  "zlib/libpng",
			kinds: []string{"@rules_license//licenses/spdx:zlib-acknowledgement"},
		},
		{
			name: "BSD-3-Clause/PSF-2.0 AND 0BSD OR MIT",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
			},
		},
		{
			name: "BSD-3-Clause OR PSF-2.0 AND MIT",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
			},
		},
		{
			name: "BSD-3-Clause AND PSF-2.0 OR MIT",
			kinds: []string{
				"@rules_license//licenses/spdx:MIT",
			},
		},
		{
			name: "BSD-3-Clause AND (PSF-2.0 OR MIT)",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
				"@rules_license//licenses/spdx:PSF-2.0",
			},
		},
		{
			name: "(BSD-3-Clause OR Apache-2.0) AND PSF-2.0 AND MIT",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
				"@rules_license//licenses/spdx:PSF-2.0",
				"@rules_license//licenses/spdx:MIT",
			},
		},
		{
			name: "(BSD-3-Clause OR Apache-2.0) AND PSF-2.0 AND MIT",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
				"@rules_license//licenses/spdx:PSF-2.0",
				"@rules_license//licenses/spdx:MIT",
			},
		},
		{
			name: "BSD-3-Clause AND Apache-2.0 AND (PSF-2.0 AND MIT)",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
				"@rules_license//licenses/spdx:Apache-2.0",
				"@rules_license//licenses/spdx:PSF-2.0",
				"@rules_license//licenses/spdx:MIT",
			},
		},
		{
			name: "(BSD-3-Clause OR Apache-2.0) AND (PSF-2.0 OR MIT)",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
				"@rules_license//licenses/spdx:PSF-2.0",
			},
		},
		{
			name: "BSD-3-Clause AND (Apache-2.0) OR Apache-2.0 AND (PSF-2.0 OR (MIT))",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
				"@rules_license//licenses/spdx:Apache-2.0",
			},
		},
		{
			name: "BSD-3-Clause AND Apache-2.0 OR BSD-3-Clause OR (PSF-2.0 OR (MIT))",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
			},
		},
		{
			name: "(BSD-3-Clause OR (Apache-2.0 AND (MIT OR BSD-3-Clause))) " +
				"AND (PSF-2.0 OR MIT) AND MIT",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
				"@rules_license//licenses/spdx:PSF-2.0",
				"@rules_license//licenses/spdx:MIT",
			},
		},
		{
			name: "BSD-3-Clause AND (Apache-2.0 OR (MIT OR BSD-3-Clause)) " +
				"AND (PSF-2.0 OR MIT) OR MIT",
			kinds: []string{
				"@rules_license//licenses/spdx:MIT",
			},
		},
		{
			name: "BSD-3-Clause/MIT",
			kinds: []string{
				"@rules_license//licenses/spdx:BSD-3-Clause",
			},
		},
		{
			name: "BSD-3-Clause AND (Apache-2.0",
			err:  unbalancedParenthesis,
		},
		{
			name: "BSD-3-Clause AND (Apache-2.0 OR foo)",
			err:  unrecognizedLicense,
		},
		{
			name: "BSD-3-Clause AND (Apache-2.0 AND foo)",
			err:  unrecognizedLicense,
		},
		{
			name: "(BSD-3-Clause AND Apache-2.0) AND foo",
			err:  unrecognizedLicense,
		},
		{
			name: "(BSD-3-Clause AND Apache-2.0) OR foo",
			err:  unrecognizedLicense,
		},
		{
			name: "foo OR BSD-3-Clause",
			err:  unrecognizedLicense,
		},
		{
			name: "foo/BSD-3-Clause",
			err:  unrecognizedLicense,
		},
		{
			name: "BSD-3-Clause/foo",
			err:  unrecognizedLicense,
		},
		{
			name: "foo AND BSD-3-Clause",
			err:  unrecognizedLicense,
		},
	}
	ctx, task := trace.NewTask(context.Background(), "Test_getLicense")
	defer task.End()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getLicence(ctx, tt.name)
			if !errors.Is(err, tt.err) {
				t.Errorf("getLicence() error = %v, wantErr %v", err, tt.err)
				return
			} else if err != nil {
				return
			}
			if got.CanonicalId == "" {
				t.Error("License was not returned.")
				return
			}
			if !reflect.DeepEqual(got.Kinds, tt.kinds) {
				t.Errorf("getLicence() = %v, want %v", got, tt.kinds)
			}
		})
	}
}

func Test_splitAndOr(t *testing.T) {
	tests := []struct {
		name   string
		first  string
		second string
		or     bool
	}{
		{
			name:   "Apache 2.0 or BSD 2-Clause",
			first:  "Apache 2.0",
			second: "BSD 2-Clause",
			or:     true,
		},
		{
			name:   "LGPL 2.1 or MPL 1.1",
			first:  "LGPL 2.1",
			second: "MPL 1.1",
			or:     true,
		},
		{
			name:   "Apache 2.0 AND BSD 2-Clause",
			first:  "Apache 2.0",
			second: "BSD 2-Clause",
		},
		{
			name:   "GNU GPL 3+ with GCC Runtime Library",
			first:  "GNU GPL 3+ with GCC Runtime Library",
			second: "",
		},
		{
			name:   "SIL Open Font License, Version 1.1",
			first:  "SIL Open Font License, Version 1.1",
			second: "",
		},
		{
			name:   "Adobe+GPLv2",
			first:  "Adobe",
			second: "GPLv2",
		},
		{
			name:   "BSD-2-Clause, PSF2",
			first:  "BSD-2-Clause",
			second: "PSF2",
		},
		{
			name:   "Apache-2.0 AND BSD-3-Clause AND PSF-2.0 AND MIT",
			first:  "Apache-2.0",
			second: "BSD-3-Clause AND PSF-2.0 AND MIT",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first, second, isOr := splitAndOr(tt.name)
			if first != tt.first {
				t.Errorf("splitAndOr() got = %s, want %s", first, tt.first)
			}
			if isOr != tt.or {
				t.Errorf("or should be %v, want %v", isOr, tt.or)
			}
			if second != tt.second {
				t.Errorf("splitAndOr() got = %s, want %s", second, tt.second)
				t.Logf("matched: %q", andOrRegexp.FindString(tt.name))
			}
		})
	}
}

func TestLicenseInfo_purl(t *testing.T) {
	tests := []struct {
		Source     string
		Name       string
		Version    string
		Qualifiers PurlQualifiers
		want       string
	}{
		// All the things.
		{
			Source:  "conda",
			Name:    "astroid",
			Version: "2.9.0",
			Qualifiers: &CondaPackageQualifiers{
				Build:   "py39h06a4308_0",
				Channel: "main",
				Subdir:  "linux-64",
				Type:    "conda",
			},
			want: "pkg:conda/astroid@2.9.0?build=py39h06a4308_0&" +
				"channel=main&subdir=linux-64&type=conda",
		},
		// No build.
		{
			Source:  "conda",
			Name:    "astroid",
			Version: "2.9.0",
			Qualifiers: &CondaPackageQualifiers{
				Channel: "main",
				Subdir:  "linux-64",
				Type:    "conda",
			},
			want: "pkg:conda/astroid@2.9.0?channel=main&subdir=linux-64&type=conda",
		},
		// No qualifiers.
		{
			Source:  "conda",
			Name:    "astroid",
			Version: "2.9.0",
			want:    "pkg:conda/astroid@2.9.0",
		},
		// Empty qualifiers.
		{
			Source:     "conda",
			Name:       "astroid",
			Version:    "2.9.0",
			Qualifiers: new(CondaPackageQualifiers),
			want:       "pkg:conda/astroid@2.9.0",
		},
		// No version.
		{
			Source: "conda",
			Name:   "astroid",
			Qualifiers: &CondaPackageQualifiers{
				Build:   "py39h06a4308_0",
				Channel: "main",
				Subdir:  "linux-64",
				Type:    "conda",
			},
			want: "pkg:conda/astroid?build=py39h06a4308_0&channel=main&subdir=linux-64&type=conda",
		},
		// @ in name
		{
			Source:  "npm",
			Name:    "@actions/core",
			Version: "1.1.9",
			want:    "pkg:npm/%40actions/core@1.1.9",
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			info := &LicenseInfo{
				Version:    tt.Version,
				Source:     tt.Source,
				Qualifiers: tt.Qualifiers,
				Name:       tt.Name,
			}
			if got := info.purl(tt.Version); got != tt.want {
				t.Errorf("LicenseInfo.purl() = %v, want %v", got, tt.want)
			}
		})
	}
}
