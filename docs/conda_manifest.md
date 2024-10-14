<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Rules for conda_package_repository to provide metadata to installer rules.

<a id="conda_deps"></a>

## conda_deps

<pre>
load("@com_github_10XGenomics_rules_conda//rules:conda_manifest.bzl", "conda_deps")

conda_deps(<a href="#conda_deps-name">name</a>, <a href="#conda_deps-deps">deps</a>)
</pre>

A set of conda package dependencies.

Unlike a simple `filegroup`, this will forward `PyInfo` and `CcInfo`.

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="conda_deps-name"></a>name |  A unique name for this target.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="conda_deps-deps"></a>deps |  The conda packages (install targets) which this package depends on.   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |


<a id="conda_files"></a>

## conda_files

<pre>
load("@com_github_10XGenomics_rules_conda//rules:conda_manifest.bzl", "conda_files")

conda_files(<a href="#conda_files-name">name</a>, <a href="#conda_files-hdrs">hdrs</a>, <a href="#conda_files-dylibs">dylibs</a>, <a href="#conda_files-hdrs_with_placeholders">hdrs_with_placeholders</a>, <a href="#conda_files-lalibs">lalibs</a>, <a href="#conda_files-link_safe_runfiles">link_safe_runfiles</a>, <a href="#conda_files-py_srcs">py_srcs</a>,
            <a href="#conda_files-runfiles">runfiles</a>, <a href="#conda_files-staticlibs">staticlibs</a>)
</pre>

Lists the files involved in the package.

These files are separated into sets based on how they are used and how they are
copied over to the environment output.
The uses are either as python sources (in `PyInfo`), C headers and libraries
(in `CcInfo`), or runfiles.
The copy mechanism is either symlink, actual copy (needed in cases where things
might otherwise escape the output tree by evaluating a symlink), or copy with
modification, generally to replace a placeholder value.

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="conda_files-name"></a>name |  A unique name for this target.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="conda_files-hdrs"></a>hdrs |  C/C++ header files.   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |
| <a id="conda_files-dylibs"></a>dylibs |  Dynamic libraries.   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |
| <a id="conda_files-hdrs_with_placeholders"></a>hdrs_with_placeholders |  C/C++ header files which require modification during install.   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |
| <a id="conda_files-lalibs"></a>lalibs |  libtool library files   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |
| <a id="conda_files-link_safe_runfiles"></a>link_safe_runfiles |  All other files which are safe to symlink into place. This is a build performance optimization. When in doubt, put it in runfiles.   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |
| <a id="conda_files-py_srcs"></a>py_srcs |  Python source files.   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |
| <a id="conda_files-runfiles"></a>runfiles |  All other files which are not known to be safe to symlink into place.   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |
| <a id="conda_files-staticlibs"></a>staticlibs |  Static libraries.   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |


<a id="conda_manifest"></a>

## conda_manifest

<pre>
load("@com_github_10XGenomics_rules_conda//rules:conda_manifest.bzl", "conda_manifest")

conda_manifest(<a href="#conda_manifest-name">name</a>, <a href="#conda_manifest-executable">executable</a>, <a href="#conda_manifest-executables">executables</a>, <a href="#conda_manifest-includes">includes</a>, <a href="#conda_manifest-index">index</a>, <a href="#conda_manifest-info_files">info_files</a>, <a href="#conda_manifest-manifest">manifest</a>, <a href="#conda_manifest-noarch">noarch</a>,
               <a href="#conda_manifest-py_stubs">py_stubs</a>, <a href="#conda_manifest-python_prefix">python_prefix</a>, <a href="#conda_manifest-symlinks">symlinks</a>)
</pre>

A rule for presenting conda metadata to downstream rules.

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="conda_manifest-name"></a>name |  A unique name for this target.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="conda_manifest-executable"></a>executable |  The path of the executable entry point, if any.   | String | optional |  `""`  |
| <a id="conda_manifest-executables"></a>executables |  The paths for files which have their executable bit set.   | List of strings | optional |  `[]`  |
| <a id="conda_manifest-includes"></a>includes |  Include directories to add for CcInfo.   | List of strings | optional |  `[]`  |
| <a id="conda_manifest-index"></a>index |  The index.json file for the repo.   | <a href="https://bazel.build/concepts/labels">Label</a> | optional |  `None`  |
| <a id="conda_manifest-info_files"></a>info_files |  Additional metadata files required during installation, specifically `info/{no_link,has_prefix}` if available.   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |
| <a id="conda_manifest-manifest"></a>manifest |  The file containing the list of files, one of `info/paths.json`, `info/files.json`, or `info/files`. Only required if some files have placeholders.   | <a href="https://bazel.build/concepts/labels">Label</a> | optional |  `None`  |
| <a id="conda_manifest-noarch"></a>noarch |  The noarch linkage type, if any.   | String | optional |  `""`  |
| <a id="conda_manifest-py_stubs"></a>py_stubs |  The `filegroup` containing the generated python stub files.   | <a href="https://bazel.build/concepts/labels">Label</a> | optional |  `None`  |
| <a id="conda_manifest-python_prefix"></a>python_prefix |  Additional prefix to prepend to installation directory, if it's a python noarch package.   | String | optional |  `""`  |
| <a id="conda_manifest-symlinks"></a>symlinks |  Symlinks and their targets.   | <a href="https://bazel.build/rules/lib/dict">Dictionary: String -> String</a> | optional |  `{}`  |


