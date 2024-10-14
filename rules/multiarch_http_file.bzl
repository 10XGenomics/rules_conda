"""A repository macro for fetching a binary for several different architectures.

Instead of fetching all of the versions into a single repository, it fetches
a separate repository for each architecture, and then, in the "primary"
repository, uses and `alias` to `select` which one to get it from, so that
`bazel` doesn't need to fetch the binaries it doesn't actually need for the
current build configuration.
"""

load("@bazel_skylib//lib:types.bzl", "types")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")

def _multiarch_http_file_impl(ctx):
    ctx.file(
        "WORKSPACE",
        content = "workspace(name = \"{name}\")\n".format(name = ctx.name),
        executable = False,
    )
    ctx.file(
        "BUILD.bazel",
        content = """
alias(
    name = "{}",
    actual = select({{
        {}
    }}),
    visibility = ["//visibility:public"],
)
""".format(
            ctx.attr.target_name or ctx.name,
            "\n        ".join([
                '"{}": "{}",'.format(plt, repo)
                for plt, repo in ctx.attr.arches.items()
            ]),
        ),
        executable = False,
    )

_multiarch_http_file = repository_rule(
    implementation = _multiarch_http_file_impl,
    attrs = {
        "arches": attr.label_keyed_string_dict(allow_empty = False),
        "target_name": attr.string(),
    },
)

def multiarch_http_file(
        *,
        name,
        target_name = None,
        archive_member = None,
        urls = {},
        sha256 = {},
        platform_shortnames = {},
        executable = True,
        **kwargs):
    """Fetch a file from http, based on target platform.

    This generates several remote repositories, each named
    `<name>_<platform suffix>`, and a proxy repository `name` with a target like

    ```starlark
    alias(
        name = "<target_name>",
        actual = select({
            "platform1": "@<name>_platform1//file",
            "platform2": "@<name>_platform2//file",
        })
    )
    ```

    This allows targets to depend on `@<name>//:<target_name>` and get the right
    version for the current platform context, while downloading lazily for
    platforms not needed for the current build.

    The remote repositories are generated with either
    [http_file](https://bazel.build/rules/lib/repo/http#http_file)
    or [http_archive](https://bazel.build/rules/lib/repo/http#http_archive).

    Args:
        name: (string) The name of the repository which will contain the
            `select` target.
        target_name: The name of the `alias` target within the repository.
            Defaults to `name`.
        archive_member: (string) If set, treat the given URLs as archive targets
            (to download with `http_archive` rather than `http_file`),
            generating a `filegroup` target for the given file.
        urls: (dict[Label, list[string]] or dict[Label, string]) Dictionary,
            keyed by `config_setting` or equivalent, of URL(s) from which
            the file may be downloaded.
            This is used for the url or urls attribute to `http_file`.
        sha256: (dict[Label, string]) Dictionary with the same keys as `urls`,
            containing the checksums for each potential target.
        platform_shortnames: (dict[Label, string]) Typically, platforms used
            as keys for `urls` and `sha256` will be specified as e.g.
            `@io_bazel_rules_go//go/platform:linux_amd64`, so only the last
            part, e.g. `linux_amd64`, will be appended to `name` for the
            platform-specific repo name.  This dictionary may be used to
            override that suffix.
        executable: (boolean) True if the target file should be executable.
            The primary use case for this rule is downloading statically-linked
            executables to use during the build, so this defaults to True.
            No effect if `archive_member` is set.
        **kwargs: Additional arguments to pass to the fetch rules, e.g.
            `auth_patterns`.  These should be attributes for
            [`http_file`](https://bazel.build/rules/lib/repo/http#http_file)
            or
            [`http_archive`](https://bazel.build/rules/lib/repo/http#http_archive)
            depending on whether `archive_member` is specified.
    """
    if not urls:
        fail("Must specify at least one platform.", attr = "urls")
    if len(urls) != len(sha256):
        fail("Must have one checksum and URL for each platform.", attr = "sha256")
    repos = {}
    for platform, url in urls.items():
        if platform in platform_shortnames:
            platform_basename = platform_shortnames[platform]
        else:
            platform_basename = platform
            i = platform.rfind(":")
            if i >= 0:
                platform_basename = platform[i + 1:]
        repo_name = "{}_{}".format(name, platform_basename)
        if archive_member:
            repos[platform] = "@{}".format(repo_name, repo_name)
            build = """
sh_binary(
    name = "{}",
    srcs = ["{}"],
    visibility = ["@{}//:__pkg__"],
)
""".format(repo_name, archive_member, name)
            if types.is_list(url):
                http_archive(
                    name = repo_name,
                    build_file_content = build,
                    urls = url,
                    sha256 = sha256[platform],
                    **kwargs
                )
            else:
                http_archive(
                    name = repo_name,
                    build_file_content = build,
                    url = url,
                    sha256 = sha256[platform],
                    **kwargs
                )
        else:
            repos[platform] = "@{}//file".format(repo_name)
            if types.is_list(url):
                http_file(
                    name = repo_name,
                    urls = url,
                    sha256 = sha256[platform],
                    executable = executable,
                    **kwargs
                )
            else:
                http_file(
                    name = repo_name,
                    url = url,
                    sha256 = sha256[platform],
                    executable = executable,
                    **kwargs
                )
    _multiarch_http_file(
        name = name,
        arches = repos,
        target_name = target_name or name,
    )
