// Package deps provides functionality for dependency management and detection
// across different JavaScript package managers and runtime environments.
package deps

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/samber/lo"
)

// ComputeNodeDepsHash computes a SHA256 hash of Node.js dependencies from package.json.
// It reads both dependencies and devDependencies, sorts them by key for deterministic
// ordering, and creates a hash from the "name@version" representation.
func ComputeNodeDepsHash(cwd string) (string, error) {
	type PackageJSONDependencies struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	packageJSONPath := filepath.Join(cwd, "package.json")
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return "", fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg PackageJSONDependencies
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", fmt.Errorf("failed to parse package.json: %w", err)
	}

	// Merge dependencies and devDependencies
	allDeps := lo.Assign(pkg.Dependencies, pkg.DevDependencies)

	// Convert to sorted slice for deterministic hashing
	depEntries := lo.Entries(allDeps)
	sort.Slice(depEntries, func(i, j int) bool {
		return depEntries[i].Key < depEntries[j].Key
	})

	// Create deterministic string representation
	depLines := lo.Map(depEntries, func(entry lo.Entry[string, string], index int) string {
		return fmt.Sprintf("%s@%s", entry.Key, entry.Value)
	})

	content := strings.Join(depLines, "\n")
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash), nil
}

// ComputeDenoImportsHash computes a SHA256 hash of Deno imports from deno.json or deno.jsonc.
// It prioritizes deno.json over deno.jsonc, normalizes JSONC if needed, sorts imports by key
// for deterministic ordering, and creates a hash from the "key=value" representation.
func ComputeDenoImportsHash(cwd string) (string, error) {
	type DenoJSONDependencies struct {
		Imports map[string]string `json:"imports"`
	}

	// Try deno.json first, then deno.jsonc
	var denoFilePath string
	denoJSONPath := filepath.Join(cwd, "deno.json")
	denoJSONCPath := filepath.Join(cwd, "deno.jsonc")

	if _, err := os.Stat(denoJSONPath); err == nil {
		denoFilePath = denoJSONPath
	} else if _, err := os.Stat(denoJSONCPath); err == nil {
		denoFilePath = denoJSONCPath
	} else {
		return "", fmt.Errorf("%w: %s", ErrDenoConfigNotFound, cwd)
	}

	data, err := os.ReadFile(denoFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", denoFilePath, err)
	}

	// If this is a deno.jsonc file, normalize it to JSON
	if filepath.Ext(denoFilePath) == ".jsonc" {
		data = NormalizeJSONCToJSON(data)
	}

	var pkg DenoJSONDependencies
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", fmt.Errorf("failed to parse %s: %w", denoFilePath, err)
	}

	// Convert to sorted slice for deterministic hashing
	importEntries := lo.Entries(pkg.Imports)
	sort.Slice(importEntries, func(i, j int) bool {
		return importEntries[i].Key < importEntries[j].Key
	})

	// Create deterministic string representation
	importLines := lo.Map(importEntries, func(entry lo.Entry[string, string], index int) string {
		return fmt.Sprintf("%s=%s", entry.Key, entry.Value)
	})

	content := strings.Join(importLines, "\n")
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash), nil
}
