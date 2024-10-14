workspace(name = "com_github_10XGenomics_rules_conda")

load("//:deps.bzl", "rules_conda_dependencies")

# gazelle:repo bazel_gazelle
rules_conda_dependencies()

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains")

go_register_toolchains(version = "1.23.1")

load("//:deps2.bzl", "second_level_dependencies")

second_level_dependencies()

load("//third-party/conda:conda_env.bzl", "conda_environment")

conda_environment()

#####################################################################
# Dependencies below here are only needed when developing within this
# repo, not required for using the rules.  Specifically, they're
# required for building the documentation.
#####################################################################

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_stardoc",
    sha256 = "fabb280f6c92a3b55eed89a918ca91e39fb733373c81e87a18ae9e33e75023ec",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/stardoc/releases/download/0.7.1/stardoc-0.7.1.tar.gz",
        "https://github.com/bazelbuild/stardoc/releases/download/0.7.1/stardoc-0.7.1.tar.gz",
    ],
)

register_toolchains(
    "//tools:java_toolchain_jdk17",
)

load("@io_bazel_stardoc//:setup.bzl", "stardoc_repositories")

stardoc_repositories()

load("@rules_jvm_external//:repositories.bzl", "rules_jvm_external_deps")

rules_jvm_external_deps()

load("@rules_jvm_external//:setup.bzl", "rules_jvm_external_setup")

rules_jvm_external_setup()

load("@io_bazel_stardoc//:deps.bzl", "stardoc_external_deps")

stardoc_external_deps()

load("@stardoc_maven//:defs.bzl", stardoc_pinned_maven_install = "pinned_maven_install")

stardoc_pinned_maven_install()
