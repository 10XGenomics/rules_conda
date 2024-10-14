# rules_conda for bazel

This repository contains [bazel][] rules for assembling a [conda][] environment.

[bazel]: https://bazel.build
[conda]: https://conda.org/

These rules

1. Generate a lock file based on a package requirements list.
1. Build a conda environment from that lock file.
1. Support depending on just the subset of that environment needed by your
   build target.
1. Register the `python` in that environment as a bazel python toolchain.
1. Support using `conda` packages a C/C++ dependencies.

## Initial setup

Add the following to your workspace:

```starlark
load(
    "@bazel_tools//tools/build_defs/repo:git.bzl",
    "git_repository",
)

git_repository(
    name = "com_github_10XGenomics_rules_conda",
    remote = "https://github.com/10XGenomics/rules_conda.git",
)

load(
    "@com_github_10XGenomics_rules_conda//:deps.bzl",
    "rules_conda_dependencies",
)

rules_conda_dependencies()

load(
    "@tenx_bazel_rules//:deps2.bzl",
    "second_level_dependencies",
)

# Set up toolchains for go, rust, etc.
second_level_dependencies()


load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains")

# Go 
go_register_toolchains(version = "1.23.1")
```

Next, in the build file for whatever package you wish
(conventionally `third-party/conda`),
add a file listing requirements in a form that `micromamba create` will
understand, e.g.

```text
coverage>=6.3
cython>=0.29
numpy
python>=3.8.8
```

and a rule

```starlark
load("@com_github_10XGenomics_rules_conda//rules:conda_package_lock.bzl", "conda_package_lock")

conda_package_lock(
    name = "generate_package_lock",
    channels = [
        "conda-forge",
    ],
    requirements = "conda_requirements.txt",
    target = "conda_env.bzl",
    visibility = ["//visibility:public"],
)
```

See the [documentation][`conda_package_lock`] for that rule
for additional options you might want to set for it.

`bazel run` that target to generate the `conda_env.bzl` file, and then add

```starlark
load("//third-party/conda:conda_env.bzl", "conda_environment")

conda_environment()
```

to your `WORKSPACE`.

