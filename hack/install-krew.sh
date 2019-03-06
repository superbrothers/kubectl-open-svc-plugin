#!/usr/bin/env bash

set -e -o pipefail

KREW_VERSION="v0.2.1"
(
  set -x; cd "$(mktemp -d)" &&
  curl -fsSLO "https://storage.googleapis.com/krew/${KREW_VERSION}/krew.{tar.gz,yaml}" &&
  tar zxvf krew.tar.gz &&
  ./krew-"$(uname | tr '[:upper:]' '[:lower:]')_amd64" install \
    --manifest=krew.yaml --archive=krew.tar.gz
)
"$HOME/.krew/bin/kubectl-krew" version
