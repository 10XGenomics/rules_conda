"""10X Toolchain registration."""

load("@bazel_skylib//:workspace.bzl", "bazel_skylib_workspace")
load("@bazel_skylib//lib:versions.bzl", "versions")
load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")
load("//rules:multiarch_http_file.bzl", "multiarch_http_file")
load("//rules:rules_bins_repository.bzl", "repo_rules_bins_repository")
load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies")

_MINIMUM_SUPPORTED_BAZEL_VERSION = "6.0.0"

def second_level_dependencies():
    """Sets up dependencies which were not set up in deps.bzl."""
    versions.check(_MINIMUM_SUPPORTED_BAZEL_VERSION)
    bazel_skylib_workspace()
    go_rules_dependencies()
    repo_rules_bins_repository(
        name = "com_github_10XGenomics_rules_conda_repository_helpers",
    )
    micromamba_version = "1.5.8"
    maybe(
        multiarch_http_file,
        name = "pm_mamba_micromamba",
        archive_member = "bin/micromamba",
        sha256 = {
            "@io_bazel_rules_go//go/platform:linux_amd64": "3376ccb2ace4bb1549659067f9f1e332fdd34e92e3be85d5968c8f45cff2b467",
            "@io_bazel_rules_go//go/platform:linux_arm64": "1812fb419da96af894a9449aff36e5b319689596b87e8a0080321780a43fba19",
            "@io_bazel_rules_go//go/platform:darwin_arm64": "d62bdc8179a485b931007d623f299ef307cdfba45438fc1f4a8d055ddc232ee1",
            "@io_bazel_rules_go//go/platform:darwin_amd64": "a520f5ae4ed6667c489f9b8635afe6632da73bdd3039d62ff91b47aeced3e4a3",
        },
        urls = {
            "@io_bazel_rules_go//go/platform:linux_amd64": [
                "https://conda.anaconda.org/conda-forge/linux-64/micromamba-1.5.8-0.tar.bz2".format(micromamba_version),
                "https://micro.mamba.pm/api/micromamba/linux-64/{}".format(micromamba_version),
            ],
            "@io_bazel_rules_go//go/platform:linux_arm64": [
                "https://conda.anaconda.org/conda-forge/linux-aarch64/micromamba-1.5.8-0.tar.bz2".format(micromamba_version),
                "https://micro.mamba.pm/api/micromamba/linux-aarch64/{}".format(micromamba_version),
            ],
            "@io_bazel_rules_go//go/platform:darwin_arm64": [
                "https://conda.anaconda.org/conda-forge/osx-arm64/micromamba-1.5.8-0.tar.bz2".format(micromamba_version),
                "https://micro.mamba.pm/api/micromamba/osx-arm64/{}".format(micromamba_version),
            ],
            "@io_bazel_rules_go//go/platform:darwin_amd64": [
                "https://conda.anaconda.org/conda-forge/osx-64/micromamba-1.5.8-0.tar.bz2".format(micromamba_version),
                "https://micro.mamba.pm/api/micromamba/osx-64/{}".format(micromamba_version),
            ],
        },
    )
