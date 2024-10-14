package conda

import "testing"

func TestTypeFromName(t *testing.T) {
	chk := func(file string, ft fileType) {
		t.Helper()
		if f := typeFromName(file, nil); f != ft {
			t.Errorf("Expected %v, got %v for %q", ft, f, file)
		}
	}
	chk("bin/h5import", otherFile)
	chk("include/python3.8/Python.h", header)
	chk("compiler_compat/README", otherFile)
	chk("include/h5_dble_interface.mod", otherFile)
	chk("info/repodata_record.json", metadataFile)
	chk("include/H5DataType.h", header)
	chk("include/hdf5_hl.h", header)
	chk("include/hdf5_hl.hpp", header)
	chk("lib/libhdf5.a", aLib)
	chk("lib/libhdf5.la", libTool)
	chk("lib/libhdf5.lo", aLib)
	chk("lib/libhdf5.settings", otherFile)
	chk("lib/libhdf5.so", soLib)
	chk("lib/libhdf5.so.10", soLib)
	chk("lib/libhdf5_cpp.a", aLib)
	chk("lib/libhdf5_cpp.so.13.0.0", soLib)
}

func TestCHeaderContentRe(t *testing.T) {
	chk := func(line string, expect bool) {
		t.Helper()
		if r := cHeaderContentRe.MatchString(line); r != expect {
			t.Errorf("Expected %v, got %v for %q", expect, r, line)
		}
	}
	chk(`#ifndef THING`, true)
	chk(`  #define THING`, true)
	chk(`# include <thing>`, true)
	chk(`//comment`, true)
	chk(`/*comment*/`, true)
	chk(`int main(void) {`, true)
	chk(`const foo = bar;`, true)
	chk(`#!/usr/bin/env bash`, false)
}

func TestTypeFromName_File(t *testing.T) {
	chk := func(file string, ft fileType) {
		t.Helper()
		if f := typeFromName(file, []string{"testdata"}); f != ft {
			t.Errorf("Expected %v, got %v for %q", ft, f, file)
		}
	}
	chk("testdata/cpp_header", header)     // Should be detected based on content.
	chk("testdata/H5DataType.h", header)   // Should be detected based on name.
	chk("testdata/record.json", otherFile) // Should be skipped based on name.
	chk("testdata/script_file", otherFile) // Should be skipped based on content.
}
