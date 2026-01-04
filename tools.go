//go:build tools
// +build tools

package tools

// This file declares dependencies on tools used during development.
// It ensures they are tracked in go.mod but not included in the final binary.

import (
	_ "github.com/air-verse/air"
)