You may now build `@conda_env//:conda_env`
(a convenience target for building all packages in the environment.
Usually you don't want other bazel targets to depend on this).
It may not work on the first try, as it is common for conda metadata to have
errors which need to be manually corrected (see below).

[`conda_package_lock`]: docs/conda_package_lock.md

## Usage

### Python

The `conda_environment` workspace macro will automatically register the conda
python package as the python toolchain for bazel.
Python targets can depend on packages from that environment as e.g.

```starlark
py_library(
    name = "uses_numpy",
    srcs = "uses_numpy.py",
    deps = [
        "@conda_env//:numpy",
    ],
)
```

This will automatically pull in transitive dependencies as well.

### C/C++

You can also depend on packages from a `cc_library`,
if the package contains C/C++ headers or libraries.
This will _not_ pull in transitive compile-time or link-time dependencies,
however.
Conda packages do not make a distinction between a dependency
that is used at runtime, e.g. as a subprocess,
versus one that supplies headers or libraries necessary at compile/link time.
You probably don't want to be linking against `libpython` just because you
needed to link against a package that contains a few python scripts.
And, if you do, adding it explicitly is easier than "un-adding" it in the cases
where you don't.
Transitive dependencies will still be propagted as if they were `data`
dependencies, however.

### Executables

If you want to be able to use a conda package as an executable target
(e.g. with `bazel run` or in the `tools` of a `genrule`),
add it to the `executable_packages` of the `conda_environment_repository`
rule.

In addition to this making the package target executable,
this will also cause two additional targets to be created:

- `@conda_env//:<package_name>_exe_file` which has only the
  the executable file as an output (so you can use it in a `$(location)`
  expansion in a `genrule`, for example)
- `@conda_env//:<package_name>_exe`, which adds back the rest of the files
  as runfiles.

## Correcting conda metadata

It is common for at least one package to have issues.

### Unrecognized license identifiers

The most common problem is an invalid or unrecognized license declaration.
That can be overridden by adding e.g.

```starlark
licenses = ["@rules_license//licenses/spdx:LGPL-2.1"],
```

to the appropriate [`conda_package_repository`][] declaration in your
`conda_env.bzl`.

> [!NOTE]
> These bazel rules make a best-effort attempt to parse the license metadata
> in the conda package, however it may misinterpret that metadata,
> or the metadata could be simply incorrect.
> These rules do not offer legal advice.

[`conda_package_repository`]: docs/conda_package_repository.md

### Missing or unspecified license file

Package metadata will frequently include an incorrect path to the license file,
or simply not specify a path at all[^1].
In this situation, you can override the metadata by using the
`license_file` attribute.
The most common failure mode of this type is for the license file to actually
be at `info/licenses/LICENSE0.txt`.

[^1]: In theory this would be fine, but there is a
      [bug in rules_license](https://github.com/bazelbuild/rules_license/issues/31)
      which means that the license rule does not work properly if a license file
      is not provided.

### File conflicts

Sometimes two packages will declare the same file.
One very common example is packages carelessly including
`lib/python*/__pycache__/_sysconfigdata*.pyc`.

You can add glob patterns like that for files or directories into the
`exclude` attribute on a [`conda_package_repository`][],
in which case those files will be skipped.

If you wish, you can also use `exclude` to get rid of unnecessary content like
man pages or test suites, which might otherwise unnecessarily bloat your build.

### Patches

The [`conda_package_repository`][] supports the same attributes as `http_archive`
for patching.

### C/C++ include path

Sometimes you may wish to use a conda package as a build dependency for a
`cc_library` target.
The rules will make a best-effort attempt to guess the include path,
but in some cases you may need to override it, e.g.

```starlark
cc_include_path = ["include/eigen3/Eigen"]
```

### Missing or invalid dependencies

Some packages will declare dependencies they don't actually need
(at least for your use case),
or may have undeclared optional dependencies which they _do_ need for your
use cases.
The `exclude_deps` and `extra_deps` attributes can be used to
cut or add dependency edges.

The [`conda_package_lock`][] rule also has an `exclude` attribute
which will cause the tool to ignore packages from the results of the solve.
It will automatically add `exclude_deps` attributes for any packages which
depend on one of those.

### Aliases

Sometimes the name of a conda package doesn't match up with the name of a
python import or corresponding `pip` package
(e.g. `scikit-learn` vs. `sklearn`).
For convenience, the [`conda_environment_repository`][] can take a dictionary
of aliases to include in the generated `BUILD` file.
It will automatically create aliases with `-` replaced by `_`,
as this is a common transformation.

This automatic aliasing may cause problems in cases
where the upstream conda packaging attempted to solve this problem by creating
e.g. both `importlib-metadata` and `importlib_metadata`.
Generally one of those is an empty package that just depend on the other.
If that happens, `exclude` the empty package in the [`conda_package_lock`][]
rule and add an alias manually.

[`conda_environment_repository`]: docs/conda_environment_repository.md

## FAQ

### Are these rules stable enough to use in production?

This is a fork of an internal version[^2] of the rules which we've been using
in production for years.
They've worked very well for us,
and we don't forsee making major changes any time soon.
So, yes
(though this should not be construed as contradicting the license terms
which make clear that we provide no warranty).

[^2]: The main differences between this public version of the rules and
      the internal version are due to divergence between our internal fork of
      `rules_license` and the public one.
      We're hoping we can eventually upstream enough changes to `rules_license`
      to allow us to eliminate those differences.

### Code coverage?

Yes!  If you include the `coverage` package in your environment,
it'll automatically be added to the generated python toolchain definition.
You can try this out in this repo - `bazel coverage //...` will produce
`bazel-out/_coverage/_coverage_report.dat` with coverage of the
go, python, and c++ targets.

### `bzlmod`?

On the roadmap, not there yet.

### Why do you need a "lock file"?

`bzlmod` version resolution rules guarantee that you'll always get the same
package resolutions for the same `MODULE.bazel`
(with caveats around local configuration).
The conda package solver makes no such guarantee, so a lock file is essential
for build reproducibility.

The solver can also be quite slow, on the order of minutes for more
complex environments.  The lock file isn't difficult to manage.
We have a github action that periodically runs the solver opens a PR
to update the lock file.

### How does this relate to `rules_python`?

In many ways this is complementary with `rules_python`.
It is probably _not_ compatible with `pip_install`, however.
`pip`, and therefore `pip_install` by extension,
is somewhat inherently problematic for reproducible builds,
and it won't integrate well with the generated conda environment.

Also, the conda ecosystem is not exclusively about python.
We have internal repositories which don't use python at all,
but which use these rules to obtain C/C++ dependencies.

### What if I can't find the package I need in conda?

If you need a package that can't be found in `conda`,
you can get it some other way (e.g. from a `.whl` or building from source)
and inject it into the generated conda environment using a e.g.
`new_conda_package_http_repository`.

Add it to `extra_packages` in the `conda_package_lock` target to
make the generated `conda_environment_repository` include it.

### Why does it generate so _many_ repositories?

By generating separate repositories for each package,
you get in some cases much faster build times because fetching of
a package can go in on parallel with analysis of targets which don't depend
transitively on that package.

Really the `conda_package_repository` calls should be thought of as private
implementation details.
Most of the time you only need to interact with `@conda_env`.

### What about support for platforms other than linux/amd64?

On the roadmap.

There's nothing fundamental preventing it,
but it hasn't been a priority for us to work on it.
Parts of it, such as  work on other platforms, including
`conda_package_lock`,
so you can update the `conda_environment.bzl` from your aarch64 Macbook,
but at the moment only if your target platform is still linux on amd64
due to a few hard-coded assumptions here and there.

Unfortunately, cross compilation doesn't really work at the moment due to
(ironically) the need to support `noarch` packages.
We need to be able to execute the python executable in order to ask it
the correct path for `lib/pythonX.Y` to prepend for those packages.
We could avoid this if we were willing to just assume that the answer is
always `lib/pythonX.Y` where `X.Y` is always the major/minor version number,
but at least so far our need for supporting cross-compilation has been
insufficient to justify making such an assumption.

### What's `go` doing in my ruleset?

First of all, conda is not exclusively a Python ecosystem.

Some aspects of putting together an environment are simply too complicated
to do purely in starlark.
`go` has several desirable properties for performing those steps:

- Go is more portable than python.  No, really.  For repository rules,
  you can't use a build target, so you're basically stuck using the system
  python, which could be anything, or even nothing.
  Once you've got your nice python environment set up using those rules,
  great!  But bootstrapping still needs to happen.
- Go is _very_ fast to compile.
  At least once you have go's module cache populated,
  it can compile the tools necessary for building the repo in a fraction of a
  second.
- Go is pretty fast at runtime, too.
  It doesn't prioritize runtime performance the way e.g. rust does,
  but it's still faster than pretty much any scripting language.
  Even if you don't have hundreds of packages in  your conda environment,
  the time it takes to compile the go-based repository helper executable used
  to generate the BUILD files is going to be far less than the it saves
  by running a compiled executable rather than an interpreted script.
- The canonical tool for formatting BUILD files, `buildifier`, is written in go.
  The libraries backing it can be used to generate build files programmatically.
- The bazel rules for go are very mature.
  The SDK repository rule knows how to use the host go,
  or download the SDK from the internet,
  and then, critically, it exposes that SDK in a way that
  can be used from another repository rule.
- Subjective reasons.  Personal preferences of the rule authors.

In our internal monorepos which use this rule set,
we're using go for other reasons anyway, so it's a non-issue.
If you _aren't_ otherwise using go, then needing to make a call to configure
the go toolchain is understandably a little annoying.
We don't expect the extra two lines of boilerplate in your WORKSPACE
to be a deal-breaker for anyone.
