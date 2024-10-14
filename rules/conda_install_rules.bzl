"""Rules to install conda packages.

These rules are intended to be used by the conda_environment rule. Users
should probably not be instantiating them directly.
"""

load("@bazel_skylib//lib:paths.bzl", "paths")
load(
    "//rules:symlink_provider.bzl",
    "ExecutableFileInfo",
    "SymlinkInfo",
    "merge_deps_symlinks",
)
load("//rules:util.bzl", "merge_pyinfo", "merge_runfiles")
load(":providers.bzl", "CondaDepsInfo", "CondaFilesInfo", "CondaManifestInfo")

def _in_path(info):
    fn = info.index or info.manifest
    i = fn.path.rfind("/")
    if i > 5 and fn.path[:i].endswith("/info"):
        return fn.path[:i - 5]
    fail("could not find input path.", attr = "manifest")

def _make_py_info(py, data, solibs, deps):
    return merge_pyinfo(
        has_py2_only_sources = py == "PY2",
        has_py3_only_sources = py == "PY3",
        sources = data,
        uses_shared_libraries = True if solibs else False,
        deps = [dep for dep in deps if PyInfo in dep],
    )

def _make_cc_info(ctx, manifest, hdrs, libs, solibs):
    """Generates CcInfo provider."""
    cc_deps = [dep[CcInfo] for dep in ctx.attr.deps if CcInfo in dep]
    if not hdrs and not libs and not solibs and not manifest.includes:
        if cc_deps:
            # Pass along the first dependency only.  Usually this case
            # comes up when we've got e.g. the :python target depending
            # on :python_tarball and a bunch of other stuff.  The correctness
            # of this rule depends on the build generator putting the tarball
            # as the first dependency, but since the generator lives in this
            # repository that should be safe.
            return cc_deps[0]
        return None

    compilation_context = None
    if hdrs or manifest.includes:
        inc = ["external/" + ctx.label.workspace_name + "/" + i for i in manifest.includes]
        inc = depset(inc + [
            ctx.genfiles_dir.path + "/" + d
            for d in inc
        ])
        compilation_context = cc_common.create_compilation_context(
            headers = depset(hdrs),
            system_includes = inc,
            quote_includes = inc,
        )
    libraries_to_link = None
    if libs or solibs:
        cc_toolchain = ctx.attr._cc_toolchain[cc_common.CcToolchainInfo]
        feature_configuration = cc_common.configure_features(
            ctx = ctx,
            cc_toolchain = cc_toolchain,
            requested_features = ["supports_pic"],
        )
        if libs:
            libraries_to_link = [
                cc_common.create_library_to_link(
                    actions = ctx.actions,
                    feature_configuration = feature_configuration,
                    cc_toolchain = cc_toolchain,
                    pic_static_library = lib,
                )
                for lib in libs
            ]
        elif solibs:
            libraries_to_link = [
                cc_common.create_library_to_link(
                    actions = ctx.actions,
                    feature_configuration = feature_configuration,
                    cc_toolchain = cc_toolchain,
                    dynamic_library = lib,
                )
                for lib in solibs
            ]
    linking_context = cc_common.create_linking_context(
        linker_inputs = depset([
            cc_common.create_linker_input(
                owner = ctx.label,
                libraries = depset(libraries_to_link),
                user_link_flags = depset(["-Lexternal/{}/lib".format(
                    ctx.label.repo_name
                )]),
            ),
        ]),
    ) if libraries_to_link else None
    cc = CcInfo(
        compilation_context = compilation_context,
        linking_context = linking_context,
    )

    return cc

def _make_symlinks(ctx, symlinks, target_path):
    if not symlinks:
        return [], merge_deps_symlinks(ctx.attr.deps)
    result = []
    result_map = {}
    for dep in ctx.attr.deps:
        result_map.update(dep[SymlinkInfo].symlinks.items())

    for link, target in symlinks.items():
        f = ctx.actions.declare_symlink(paths.join(target_path, link))
        ctx.actions.symlink(output = f, target_path = target)
        result.append(f)
        result_map[f] = target
    return result, result_map

def _find_exe(manifest, data):
    for f in data:
        if f.path.endswith(manifest.executable):
            return f
    fail("executable {} not found".format(manifest.executable), attr = "manifest.executable")

def _strip_root(fn, in_path):
    p = fn.path
    if fn.root.path:
        p = p[len(fn.root.path) + 1:]
    if p.startswith(in_path):
        p = p[len(in_path) + 1:]
    return p

def _out_path(fn, in_path, target_path):
    return paths.join(target_path, _strip_root(fn, in_path))

