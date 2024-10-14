"""
Loader method for the repositories which are direct dependencies
of this workspace.
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")

def rules_conda_dependencies():
    """Loads repositories which can be loaded without other tools.

    After invoking this macro in a workspace, one should then
    ```python
    load("@com_github_10XGenomics_rules_conda//:toolchains.bzl", "register_pipeline_toolchains")
    register_pipeline_toolchains()
    ```
    in order to pick up transitive dependencies and set up the compiler
    toolchains.
    """

    go_rules_version = "v0.50.1"
    go_rules_sha = "f4a9314518ca6acfa16cc4ab43b0b8ce1e4ea64b81c38d8a3772883f153346b8"
    maybe(
        http_archive,
        name = "io_bazel_rules_go",
        sha256 = go_rules_sha,
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/" +
            "rules_go/releases/download/{ver}/rules_go-{ver}.zip".format(
                ver = go_rules_version,
            ),
            "https://github.com/bazelbuild/rules_go/releases/download/" +
            "{ver}/rules_go-{ver}.zip".format(
                ver = go_rules_version,
            ),
        ],
    )

    buildtools_version = "7.3.1"
    maybe(
        http_archive,
        name = "com_github_bazelbuild_buildtools",
        integrity = "sha256-BRlRwQ/4rd608QvjsM9HSzBLLM1nXyzHaDzdkBAyDKk=",
        strip_prefix = "buildtools-" + buildtools_version,
        urls = [
            "https://github.com/bazelbuild/buildtools/archive/refs/tags/v{}.tar.gz".format(buildtools_version),
        ],
    )

    maybe(
        http_archive,
        name = "bazel_skylib",
        # 1.6.1, latest as of 2024-05-20
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.6.1/bazel-skylib-1.6.1.tar.gz",
            "https://github.com/bazelbuild/bazel-skylib/releases/download/1.6.1/bazel-skylib-1.6.1.tar.gz",
        ],
        sha256 = "9f38886a40548c6e96c106b752f242130ee11aaa068a56ba7e56f4511f33e4f2",
    )

    maybe(
        http_archive,
        name = "rules_license",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/rules_license/releases/download/1.0.0/rules_license-1.0.0.tar.gz",
            "https://github.com/bazelbuild/rules_license/releases/download/1.0.0/rules_license-1.0.0.tar.gz",
        ],
        sha256 = "26d4021f6898e23b82ef953078389dd49ac2b5618ac564ade4ef87cced147b38",
    )
