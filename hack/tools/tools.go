// +build tools
package main

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "sigs.k8s.io/krew/cmd/validate-krew-manifest"
)
