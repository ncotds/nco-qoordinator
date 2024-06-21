//go:build tools
// +build tools

package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/michurin/human-readable-json-logging/cmd/pplog"
	_ "github.com/vektra/mockery/v2"
	_ "golang.org/x/tools/cmd/goimports"
	_ "gotest.tools/gotestsum"
)
