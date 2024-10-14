"""Providers for conda packages."""

CondaManifestInfo = provider(
    doc = "Provides information about the files in a conda package.",
    fields = {
        "manifest": "File: The package manifest file, one of `paths.json`, `files`, or `files.json`.",
        "symlinks": "dict of str to str: paths which are symlinks, " +
                    "and where they symlink to.",
        "executable": "str: path to the file, if any, to be used as the " +
                      "executable entry point for the package.",
        "executables": "list[str]: all paths which should have their " +
                       "executable bit set.",
        "includes": "List of string: include directories to be added to the " +
                    "compilation context for CcInfo.",
        "noarch": "str: The type of noarch linkage to use, if any.",
        "py_stubs": "depset[File]: generated python stub files.",
        "python_prefix": "str: prefix to prepend to installation directory, " +
                         "if it's a python noarch package.",
        "index": "File: the `index.json` file.",
    },
)

CondaDepsInfo = provider(
    doc = "Provides information about transitive dependencies of conda packages.",
    fields = {
        "hdrs": "depset[File]: transitive C/C++ header files.",
        "data": "depset[File]: transitive other files.",
        "libs": "depset[File]: transitive .a lib files.",
        "solibs": "depset[File]: transitive .so lib files.",
    },
)

CondaFilesInfo = provider(
    doc = """Provides the file depsets for a conda package.

Files are divided up based on essentially two axes.
First, where is this file used?

1. Python sources
2. C header files
3. C static libraries
4. C dynamic libraries.  A very frequent pattern with these is to have
   a `libfoo.so.1` and `libfoo.so` being symlinks to a `libfoo.so.1.2`.
   We don't want to include the symlinks in the dynamic linking context for
   the resulting `CcInfo`, however, as they'd of course be redundant.
   Some care must be taken in this case, however, because often the library
   will set `DT_SONAME=libfoo.so.1`.
   This means that we need to have `libfoo.so.1` in the linking context for the
   `CcInfo`, and we do _not_ want to include the other names,
   which means we need to fudge things a bit reverse the symlink so that
   `libfoo.so.1.2` is as symlink to `libfoo.so.1`.
5. Runfiles

When used as a python dependency, python sources, dynamic libraries,
and runfiles are required.
When used as a C dependency, headers and libraries are needed
(separately in CcInfo), and runfiles are assumed needed at runtime.

The second axis is how must the file be placed into the repo directory?

1. Files which can be symlinked in from the package repo
   (with ctx.actions.symlink).
   These are "cheap" because you don't need to spin up a sandbox just to copy a
   file, but can cause problems if something is resolving symlinks,
   as is often the case for python sources, because then instead of looking in
   external/conda_env/lib/python3.xx/site-packages it starts looking in
   external/conda_package_python/lib/python3.xx/site-packages, which will
   be basically empty.
   This category includes everything except
   - `.a` static libraries (not `.la`)
   - `.so` libraries, but not their symlink aliases.
2. Files which need to be actually copied.
   With remote execution, ctx.actions.symlink will behave effectively the same
   way, but in order to work correctly in local mode we have to actually copy
   the file.
3. Files which need to have strings substituted.
4. Files which are relative symlinks.  These are usually `.so` libraries, but
   are sometimes executables.

We don't actually make a distinction between 2 and 3, however, because the same
process can handle both cases.
""",
    fields = {
        "py_srcs": "depset[File]: Python sources to be copied.",
        "lalibs": "depset[File]: `.la` library files.  These will almost always have placeholders.",
        "staticlibs": "depset[File]: `.a` library files.",
        "dylibs": "depset[File]: `.so` library files.",
        "hdrs": "depset[File]: `.h` files without any placeholders to be replaced.",
        "hdrs_with_placeholders": "depset[File]: `.h` files with placeholders to be replaced.",
        "link_safe_runfiles": "depset[File]: Any other files matching " +
                              "patterns known to be safe to use symlinks with.",
        "runfiles": "depset[File]: Any other files.",
    },
)
