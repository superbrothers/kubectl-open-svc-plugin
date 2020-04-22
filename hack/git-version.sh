#!/usr/bin/env bash

set -e -o pipefail; [[ -n "$DEBUG" ]] && set -x

# parse the current git commit hash
COMMIT=`git rev-parse HEAD`

# check if the current commit has a matching tag
TAG=$(git describe --exact-match --abbrev=0 --tags ${COMMIT} 2> /dev/null || true)

# use the matching tag as the version, if available
if [ -z "$TAG" ]; then
    VERSION="v0.0.0-$COMMIT"
else
    VERSION=$TAG
fi

echo $VERSION
