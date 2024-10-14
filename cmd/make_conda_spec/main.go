// Generate complete conda spec from a requirements file.
//
// The requirements file will be given to `conda create -F` so must
// be formatted appropriately for that parser.
//
// The requirements will be output as a .bzl file with a macro that can
// be used to initialize the repository.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

func getEnv(condaDir string) []string {
	oldEnvs := os.Environ()
	envs := make([]string, 0, len(oldEnvs)+2)
	for _, e := range oldEnvs {
		name, value, _ := strings.Cut(e, "=")
		switch name {
		case "PYTHONHASHSEED", "PYTHONNOUSERSITE", "PYTHONHOME":
		case "PATH":
			if condaDir != "" {
				envs = append(envs, string(append(append(append(append(append(
					make([]byte, 0, len(e)+len(condaDir)+1),
					name...), '='),
					condaDir...),
					os.PathListSeparator),
					value...)))
			} else {
				envs = append(envs, e)
			}
		default:
			envs = append(envs, e)
		}
	}
	envs = append(envs, "PYTHONHASHSEED=0", "PYTHONNOUSERSITE=0")
	return envs
}

func splitList(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func main() {
	var buildFile, requirements, conda, outName, channels, exclude, extra, arch string
	flag.StringVar(&buildFile, "build", "",
		"The path to a sibling file of the target location. "+
			"This is used if the output file doesn't already exist.")
	flag.StringVar(&requirements, "requirements", "",
		"Specifies the requirements to use in generating the spec.")
	flag.StringVar(&conda, "conda", "",
		"The path to the conda executable to use for fetching.")
	flag.StringVar(&outName, "o", "",
		"The output file target.")
	flag.StringVar(&channels, "chan", "defaults,bioconda,conda-forge",
		"Specify an ordered, comma-separated list of channels to search.")
	flag.StringVar(&extra, "extra", "",
		"Specify the comma-separated list of workspace names for "+
			"additional conda package repositories to add to the "+
			"conda_packages list of the generated "+
			"conda_environment rule.")
	flag.StringVar(&exclude, "exclude", "",
		"A comma-separated list of packages to exclude from the generated "+
			"lock file, if they are present in the solution returned by conda.")
	flag.StringVar(&arch, "arch", "linux-64",
		"The architecture to pass to the solver.")
	flag.Parse()

	if outName == "" {
		log.Fatalln("Missing outName")
	}
	if resolved, err := filepath.EvalSymlinks(outName); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if buildFile == "" {
				log.Fatalln(
					"The target file doesn't exist, and no sibling file was specified.")
			}
			if resolved, err := filepath.EvalSymlinks(buildFile); err != nil {
				log.Fatalln(
					"Error resolving sibling file for target location:",
					err)
			} else {
				outName = filepath.Join(filepath.Dir(resolved), outName)
			}
		} else {
			log.Fatalln(
				"Could not resolve location for target file.")
		}
	} else {
		outName = resolved
	}
	if conda == "" {
		log.Fatalln("Path to conda is required.")
	}
	if resolved, err := filepath.Abs(conda); err == nil {
		conda = resolved
	}
	tempdir, err := os.MkdirTemp("", "micromamba")
	if err != nil {
		log.Fatalln("Can't create temp dir for mamba root.")
	}
	defer os.RemoveAll(tempdir)
	channelList := splitList(channels)
	extrasList := splitList(extra)
	excludeList := splitList(exclude)
	cmd := condaCreate(requirements, conda, channelList, arch, tempdir)
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalln("Failed making output pipe for conda:\n", err)
	}
	cmd.Stderr = os.Stderr
	cmd.Env = getEnv(filepath.Dir(conda))
	fmt.Fprintln(os.Stderr, "Solving dependencies...")
	if err := cmd.Start(); err != nil {
		log.Fatalln("Failed starting conda:\n", err)
	}
	specs, err := readSpecs(out, excludeList)
	if err != nil {
		log.Fatalln("Failed reading conda output:\n", err)
	}
	if err := cmd.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, "Conda failed, rerunning to show output...")
		showCondaError(requirements, conda, channelList, arch, tempdir)
		log.Fatalln("Original conda failure:\n", err)
	}
	fmt.Fprintln(os.Stderr, "Getting package URLs and hashes...")
	if err := fillSpecs(specs, conda, requirements, channelList, arch, tempdir); err != nil {
		log.Fatalln("Failed getting hashes:\n", err)
	}
	if err := writeSpecs(specs, extrasList, outName); err != nil {
		log.Fatalln("Failed writing spec:\n", err)
	}
}

