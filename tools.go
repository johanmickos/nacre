// Package for defining build tools used in this project
// that are not critical for actually running the service.

//go:build tools
// +build tools

package tools

import (
	_ "golang.org/x/lint/golint"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
