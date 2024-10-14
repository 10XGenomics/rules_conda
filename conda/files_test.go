package conda

import (
	"strings"
	"testing"
)

func TestTranslateBytes(t *testing.T) {
	p := condaFilePath{
		Path:        `lib/python2.7/_sysconfigdata.py`,
		Mode:        "text",
		Placeholder: `/opt/anaconda1anaconda2anaconda3`,
	}
	source := []byte(`# -*- coding: utf-8 -*-
	# system configuration generated and used by the sysconfig module
	build_time_vars = {'AC_APPLE_UNIVERSAL_BUILD': 0,
	  'AR': 'ar',
	  'AST_H': 'Include/Python-ast.h',
	  'BINDIR': '/opt/anaconda1anaconda2anaconda3/bin',
	  'BLDSHARED': 'gcc -pthread -B /opt/anaconda1anaconda2anaconda3/compiler_compat '
				   '-shared -L/opt/anaconda1anaconda2anaconda3/lib '
				   '-Wl,-rpath=/opt/anaconda1anaconda2anaconda3/lib -Wl,--no-as-needed '
				   '-Wl,--sysroot=/',
	  'BUILDEXE': '',
	  'srcdir': '.'}`)
	const expect = `# -*- coding: utf-8 -*-
	# system configuration generated and used by the sysconfig module
	build_time_vars = {'AC_APPLE_UNIVERSAL_BUILD': 0,
	  'AR': 'ar',
	  'AST_H': 'Include/Python-ast.h',
	  'BINDIR': 'external/conda_env/bin',
	  'BLDSHARED': 'gcc -pthread -B external/conda_env/compiler_compat '
				   '-shared -Lexternal/conda_env/lib '
				   '-Wl,-rpath=external/conda_env/lib -Wl,--no-as-needed '
				   '',
	  'BUILDEXE': '',
	  'srcdir': '.'}`
	compare := func(result, expect string) {
		display := strings.NewReplacer(" ", "·", "\t", "\\t  ")
		if result == expect {
			return
		}
		t.Error("incorrect output")
		line := 1
		i := strings.IndexRune(result, '\n')
		j := strings.IndexRune(expect, '\n')
		for i >= 0 && j >= 0 {
			if result[:i] != expect[:j] {
				t.Error("line", line, ":",
					display.Replace(result[:i]),
					"!=", display.Replace(expect[:j]))
			}
			result = result[i+1:]
			expect = expect[j+1:]
			i = strings.IndexRune(result, '\n')
			j = strings.IndexRune(expect, '\n')
		}
		if i == 0 {
			t.Error("result has blank line at end")
		} else if j == 0 {
			t.Error("result missing blank line at end")
		} else if i > 0 {
			t.Error("result has extra lines:\n", display.Replace(result))
		} else if j > 0 {
			t.Error("result missing lines:\n", display.Replace(expect))
		}
	}
	result := string(p.translateBytes(source))
	if result != expect {
		compare(result, expect)
	}
	p.Path = "lib/python3.7/_sysconfigdata_m_linux_x86_64-linux-gnu.py"
	result = string(p.translateBytes(source))
	if result != expect {
		compare(result, expect)
	}
}