func condaArgs(requirements string, channels []string, arch, tempdir string) []string {
	args := []string{
		"create",
		"-y", // Non-interactive, assume "yes" for all questions
		"-n", // Name of the environment we're not actually creating
		"spec",
		// Path to mamba root. Used to cache repo data.
		// We don't actually _want_ to use a persistent cache here because
		// corrupted state can result in weird solve issues.
		"-r",
		tempdir,
		"--no-rc", // Ignore weird stuff might be in the user's home dir.
		"--file",  // Specify the requirements file
		requirements,
		"--platform", // Specify the architecture of packages to fetch
		arch,
		"--override-channels", // Don't use "default" channels.
	}
	for _, c := range channels {
		args = append(args, "--channel", c)
	}
	return args
}

func condaCreate(requirements, conda string, channels []string, arch, tempdir string) *exec.Cmd {
	condaPath := path.Dir(conda) + string([]rune{os.PathListSeparator}) + os.Getenv("PATH")
	args := append(condaArgs(requirements, channels, arch, tempdir),
		"--dry-run", // Don't actually create the environment
		"--json",    // Output as json so we can parse it
	)
	return makeCmd(conda, condaPath, args...)
}

func showCondaError(requirements, conda string, channels []string, arch, tempdir string) {
	condaPath := path.Dir(conda) + string([]rune{os.PathListSeparator}) + os.Getenv("PATH")
	args := append(condaArgs(requirements, channels, arch, tempdir),
		"--dry-run", // Don't actually create the environment
	)
	cmd := makeCmd(conda, condaPath, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Println("Conda failed: ", err)
	}
}

func condaDownload(requirements, conda string, channels []string, arch, tempdir string) *exec.Cmd {
	condaPath := path.Dir(conda) + string([]rune{os.PathListSeparator}) + os.Getenv("PATH")
	args := append(condaArgs(requirements, channels, arch, tempdir),
		"--download-only",
	)
	return makeCmd(conda, condaPath, args...)
}

type PkgSpec struct {
	Name     string   `json:"name"`
	Channel  string   `json:"channel"`
	Depends  []string `json:"depends"`
	DistName string   `json:"dist_name"`
	Platform string   `json:"platform"`
	Version  string   `json:"version"`
	BaseUrl  string   `json:"base_url"`
	Url      string   `json:"url,omitempty"`
	Sha256   string   `json:"sha256"`
	// In the output of conda create, the build string is "build_string", but in
	// but in the package info or index "build" is the build string.
	BuildStr string `json:"build_string,omitempty"`
	Build    string `json:"build,omitempty"`
}

