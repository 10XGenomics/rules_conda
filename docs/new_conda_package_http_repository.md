<!-- Generated with Stardoc: http://skydoc.bazel.build -->

This file define the repository rule for fetching git repositories
which can then be used in the same way as a `conda_package_repository`
in an conda_environment rule.

<a id="new_conda_package_http_repository"></a>

## new_conda_package_http_repository

<pre>
load("@com_github_10XGenomics_rules_conda//rules:new_conda_package_http_repository.bzl", "new_conda_package_http_repository")

new_conda_package_http_repository(<a href="#new_conda_package_http_repository-name">name</a>, <a href="#new_conda_package_http_repository-add_prefix">add_prefix</a>, <a href="#new_conda_package_http_repository-auth_patterns">auth_patterns</a>, <a href="#new_conda_package_http_repository-build_file">build_file</a>, <a href="#new_conda_package_http_repository-build_file_content">build_file_content</a>,
                                  <a href="#new_conda_package_http_repository-canonical_id">canonical_id</a>, <a href="#new_conda_package_http_repository-netrc">netrc</a>, <a href="#new_conda_package_http_repository-package_name">package_name</a>, <a href="#new_conda_package_http_repository-patch_args">patch_args</a>, <a href="#new_conda_package_http_repository-patch_cmds">patch_cmds</a>,
                                  <a href="#new_conda_package_http_repository-patch_cmds_win">patch_cmds_win</a>, <a href="#new_conda_package_http_repository-patch_tool">patch_tool</a>, <a href="#new_conda_package_http_repository-patches">patches</a>, <a href="#new_conda_package_http_repository-repo_mapping">repo_mapping</a>, <a href="#new_conda_package_http_repository-sha256">sha256</a>,
                                  <a href="#new_conda_package_http_repository-strip_prefix">strip_prefix</a>, <a href="#new_conda_package_http_repository-type">type</a>, <a href="#new_conda_package_http_repository-url">url</a>, <a href="#new_conda_package_http_repository-urls">urls</a>, <a href="#new_conda_package_http_repository-workspace_file">workspace_file</a>,
                                  <a href="#new_conda_package_http_repository-workspace_file_content">workspace_file_content</a>)
</pre>

Download and extract an archive over http and add conda metadata`.

Dowwnloads an archive over http, verifies its hash, extracts the archive, and
makes its targets available for binding. Generates `info/index.json` and
`info/files` metadata files so it can be used as a source for a `conda_package`
rule in an `conda_environment` rule.

It supports the following file extensions: `"zip"`, `"jar"`, `"war"`, `"aar"`, `"tar"`,
`"tar.gz"`, `"tgz"`, `"tar.xz"`, `"txz"`, `"tar.zst"`, `"tzst"`, `tar.bz2`, `"ar"`,
or `"deb"`.

There are some additional requirements on the BUILD file:

There must be a `conda_files` target rule named `files`.
See the documentation for the `conda_files` rule to determine which
attribute to use for each file.

There must be a `conda_manifest` target named `conda_metadata`, which must
include `index = "info/index.json"` (which is genereated by the repo rule).

There must be a `conda_deps` target named `conda_deps`,
which should declare the targets which this package depends on.

The following is an example template for the `BUILD` file:

```starlark
load(
    "@com_github_10XGenomics_rules_conda//rules:conda_manifest.bzl",
    "conda_deps",
    "conda_files",
    "conda_manifest",
)
load("@conda_package_python//:vars.bzl", "PYTHON_PREFIX")

package(default_applicable_licenses = ["license"])

site_packages = PYTHON_PREFIX + "/site-packages"

license(
    name = "license",
    package_name = "<package name>",
    additional_info = {
        "homepage": "<package homepage>",
        "version": "<version>",
    },
    copyright_notice = "Copyright (c) 2016 ...",
    license_kinds = [
        "@rules_license//licenses/spdx:MIT",
    ],
    license_text = site_packages + "/LICENSE",
)


