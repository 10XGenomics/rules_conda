"""This declares a workspace rule to build binaries used in other workspace rules.

If this seems a bit convoluted, that is because it is.  The issue is that
bazel does not allow workspace rules to depend on built targets (for reasons
which are fairly obvious).  Therefore if you need to depend on a build binary,
you have to either fetch it from an archive somewhere (and make sure you're
getting the right one for your platform) or build it during the repository rule
action.  The latter approach is mostly fine for repository rules which are going
to be used once or twice, but for rules like the conda package repository rule
it can add up quickly.
"""

load("@bazel_skylib//lib:paths.bzl", "paths")

def _repo_rules_bins_impl(ctx):
    go = ctx.path(ctx.attr._go)
    for dep in ctx.attr.deps:
        # Pre-fetch to prevent restarts.
        ctx.path(dep)
    conda_package_repo = str(ctx.path(ctx.attr.conda_package_repo))
    ctx.file(
        "WORKSPACE",
        """workspace(name = \"{name}\")

# This repository contains generated binaries used by other repository rules.
""".format(name = ctx.name),
        executable = False,
    )
    ctx.symlink(ctx.attr._generator_script, "generate_repo_bins.bash")
    ctx.report_progress("Generating binaries for use in other fetches...")
    generate_build = ctx.execute(
        [
            "/usr/bin/env",
            "bash",
            "generate_repo_bins.bash",
            go,
            conda_package_repo,
        ],
        quiet = True,
    )
    ctx.file(
        "BUILD.bazel",
        """
exports_files(
    [
        "{}",
    ],
    visibility = ["//visibility:public"],
)    
""".format(
            paths.basename(paths.dirname(conda_package_repo)),
        ),
        executable = False,
    )
    if generate_build.return_code != 0:
        fail("Failed to generate binaries: " + generate_build.stderr)

repo_rules_bins_repository = repository_rule(
    attrs = {
        "conda_package_repo": attr.label(
            default = Label("@com_github_10XGenomics_rules_conda//cmd/generate_conda_package_repo:main.go"),
            allow_files = True,
            doc = "The go program used to generate the BUILD file for the individual conda packages.",
        ),
        # These labels are to force a rebuild if those files have changed.
        "deps": attr.label_list(
            default = [
                Label("@com_github_10XGenomics_rules_conda//conda:files.go"),
                Label("@com_github_10XGenomics_rules_conda//conda:metadata.go"),
                Label("@com_github_10XGenomics_rules_conda//conda:package_tarball.go"),
                Label("@com_github_10XGenomics_rules_conda//conda:packages.go"),
                Label("@com_github_10XGenomics_rules_conda//conda:python_package.go"),
                Label("@com_github_10XGenomics_rules_conda//buildutil:common.go"),
                Label("@com_github_10XGenomics_rules_conda//buildutil:repos.go"),
                Label("@com_github_10XGenomics_rules_conda//buildutil:util.go"),
                Label("@com_github_10XGenomics_rules_conda//licensing:known_licenses.go"),
                Label("@com_github_10XGenomics_rules_conda//licensing:licenses.go"),
                Label("@com_github_10XGenomics_rules_conda//licensing:normalize_license.go"),
                Label("@com_github_10XGenomics_rules_conda//licensing:spdx.go"),
                Label("@com_github_10XGenomics_rules_conda//:go.mod"),
            ],
            allow_files = True,
        ),
        "_go": attr.label(
            default = Label("@go_sdk//:bin/go"),
            allow_single_file = True,
        ),
        "_generator_script": attr.label(
            default = Label("@com_github_10XGenomics_rules_conda//scripts:generate_repo_bins.bash"),
            executable = True,
            cfg = "exec",
            allow_single_file = True,
        ),
    },
    local = False,
    implementation = _repo_rules_bins_impl,
)
