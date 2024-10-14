<!-- Generated with Stardoc: http://skydoc.bazel.build -->

Defines a macro rule to generate a conda lockfile.

A user would define this rule, referring to a `requirements.txt` file
and a target bzl file, to regenerate that file with `bazel run`.  The
target file contains an `conda_environment` which can be loaded and
invoked in the repository WORKSPACE file to make the @conda_env target.

<a id="conda_package_lock"></a>

## conda_package_lock

<pre>
load("@com_github_10XGenomics_rules_conda//rules:conda_package_lock.bzl", "conda_package_lock")

conda_package_lock(<a href="#conda_package_lock-name">name</a>, <a href="#conda_package_lock-requirements">requirements</a>, <a href="#conda_package_lock-channels">channels</a>, <a href="#conda_package_lock-exclude">exclude</a>, <a href="#conda_package_lock-extra_packages">extra_packages</a>, <a href="#conda_package_lock-target">target</a>, <a href="#conda_package_lock-glibc_version">glibc_version</a>,
                   <a href="#conda_package_lock-build_file_name">build_file_name</a>, <a href="#conda_package_lock-kwargs">kwargs</a>)
</pre>

Defines a build target for regenerating the conda package lock.

Consums the requirements.txt file, and produces the package lock file.
Requires an existing package lock that includes conda itself.

Once the make_conda_spec tool has been run manually to bootstrap the
initial repository, a build target defined with this rule can be used
to maintain the package lock, by running
`bazel run //:generate_package_lock`.


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="conda_package_lock-name"></a>name |  The name of the generator target to be invoked with `bazel run`.   |  `"generate_package_lock"` |
| <a id="conda_package_lock-requirements"></a>requirements |  The requirements.txt source file, formatted for `conda`.   |  `"requirements.txt"` |
| <a id="conda_package_lock-channels"></a>channels |  A list of conda channels in which to look for packages.   |  `["conda-forge"]` |
| <a id="conda_package_lock-exclude"></a>exclude |  Packages to omit from the generated package lock file.   |  `[]` |
| <a id="conda_package_lock-extra_packages"></a>extra_packages |  Additional conda_package repository targets to include.   |  `[]` |
| <a id="conda_package_lock-target"></a>target |  The name of the output file, from which the `WORKSPACE` can load and call the `conda_environment` method.   |  `"conda_package_lock.bzl"` |
| <a id="conda_package_lock-glibc_version"></a>glibc_version |  The glibc version to tell `conda` to use when solving dependencies.   |  `""` |
| <a id="conda_package_lock-build_file_name"></a>build_file_name |  The name of this build file, used for finding the source repository to modify.   |  `"BUILD.bazel"` |
| <a id="conda_package_lock-kwargs"></a>kwargs |  additional arguments to the rule, e.g. `visibility`.   |  none |


