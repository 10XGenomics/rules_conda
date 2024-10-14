"""
Shared functionality for new_conda_package_*_repository rules.
"""

load(
    "@bazel_tools//tools/build_defs/repo:utils.bzl",
    "patch",
    "workspace_and_buildfile",
)

def _generate_index_json(ctx):
    ctx.file(
        ctx.path("info/index.json"),
        content = json.encode_indent(struct(
            name = ctx.attr.package_name or ctx.name[len("conda_package_"):],
            version = ctx.attr.version,
            license = ctx.attr.license,
        ), indent = "    "),
        executable = False,
    )

def _generate_paths_list(ctx):
    ctx.file(
        ctx.path("info/files"),
        content = "\n".join(ctx.attr.exported_files) + "\n",
        executable = False,
    )

def generate_metadata(ctx):
    """Generate conda metadata files for the repository.

    Args:
        ctx: The workspace context.
    """
    ctx.report_progress("Generating metadata")
    if ctx.attr.exported_files:
        _generate_paths_list(ctx)
    _generate_index_json(ctx)

def setup_workspace(ctx):
    """Apply patches, add bazel and conda metadata files for the repository.

    Args:
        ctx: The workspace context.
    """
    workspace_and_buildfile(ctx)
    patch(ctx)

    generate_metadata(ctx)

CONDA_BUILD_DOC = """
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
"""