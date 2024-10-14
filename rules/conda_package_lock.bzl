"""Defines a macro rule to generate a conda lockfile.

A user would define this rule, referring to a `requirements.txt` file
and a target bzl file, to regenerate that file with `bazel run`.  The
target file contains an `conda_environment` which can be loaded and
invoked in the repository WORKSPACE file to make the @conda_env target.
"""

load(":util.bzl", "merge_runfiles")

def _conda_package_lock_generator_impl(ctx):
    ctx.actions.expand_template(
        template = ctx.file._generator_script,
        output = ctx.outputs.executable,
        substitutions = {
            "{conda}": ctx.executable._conda.short_path,
            "{generator}": ctx.executable._generator.short_path,
            "{glibc_version}": ctx.attr.glibc_version,
            "{requirements}": ctx.file.requirements.short_path,
            "{build}": ctx.file.root.short_path,
            "{target}": ctx.attr.target,
            "{channels}": ",".join(ctx.attr.channels),
            "{extra}": ",".join(ctx.attr.extra_packages),
            "{exclude}": ",".join(ctx.attr.exclude),
            "{architecture}": ctx.attr.architecture,
        },
        is_executable = True,
    )
    rf = merge_runfiles(
        ctx,
        [ctx.attr._conda],
        collect_data = True,
        collect_default = True,
        files = [
            ctx.outputs.executable,
            ctx.executable._conda,
            ctx.file.requirements,
            ctx.executable._generator,
            ctx.file.root,
        ],
    )
    return [DefaultInfo(
        executable = ctx.outputs.executable,
        files = depset([ctx.outputs.executable]),
        runfiles = rf,
    )]

_conda_package_lock_generator = rule(
    attrs = {
        "requirements": attr.label(
            allow_single_file = True,
            doc = "The requirements.txt source file, formatted for `conda`.",
        ),
        "target": attr.string(
            mandatory = True,
            doc = "The name of the output file, from which the " +
                  "`WORKSPACE` can load and call the " +
                  "`conda_environment` method.",
        ),
        "channels": attr.string_list(
            mandatory = True,
            allow_empty = False,
            doc = "A list of conda channels to use when solving the requirements.",
        ),
        "root": attr.label(
            allow_single_file = True,
            doc = "A file in the directory where the rule will be generated.",
        ),
        "exclude": attr.string_list(
            allow_empty = True,
            doc = "Packages to omit from the generated package lock file.",
        ),
        "extra_packages": attr.string_list(
            doc =
                """Extra packages to include.

List the "@conda_package_*" targets for any conda package remotes defined
separately from those defined by this rule here.  Generally these will
have been declared with `new_conda_package_repository` rules.""",
            default = [],
        ),
        "glibc_version": attr.string(
            doc = "The glibc version to tell `conda` to use when solving dependencies.",
            default = "2.17",
        ),
        "architecture": attr.string(
            doc = "The conda architecture to use for package solving.",
            default = "linux-64",
            values = [
                "linux-64",
                "win-64",
                "osx-64",
                "osx-arm64",
                "linux-aarch64",
            ],
        ),
        "_conda": attr.label(
            executable = True,
            cfg = "target",
            default = Label("@pm_mamba_micromamba//:pm_mamba_micromamba"),
        ),
        "_generator": attr.label(
            executable = True,
            cfg = "target",
            default = Label("@com_github_10XGenomics_rules_conda//cmd/make_conda_spec"),
        ),
        "_generator_script": attr.label(
            default = Label("@com_github_10XGenomics_rules_conda//scripts:conda_package_lock_generator"),
            allow_single_file = True,
        ),
    },
    executable = True,
    implementation = _conda_package_lock_generator_impl,
)

def conda_package_lock(
        name = "generate_package_lock",
        requirements = "requirements.txt",
        channels = [
            "conda-forge",
        ],
        exclude = [],
        extra_packages = [],
        target = "conda_package_lock.bzl",
        glibc_version = "",
        build_file_name = "BUILD.bazel",
        **kwargs):
    """Defines a build target for regenerating the conda package lock.

    Consums the requirements.txt file, and produces the package lock file.
    Requires an existing package lock that includes conda itself.

    Once the make_conda_spec tool has been run manually to bootstrap the
    initial repository, a build target defined with this rule can be used
    to maintain the package lock, by running
    `bazel run //:generate_package_lock`.

    Args:
      name: The name of the generator target to be invoked with
            `bazel run`.
      requirements: The requirements.txt source file, formatted for
                    `conda`.
      target: The name of the output file, from which the
              `WORKSPACE` can load and call the
              `conda_environment` method.
      build_file_name: The name of this build file, used for finding the source
                       repository to modify.
      channels: A list of conda channels in which to look for packages.
      exclude: Packages to omit from the generated package lock file.
      extra_packages: Additional conda_package repository targets to include.
      glibc_version (str): The glibc version to tell `conda` to use when solving
                           dependencies.
      **kwargs: additional arguments to the rule, e.g. `visibility`.
    """
    _conda_package_lock_generator(
        name = name,
        channels = channels,
        exclude = exclude,
        extra_packages = extra_packages,
        glibc_version = glibc_version,
        requirements = requirements,
        root = build_file_name,
        tags = ["no-sandbox"],
        target = target,
        **kwargs
    )
