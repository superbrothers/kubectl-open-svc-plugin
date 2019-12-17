#!/usr/bin/env bash

set -e -o pipefail; [[ -n "$DEBUG" ]] && set -x

dependencies=(
  "k8s.io/client-go@kubernetes-1.17.0" \
  "k8s.io/cli-runtime@kubernetes-1.17.0" \
  "k8s.io/apimachinery@kubernetes-1.17.0" \
  "k8s.io/kubectl@kubernetes-1.17.0" \
  "k8s.io/klog@v1.0.0" \
  "github.com/spf13/cobra@v0.0.5" \
  "github.com/pkg/browser@master" \
)

GO111MODULE=on go get "${dependencies[@]}"
GO111MODULE=on go mod tidy
