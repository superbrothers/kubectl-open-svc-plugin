#!/usr/bin/env bash

set -e -o pipefail

KREW_VERSION="v0.3.4"
(
  set -x; cd "$(mktemp -d /tmp/krew-XXXXXXXXX)" &&
  curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/download/${KREW_VERSION}/krew.{tar.gz,yaml}" &&
  tar zxvf krew.tar.gz &&
  ./krew-"$(uname | tr '[:upper:]' '[:lower:]')_amd64" install \
    --manifest=krew.yaml --archive=krew.tar.gz
)
"$HOME/.krew/bin/kubectl-krew" version
