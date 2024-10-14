"""
This file define the repository rule for fetching git repositories
which can then be used in the same way as a `conda_package_repository`
in an conda_environment rule.
"""

load(
    "@bazel_tools//tools/build_defs/repo:git_worker.bzl",
    "git_repo",
)
load(
    "@bazel_tools//tools/build_defs/repo:utils.bzl",
    "update_attrs",
)
load(
    "//rules/private:new_conda_package_repository_utils.bzl",
    "CONDA_BUILD_DOC",
    "setup_workspace",
)

def _recursive_get_child(root, suffix):
    directory = root
    for part in suffix.split("/"):
        if part:
            directory = directory.get_child(part)
    return directory

# Copied from
# https://github.com/bazelbuild/bazel/blob/1.0.0/tools/build_defs/repo/git.bzl
# except with support for add_prefix
def _clone_or_update(ctx):
    if ((not ctx.attr.tag and not ctx.attr.commit and not ctx.attr.branch) or
        (ctx.attr.tag and ctx.attr.commit) or
        (ctx.attr.tag and ctx.attr.branch) or
        (ctx.attr.commit and ctx.attr.branch)):
        fail("Exactly one of commit, tag, or branch must be provided")

    root = ctx.path(".")
    directory = str(root)
    if ctx.attr.strip_prefix:
        directory = directory + "-tmp"
    elif ctx.attr.add_prefix:
        directory = str(_recursive_get_child(root, ctx.attr.add_prefix))

    git_ = git_repo(ctx, directory)

    if ctx.attr.strip_prefix:
        dest_link = "{}/{}".format(directory, ctx.attr.strip_prefix)
        if not ctx.path(dest_link).exists:
            fail("strip_prefix at {} does not exist in repo".format(ctx.attr.strip_prefix))
        if ctx.attr.add_prefix:
            ctx.symlink(dest_link, _recursive_get_child(root, ctx.attr.add_prefix))
        else:
            ctx.delete(root)
            ctx.symlink(dest_link, root)
    return {"commit": git_.commit, "shallow_since": git_.shallow_since}

# Copied from
# https://github.com/bazelbuild/bazel/blob/1.0.0/tools/build_defs/repo/git.bzl
def _update_git_attrs(orig, keys, override):
    result = update_attrs(orig, keys, override)

    # if we found the actual commit, remove all other means of specifying it,
    # like tag or branch.
    if "commit" in result:
        result.pop("tag", None)
        result.pop("branch", None)
    return result

def _new_conda_package_git_repository_implementation(ctx):
    if not ctx.name.startswith("conda_package_"):
        fail("repository name must begin with 'conda_package_'")
    if ((not ctx.attr.build_file and not ctx.attr.build_file_content) or
        (ctx.attr.build_file and ctx.attr.build_file_content)):
        fail("Exactly one of build_file and build_file_content must be provided.")
    update = _clone_or_update(ctx)
    setup_workspace(ctx)

    if ctx.attr.add_prefix:
        ctx.delete(ctx.path(ctx.attr.add_prefix).get_child(".git"))
    else:
        ctx.delete(ctx.path(".git"))
    return _update_git_attrs(
        ctx.attr,
        _new_conda_package_git_repository_attrs.keys(),
        update,
    )

