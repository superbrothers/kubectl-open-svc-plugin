#!/usr/bin/env bash

set -e -o pipefail; [[ -n "$DEBUG" ]] && set -x

VERSION="$(./hack/git-version.sh)"
DARWIN_SHA256="$(cat ./_dist/kubectl-open_svc-darwin-amd64.zip.sha256)"
LINUX_SHA256="$(cat ./_dist/kubectl-open_svc-linux-amd64.zip.sha256)"
WINDOWS_SHA256="$(cat ./_dist/kubectl-open_svc-windows-amd64.zip.sha256)"

sed -e "s/VERSION/${VERSION}/g" \
    -e "s/DARWIN_SHA256/${DARWIN_SHA256}/g" \
    -e "s/LINUX_SHA256/${LINUX_SHA256}/g" \
    -e "s/WINDOWS_SHA256/${WINDOWS_SHA256}/g" \
    ./hack/open-svc.yaml
