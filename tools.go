//go:build tools

package tools

import (
	_ "github.com/atombender/go-jsonschema"
)

// This file ensures that go mod tidy doesn't remove tool dependencies
