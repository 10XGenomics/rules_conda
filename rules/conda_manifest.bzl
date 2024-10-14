"""Rules for conda_package_repository to provide metadata to installer rules."""

load(
    "@com_github_10XGenomics_rules_conda//rules:symlink_provider.bzl",
    "ExecutableFileInfo",
    "SymlinkInfo",
    "merge_deps_symlinks",
)
load(":providers.bzl", "CondaDepsInfo", "CondaFilesInfo", "CondaManifestInfo")

def _conda_manifest_impl(ctx):
    info = []
    manifest = ctx.file.manifest
    index = ctx.file.index
    if index:
        info.append(index)
    else:
        for fn in ctx.files.info_files:
            if fn.basename == "index.json":
                index = fn
                break
    if manifest:
        info.append(manifest)
    else:
        for fn in ctx.files.info_files:
            if fn.basename == "paths.json":
                index = fn
                break
    info.extend(ctx.files.info_files)
    return [
        DefaultInfo(
            files = depset(info),
        ),
        CondaManifestInfo(
            manifest = manifest,
            symlinks = ctx.attr.symlinks,
            executable = ctx.attr.executable,
            executables = ctx.attr.executables,
            includes = ctx.attr.includes,
            noarch = ctx.attr.noarch,
            py_stubs = ctx.attr.py_stubs[DefaultInfo].files if (
                ctx.attr.py_stubs and ctx.attr.py_stubs[DefaultInfo]
            ) else None,
            python_prefix = ctx.attr.python_prefix,
            index = index,
        ),
    ]

conda_manifest = rule(
    attrs = {
        "manifest": attr.label(
            doc = "The file containing the list of files, one of " +
                  "`info/paths.json`, `info/files.json`, or `info/files`. " +
                  "Only required if some files have placeholders.",
            allow_single_file = True,
        ),
        "index": attr.label(
            doc = "The index.json file for the repo.",
            allow_single_file = [".json"],
        ),
        "info_files": attr.label_list(
            allow_files = True,
            doc = "Additional metadata files required during installation, " +
                  "specifically `info/{no_link,has_prefix}` if available.",
        ),
        "symlinks": attr.string_dict(
            doc = "Symlinks and their targets.",
        ),
        "includes": attr.string_list(
            doc = "Include directories to add for CcInfo.",
        ),
        "executable": attr.string(
            doc = "The path of the executable entry point, if any.",
        ),
        "executables": attr.string_list(
            doc = "The paths for files which have their executable bit set.",
        ),
        "noarch": attr.string(
            doc = "The noarch linkage type, if any.",
        ),
        "py_stubs": attr.label(
            doc = "The `filegroup` containing the generated python stub files.",
        ),
        "python_prefix": attr.string(
            doc = "Additional prefix to prepend to installation directory, " +
                  "if it's a python noarch package.",
        ),
    },
    doc = "A rule for presenting conda metadata to downstream rules.",
    provides = [CondaManifestInfo],
    implementation = _conda_manifest_impl,
)

def _merge_pyinfo(deps):
    if len(deps) == 1:
        return deps[0][PyInfo]
    uses_shared_libraries = False
    has_py2_only_sources = False
    has_py3_only_sources = False
    for dep in deps:
        info = dep[PyInfo]
        if info.uses_shared_libraries:
            uses_shared_libraries = True
        if info.has_py2_only_sources:
            has_py2_only_sources = True
        if info.has_py3_only_sources:
            has_py3_only_sources = True
    return PyInfo(
        has_py2_only_sources = has_py2_only_sources,
        has_py3_only_sources = has_py3_only_sources,
        imports = depset(transitive = [
            dep[PyInfo].imports
            for dep in deps
        ]),
        transitive_sources = depset(transitive = [
            dep[PyInfo].transitive_sources
            for dep in deps
        ]),
        uses_shared_libraries = uses_shared_libraries,
    )

def _merge_runfiles(deps):
    return deps[0][DefaultInfo].default_runfiles.merge_all(
        [dep[DefaultInfo].default_runfiles for dep in deps[1:]],
    )

def _conda_deps_impl(ctx):
    symlinks = SymlinkInfo(
        symlinks = merge_deps_symlinks(ctx.attr.deps),
    )
    if not ctx.attr.deps:
        return [
            DefaultInfo(),
            CondaDepsInfo(
                hdrs = depset(),
                data = depset(),
                libs = depset(),
                solibs = depset(),
            ),
            symlinks,
            ExecutableFileInfo(files = depset()),
        ]
    transitive = CondaDepsInfo(
        hdrs = depset(transitive = [dep[CondaDepsInfo].hdrs for dep in ctx.attr.deps]),
        data = depset(transitive = [dep[CondaDepsInfo].data for dep in ctx.attr.deps]),
        libs = depset(transitive = [dep[CondaDepsInfo].libs for dep in ctx.attr.deps]),
        solibs = depset(transitive = [dep[CondaDepsInfo].solibs for dep in ctx.attr.deps]),
    )
    if len(ctx.attr.deps) == 1:
        info = ctx.attr.deps[0][DefaultInfo]
        result = [
            DefaultInfo(
                files = info.files,
                runfiles = info.default_runfiles,
            ),
            transitive,
            symlinks,
            ctx.attr.deps[0][ExecutableFileInfo],
        ]
    else:
        result = [
            DefaultInfo(
                files = depset(transitive = [
                    dep[DefaultInfo].files
                    for dep in ctx.attr.deps
                ]),
                runfiles = _merge_runfiles(
                    ctx.attr.deps,
                ),
            ),
            transitive,
            symlinks,
            ExecutableFileInfo(
                files = depset(transitive = [
                    dep[ExecutableFileInfo].files
                    for dep in ctx.attr.deps
                ]),
            ),
        ]
    for dep in ctx.attr.deps:
        if CcInfo in dep:
            result.append(dep[CcInfo])
            break
    py = [dep for dep in ctx.attr.deps if PyInfo in dep]
    if py:
        result.append(_merge_pyinfo(py))
    return result