_new_conda_package_git_repository_attrs = {
    # attributes from git_repository
    "remote": attr.string(
        mandatory = True,
        doc = "The URI of the remote Git repository",
    ),
    "commit": attr.string(
        default = "",
        doc =
            "specific commit to be checked out." +
            " Precisely one of branch, tag, or commit must be specified.",
    ),
    "shallow_since": attr.string(
        default = "",
        doc =
            "an optional date, not after the specified commit; the " +
            "argument is not allowed if a tag is specified (which allows " +
            "cloning with depth 1). Setting such a date close to the " +
            "specified commit allows for a more shallow clone of the " +
            "repository, saving bandwidth and wall-clock time.",
    ),
    "tag": attr.string(
        default = "",
        doc =
            "tag in the remote repository to checked out." +
            " Precisely one of branch, tag, or commit must be specified.",
    ),
    "branch": attr.string(
        default = "",
        doc =
            "branch in the remote repository to checked out." +
            " Precisely one of branch, tag, or commit must be specified.",
    ),
    "init_submodules": attr.bool(
        default = False,
        doc = "Whether to clone submodules in the repository.",
    ),
    "recursive_init_submodules": attr.bool(
        default = False,
        doc = "Whether to clone submodules recursively in the repository.",
    ),
    "verbose": attr.bool(default = False),
    "strip_prefix": attr.string(
        default = "",
        doc = "A directory prefix to strip from the extracted files.",
    ),
    "add_prefix": attr.string(
        default = "",
        doc = "A directory prefix to add to the extracted files. " +
              "If both this and strip_prefix are present, this is " +
              "applied after strip_prefix.",
    ),
    "patches": attr.label_list(
        default = [],
        doc =
            "A list of files that are to be applied as patches after " +
            "extracting the archive. By default, it uses the " +
            "Bazel-native patch implementation which doesn't support fuzz " +
            "match and binary patch, but Bazel will fall back to use " +
            "patch command line tool if `patch_tool` attribute is " +
            "specified or there are arguments other than `-p` in " +
            "`patch_args` attribute.",
    ),
    "patch_tool": attr.string(
        default = "",
        doc = "The patch(1) utility to use. If this is specified, Bazel " +
              "will use the specified patch tool instead of the " +
              "Bazel-native patch implementation.",
    ),
    "patch_args": attr.string_list(
        default = ["-p0"],
        doc =
            "The arguments given to the patch tool. Defaults to -p0, " +
            "however -p1 will usually be needed for patches generated by " +
            "git. If multiple -p arguments are specified, the last one " +
            "will take effect. If arguments other than -p are specified, " +
            "Bazel will fall back to use patch command line tool instead " +
            "of the Bazel-native patch implementation. When falling back " +
            "to patch command line tool and patch_tool attribute is not " +
            "specified, `patch` will be used.",
    ),
    "patch_cmds": attr.string_list(
        default = [],
        doc = "Sequence of Bash commands to be applied on Linux/Macos " +
              "after patches are applied.",
    ),
    "patch_cmds_win": attr.string_list(
        default = [],
        doc =
            "Sequence of Powershell commands to be applied on Windows " +
            "after patches are applied. If this attribute is not set, " +
            "patch_cmds will be executed on Windows, which requires " +
            "Bash binary to exist.",
    ),
    # Attributes from new_git_repository
    "build_file": attr.label(
        allow_single_file = True,
        doc =
            "The file to use as the BUILD file for this repository." +
            "This attribute is an absolute label (use '@//' for the main " +
            "repo). The file does not need to be named BUILD, but can " +
            "be (something like BUILD.new-repo-name may work well for " +
            "distinguishing it from the repository's actual BUILD files. " +
            "Either build_file or build_file_content must be specified.",
    ),
    "build_file_content": attr.string(
        doc =
            "The content for the BUILD file for this repository. " +
            "Either build_file or build_file_content must be specified.",
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
    "exported_files": attr.string_list(
        doc = "Does nothing.  Retained for backwards compatibility.",
    ),
    "version": attr.string(
        doc = "The version of the package to report in the `index.json`. " +
              "This is completely optional and does not affect SBOM metadata " +
              "computed by bazel. It is retained for backwards compatibility.",
    ),
    "license": attr.string(
        doc = "The package license to report in the `index.json`. " +
              "This is completely optional and does not affect SBOM metadata " +
              "computed by bazel. It is retained for backwards compatibility.",
    ),
}

new_conda_package_git_repository = repository_rule(
    implementation = _new_conda_package_git_repository_implementation,
    attrs = _new_conda_package_git_repository_attrs,
    doc = """Clone an external git repository and add conda metadata`.

Clones a Git repository, checks out the specified tag, or commit, and
makes its targets available for binding. Generates `info/index.json` and
`info/files` metadata files so it can be used as a source for a `conda_package`
rule in an `conda_environment` rule.
""" + CONDA_BUILD_DOC,
)
