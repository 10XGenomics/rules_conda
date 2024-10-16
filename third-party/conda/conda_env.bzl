# Code generated by make_conda_spec.  DO NOT EDIT.
"""This file contains the workspace rules for the conda environment.

To use, add
```
    load(":conda_env.bzl", "conda_environment")
    conda_environment()
```
to your `WORKSPACE` file.

This file contains the workspace rules for the conda environment.

To use, add
```
    load(":conda_spec.bzl", "conda_environment")
    conda_environment()
```
to your `WORKSPACE` file.
"""

load(
    "@com_github_10XGenomics_rules_conda//rules:conda_environment.bzl",
    "conda_environment_repository",
)
load(
    "@com_github_10XGenomics_rules_conda//rules:conda_package_repository.bzl",
    "conda_package_repository",
)

def conda_environment(name = "conda_env"):
    """Create remote repositories to download each conda package.

    Also create the repository rule to generate the complete distribution.  In
    general, other targets should depend on targets in the `@conda_env`
    repository, rather than individual package repositories.

    Args:
        name (string): The name of the top level distribution repo.
    """

    conda_package_repository(
        name = "conda_package__libgcc_mutex",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "_libgcc_mutex-0.1-conda_forge",
        sha256 = "fe51de6107f9edc7aa4f786a70f4a883943bc9d39b3bb7307c04c41410990726",
        conda_repo = name,
    )

    conda_package_repository(
        name = "conda_package_black",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "black-24.10.0-py312h7900ff3_0",
        sha256 = "2b4344d18328b3e8fd9b5356f4ee15556779766db8cb21ecf2ff818809773df6",
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_bzip2",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "bzip2-1.0.8-hd590300_5",
        sha256 = "242c0c324507ee172c0e0dd2045814e746bb303d1eb78870d182ceb0abc726a8",
        exclude = [
            "man",
        ],
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_ca_certificates",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "ca-certificates-2024.8.30-hbcca054_0",
        sha256 = "afee721baa6d988e27fef1832f68d6f32ac8cc99cdf6015732224c2841a09cea",
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_click",
        base_urls = ["https://conda.anaconda.org/conda-forge/noarch"],
        dist_name = "click-8.1.7-unix_pyh707e725_0",
        sha256 = "f0016cbab6ac4138a429e28dbcb904a90305b34b3fe41a9b89d697c90401caec",
        archive_type = "conda",
        conda_repo = name,
    )

    conda_package_repository(
        name = "conda_package_coverage",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "coverage-7.5.4-py312h9a8786e_0",
        sha256 = "2902fb27f4a6b16512264973b13fea7ddb81811ba8599906a737d3cc1ee24db0",
        archive_type = "conda",
        conda_repo = name,
    )

    conda_package_repository(
        name = "conda_package_eigen",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "eigen-3.4.0-h00ab1b0_0",
        sha256 = "53b15a98aadbe0704479bacaf7a5618fcb32d1577be320630674574241639b34",
        archive_type = "conda",
        cc_include_path = ["include/eigen3"],
        conda_repo = name,
    )

    conda_package_repository(
        name = "conda_package_libexpat",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "libexpat-2.6.2-h59595ed_0",
        sha256 = "331bb7c7c05025343ebd79f86ae612b9e1e74d2687b8f3179faec234f986ce19",
        archive_type = "conda",
        conda_repo = name,
    )

    conda_package_repository(
        name = "conda_package_libffi",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "libffi-3.4.2-h7f98852_5",
        sha256 = "ab6e9856c21709b7b517e940ae7028ae0737546122f83c2aa5d692860c3b149e",
        licenses = ["@rules_license//licenses/spdx:MIT"],
        exclude = [
            "share/man",
            "share/info",
        ],
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_libgcc",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "libgcc-14.1.0-h77fa898_1",
        sha256 = "10fa74b69266a2be7b96db881e18fa62cfa03082b65231e8d652e897c4b335a3",
        archive_type = "conda",
        license_file = "share/licenses/gcc-libs/RUNTIME.LIBRARY.EXCEPTION",
        exclude_deps = ["_openmp_mutex"],
        conda_repo = name,
    )

    conda_package_repository(
        name = "conda_package_libgcc_ng",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "libgcc-ng-14.1.0-h69a702a_1",
        sha256 = "b91f7021e14c3d5c840fbf0dc75370d6e1f7c7ff4482220940eaafb9c64613b7",
        exclude = [
            "lib/libgomp.so*",
            "lib/lib*san.so*",
            "share/info",
            "x86_64-conda_cos6-linux-gnu/sysroot/lib/libgomp.so*",
        ],
        exclude_deps = [
            "_openmp_mutex",
        ],
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_libnsl",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "libnsl-2.0.1-hd590300_0",
        sha256 = "26d77a3bb4dceeedc2a41bd688564fe71bf2d149fdcf117049970bc02ff1add6",
        archive_type = "conda",
        conda_repo = name,
    )

    conda_package_repository(
        name = "conda_package_libsqlite",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "libsqlite-3.46.0-hde9e2c9_0",
        sha256 = "daee3f68786231dad457d0dfde3f7f1f9a7f2018adabdbb864226775101341a8",
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_libstdcxx",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "libstdcxx-14.1.0-hc0a3c3a_1",
        sha256 = "44decb3d23abacf1c6dd59f3c152a7101b7ca565b4ef8872804ceaedcc53a9cd",
        archive_type = "conda",
        license_file = "share/licenses/libstdc++/RUNTIME.LIBRARY.EXCEPTION",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_libstdcxx_ng",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "libstdcxx-ng-14.1.0-h4852527_1",
        sha256 = "a2dc44f97290740cc187bfe94ce543e6eb3c2ea8964d99f189a1d8c97b419b8c",
        archive_type = "conda",
        conda_repo = name,
    )

    conda_package_repository(
        name = "conda_package_libuuid",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "libuuid-2.38.1-h0b41bf4_0",
        sha256 = "787eb542f055a2b3de553614b25f09eefb0a0931b0c87dbcce6efdfd92f04f18",
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_libxcrypt",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "libxcrypt-4.4.36-hd590300_1",
        sha256 = "6ae68e0b86423ef188196fff6207ed0c8195dd84273cb5623b85aa08033a410c",
        archive_type = "conda",
        conda_repo = name,
    )

    conda_package_repository(
        name = "conda_package_libzlib",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "libzlib-1.3.1-h4ab18f5_1",
        sha256 = "adf6096f98b537a11ae3729eaa642b0811478f0ea0402ca67b5108fe2cb0010d",
        archive_type = "conda",
        conda_repo = name,
    )

    conda_package_repository(
        name = "conda_package_mypy_extensions",
        base_urls = ["https://conda.anaconda.org/conda-forge/noarch"],
        dist_name = "mypy_extensions-1.0.0-pyha770c72_0",
        sha256 = "f240217476e148e825420c6bc3a0c0efb08c0718b7042fae960400c02af858a3",
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_ncurses",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "ncurses-6.5-h59595ed_0",
        sha256 = "4fc3b384f4072b68853a0013ea83bdfd3d66b0126e2238e1d6e1560747aa7586",
        licenses = ["@rules_license//licenses/spdx:MIT-open-group"],
        exclude = [
            "share/terminfo/[A-Z]",
            "share/terminfo/d/darwin*",
            "share/terminfo/n/n7900",
            "share/terminfo/n/ncr*",
        ],
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_openssl",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "openssl-3.3.1-h4ab18f5_1",
        sha256 = "ff3faf8d4c1c9aa4bd3263b596a68fcc6ac910297f354b2ce28718a3509db6d9",
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_packaging",
        base_urls = ["https://conda.anaconda.org/conda-forge/noarch"],
        dist_name = "packaging-24.1-pyhd8ed1ab_0",
        sha256 = "36aca948219e2c9fdd6d80728bcc657519e02f06c2703d8db3446aec67f51d81",
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_pathspec",
        base_urls = ["https://conda.anaconda.org/conda-forge/noarch"],
        dist_name = "pathspec-0.12.1-pyhd8ed1ab_0",
        sha256 = "4e534e66bfe8b1e035d2169d0e5b185450546b17e36764272863e22e0370be4d",
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_pip",
        base_urls = ["https://conda.anaconda.org/conda-forge/noarch"],
        dist_name = "pip-24.2-pyh8b19718_1",
        sha256 = "d820e5358bcb117fa6286e55d4550c60b0332443df62121df839eab2d11c890b",
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_platformdirs",
        base_urls = ["https://conda.anaconda.org/conda-forge/noarch"],
        dist_name = "platformdirs-4.3.6-pyhd8ed1ab_0",
        sha256 = "c81bdeadc4adcda216b2c7b373f0335f5c78cc480d1d55d10f21823590d7e46f",
        archive_type = "conda",
        conda_repo = name,
    )

    conda_package_repository(
        name = "conda_package_python",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "python-3.12.4-h194c7f8_0_cpython",
        sha256 = "97a78631e6c928bf7ad78d52f7f070fcf3bd37619fa48dc4394c21cf3058cdee",
        exclude_deps = [
            "ld_impl_linux-64",
            "readline",
        ],
        archive_type = "conda",
        exclude = [
            "lib/python*/distutils/tests/__pycache__",
            "lib/python*/idlelib/idle_test/__pycache__",
            "lib/python*/lib2to3/fixes/__pycache__",
            "lib/python*/lib2to3/tests",
            "share/man",
        ],
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_python_abi",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "python_abi-3.12-5_cp312",
        sha256 = "d10e93d759931ffb6372b45d65ff34d95c6000c61a07e298d162a3bc2accebb0",
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_setuptools",
        base_urls = ["https://conda.anaconda.org/conda-forge/noarch"],
        dist_name = "setuptools-75.1.0-pyhd8ed1ab_0",
        sha256 = "6725235722095c547edd24275053c615158d6163f396550840aebd6e209e4738",
        exclude = [
            "lib/python*/site-packages/setuptools/command/launcher manifest.xml",
            "lib/python*/site-packages/setuptools/script (dev).tmpl",
            "lib/python*/site-packages/setuptools/*.exe",
        ],
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_shellcheck",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "shellcheck-0.10.0-ha770c72_0",
        sha256 = "6809031184c07280dcbaed58e15020317226a3ed234b99cb1bd98384ea5be813",
        archive_type = "conda",
        conda_repo = name,
    )

    conda_package_repository(
        name = "conda_package_tk",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "tk-8.6.13-noxft_h4845f30_101",
        sha256 = "e0569c9caa68bf476bead1bed3d79650bb080b532c64a4af7d8ca286c08dea4e",
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_tomli",
        base_urls = ["https://conda.anaconda.org/conda-forge/noarch"],
        dist_name = "tomli-2.0.2-pyhd8ed1ab_0",
        sha256 = "5e742ba856168b606ac3c814d247657b1c33b8042371f1a08000bdc5075bc0cc",
        conda_repo = name,
        archive_type = "conda",
    )

    conda_package_repository(
        name = "conda_package_tzdata",
        base_urls = ["https://conda.anaconda.org/conda-forge/noarch"],
        dist_name = "tzdata-2024b-hc8b5060_0",
        sha256 = "4fde5c3008bf5d2db82f2b50204464314cc3c91c1d953652f7bd01d9e52aefdf",
        archive_type = "conda",
        conda_repo = name,
    )
    conda_package_repository(
        name = "conda_package_wheel",
        base_urls = ["https://conda.anaconda.org/conda-forge/noarch"],
        dist_name = "wheel-0.44.0-pyhd8ed1ab_0",
        sha256 = "d828764736babb4322b8102094de38074dedfc71f5ff405c9dfee89191c14ebc",
        archive_type = "conda",
        conda_repo = name,
    )

    conda_package_repository(
        name = "conda_package_xz",
        base_urls = ["https://conda.anaconda.org/conda-forge/linux-64"],
        dist_name = "xz-5.2.6-h166bdaf_0",
        sha256 = "03a6d28ded42af8a347345f82f3eebdd6807a08526d47899a42d62d319609162",
        exclude = [
            "share",
        ],
        conda_repo = name,
    )

    conda_environment_repository(
        name = name,
        conda_packages = [
            "_libgcc_mutex",
            "black",
            "bzip2",
            "ca-certificates",
            "click",
            "coverage",
            "eigen",
            "libexpat",
            "libffi",
            "libgcc",
            "libgcc-ng",
            "libnsl",
            "libsqlite",
            "libstdcxx",
            "libstdcxx-ng",
            "libuuid",
            "libxcrypt",
            "libzlib",
            "mypy_extensions",
            "ncurses",
            "openssl",
            "packaging",
            "pathspec",
            "pip",
            "platformdirs",
            "python",
            "python_abi",
            "setuptools",
            "shellcheck",
            "tk",
            "tomli",
            "tzdata",
            "wheel",
            "xz",
        ],
        executable_packages = [
            "conda",
            "coverage",
            "python",
            "shellcheck",
        ],
        aliases = {
            "importlib_metadata": "importlib-metadata",
            "typing-extensions": "typing_extensions",
        },
        py_version = 3,
    )
    native.register_toolchains("@{}//:python_toolchain".format(name))
