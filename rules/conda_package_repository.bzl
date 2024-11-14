"""
This file define the repository rule for fetching conda packages.
"""

load(
    "@bazel_tools//tools/build_defs/repo:utils.bzl",
    "patch",
    "read_netrc",
    "read_user_netrc",
    "update_attrs",
    "use_netrc",
)

def _is_empty(ctx):
    """Returns true if the metadata indicates that this is an empty archive."""
    files_list = ctx.path("info/files")
    if files_list.exists:
        return not ctx.read(files_list).strip()
    paths_json = ctx.path("info/paths.json")
    if not paths_json.exists:
        fail("Archive contained neither info/files nor info/paths.json")
    return not json.decode(ctx.read(paths_json)).get("paths")

def _conda_package_repository_impl(ctx):
    # Get this path here, because it might trigger a restart of the fetch,
    # which would be better to have happen before we download anything.
    generator = ctx.path(ctx.attr._generator)
    dest = ctx.path("")

    if not ctx.attr.base_urls:
        base_urls = [ctx.attr.base_url]
    else:
        base_urls = ctx.attr.base_urls

    url = [
        "{}/{}.{}".format(url, ctx.attr.dist_name, ctx.attr.archive_type)
        for url in ctx.attr.base_urls
    ]
    if not url:
        fail("At least one URL must be provided.", attr = "base_urls")

    if ctx.attr.netrc:
        netrc = read_netrc(ctx, ctx.attr.netrc)
    else:
        netrc = read_user_netrc(ctx)
    auth = use_netrc(netrc, url, ctx.attr.auth_patterns)

    ctx.file(
        "WORKSPACE",
        "workspace(name = '{name}')\n".format(name = ctx.name),
        executable = False,
    )

    if ctx.attr.archive_type == "conda":
        download_info = ctx.download_and_extract(
            url = url,
            sha256 = ctx.attr.sha256,
            auth = auth,
            type = "zip",
        )
        ctx.delete("metadata.json")
        ctx.extract("info-{}.tar.zst".format(ctx.attr.dist_name))
        ctx.delete("info-{}.tar.zst".format(ctx.attr.dist_name))

        # Do not attempt to untar empty archives.
        if not _is_empty(ctx):
            ctx.extract("pkg-{}.tar.zst".format(ctx.attr.dist_name))
        ctx.delete("pkg-{}.tar.zst".format(ctx.attr.dist_name))
    else:
        download_info = ctx.download_and_extract(
            url = url,
            sha256 = ctx.attr.sha256,
            auth = auth,
        )
    patch(ctx)
    i = base_urls[0].rfind("/")
    if i > 0:
        channel = base_urls[0][:i]
        i = channel.rfind("/")
        if i > 0:
            channel = channel[i + 1:]
        else:
            channel = ""
    else:
        channel = ""
    ctx.report_progress("Generating BUILD file...")
    generate_build = ctx.execute(
        [
            generator,
            "-dir",
            dest,
            "-distname",
            ctx.attr.dist_name,
            "-licenses",
            " ".join(ctx.attr.licenses),
            "-license_file",
            ctx.attr.license_file or "",
            "-exclude_deps",
            ",".join(ctx.attr.exclude_deps),
            "-extra_deps",
            ",".join(ctx.attr.extra_deps),
            "-cc_include_path",
            ":".join(ctx.attr.cc_include_path),
            "-channel",
            channel,
            "-url",
            url[0],
            "-type",
            ctx.attr.archive_type,
            "-conda",
            ctx.attr.conda_repo,
        ] + ctx.attr.exclude,
        quiet = True,
    )
    if generate_build.return_code != 0:
        fail("Failed to generate BUILD file: " + generate_build.stderr)
    return update_attrs(ctx.attr, _conda_package_repository_attrs.keys(), {
        "sha256": download_info.sha256,
        "base_urls": base_urls,
    })

