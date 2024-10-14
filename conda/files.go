package conda

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/10XGenomics/rules_conda/buildutil"
)

type condaPathFile struct {
	Paths    []condaFilePath `json:"paths"`
	Manifest []string        `json:"-"`
}

// Load path information from files.json.
func (c *condaPathFile) loadFromFilesJson(dir string) error {
	if b, err := os.ReadFile(path.Join(dir, "info", "files.json")); err != nil {
		return err
	} else {
		c.Manifest = append(c.Manifest, path.Join("info", "files.json"))
		var f condaFilesFile
		if err := json.Unmarshal(b, &f); err != nil {
			return err
		} else {
			f.CopyTo(c)
			return nil
		}
	}
}

// Load path information from the "files" file.
func (c *condaPathFile) loadFromFilesList(dir string) error {
	if b, err := os.ReadFile(path.Join(dir, "info", "files")); err != nil {
		return err
	} else {
		c.Manifest = append(c.Manifest, path.Join("info", "files"))
		for _, f := range strings.Split(string(b), "\n") {
			f = strings.TrimSpace(f)
			if f != "" {
				c.Paths = append(c.Paths, condaFilePath{
					Path: f,
					Mode: "binary",
					Type: "hardlink",
				})
			}
		}
		if err := c.readPrefixes(dir); err != nil {
			return err
		}
		if err := c.readNolink(dir); err != nil {
			return err
		}
		return nil
	}
}

// Load metadata from no_link file.  This is not required when loading
// from files.json or paths.json.
func (c *condaPathFile) readNolink(dir string) error {
	if b, err := os.ReadFile(path.Join(dir, "info", "no_link")); err != nil &&
		!os.IsNotExist(err) {
		return err
	} else if err == nil {
		if len(b) > 0 {
			c.Manifest = append(c.Manifest, path.Join("info", "no_link"))
		}
		nolinkFiles := strings.Split(string(b), "\n")
		set := make(map[string]struct{}, len(nolinkFiles))
		for _, f := range nolinkFiles {
			set[strings.TrimSpace(f)] = struct{}{}
		}
		for i := range c.Paths {
			if _, ok := set[c.Paths[i].Path]; ok {
				c.Paths[i].NoLink = true
			}
		}
	}
	return nil
}

// Load metadata from has_prefix file.  This is not required when loading
// from files.json or paths.json.
func (c *condaPathFile) readPrefixes(dir string) error {
	if b, err := os.ReadFile(path.Join(dir, "info", "has_prefix")); err != nil &&
		!os.IsNotExist(err) {
		return err
	} else if err == nil {
		if len(b) > 0 {
			c.Manifest = append(c.Manifest, path.Join("info", "has_prefix"))
		}
		prefixLines := bytes.Split(b, []byte("\n"))
		set := make(map[string]string, len(prefixLines))
		for _, line := range prefixLines {
			fields := bytes.Fields(line)
			if len(fields) == 3 && bytes.Equal(fields[1], []byte("text")) {
				set[string(bytes.TrimSpace(fields[2]))] = string(fields[0])
			}
		}
		for i := range c.Paths {
			if ph, ok := set[c.Paths[i].Path]; ok {
				c.Paths[i].Mode = "text"
				c.Paths[i].Placeholder = ph
			}
		}
	}
	return nil
}