def _symlink_install(ctx, sources, in_path, target_path):
    """Install a set of files as symlinks.

    Uses `ctx.actions.symlink(output, target_file = src)`.

    Returns:
        list: The output file objects.
    """
    srcs = [
        ctx.actions.declare_file(_out_path(fn, in_path, target_path))
        for fn in sources
    ]
    for src, target in zip(sources, srcs):
        ctx.actions.symlink(output = target, target_file = src)
    return srcs

def _copy_files(ctx, in_files, executable_in, in_path, target_path, executable_out, manifest):
    files = []
    for fn in in_files:
        f = ctx.actions.declare_file(_out_path(fn, in_path, target_path))
        files.append(f)
        if _strip_root(fn, in_path) in executable_in:
            executable_out.append(f)
    if files:
        prefix = files[0].path
        prefix = prefix[:prefix.index(
            "/" + ctx.label.workspace_name + "/",
        ) + len(ctx.label.workspace_name) + 1]
        if target_path != ".":
            prefix = prefix + "/" + target_path
        args = ctx.actions.args()
        args.add("-install", in_path)
        args.add("-dest", prefix)
        args.add_joined("-roots", depset([
            fn.root.path
            for fn in in_files
            if fn.root.path
        ]), join_with = ",")
        file_list = ctx.actions.args()
        file_list.use_param_file("@%s")
        file_list.add_all(in_files)
        info_files = []
        if manifest.index:
            info_files.append(manifest.index)
        if manifest.manifest:
            info_files.append(manifest.manifest)
        ctx.actions.run(
            inputs = in_files + info_files,
            outputs = files,
            executable = ctx.executable._installer,
            arguments = [args, file_list],
            mnemonic = "CondaInstall",
            progress_message = "install {name}".format(name = ctx.attr.name),
            exec_group = "trivial",
        )
    return files

def _base_conda_package_impl(ctx, is_exe):
    """Implementation for conda_package and conda_exe rules.

    The only difference is `is_exe`.
    """
    manifest = ctx.attr.manifest[CondaManifestInfo]
    in_path = _in_path(manifest)
    path_prefix = manifest.python_prefix if manifest.noarch == "python" else ""
    target_path = path_prefix or "."
    executable_in = {f: None for f in manifest.executables}
    executable_files = []
    pkg_files = ctx.attr.srcs[CondaFilesInfo]
    py_srcs = _copy_files(
        ctx,
        pkg_files.py_srcs.to_list(),
        executable_in,
        in_path,
        target_path,
        executable_files,
        manifest,
    )
    lalibs = _copy_files(
        ctx,
        pkg_files.lalibs.to_list(),
        executable_in,
        in_path,
        target_path,
        executable_files,
        manifest,
    )
    staticlibs = _symlink_install(
        ctx,
        pkg_files.staticlibs.to_list(),
        in_path,
        target_path,
    )
    solibs = _symlink_install(
        ctx,
        pkg_files.dylibs.to_list(),
        in_path,
        target_path,
    )
    hdrs = _symlink_install(
        ctx,
        pkg_files.hdrs.to_list(),
        in_path,
        target_path,
    ) + _copy_files(
        ctx,
        pkg_files.hdrs_with_placeholders.to_list(),
        executable_in,
        in_path,
        target_path,
        executable_files,
        manifest,
    )
    run_data = _symlink_install(
        ctx,
        pkg_files.link_safe_runfiles.to_list(),
        in_path,
        target_path,
    ) + _copy_files(
        ctx,
        pkg_files.runfiles.to_list(),
        executable_in,
        in_path,
        target_path,
        executable_files,
        manifest,
    )
    symlinks, symlink_map = _make_symlinks(ctx, manifest.symlinks, target_path)
    if manifest.py_stubs:
        for stub in manifest.py_stubs.to_list():
            stub_dest = ctx.actions.declare_file("bin/" + stub.basename)
            ctx.actions.symlink(
                output = stub_dest,
                target_file = stub,
                is_executable = True,
            )
            run_data.append(stub_dest)
            executable_files.append(stub_dest)
    data = py_srcs + run_data
    out_files = data + lalibs + staticlibs + solibs + hdrs + symlinks
    runfiles = merge_runfiles(
        ctx,
        ctx.attr.deps,
        files = data + solibs + [
            link
            for link in symlinks
            if not link.basename.endswith(".a") and not link.basename.endswith(".h")
        ],
        transitive_files = False,
    )
    exe_file = None
    extra_groups = {lib.basename: [lib] for lib in staticlibs}
    if is_exe:
        exe_file = _find_exe(manifest, out_files)
        extra_groups["exe_file"] = [exe_file]
    all_solibs = solibs + [
        link
        for link in symlinks
        if link.basename.endswith(".so") or ".so." in link.basename
    ]
    transitive_solibs = [dep[CondaDepsInfo].solibs for dep in ctx.attr.deps]
    all_transitive_solibs = depset(all_solibs, transitive = transitive_solibs)
    result = [
        DefaultInfo(
            executable = exe_file,
            files = depset(out_files, transitive = transitive_solibs),
            runfiles = runfiles,
        ),
        OutputGroupInfo(
            hdrs = hdrs,
            all_files = out_files,
            data = data,
            libs = staticlibs,
            solibs = all_solibs,
            **extra_groups
        ),
        CondaDepsInfo(
            hdrs = depset(hdrs, transitive = [dep[CondaDepsInfo].hdrs for dep in ctx.attr.deps]),
            data = depset(data, transitive = [dep[CondaDepsInfo].data for dep in ctx.attr.deps]),
            libs = depset(staticlibs, transitive = [dep[CondaDepsInfo].libs for dep in ctx.attr.deps]),
            solibs = all_transitive_solibs,
        ),
        SymlinkInfo(
            symlinks = symlink_map,
        ),
        ExecutableFileInfo(
            files = depset(executable_files, transitive = [
                dep[ExecutableFileInfo].files
                for dep in ctx.attr.deps
            ]),
        ),
    ]
    cc_info = _make_cc_info(ctx, manifest, hdrs, staticlibs, solibs)
    if cc_info:
        result.append(cc_info)
    if ctx.files.srcs or any([PyInfo in dep for dep in ctx.attr.deps]):
        result.append(_make_py_info(
            ctx.attr.py,
            data,
            all_solibs,
            ctx.attr.deps,
        ))

    return result