conda_deps = rule(
    attrs = {
        "deps": attr.label_list(
            doc = "The conda packages (install targets) which this package depends on.",
            providers = [
                [CondaDepsInfo, SymlinkInfo, ExecutableFileInfo],
                [CondaDepsInfo, SymlinkInfo, ExecutableFileInfo, CcInfo],
                [CondaDepsInfo, SymlinkInfo, ExecutableFileInfo, PyInfo],
            ],
        ),
    },
    doc = """A set of conda package dependencies.

Unlike a simple `filegroup`, this will forward `PyInfo` and `CcInfo`.""",
    implementation = _conda_deps_impl,
    provides = [CondaDepsInfo, SymlinkInfo, ExecutableFileInfo],
)

def _conda_files_impl(ctx):
    py_srcs = depset(transitive = [dep[DefaultInfo].files for dep in ctx.attr.py_srcs])
    lalibs = depset(transitive = [dep[DefaultInfo].files for dep in ctx.attr.lalibs])
    staticlibs = depset(transitive = [dep[DefaultInfo].files for dep in ctx.attr.staticlibs])
    dylibs = depset(transitive = [dep[DefaultInfo].files for dep in ctx.attr.dylibs])
    hdrs = depset(transitive = [dep[DefaultInfo].files for dep in ctx.attr.hdrs])
    hdrs_with_placeholders = depset(transitive = [
        dep[DefaultInfo].files
        for dep in ctx.attr.hdrs_with_placeholders
    ])
    runfiles = depset(transitive = [dep[DefaultInfo].files for dep in ctx.attr.runfiles])
    link_safe_runfiles = depset(transitive = [dep[DefaultInfo].files for dep in ctx.attr.link_safe_runfiles])
    return [
        DefaultInfo(files = depset(transitive = [
            py_srcs,
            lalibs,
            staticlibs,
            dylibs,
            hdrs,
            hdrs_with_placeholders,
            runfiles,
        ])),
        CondaFilesInfo(
            py_srcs = py_srcs,
            lalibs = lalibs,
            staticlibs = staticlibs,
            dylibs = dylibs,
            hdrs = hdrs,
            hdrs_with_placeholders = hdrs_with_placeholders,
            runfiles = runfiles,
            link_safe_runfiles = link_safe_runfiles,
        ),
    ]

conda_files = rule(
    attrs = {
        "py_srcs": attr.label_list(
            doc = "Python source files.",
            allow_files = [".py"],
        ),
        "lalibs": attr.label_list(
            doc = "libtool library files",
            allow_files = [".la"],
        ),
        "staticlibs": attr.label_list(
            doc = "Static libraries.",
            allow_files = [".a", ".lib"],
        ),
        "dylibs": attr.label_list(
            doc = "Dynamic libraries.",
            allow_files = True,
        ),
        "hdrs": attr.label_list(
            doc = "C/C++ header files.",
            allow_files = [
                ".c",
                ".C",
                ".c++",
                ".cc",
                ".cpp",
                ".cxx",
                ".h",
                ".hh",
                ".H",
                ".hpp",
                ".hxx",
                ".inc",
                ".ipp",
            ],
        ),
        "hdrs_with_placeholders": attr.label_list(
            doc = "C/C++ header files which require modification during install.",
            allow_files = [
                ".c",
                ".C",
                ".c++",
                ".cc",
                ".cpp",
                ".cxx",
                ".h",
                ".hh",
                ".H",
                ".hpp",
                ".hxx",
                ".inc",
                ".ipp",
            ],
        ),
        "runfiles": attr.label_list(
            doc = "All other files which are not known to be " +
                  "safe to symlink into place.",
            allow_files = True,
        ),
        "link_safe_runfiles": attr.label_list(
            doc = "All other files which are safe to symlink into place. " +
                  "This is a build performance optimization. " +
                  "When in doubt, put it in runfiles.",
            allow_files = True,
        ),
    },
    implementation = _conda_files_impl,
    doc = """Lists the files involved in the package.

These files are separated into sets based on how they are used and how they are
copied over to the environment output.
The uses are either as python sources (in `PyInfo`), C headers and libraries
(in `CcInfo`), or runfiles.
The copy mechanism is either symlink, actual copy (needed in cases where things
might otherwise escape the output tree by evaluating a symlink), or copy with
modification, generally to replace a placeholder value.
""",
    provides = [CondaFilesInfo],
)