conda_files(
    name = "files",
    py_srcs = [site_packages + "/<package>/" + f for f in [
        "__init__.py",
        "foo.py",
    ]],
    dylibs = [
        "foo.so",
    ],
    visibility = ["@conda_env//:__pkg__"],
)

conda_manifest(
    name = "conda_metadata",
    info_files = ["info/index.json"],
    manifest = "info/files",
    visibility = ["//visibility:public"],
)

conda_deps(
    name = "conda_deps",
    visibility = ["@conda_env//:__pkg__"],
    deps = [
        "@conda_env//:numpy",
        "@conda_env//:sklearn",
    ],
)
```

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="new_conda_package_http_repository-name"></a>name |  A unique name for this repository.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="new_conda_package_http_repository-add_prefix"></a>add_prefix |  A directory prefix to add to the extracted files. If both this and strip_prefix are present, this is applied after strip_prefix.   | String | optional |  `""`  |
| <a id="new_conda_package_http_repository-auth_patterns"></a>auth_patterns |  See [http_archive](https://docs.bazel.build/repo/http.html#http_archive-auth_patterns).   | <a href="https://bazel.build/rules/lib/dict">Dictionary: String -> String</a> | optional |  `{}`  |
| <a id="new_conda_package_http_repository-build_file"></a>build_file |  The file to use as the BUILD file for this repository.This attribute is an absolute label (use '@//' for the main repo). The file does not need to be named BUILD, but can be (something like BUILD.new-repo-name may work well for distinguishing it from the repository's actual BUILD files. Exactly one of build_file or build_file_content must be specified.   | <a href="https://bazel.build/concepts/labels">Label</a> | optional |  `None`  |
| <a id="new_conda_package_http_repository-build_file_content"></a>build_file_content |  The content for the BUILD file for this repository. Exactly one of build_file or build_file_content must be specified.   | String | optional |  `""`  |
| <a id="new_conda_package_http_repository-canonical_id"></a>canonical_id |  A canonical id of the archive downloaded. If specified and non-empty, bazel will not take the archive from cache, unless it was added to the cache by a request with the same canonical id.   | String | optional |  `""`  |
| <a id="new_conda_package_http_repository-netrc"></a>netrc |  Location of the .netrc file to use for authentication   | String | optional |  `""`  |
| <a id="new_conda_package_http_repository-package_name"></a>package_name |  The conda package name for this repository. Even if this repository doesn't exist as a conda package, it will act like one.  It must not conflict with another package name. The default is the repository name.   | String | optional |  `""`  |
| <a id="new_conda_package_http_repository-patch_args"></a>patch_args |  The arguments given to the patch tool. Defaults to -p0, however -p1 will usually be needed for patches generated by git. If multiple -p arguments are specified, the last one will take effect.If arguments other than -p are specified, Bazel will fall back to use patch command line tool instead of the Bazel-native patch implementation. When falling back to patch command line tool and patch_tool attribute is not specified, `patch` will be used.   | List of strings | optional |  `["-p0"]`  |
| <a id="new_conda_package_http_repository-patch_cmds"></a>patch_cmds |  Sequence of Bash commands to be applied on Linux/Macos after patches are applied.   | List of strings | optional |  `[]`  |
| <a id="new_conda_package_http_repository-patch_cmds_win"></a>patch_cmds_win |  Sequence of Powershell commands to be applied on Windows after patches are applied. If this attribute is not set, patch_cmds will be executed on Windows, which requires Bash binary to exist.   | List of strings | optional |  `[]`  |
| <a id="new_conda_package_http_repository-patch_tool"></a>patch_tool |  The patch(1) utility to use. If this is specified, Bazel will use the specified patch tool instead of the Bazel-native patch implementation.   | String | optional |  `""`  |
| <a id="new_conda_package_http_repository-patches"></a>patches |  A list of files that are to be applied as patches after extracting the archive. By default, it uses the Bazel-native patch implementation which doesn't support fuzz match and binary patch, but Bazel will fall back to use patch command line tool if `patch_tool` attribute is specified or there are arguments other than `-p` in `patch_args` attribute.   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |
| <a id="new_conda_package_http_repository-repo_mapping"></a>repo_mapping |  In `WORKSPACE` context only: a dictionary from local repository name to global repository name. This allows controls over workspace dependency resolution for dependencies of this repository.<br><br>For example, an entry `"@foo": "@bar"` declares that, for any time this repository depends on `@foo` (such as a dependency on `@foo//some:target`, it should actually resolve that dependency within globally-declared `@bar` (`@bar//some:target`).<br><br>This attribute is _not_ supported in `MODULE.bazel` context (when invoking a repository rule inside a module extension's implementation function).   | <a href="https://bazel.build/rules/lib/dict">Dictionary: String -> String</a> | optional |  |
| <a id="new_conda_package_http_repository-sha256"></a>sha256 |  The expected SHA-256 of the file downloaded. This must match the SHA-256 of the file downloaded. _It is a security risk to omit the SHA-256 as remote files can change._ At best omitting this field will make your build non-hermetic. It is optional to make development easier but should be set before shipping.   | String | optional |  `""`  |
| <a id="new_conda_package_http_repository-strip_prefix"></a>strip_prefix |  A directory prefix to strip from the extracted files. Many archives contain a top-level directory that contains all of the useful files in archive. Instead of needing to specify this prefix over and over in the `build_file`, this field can be used to strip it from all of the extracted files. For example, suppose you are using `foo-lib-latest.zip`, which contains the directory `foo-lib-1.2.3/` under which there is a `WORKSPACE` file and are `src/`, `lib/`, and `test/` directories that contain the actual code you wish to build. Specify `strip_prefix = "foo-lib-1.2.3"` to use the `foo-lib-1.2.3` directory as your top-level directory. Note that if there are files outside of this directory, they will be discarded and inaccessible (e.g., a top-level license file). This includes files/directories that start with the prefix but are not in the directory (e.g., `foo-lib-1.2.3.release-notes`). If the specified prefix does not match a directory in the archive, Bazel will return an error.   | String | optional |  `""`  |
| <a id="new_conda_package_http_repository-type"></a>type |  The archive type of the downloaded file. By default, the archive type is determined from the file extension of the URL. If the file has no extension, you can explicitly specify one of the following: `"zip"`, `"jar"`, `"war"`, `"tar"`, `"tar.gz"`, `"tgz"`, `"tar.xz"`, or `tar.bz2`.   | String | optional |  `""`  |
| <a id="new_conda_package_http_repository-url"></a>url |  A URL to a file that will be made available to Bazel. This must be a file, http or https URL. Redirections are followed. Authentication is not supported. This parameter is to simplify the transition from the native http_archive rule. More flexibility can be achieved by the urls parameter that allows to specify alternative URLs to fetch from.   | String | optional |  `""`  |
| <a id="new_conda_package_http_repository-urls"></a>urls |  A list of URLs to a file that will be made available to Bazel. Each entry must be a file, http or https URL. Redirections are followed. Authentication is not supported.   | List of strings | optional |  `[]`  |
| <a id="new_conda_package_http_repository-workspace_file"></a>workspace_file |  The file to use as the `WORKSPACE` file for this repository. Either `workspace_file` or `workspace_file_content` can be specified, or neither, but not both.   | <a href="https://bazel.build/concepts/labels">Label</a> | optional |  `None`  |
| <a id="new_conda_package_http_repository-workspace_file_content"></a>workspace_file_content |  The content for the WORKSPACE file for this repository. Either `workspace_file` or `workspace_file_content` can be specified, or neither, but not both.   | String | optional |  `""`  |


