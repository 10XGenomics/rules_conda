#!/usr/bin/env bash

set -e

GO="$1"
shift
conda_package_repo=$(dirname -- "$1")
REPO=$(pwd -P)

export GOBIN="${REPO}"
export CGO_ENABLED=0

DIR=$(dirname "$conda_package_repo")
cd -- "$DIR"
conda_package_repo=$(basename "${conda_package_repo}")
exec "${GO}" install -buildvcs=false -trimpath -ldflags="-s -w" \
    "./${conda_package_repo}"
