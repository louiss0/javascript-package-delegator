// Package deps provides functionality for dependency management and detection
// across different JavaScript package managers and runtime environments.
package deps

import "errors"

// ErrHashStorageUnavailable indicates that the dependency hash cannot be persisted
// for the given project (for example, when node_modules is absent as in Deno projects).
var ErrHashStorageUnavailable = errors.New("dependency hash storage unavailable")

// ErrDenoConfigNotFound indicates that no deno.json or deno.jsonc configuration file
// exists for the project in the requested directory.
var ErrDenoConfigNotFound = errors.New("deno config not found")