func readSpecs(out io.ReadCloser, excludeList []string) (map[string]*PkgSpec, error) {
	defer out.Close()
	dec := json.NewDecoder(out)
	var condaResult struct {
		Actions struct {
			Fetch []PkgSpec `json:"FETCH"`
			Link  []PkgSpec `json:"LINK"`
		} `json:"actions"`
	}
	if err := dec.Decode(&condaResult); err != nil {
		return nil, err
	}
	specs := make(map[string]*PkgSpec, len(condaResult.Actions.Link))
	for i := range condaResult.Actions.Link {
		pkg := &condaResult.Actions.Link[i]
		if pkg.BaseUrl != "" && pkg.Platform != "" {
			pkg.BaseUrl = pkg.BaseUrl + "/" + pkg.Platform
		}
		if pkg.Url == "" && pkg.BaseUrl != "" {
			pkg.Url = pkg.BaseUrl + "/" + pkg.DistName + ".tar.bz2"
		}
		if pkg.DistName == "" {
			if pkg.BuildStr != "" {
				pkg.DistName = strings.Join([]string{
					pkg.Name, pkg.Version, pkg.BuildStr}, "-")
			}
		}
		specs[pkg.Name] = pkg
	}
	for _, pkg := range condaResult.Actions.Fetch {
		spec := specs[pkg.Name]
		if pkg.Url != "" {
			spec.Url = pkg.Url
			if pkg.Sha256 != "" {
				spec.Sha256 = pkg.Sha256
			}
		}
		if pkg.BaseUrl != "" && pkg.Platform != "" {
			spec.BaseUrl = pkg.BaseUrl + "/" + pkg.Platform
		} else if pkg.Url != "" {
			i := strings.LastIndexByte(pkg.Url, '/')
			spec.BaseUrl = pkg.Url[:i]
		}
		for i, d := range pkg.Depends {
			if j := strings.IndexByte(d, ' '); j > 0 {
				pkg.Depends[i] = d[:j]
			}
		}
		spec.Depends = pkg.Depends
	}
	for _, e := range excludeList {
		delete(specs, e)
	}
	return specs, nil
}

func urlRank(url string) int {
	if strings.HasPrefix(url, "https://conda.anaconda.org/") {
		return 0
	}
	if url == "https://conda.anaconda.org/main/noarch" {
		return 1
	}
	if strings.HasPrefix(url, "https://conda.anaconda.org/main") {
		return 2
	}
	if url == "https://repo.anaconda.com/pkgs/main/noarch" {
		return 3
	}
	if strings.HasPrefix(url, "https://repo.anaconda.com/pkgs/main") {
		return 4
	}
	if strings.HasPrefix(url, "https://repo.anaconda.com/") {
		return 5
	}
	if strings.HasPrefix(url, "https://") {
		return 6
	}
	return 7
}

