#!/usr/bin/env bash

set -e -o pipefail

version="$1"
if [[ -z "$version" ]]; then
  echo "Usage: $0 <version>" >&2
  exit 1
fi

set -x
echo "$version" > version.txt
git commit -a -m "Bump version to $version"
git tag "$version"
git --no-pager show "$version"
# vim: ai ts=2 sw=2 et sts=2 ft=sh