func (c *condaPathFile) Load(dir string) error {
	if b, err := os.ReadFile(path.Join(dir, "info", "paths.json")); os.IsNotExist(err) {
		if err := c.loadFromFilesJson(dir); os.IsNotExist(err) {
			if err := c.loadFromFilesList(dir); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if err := json.Unmarshal(b, c); err != nil {
		return err
	} else {
		c.Manifest = append(c.Manifest, path.Join("info", "paths.json"))
	}
	return nil
}

type condaFilePath struct {
	Path string `json:"_path"`

	// "binary" or "text"
	Mode string `json:"file_mode"`

	// "hardlink" or "softlink"
	Type string `json:"path_type"`

	// For text mode files, the string to replace.
	Placeholder string `json:"prefix_placeholder,omitempty"`

	NoLink bool `json:"no_link,omitempty"`
}

func (p *condaFilePath) NeedsTranslate() bool {
	return p.Mode == "text" && p.Placeholder != ""
}

type oldCondaFilePath struct {
	// For compatibility with older files.json format
	OldPath string `json:"path,omitempty"`
	OldType string `json:"file_type"`

	condaFilePath
}

type condaFilesFile struct {
	Files []oldCondaFilePath `json:"files"`
}

func (f *condaFilesFile) CopyTo(c *condaPathFile) {
	for _, file := range f.Files {
		p := file.condaFilePath
		if file.OldPath != "" {
			p.Path = file.OldPath
		}
		if file.OldType != "" {
			p.Type = file.OldType
		}
		c.Paths = append(c.Paths, p)
	}
}

// See https://github.com/bazelbuild/bazel/issues/374
const bazelBannedFilenameCharacters = " :"

type symlinkEntry struct {
	location string
	relPath  string
}

// filesList returns the list of files from the manifest which are present and
// not symlinks, as well as the list of non-broken symlinks with their targets.
func (paths *condaPathFile) filesList(dir string) (
	outs []*condaFilePath, symlinks []symlinkEntry, executables []string) {
	outs = make([]*condaFilePath, 0, len(paths.Paths))
	for i, p := range paths.Paths {
		// Omit missing files so that bazel doesn't complain when they
		// don't show up.  This is to deal with for example the broken
		// symlinks which matplotlib installs for libtcl.so and libtk.so
		fullPath := path.Join(dir, p.Path)
		if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
			if strings.ContainsAny(p.Path, bazelBannedFilenameCharacters) {
				// See https://github.com/bazelbuild/bazel/issues/374
				fmt.Fprintf(os.Stderr,
					"WARNING: omitting file %q from build because the "+
						"filename contains characters which are not permitted.\n",
					p.Path)
			} else if info, err := os.Lstat(fullPath); err != nil {
				panic("error accessing file " + fullPath)
			} else if info.Mode()&os.ModeSymlink == 0 {
				outs = append(outs, &paths.Paths[i])
				if !info.IsDir() && info.Mode()&0o111 != 0 {
					executables = append(executables, p.Path)
				}
			} else if content, err := os.Readlink(fullPath); err != nil {
				panic("error reading link " + fullPath)
			} else {
				symlinks = append(symlinks, symlinkEntry{
					location: p.Path,
					relPath:  content,
				})
			}
		}
	}
	return
}

func (p *condaFilePath) Install(src, dest string) error {
	target := p.Path
	if dest != "" {
		target = path.Join(dest, target)
	}
	if d := path.Dir(target); d != "" {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	if p.Type == "softlink" {
		// These are handled in starlark.
		return nil
	} else if p.NeedsTranslate() {
		return p.Translate(src, target)
	} else {
		return p.Copy(src, target)
	}
}

func (p *condaFilePath) verify(src string) ([]byte, os.FileMode, error) {
	if f, err := os.Open(src); err != nil {
		return nil, 0, err
	} else {
		defer f.Close()
		if info, err := f.Stat(); err != nil {
			return nil, 0, err
		} else if b, err := io.ReadAll(f); err != nil {
			return b, info.Mode(), err
		} else {
			return b, info.Mode(), f.Close()
		}
	}
}

func (p *condaFilePath) Copy(src, dest string) error {
	if !p.NoLink {
		if p, err := filepath.EvalSymlinks(src); err == nil {
			if os.Link(p, dest) == nil {
				return nil
			}
		}
	}
	// link failed, fall back on copy
	return copyFile(src, dest)
}

// Copy a file from one path to another, keeping the mode flags.
//
// Uses copy_file_range system call if possible.
func copyFile(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if fDest, err := os.OpenFile(dest,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		info.Mode()); err != nil {
		return err
	} else {
		defer fDest.Close()
		if _, err := fDest.ReadFrom(f); err != nil {
			return err
		}
		return fDest.Close()
	}
}

func write(dest string, b []byte, mode os.FileMode) error {
	if f, err := os.OpenFile(dest,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		mode); err != nil {
		return err
	} else {
		defer f.Close()
		if _, err := f.Write(b); err != nil {
			return err
		}
		return f.Close()
	}
}

func (p *condaFilePath) Translate(src, target string) error {
	if b, mode, err := p.verify(src); err != nil {
		return err
	} else {
		// Replace shebang lines
		b = p.translateBytes(b)
		return write(target, b, mode)
	}
}

var (
	shebang       = regexp.MustCompile(`^#!/.*/bin/python`)
	newShebang    = []byte(`#!/usr/bin/env python`)
	newArgvPython = []byte(`"/usr/bin/env", "python"`)
	wlSysroot     = []byte(`-Wl,--sysroot=/`)
	placeHeldFor  = []byte(`external/` + buildutil.DefaultCondaRepo)

	// sysconfigdata in conda has a placehold for `TZPATH`, but setting
	// a relative path for `TZPATH` causes annoying warnings.
	tzPath    = regexp.MustCompile(`(\n\s*['"]TZPATH['"]: ['"])[^'"]*(['"])`)
	newTzPath = []byte(`$1$2`)
)

func SetCondaRepo(condaRepo string) {
	placeHeldFor = append(placeHeldFor[:len(`external/`)], condaRepo...)
}

func (p *condaFilePath) translateBytes(b []byte) []byte {
	b = shebang.ReplaceAllLiteral(b, newShebang)

	if strings.HasSuffix(p.Path, ".json") {
		// This is for kernel.json, which has
		// 'argv = ["/placeholder/bin/python"...' in it.
		ph := make([]byte, 1, 2+len(p.Placeholder)+len("/bin/python"))
		ph[0] = '"'
		ph = append(ph, p.Placeholder...)
		ph = append(ph, `/bin/python"`...)
		b = bytes.Replace(b, ph, newArgvPython, -1)
	}
	// Replace other paths
	b = bytes.Replace(b,
		[]byte(p.Placeholder),
		placeHeldFor, -1)
	if strings.HasSuffix(p.Path, ".py") && strings.HasPrefix(
		filepath.Base(p.Path), "_sysconfigdata") {
		b = bytes.Replace(b, wlSysroot, nil, -1)
		b = tzPath.ReplaceAll(b, newTzPath)
	}
	return b
}

func isLinkPython(dir string) (bool, []string) {
	b, err := os.ReadFile(path.Join(dir, "info", "link.json"))
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		return false, nil
	}
	var f struct {
		NoArch struct {
			Type        string   `json:"type"`
			EntryPoints []string `json:"entry_points"`
		} `json:"noarch"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		panic(err)
	}
	return f.NoArch.Type == "python", f.NoArch.EntryPoints
}