func fillSpecs(specs map[string]*PkgSpec, conda, requirements string,
	channels []string, arch, tempdir string) error {
	pkgDir := path.Join(path.Dir(path.Dir(conda)), "pkgs")
	cacheFiles, err := filepath.Glob(path.Join(pkgDir, "cache/*.json"))
	if err != nil {
		return err
	}
	caches := make([]*struct {
		Url      string             `json:"_url"`
		Packages map[string]PkgSpec `json:"packages"`
	}, len(cacheFiles))
	for i, cachefile := range cacheFiles {
		if err := func(cachefile string, cache interface{}) error {
			f, err := os.Open(cachefile)
			if err != nil {
				return err
			}
			defer f.Close()
			dec := json.NewDecoder(f)
			return dec.Decode(cache)
		}(cachefile, &caches[i]); err != nil {
			return err
		}
	}
	sort.SliceStable(caches, func(i, j int) bool {
		r1, r2 := urlRank(caches[i].Url), urlRank(caches[j].Url)
		if r1 < r2 {
			return true
		}
		if r2 < r1 {
			return false
		}
		return caches[i].Url < caches[j].Url
	})
	for _, cache := range caches {
		for tarballName, pkgCache := range cache.Packages {
			if pkg := specs[pkgCache.Name]; pkg != nil &&
				(pkg.Sha256 == "" || pkg.BaseUrl == "") &&
				(pkg.DistName == strings.TrimSuffix(tarballName, ".tar.bz2") ||
					pkg.DistName == strings.TrimSuffix(tarballName, ".conda")) &&
				pkg.Version == pkgCache.Version &&
				(pkg.BuildStr == pkgCache.Build || pkg.Build == pkgCache.Build) &&
				(pkg.BaseUrl == "" || cache.Url == pkg.BaseUrl) {
				if pkgCache.Sha256 != "" {
					pkg.Sha256 = pkgCache.Sha256
				}
				if pkg.BaseUrl == "" {
					pkg.BaseUrl = cache.Url
				}
			}
		}
	}
	for _, pkg := range specs {
		if pkg.Sha256 == "" {
			if sha, err := computeSha256(path.Join(pkgDir, path.Base(pkg.Url))); err == nil {
				if sha != "" {
					pkg.Sha256 = sha
				} else {
					fmt.Fprintf(os.Stderr,
						`WARNING: Computed hash for %s was empty!
WARNING: The spec is incomplete. Continuing anyway to allow for debugging.
`, pkg.DistName)
				}
			} else if requirements == "" {
				fmt.Fprintf(os.Stderr,
					`WARNING: Error computing hash for %s: %v
WARNING: The spec is incomplete. Continuing anyway to allow for debugging.
`, pkg.DistName, err)
			} else {
				if pkg.Url != "" {
					fmt.Fprintln(os.Stderr, "Downloading", pkg.Url,
						"to compute checksum.")
					if sha, err := computeSha256http(pkg.Url); err != nil {
						fmt.Fprintf(os.Stderr,
							"WARNING: Error computing hash for %s: %v\n",
							pkg.DistName, err)
					} else if sha != "" {
						pkg.Sha256 = sha
					} else {
						fmt.Fprintf(os.Stderr,
							"WARNING: Computed hash for %s was empty!\n",
							pkg.DistName)
					}
				}
				if pkg.Sha256 == "" {
					fmt.Fprintln(os.Stderr, "No known checksum for", pkg.Name)
					return download(specs, conda, requirements, channels, arch, tempdir)
				}
			}
		}
	}
	for _, pkg := range specs {
		if pkg.BaseUrl == "" {
			if url, err := getBaseUrl(path.Join(pkgDir, pkg.DistName)); err == nil && url != "" {
				pkg.BaseUrl = url
			} else if requirements == "" {
				fmt.Fprintf(os.Stderr,
					`WARNING: Error getting base URL for %s: %v
WARNING: The spec is incomplete. Continuing anyway to allow for debugging.
It is recommended that you run 'git checkout -p' on the generated package
lock file before committing to revert broken changes.
`, pkg.DistName, err)
			} else {
				fmt.Fprintln(os.Stderr, "No known URL for", pkg.Name)
				return download(specs, conda, requirements, channels, arch, tempdir)
			}
		}
	}
	return nil
}

func computeSha256(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return computeSha256reader(f)
}

func computeSha256http(url string) (string, error) {
	r, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	return computeSha256reader(r.Body)
}

func computeSha256reader(f io.Reader) (string, error) {
	var buffer [4096]byte
	sum := sha256.New()
	n, err := f.Read(buffer[:])
	for n > 0 && err == nil {
		if _, err := sum.Write(buffer[:n]); err != nil {
			return "", err
		}
		n, err = f.Read(buffer[:])
	}
	if err != nil && err != io.EOF {
		return "", err
	}
	if n > 0 {
		if _, err := sum.Write(buffer[:n]); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(sum.Sum(nil)), nil
}

func getBaseUrl(pkgDir string) (string, error) {
	b, err := os.ReadFile(path.Join(pkgDir, "info/repodata_record.json"))
	if err != nil {
		return "", err
	}
	var data struct {
		BaseUrl string `json:"channel"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return "", err
	}
	return data.BaseUrl, nil
}

func download(specs map[string]*PkgSpec, conda, requirements string,
	channels []string, arch, tempdir string) error {
	fmt.Fprintln(os.Stderr, "Downloading missing packages to compute hashes...")
	cmd := condaDownload(requirements, conda, channels, arch, tempdir)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	return fillSpecs(specs, conda, "", channels, arch, tempdir)
}
