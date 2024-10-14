"""
This file defines the repository rule for assembling a conda environment.
"""

load("@bazel_skylib//lib:sets.bzl", "sets")
load("@bazel_skylib//lib:versions.bzl", "versions")

def _default_alias(pkg):
    """Automatically alias - to _

    That's more like to be the name the package is imported as in python.
    Adding the alias reduces confusion.
    """
    return pkg.replace(".", "_").replace("-", "_")

def _alias(pkg, actual):
    return """alias(
    name = "{name}",
    actual = "{actual}",
    visibility = ["//visibility:public"],
)""".format(name = pkg, actual = actual)

def _conda_exe(pkg, py):
    """Executable package.

    Export the package as an executable target.
    Also export file groups for ease in integrating the executable into
    e.g. genrules, which need a single-file target they can pass to
    `$(location)`.
    """
    return """conda_exe(
    name = "{name}",
    srcs = "@conda_package_{name}//:files",
    manifest = "@conda_package_{name}//:conda_metadata",
    py = "{py}",
    visibility = ["//visibility:public"],
    deps = ["@conda_package_{name}//:conda_deps"],
)

filegroup(
    name = "{name}_exe_file",
    srcs = [":{name}"],
    output_group = "exe_file",
    visibility = ["//visibility:public"],
)

filegroup(
    name = "{name}_exe",
    srcs = ["{name}_exe_file"],
    data = [":{name}"],
    visibility = ["//visibility:public"],
)""".format(name = pkg, py = py)

def _conda_package(pkg, py):
    return """conda_package(
    name = "{name}",
    srcs = "@conda_package_{name}//:files",
    manifest = "@conda_package_{name}//:conda_metadata",
    py = "{py}",
    visibility = ["//visibility:public"],
    deps = ["@conda_package_{name}//:conda_deps"],
)""".format(name = pkg, py = py)

def _python_package_target(coverage, py):
    return """conda_exe(
    name = "python",
    srcs = "@conda_package_python//:files",
    manifest = "@conda_package_python//:conda_metadata",
    visibility = ["//visibility:public"],
    deps = ["@conda_package_python//:conda_deps"],
)

config_setting(
    name = "coverage_enabled",
    values = {{"collect_code_coverage": "true"}},
)

PYTHON_RUNTIME_NAME = "python_runtime_{py}"

py_runtime(
    name = PYTHON_RUNTIME_NAME,
    coverage_tool = select({{
        ":coverage_enabled": "{coverage}",
        "//conditions:default": None,
    }}),
    files = [":python"],
    interpreter = ":python_interpreter_exe",
    python_version = "{py}",
    visibility = ["//visibility:public"],
)

alias(
    name = "python_interpreter",
    actual = PYTHON_RUNTIME_NAME,
    visibility = ["//visibility:public"],
)

py_runtime_pair(
    name = "python_runtime_pair",
    py3_runtime = ":python_interpreter",
)

toolchain(
    name = "python_toolchain",
    exec_compatible_with = ["@platforms//os:linux"],
    target_compatible_with = [
        "@platforms//cpu:x86_64",
        "@platforms//os:linux",
    ],
    toolchain = ":python_runtime_pair",
    toolchain_type = "@bazel_tools//tools/python:toolchain_type",
    visibility = ["//visibility:public"],
)

filegroup(
    name = "python_interpreter_exe",
    srcs = [":python"],
    output_group = "exe_file",
    visibility = ["//visibility:public"],
)""".format(py = py, coverage = ":coverage" if coverage else "None")

def _package_targets(pkg, actual, executable, coverage, py):
    if actual:
        return _alias(pkg, actual)
    if pkg == "python":
        return _python_package_target(coverage, py)
    if executable:
        return _conda_exe(pkg, py)
    return _conda_package(pkg, py)

def _generate_build_content(
        py_version,
        packages,
        executables,
        aliases,
        name,
        coverage):
    all_packages = {}
    for pkg in packages:
        if pkg in all_packages:
            fail("multiple packages named " + pkg, attr = "conda_packages")
        if pkg == name:
            fail(
                "package name must not match repository name " + pkg,
                attr = "conda_packages",
            )
        auto_alias = _default_alias(pkg)
        if auto_alias != pkg:
            all_packages[pkg] = auto_alias
            if auto_alias in all_packages:
                fail(
                    "multiple packages named {} (which is an automatic alias of {})".format(
                        auto_alias,
                        pkg,
                    ),
                    attr = "conda_packages",
                )
        all_packages[auto_alias] = None
    for alias, actual in aliases.items():
        if alias == name:
            fail(
                "alias must not match repository name " + pkg,
                attr = "conda_packages",
            )
        if alias in all_packages:
            other = all_packages[alias]
            if other:
                if other != actual:
                    fail(
                        "alias {} conflicts with existing alias {}".format(alias, other),
                        attr = "aliases",
                    )
            elif _default_alias(actual) != alias:
                # It can be difficult to keep track of which one is the real one
                # and which one is the alias, so be tolerant of this sort of
                # harmless redundancy.
                fail(
                    "alias {} conflicts with a package name".format(alias),
                    attr = "aliases",
                )
        else:
            all_packages[alias] = actual

    executables = sets.make(executables)
    coverage = coverage and sets.contains(executables, "coverage")
    py = "PY2" if py_version == 2 else "PY3"

    return "\n\n".join([
        "# Generated BUILD file for {}".format(name),
        """load(
    "@bazel_tools//tools/python:toolchain.bzl",
    "py_runtime_pair",
)
load(
    "@com_github_10XGenomics_rules_conda//rules:conda_install_rules.bzl",
    "conda_exe",
    "conda_package",
)""",
    ] + [
        _package_targets(pkg, actual, sets.contains(executables, pkg), coverage, py)
        for pkg, actual in sorted(all_packages.items())
    ] + [
        """filegroup(
    name = "{name}",
    srcs = [
        ":{pkgs}",
    ],
    visibility = ["//visibility:public"],
)""".format(name = name, pkgs = '''",
        ":'''.join(sorted(packages))),
    ])

def _conda_environment_impl(ctx):
    ctx.file(
        "WORKSPACE",
        "workspace(name = '{name}')\n".format(name = ctx.name),
        executable = False,
    )
    ctx.file(
        "BUILD.bazel",
        _generate_build_content(
            py_version = ctx.attr.py_version,
            packages = ctx.attr.conda_packages,
            executables = ctx.attr.executable_packages,
            aliases = ctx.attr.aliases,
            name = ctx.attr.name,
            coverage = versions.is_at_least(
                threshold = "6.0.0",
                version = versions.get(),
            ),
        ),
    )

conda_environment_repository = repository_rule(
    attrs = {
        "conda_packages": attr.string_list(
            mandatory = True,
            allow_empty = False,
            doc = "The list of conda package repositories.",
        ),
        "executable_packages": attr.string_list(
            default = ["conda", "python"],
            doc = "The list of conda packages which have executable entry points.",
        ),
        "py_version": attr.int(
            default = 3,
            doc = "The assumed python version for python libraries when generating BUILD files.",
        ),
        "aliases": attr.string_dict(
            doc = "A set of target aliases to add, mapping from alias to actual.",
        ),
    },
    doc = "Assembles a collection of conda packages together into one place.",
    local = False,
    implementation = _conda_environment_impl,
)
