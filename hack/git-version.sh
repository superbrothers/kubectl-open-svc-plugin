#!/usr/bin/env bash

set -e -o pipefail; [[ -n "$DEBUG" ]] && set -x

# parse the current git commit hash
commit="$(git rev-parse HEAD)"

# check if the current commit has a matching tag
tag="$(git describe --exact-match --abbrev=0 --tags "${commit}" 2> /dev/null ||:)"

# use the matching tag as the version, if available
if [[ -z "$tag" ]]; then
  version="v0.0.0-${commit}"
else
  version="$tag"
fi

# check for changed files
if [[ -n "$(git diff --shortstat 2> /dev/null | tail -n1)" ]]; then
  version="${version}-dirty"
fi

echo "$version"
# vim: ai ts=2 sw=2 et sts=2 ft=sh
