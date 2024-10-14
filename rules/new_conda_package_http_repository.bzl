"""
This file define the repository rule for fetching git repositories
which can then be used in the same way as a `conda_package_repository`
in an conda_environment rule.
"""

load(
    "@bazel_tools//tools/build_defs/repo:utils.bzl",
    "read_netrc",
    "read_user_netrc",
    "use_netrc",
    "update_attrs",
)
load(
    "//rules/private:new_conda_package_repository_utils.bzl",
    "CONDA_BUILD_DOC",
    "setup_workspace",
)

# mostly copied from
# https://github.com/bazelbuild/bazel/blob/3.7.2/tools/build_defs/repo/http.bzl
def _new_conda_package_http_repository_implementation(ctx):
    if not ctx.name.startswith("conda_package_"):
        fail("repository name must begin with 'conda_package_'")
    if not ctx.attr.url and not ctx.attr.urls:
        fail("At least one of url and urls must be provided")
    if ((not ctx.attr.build_file and not ctx.attr.build_file_content) or
        (ctx.attr.build_file and ctx.attr.build_file_content)):
        fail("Exactly one of build_file and build_file_content must be provided.")

    all_urls = []
    if ctx.attr.urls:
        all_urls = ctx.attr.urls
    if ctx.attr.url:
        all_urls = [ctx.attr.url] + all_urls
    if ctx.attr.netrc:
        netrc = read_netrc(ctx, ctx.attr.netrc)
    else:
        netrc = read_user_netrc(ctx)
    auth = use_netrc(netrc, all_urls, ctx.attr.auth_patterns)


    download_info = ctx.download_and_extract(
        all_urls,
        ctx.attr.add_prefix,
        ctx.attr.sha256,
        ctx.attr.type,
        ctx.attr.strip_prefix,
        canonical_id = ctx.attr.canonical_id,
        auth = auth,
    )
    setup_workspace(ctx)

    return update_attrs(
        ctx.attr,
        _new_conda_package_http_repository_attrs.keys(),
        {"sha256": download_info.sha256},
    )

_new_conda_package_http_repository_attrs = {
    # attributes from http_repository
    "url": attr.string(
        doc =
            """A URL to a file that will be made available to Bazel.
This must be a file, http or https URL. Redirections are followed.
Authentication is not supported.
This parameter is to simplify the transition from the native http_archive
rule. More flexibility can be achieved by the urls parameter that allows
to specify alternative URLs to fetch from.
""",
    ),
    "urls": attr.string_list(
        doc =
            """A list of URLs to a file that will be made available to Bazel.
Each entry must be a file, http or https URL. Redirections are followed.
Authentication is not supported.""",
    ),
    "sha256": attr.string(
        doc = """The expected SHA-256 of the file downloaded.
This must match the SHA-256 of the file downloaded. _It is a security risk
to omit the SHA-256 as remote files can change._ At best omitting this
field will make your build non-hermetic. It is optional to make development
easier but should be set before shipping.""",
    ),
    "netrc": attr.string(
        doc = "Location of the .netrc file to use for authentication",
    ),
    "auth_patterns": attr.string_dict(
        doc = "See [http_archive](https://docs.bazel.build/" +
              "repo/http.html#http_archive-auth_patterns).",
    ),
    "canonical_id": attr.string(
        doc = """A canonical id of the archive downloaded.
If specified and non-empty, bazel will not take the archive from cache,
unless it was added to the cache by a request with the same canonical id.
""",
    ),
    "strip_prefix": attr.string(
        doc = """A directory prefix to strip from the extracted files.
Many archives contain a top-level directory that contains all of the useful
files in archive. Instead of needing to specify this prefix over and over
in the `build_file`, this field can be used to strip it from all of the
extracted files.
For example, suppose you are using `foo-lib-latest.zip`, which contains the
directory `foo-lib-1.2.3/` under which there is a `WORKSPACE` file and are
`src/`, `lib/`, and `test/` directories that contain the actual code you
wish to build. Specify `strip_prefix = "foo-lib-1.2.3"` to use the
`foo-lib-1.2.3` directory as your top-level directory.
Note that if there are files outside of this directory, they will be
discarded and inaccessible (e.g., a top-level license file). This includes
files/directories that start with the prefix but are not in the directory
(e.g., `foo-lib-1.2.3.release-notes`). If the specified prefix does not
match a directory in the archive, Bazel will return an error.""",
    ),
    "add_prefix": attr.string(
        default = "",
        doc = "A directory prefix to add to the extracted files. " +
              "If both this and strip_prefix are present, this is " +
              "applied after strip_prefix.",
    ),
    "type": attr.string(
        doc = """The archive type of the downloaded file.
By default, the archive type is determined from the file extension of the
URL. If the file has no extension, you can explicitly specify one of the
following: `"zip"`, `"jar"`, `"war"`, `"tar"`, `"tar.gz"`, `"tgz"`,
`"tar.xz"`, or `tar.bz2`.""",
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
    "build_file": attr.label(
        allow_single_file = True,
        doc =
            "The file to use as the BUILD file for this repository." +
            "This attribute is an absolute label (use '@//' for the main " +
            "repo). The file does not need to be named BUILD, but can " +
            "be (something like BUILD.new-repo-name may work well for " +
            "distinguishing it from the repository's actual BUILD files. " +
            "Exactly one of build_file or build_file_content must be specified.",
    ),
    "build_file_content": attr.string(
        doc =
            "The content for the BUILD file for this repository. " +
            "Exactly one of build_file or build_file_content must be specified.",
    ),
    "workspace_file": attr.label(
        doc =
            "The file to use as the `WORKSPACE` file for this repository. " +
            "Either `workspace_file` or `workspace_file_content` can be " +
            "specified, or neither, but not both.",
    ),
    "workspace_file_content": attr.string(
        doc =
            "The content for the WORKSPACE file for this repository. " +
            "Either `workspace_file` or `workspace_file_content` can be " +
            "specified, or neither, but not both.",
    ),
    # These attributes are specific to this rule.
    "package_name": attr.string(
        doc =
            "The conda package name for this repository. Even if this " +
            "repository doesn't exist as a conda package, it will act " +
            "like one.  It must not conflict with another package name. " +
            "The default is the repository name.",
    ),
}

new_conda_package_http_repository = repository_rule(
    implementation = _new_conda_package_http_repository_implementation,
    attrs = _new_conda_package_http_repository_attrs,
    doc = """Download and extract an archive over http and add conda metadata`.

Dowwnloads an archive over http, verifies its hash, extracts the archive, and
makes its targets available for binding. Generates `info/index.json` and
`info/files` metadata files so it can be used as a source for a `conda_package`
rule in an `conda_environment` rule.

It supports the following file extensions: `"zip"`, `"jar"`, `"war"`, `"aar"`, `"tar"`,
`"tar.gz"`, `"tgz"`, `"tar.xz"`, `"txz"`, `"tar.zst"`, `"tzst"`, `tar.bz2`, `"ar"`,
or `"deb"`.
""" + CONDA_BUILD_DOC,
)
