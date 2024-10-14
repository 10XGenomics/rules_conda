<!-- Generated with Stardoc: http://skydoc.bazel.build -->

This file defines the repository rule for assembling a conda environment.

<a id="conda_environment_repository"></a>

## conda_environment_repository

<pre>
load("@com_github_10XGenomics_rules_conda//rules:conda_environment.bzl", "conda_environment_repository")

conda_environment_repository(<a href="#conda_environment_repository-name">name</a>, <a href="#conda_environment_repository-aliases">aliases</a>, <a href="#conda_environment_repository-conda_packages">conda_packages</a>, <a href="#conda_environment_repository-executable_packages">executable_packages</a>, <a href="#conda_environment_repository-py_version">py_version</a>,
                             <a href="#conda_environment_repository-repo_mapping">repo_mapping</a>)
</pre>

Assembles a collection of conda packages together into one place.

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="conda_environment_repository-name"></a>name |  A unique name for this repository.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="conda_environment_repository-aliases"></a>aliases |  A set of target aliases to add, mapping from alias to actual.   | <a href="https://bazel.build/rules/lib/dict">Dictionary: String -> String</a> | optional |  `{}`  |
| <a id="conda_environment_repository-conda_packages"></a>conda_packages |  The list of conda package repositories.   | List of strings | required |  |
| <a id="conda_environment_repository-executable_packages"></a>executable_packages |  The list of conda packages which have executable entry points.   | List of strings | optional |  `["conda", "python"]`  |
| <a id="conda_environment_repository-py_version"></a>py_version |  The assumed python version for python libraries when generating BUILD files.   | Integer | optional |  `3`  |
| <a id="conda_environment_repository-repo_mapping"></a>repo_mapping |  In `WORKSPACE` context only: a dictionary from local repository name to global repository name. This allows controls over workspace dependency resolution for dependencies of this repository.<br><br>For example, an entry `"@foo": "@bar"` declares that, for any time this repository depends on `@foo` (such as a dependency on `@foo//some:target`, it should actually resolve that dependency within globally-declared `@bar` (`@bar//some:target`).<br><br>This attribute is _not_ supported in `MODULE.bazel` context (when invoking a repository rule inside a module extension's implementation function).   | <a href="https://bazel.build/rules/lib/dict">Dictionary: String -> String</a> | optional |  |


