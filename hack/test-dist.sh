#!/usr/bin/env bash

set -e -x -o pipefail

export KREW_ROOT="$(mktemp -d)"
trap "rm -rf $KREW_ROOT" EXIT

"$HOME/.krew/bin/kubectl-krew" install \
    --manifest _dist/open-svc.yaml \
    --archive _dist/kubectl-open_svc-linux-amd64.zip
