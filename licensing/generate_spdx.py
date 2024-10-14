#!/usr/bin/env python3

r"""SPDX license list generator.

Generates spdx.go from a list bazel license targets.

Add more locations to the query if desired.

Usage:
    generate_spdx.py | gofmt > spdx.go
"""

import argparse
import subprocess
import sys
from typing import IO


HEADER = """// Code generated by generate_spdx.py. DO NOT EDIT.

package licensing

var knownLicenses = map[string]string{"""

_THIS_REPO = "@com_github_10XGenomics_rules_conda"


def generate(stdin: IO[str], stdout: IO[str]):
    """Generate go code for the license mapping."""
    targets = sorted(stdin)
    print(HEADER, file=stdout)
    for target in targets:
        target = target.strip()
        if target.startswith("//"):
            target = "@com_github_10XGenomics_rules_conda" + target
        key = target[target.rfind(":") + 1 :]
        print(
            '\t"{key}": "{target}",'.format(key=key, target=target), file=stdout
        )
    print("}", file=stdout)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "-o",
        metavar="FILE",
        type=argparse.FileType("w", encoding="utf-8"),
        default=sys.stdout,
    )
    parser.add_argument(
        "--stdin",
        action="store_true",
        help="Take target list from standard input rather "
        "than runninig `bazel query`.",
    )
    parser.add_argument(
        "target",
        nargs="*",
        default=[
            "@rules_license//licenses/spdx:all",
            _THIS_REPO + "//licensing/known:all",
        ],
    )
    args = parser.parse_args()
    if args.stdin:
        generate(sys.stdin, args.o)
        return
    assert args.target
    with subprocess.Popen(
        [
            "bazel",
            "query",
            "--output=label",
            """
            let all_targets = set({}) in
            kind(license_kind, $all_targets) +
            kind(alias, $all_targets)
            """.format(
                " ".join(args.target)
            ),
        ],
        encoding="utf-8",
        stdout=subprocess.PIPE,
    ) as in_proc:
        assert in_proc.stdout
        generate(in_proc.stdout, args.o)


if __name__ == "__main__":
    main()