# buildifier: disable=attr-licenses
_conda_package_repository_attrs = {
    "base_url": attr.string(
        default = "https://conda.anaconda.org/conda-forge/linux-64",
        doc = "The base URL for fetching this package from conda, e.g. " +
              "`https://conda.anaconda.org/conda-forge/linux-64`",
    ),
    "base_urls": attr.string_list(
        doc = """List of mirror URLs where the requested package can be found, e.g
["http://mirror.example.com/pkgs/conda-forge/linux-64",
"https://conda.anaconda.org/conda-forge/linux-64"]`.""",
    ),
    "dist_name": attr.string(
        doc = "The fully-qualified (including build ID) name of the package.",
    ),
    "sha256": attr.string(
        doc = "The sha256 checksum of the tarball to be downloaded.",
    ),
    "archive_type": attr.string(
        doc = "The archive type (filename suffix) for the download.",
        default = "tar.bz2",
    ),
    # These match the attributes for bazel's http_archive repository rule,
    # and share their implementation. See
    # https://github.com/bazelbuild/bazel/blob/2.0.0/tools/build_defs/repo/http.bzl
    "netrc": attr.string(
        doc = "Location of the .netrc file to use for authentication",
    ),
    "patches": attr.label_list(
        default = [],
        doc =
            "A list of files that are to be applied as patches after " +
            "extracting the archive. By default, it uses the Bazel-native patch implementation " +
            "which doesn't support fuzz match and binary patch, but Bazel will fall back to use " +
            "patch command line tool if `patch_tool` attribute is specified or there are " +
            "arguments other than `-p` in `patch_args` attribute.",
    ),
    "patch_tool": attr.string(
        default = "",
        doc = "The patch(1) utility to use. If this is specified, Bazel will use the specified " +
              "patch tool instead of the Bazel-native patch implementation.",
    ),
    "patch_args": attr.string_list(
        default = ["-p0"],
        doc =
            "The arguments given to the patch tool. Defaults to -p0, " +
            "however -p1 will usually be needed for patches generated by " +
            "git. If multiple -p arguments are specified, the last one will take effect." +
            "If arguments other than -p are specified, Bazel will fall back to use patch " +
            "command line tool instead of the Bazel-native patch implementation. When falling " +
            "back to patch command line tool and patch_tool attribute is not specified, " +
            "`patch` will be used.",
    ),
    "patch_cmds": attr.string_list(
        default = [],
        doc = "Sequence of Bash commands to be applied on Linux/Macos after patches are applied.",
    ),
    "patch_cmds_win": attr.string_list(
        default = [],
        doc = "Sequence of Powershell commands to be applied on Windows after patches are " +
              "applied. If this attribute is not set, patch_cmds will be executed on Windows, " +
              "which requires Bash binary to exist.",
    ),
    "auth_patterns": attr.string_dict(
        doc = "See [http_archive](https://docs.bazel.build/" +
              "repo/http.html#http_archive-auth_patterns).",
    ),
    "licenses": attr.string_list(
        doc = "One or more `license_kind` targets to use for the package license. " +
              "If not specified, the appropriate taraget will be guessed from " +
              "the license field in the package's `about.json` file.",
    ),
    "license_file": attr.string(
        doc = "The tarball-relative path to the license file for this package. " +
              "If not specified, the path found in the package's `about.json` " +
              "file will be used.",
    ),
    "exclude": attr.string_list(
        doc = "Glob patterns for files to ignore.",
    ),
    "exclude_deps": attr.string_list(
        doc = "A list of dependencies to exclude from the set declared in metadata.",
    ),
    "extra_deps": attr.string_list(
        doc = "A list of dependencies to add to the set declared in metadata.",
    ),
    "cc_include_path": attr.string_list(
        doc = "A list of include paths to add for C/C++ targets which depend on " +
              "this package.  If left unspecified, any directory named `includes` " +
              "and which contains `.h` file will be used.",
    ),
    "conda_repo": attr.string(
        doc = "The name of the merged repository, " +
              "to use when referring to dependencies.",
        default = "conda_env",
    ),
    # Tool dependencies
    "_generator": attr.label(
        default = Label("@com_github_10XGenomics_rules_conda_repository_helpers//:generate_conda_package_repo"),
        allow_files = True,
        doc = "The go program used to generate the BUILD file for the package.",
    ),
}

conda_package_repository = repository_rule(
    attrs = _conda_package_repository_attrs,
    doc = "Fetches a conda package and sets up its BUILD file.",
    local = False,
    implementation = _conda_package_repository_impl,
)
