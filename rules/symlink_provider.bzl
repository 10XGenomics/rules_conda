"""Functionality for propagating information about symlinks through dependencies.

This is mostly a hack to get around lack of support for
`ctx.actions.declare_symlink`.
"""

SymlinkInfo = provider(
    doc = "Provides information about files which are symlinks",
    fields = {
        "symlinks": "dict[File, str]: transitive set of symlinks.",
    },
)

ExecutableFileInfo = provider(
    doc = "Provides information about files which should have their executable bit set.",
    fields = {
        "files": "depset[File]: The transitive set of executable files.",
    },
)

CollectedSymlinkInfo = provider(
    doc = "Provides information about a transitive set of symlink maps " +
          "and executable files.",
    fields = {
        "symlinks": "depset(SymlinkInfo): transitive set of symlinks.",
        "executables": "depset(File): transitive set of executables.",
    },
)

_EMPTY_DICT = {}

def merge_deps_symlinks(deps):
    """Merge dictionaries from SimlinkInfo providers in a set of dependencies.

    Args:
        deps: (list[targets]) The targets with SimlinkInfo providers to merge.

    Returns:
        dict[File, str]: Files which are symlinks, and their targets.
    """
    if len(deps) == 0:
        return _EMPTY_DICT
    dep_links = [dep[SymlinkInfo].symlinks for dep in deps if dep[SymlinkInfo].symlinks]
    return _collect_symlinks(dep_links)

def _collect_symlinks(dep_links):
    """Coolect a set of symlink dictionaries into a single dictionary.

    This is intended to be used to merge symlink infos found using the
    `gather_symlink_info` aspect, e.g.

    ```starlark
    collect_symlinks([
        info.symlinks
        for info in depset(transitive = [
            dep[CollectedSymlinkInfo].symlinks
            for dep in deps
            if CollectedSymlinkInfo in dep
        ]).to_list()
    ])
    ```

    Args:
        dep_links: (dict[File, str]) The dictionaries to merge.

    Returns:
        dict[File, str]: The merged dictionary.
    """
    if not dep_links:
        return _EMPTY_DICT
    if len(dep_links) == 1:
        # Don't allocate a new dictionary if we don't have to.
        return dep_links[0]
    longest = dep_links[0]
    for links in dep_links[1:]:
        if len(links) > len(longest):
            longest = links
    result = {}
    for links in dep_links:
        result.update(links.items())
    if len(result) == len(longest):
        # One of them was a superset of the others.  Save memory by just
        # returning that one.
        return longest
    return result

_ATTR_BLACKLIST = (
    "applicable_licenses",
    "struct.output_licenses",
    "struct.to_json",
    "struct.to_proto",
)

def merge_collected_symlinks(deps):
    """Gets a merged dictionary of symlinks from a set of targets.

    Args:
        deps: (list[Target]) A set of targets to check. These targets should
            come via an attribute using the `gather_symlink_info` attribute.

    Returns:
        dict[File, str]: The merged dictionary.
    """
    return _collect_symlinks([
        info.symlinks
        for info in depset(transitive = [
            dep[CollectedSymlinkInfo].symlinks
            for dep in deps
            if CollectedSymlinkInfo in dep
        ]).to_list()
    ])

def _is_target(dep):
    if type(dep) == "Target":
        return CollectedSymlinkInfo in dep
    return False

def _gather_symlink_info_impl(target, ctx):
    if SymlinkInfo in target:
        return [CollectedSymlinkInfo(
            symlinks = depset([target[SymlinkInfo]]),
            executables = target[ExecutableFileInfo].files if ExecutableFileInfo in target else depset(),
        )]
    deps = []
    executables = []
    if ExecutableFileInfo in target:
        executables.append(target[ExecutableFileInfo].files)
    for attr in dir(ctx.rule.attr):
        if attr in _ATTR_BLACKLIST:
            continue
        value = getattr(ctx.rule.attr, attr)
        if value == None:
            continue
        if type(value) in ("list", "dict"):
            for dep in value:
                if _is_target(dep):
                    deps.append(dep[CollectedSymlinkInfo].symlinks)
                    executables.append(dep[CollectedSymlinkInfo].executables)
        elif _is_target(value):
            deps.append(value[CollectedSymlinkInfo].symlinks)
            executables.append(value[CollectedSymlinkInfo].executables)

    return [CollectedSymlinkInfo(
        symlinks = depset(transitive = deps),
        executables = depset(transitive = executables),
    )]

gather_symlink_info = aspect(
    doc = """Collects SymlinkInfo providers from transitive dependencies.""",
    implementation = _gather_symlink_info_impl,
    attr_aspects = ["*"],
    apply_to_generating_rules = True,
    provides = [CollectedSymlinkInfo],
)