def _conda_package_impl(ctx):
    return _base_conda_package_impl(ctx, False)

def _conda_exe_impl(ctx):
    return _base_conda_package_impl(ctx, True)

_COMMON_ATTRS = {
    "manifest": attr.label(
        doc = "Target with conda metadata.",
        providers = [[CondaManifestInfo]],
        mandatory = True,
    ),
    "srcs": attr.label(
        doc = "File group for python sources.",
        providers = [[CondaFilesInfo]],
    ),
    "py": attr.string(
        doc = "PY2AND3, PY2 or PY3 if this package exposes python sources, " +
              "including transitively, depending on python version.",
        default = "",
    ),
    "deps": attr.label_list(
        doc = "Packages which this package depends on.",
        providers = [
            [CondaDepsInfo, SymlinkInfo, ExecutableFileInfo],
            [CondaDepsInfo, SymlinkInfo, ExecutableFileInfo, CcInfo],
            [CondaDepsInfo, SymlinkInfo, ExecutableFileInfo, PyInfo],
        ],
    ),
    "_installer": attr.label(
        default = Label("@com_github_10XGenomics_rules_conda//cmd/conda_pkg_install"),
        allow_single_file = True,
        executable = True,
        cfg = "exec",
    ),
    "_cc_toolchain": attr.label(
        default = Label("@bazel_tools//tools/cpp:current_cc_toolchain"),
    ),
}

conda_package = rule(
    attrs = _COMMON_ATTRS,
    doc = """Installs a conda package for the repository.

This rule is used by the `conda_environment_repository` workspace rule, and is not
intended to be instantiated by users.
""",
    exec_groups = {
        "trivial": exec_group(
            toolchains = ["@bazel_tools//tools/cpp:toolchain_type"],
        ),
    },
    fragments = ["cpp"],
    toolchains = ["@bazel_tools//tools/cpp:toolchain_type"],
    implementation = _conda_package_impl,
    provides = [
        OutputGroupInfo,
        CondaDepsInfo,
        SymlinkInfo,
        ExecutableFileInfo,
    ],
)

conda_exe = rule(
    attrs = _COMMON_ATTRS,
    doc = """Installs a conda package with an executable for the repository.

This rule is used by the `conda_environment_repository` workspace rule, and is not
intended to be instantiated by users.
""",
    executable = True,
    exec_groups = {
        "trivial": exec_group(
            toolchains = ["@bazel_tools//tools/cpp:toolchain_type"],
        ),
    },
    fragments = ["cpp"],
    toolchains = ["@bazel_tools//tools/cpp:toolchain_type"],
    implementation = _conda_exe_impl,
    provides = [
        OutputGroupInfo,
        CondaDepsInfo,
        SymlinkInfo,
        ExecutableFileInfo,
    ],
)
