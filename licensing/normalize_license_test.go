package licensing

import "testing"

func Test_normalizeLicenseId(t *testing.T) {
	// These test cases are a (large) subset of the license IDs in the conda
	// dependencies for cellranger and turing.
	tests := []struct {
		id   string
		want string
	}{
		{
			id:   "3-clause BSD",
			want: "BSD-3-Clause",
		},
		{
			id:   "Affero GPL",
			want: "AGPL-3.0",
		},
		{
			id:   "Apache 2.0",
			want: "Apache-2.0",
		},
		{
			id:   "Apache License 2.0",
			want: "Apache-2.0",
		},
		{
			id:   "Apache-2.0",
			want: "Apache-2.0",
		},
		{
			id:   "Apache",
			want: "Apache-2.0",
		},
		{
			id:   "Biopython License Agreement",
			want: "Biopython",
		},
		{
			id:   "Boost-1.0",
			want: "BSL-1.0",
		},
		{
			id:   "BSD 2-Clause",
			want: "BSD-2-Clause",
		},
		{
			id:   "BSD 3 Clause",
			want: "BSD-3-Clause",
		},
		{
			id:   "BSD 3-clause",
			want: "BSD-3-Clause",
		},
		{
			id:   "BSD 3-Clause",
			want: "BSD-3-Clause",
		},
		{
			id:   "BSD License",
			want: "BSD-3-Clause",
		},
		{
			id:   "BSD Like",
			want: "BSD-3-Clause",
		},
		{
			id:   "BSD-2-clause",
			want: "BSD-2-Clause",
		},
		{
			id:   "BSD-2-Clause",
			want: "BSD-2-Clause",
		},
		{
			id:   "BSD-3-clause",
			want: "BSD-3-Clause",
		},
		{
			id:   "BSD-3-Clause",
			want: "BSD-3-Clause",
		},
		{
			id:   "BSD-3",
			want: "BSD-3-Clause",
		},
		{
			id:   "BSD-like",
			want: "BSD-3-Clause",
		},
		{
			id:   "BSD",
			want: "BSD-3-Clause",
		},
		{
			id:   "BSL-1.0",
			want: "BSL-1.0",
		},
		{
			id:   "bzip2",
			want: "bzip2-1.0.6",
		},
		{
			id:   "C News-like",
			want: "BSD-3-Clause",
		},
		{
			id:   "Commercial",
			want: "Commercial",
		},
		{
			id:   "curl",
			want: "curl",
		},
		{
			id:   "EPL v1.0",
			want: "EPL-1.0",
		},
		{
			id:   "fitsio",
			want: "Public Domain",
		},
		{
			id:   "Free software (MIT-like)",
			want: "Free software (MIT-like)",
		},
		{
			id:   "GD",
			want: "GD",
		},
		{
			id:   "GNU GENERAL PUBLIC LICENSE",
			want: "GPL-2.0-or-later",
		},
		{
			id:   "GNU GPL 3+ with GCC Runtime Library",
			want: "GPL-3.0-with-GCC-exception",
		},
		{
			id:   "LGPL",
			want: "LGPL-2.1",
		},
		{
			id:   "GPL 3.0",
			want: "GPL-3.0",
		},
		{
			id:   "GPL 3",
			want: "GPL-3.0",
		},
		{
			id:   "GPL v2",
			want: "GPL-2.0",
		},
		{
			id:   "GPL v2+",
			want: "GPL-2.0-or-later",
		},
		{
			id:   "GPL-2.0-only",
			want: "GPL-2.0-only",
		},
		{
			id:   "GPL-2.0-or-later",
			want: "GPL-2.0-or-later",
		},
		{
			id:   "GPL-2.0",
			want: "GPL-2.0",
		},
		{
			id:   "GPL-3.0-only WITH GCC-exception-3.1",
			want: "GPL-3.0-with-GCC-exception",
		},
		{
			id:   "GPL-3.0-only",
			want: "GPL-3.0-only",
		},
		{
			id:   "GPL-3.0-or-later",
			want: "GPL-3.0-or-later",
		},
		{
			id:   "GPL-3.0",
			want: "GPL-3.0",
		},
		{
			id:   "GPL",
			want: "GPL-2.0-or-later",
		},
		{
			id:   "GPL2",
			want: "GPL-2.0",
		},
		{
			id:   "GPLv2",
			want: "GPL-2.0",
		},
		{
			id:   "HDF5",
			want: "HDF5",
		},
		{
			id:   "HPND",
			want: "HPND",
		},
		{
			id:   "ISC (ISCL)",
			want: "ISC",
		},
		{
			id:   "ISC",
			want: "ISC",
		},
		{
			id:   "LGPL 2.1",
			want: "LGPL-2.1",
		},
		{
			id:   "LGPL-2.0-or-later",
			want: "LGPL-2.0-or-later",
		},
		{
			id:   "LGPL-2.1-only",
			want: "LGPL-2.1-only",
		},
		{
			id:   "LGPL-2.1-or-later",
			want: "LGPL-2.1-or-later",
		},
		{
			id:   "LGPL-2.1",
			want: "LGPL-2.1",
		},
		{
			id:   "LGPL-3.0-or-later",
			want: "LGPL-3.0-or-later",
		},
		{
			id:   "LGPL-3.0",
			want: "LGPL-3.0",
		},
		{
			id:   "LGPL",
			want: "LGPL-2.1",
		},
		{
			id:   "LGPL2",
			want: "LGPL-2.0",
		},
		{
			id:   "LGPLv2",
			want: "LGPL-2.0",
		},
		{
			id:   "AFL-2.1",
			want: "AFL-2.1",
		},
		{
			id:   "MIT License",
			want: "MIT",
		},
		{
			id:   "MIT",
			want: "MIT",
		},
		{
			// nolint: misspell
			id:   "MIT/X derivate (http://curl.haxx.se/docs/copyright.html)",
			want: "curl",
		},
		{
			id:   "MPL 1.1",
			want: "MPL-1.1",
		},
		{
			id:   "MPL 2.0",
			want: "MPL-2.0",
		},
		{
			id:   "MPL-1.1",
			want: "MPL-1.1",
		},
		{
			id:   "MPL-2.0",
			want: "MPL-2.0",
		},
		{
			id:   "MPL2",
			want: "MPL-2.0",
		},
		{
			id:   "OFL",
			want: "OFL-1.1",
		},
		{
			id:   "OpenSSL",
			want: "OpenSSL",
		},
		{
			id:   "Perl Artistic",
			want: "Artistic-1.0-Perl",
		},
		{
			id:   "PSF 2.0",
			want: "PSF-2.0",
		},
		{
			id:   "PSF 2.1.1",
			want: "PSF-2.1.1",
		},
		{
			id:   "PSF-2.0",
			want: "PSF-2.0",
		},
		{
			id:   "PSF",
			want: "PSF-2.0",
		},
		{
			id:   "PSF2",
			want: "PSF-2.0",
		},
		{
			id:   "Public Domain Dedictation",
			want: "Public Domain",
		},
		{
			id:   "Public Domain",
			want: "Public Domain",
		},
		{
			id:   "Public-Domain (http://www.sqlite.org/copyright.html)",
			want: "Public Domain",
		},
		{
			id:   "Public-Domain",
			want: "Public Domain",
		},
		{
			id:   "Python-2.0",
			want: "Python-2.0",
		},
		{
			id:   "SIL Open Font License, Version 1.1",
			want: "OFL-1.1",
		},
		{
			id:   "Tcl/Tk",
			want: "TCL",
		},
		{
			// nolint: misspell
			id:   "Ubuntu Font Licence Version 1.0",
			want: "UFL-1.0",
		},
		{
			id:   "Unlicense",
			want: "Unlicense",
		},
		{
			id:   "zlib",
			want: "zlib",
		},
		{
			id:   "zlib/libpng",
			want: "zlib-acknowledgement",
		},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := normalizeLicenseId(tt.id); got != tt.want {
				t.Errorf("normalizeLicenseId() = %q, want %q", got, tt.want)
			}
		})
	}
}
