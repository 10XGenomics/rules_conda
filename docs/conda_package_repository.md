<!-- Generated with Stardoc: http://skydoc.bazel.build -->

This file define the repository rule for fetching conda packages.

<a id="conda_package_repository"></a>

## conda_package_repository

<pre>
load("@com_github_10XGenomics_rules_conda//rules:conda_package_repository.bzl", "conda_package_repository")

conda_package_repository(<a href="#conda_package_repository-name">name</a>, <a href="#conda_package_repository-archive_type">archive_type</a>, <a href="#conda_package_repository-auth_patterns">auth_patterns</a>, <a href="#conda_package_repository-base_url">base_url</a>, <a href="#conda_package_repository-base_urls">base_urls</a>, <a href="#conda_package_repository-cc_include_path">cc_include_path</a>,
                         <a href="#conda_package_repository-conda_repo">conda_repo</a>, <a href="#conda_package_repository-dist_name">dist_name</a>, <a href="#conda_package_repository-exclude">exclude</a>, <a href="#conda_package_repository-exclude_deps">exclude_deps</a>, <a href="#conda_package_repository-extra_deps">extra_deps</a>, <a href="#conda_package_repository-license_file">license_file</a>,
                         <a href="#conda_package_repository-licenses">licenses</a>, <a href="#conda_package_repository-netrc">netrc</a>, <a href="#conda_package_repository-patch_args">patch_args</a>, <a href="#conda_package_repository-patch_cmds">patch_cmds</a>, <a href="#conda_package_repository-patch_cmds_win">patch_cmds_win</a>, <a href="#conda_package_repository-patch_tool">patch_tool</a>, <a href="#conda_package_repository-patches">patches</a>,
                         <a href="#conda_package_repository-repo_mapping">repo_mapping</a>, <a href="#conda_package_repository-sha256">sha256</a>)
</pre>

Fetches a conda package and sets up its BUILD file.

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="conda_package_repository-name"></a>name |  A unique name for this repository.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="conda_package_repository-archive_type"></a>archive_type |  The archive type (filename suffix) for the download.   | String | optional |  `"tar.bz2"`  |
| <a id="conda_package_repository-auth_patterns"></a>auth_patterns |  See [http_archive](https://docs.bazel.build/repo/http.html#http_archive-auth_patterns).   | <a href="https://bazel.build/rules/lib/dict">Dictionary: String -> String</a> | optional |  `{}`  |
| <a id="conda_package_repository-base_url"></a>base_url |  The base URL for fetching this package from conda, e.g. `https://conda.anaconda.org/conda-forge/linux-64`   | String | optional |  `"https://conda.anaconda.org/conda-forge/linux-64"`  |
| <a id="conda_package_repository-base_urls"></a>base_urls |  List of mirror URLs where the requested package can be found, e.g ["http://mirror.example.com/pkgs/conda-forge/linux-64", "https://conda.anaconda.org/conda-forge/linux-64"]`.   | List of strings | optional |  `[]`  |
| <a id="conda_package_repository-cc_include_path"></a>cc_include_path |  A list of include paths to add for C/C++ targets which depend on this package.  If left unspecified, any directory named `includes` and which contains `.h` file will be used.   | List of strings | optional |  `[]`  |
| <a id="conda_package_repository-conda_repo"></a>conda_repo |  The name of the merged repository, to use when referring to dependencies.   | String | optional |  `"conda_env"`  |
| <a id="conda_package_repository-dist_name"></a>dist_name |  The fully-qualified (including build ID) name of the package.   | String | optional |  `""`  |
| <a id="conda_package_repository-exclude"></a>exclude |  Glob patterns for files to ignore.   | List of strings | optional |  `[]`  |
| <a id="conda_package_repository-exclude_deps"></a>exclude_deps |  A list of dependencies to exclude from the set declared in metadata.   | List of strings | optional |  `[]`  |
| <a id="conda_package_repository-extra_deps"></a>extra_deps |  A list of dependencies to add to the set declared in metadata.   | List of strings | optional |  `[]`  |
| <a id="conda_package_repository-license_file"></a>license_file |  The tarball-relative path to the license file for this package. If not specified, the path found in the package's `about.json` file will be used.   | String | optional |  `""`  |
| <a id="conda_package_repository-licenses"></a>licenses |  One or more `license_kind` targets to use for the package license. If not specified, the appropriate taraget will be guessed from the license field in the package's `about.json` file.   | List of strings | optional |  `[]`  |
| <a id="conda_package_repository-netrc"></a>netrc |  Location of the .netrc file to use for authentication   | String | optional |  `""`  |
| <a id="conda_package_repository-patch_args"></a>patch_args |  The arguments given to the patch tool. Defaults to -p0, however -p1 will usually be needed for patches generated by git. If multiple -p arguments are specified, the last one will take effect.If arguments other than -p are specified, Bazel will fall back to use patch command line tool instead of the Bazel-native patch implementation. When falling back to patch command line tool and patch_tool attribute is not specified, `patch` will be used.   | List of strings | optional |  `["-p0"]`  |
| <a id="conda_package_repository-patch_cmds"></a>patch_cmds |  Sequence of Bash commands to be applied on Linux/Macos after patches are applied.   | List of strings | optional |  `[]`  |
| <a id="conda_package_repository-patch_cmds_win"></a>patch_cmds_win |  Sequence of Powershell commands to be applied on Windows after patches are applied. If this attribute is not set, patch_cmds will be executed on Windows, which requires Bash binary to exist.   | List of strings | optional |  `[]`  |
| <a id="conda_package_repository-patch_tool"></a>patch_tool |  The patch(1) utility to use. If this is specified, Bazel will use the specified patch tool instead of the Bazel-native patch implementation.   | String | optional |  `""`  |
| <a id="conda_package_repository-patches"></a>patches |  A list of files that are to be applied as patches after extracting the archive. By default, it uses the Bazel-native patch implementation which doesn't support fuzz match and binary patch, but Bazel will fall back to use patch command line tool if `patch_tool` attribute is specified or there are arguments other than `-p` in `patch_args` attribute.   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |
| <a id="conda_package_repository-repo_mapping"></a>repo_mapping |  In `WORKSPACE` context only: a dictionary from local repository name to global repository name. This allows controls over workspace dependency resolution for dependencies of this repository.<br><br>For example, an entry `"@foo": "@bar"` declares that, for any time this repository depends on `@foo` (such as a dependency on `@foo//some:target`, it should actually resolve that dependency within globally-declared `@bar` (`@bar//some:target`).<br><br>This attribute is _not_ supported in `MODULE.bazel` context (when invoking a repository rule inside a module extension's implementation function).   | <a href="https://bazel.build/rules/lib/dict">Dictionary: String -> String</a> | optional |  |
| <a id="conda_package_repository-sha256"></a>sha256 |  The sha256 checksum of the tarball to be downloaded.   | String | optional |  `""`  |


