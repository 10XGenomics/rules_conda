//go:generate python3 generate_spdx.py -o spdx.go
//go:generate gofmt -w spdx.go

package licensing

import "strings"

// Get the canonical capitalization and kind for the given ID, ignoring case.
func getKnown(id string) *License {
	if strings.EqualFold(id, "Public Domain") || strings.EqualFold(id, "10X Genomics") {
		return &License{
			CanonicalId: "Public Domain",
		}
	}
	// Check for exact match
	if kind := knownLicenses[id]; kind != "" {
		return &License{
			CanonicalId: id,
			Kinds:       []string{kind},
		}
	}
	// Check for case-insensitive match
	for cid, kind := range knownLicenses {
		if strings.EqualFold(id, cid) {
			return &License{
				CanonicalId: id,
				Kinds:       []string{kind},
			}
		}
	}
	return nil
}
