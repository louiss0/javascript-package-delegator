// Package deps provides functionality for dependency management and detection
// across different JavaScript package managers and runtime environments.
package deps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DepsHashFile is the filename for storing the computed dependency hash
const DepsHashFile = ".jpd-deps-hash"

// DenoDepsHashFile stores the computed Deno import hash alongside deno.json.
const DenoDepsHashFile = ".jpd-deno-deps-hash"

// ReadStoredDepsHash reads the stored dependency hash from the node_modules directory.
// If the storage directory does not exist, ErrHashStorageUnavailable is returned.
// If the file itself does not exist, returns empty string and nil error (not an error condition).
func ReadStoredDepsHash(cwd string) (string, error) {
	nodeModulesPath := filepath.Join(cwd, "node_modules")

	if _, err := os.Stat(nodeModulesPath); err != nil {
		if os.IsNotExist(err) {
			return "", ErrHashStorageUnavailable
		}
		return "", err
	}

	hashFilePath := filepath.Join(nodeModulesPath, DepsHashFile)

	data, err := os.ReadFile(hashFilePath)
	if os.IsNotExist(err) {
		// File not existing is not an error - just means no hash stored yet
		return "", nil
	}
	if err != nil {
		return "", err
	}

	// Trim whitespace and return the hash
	hash := strings.TrimSpace(string(data))
	return hash, nil
}

// WriteStoredDepsHash writes the dependency hash to the node_modules directory.
// Creates or overwrites the hash file with the provided hash value.
// Requires that the node_modules directory already exists.
func WriteStoredDepsHash(cwd, hash string) error {
	nodeModulesPath := filepath.Join(cwd, "node_modules")

	// Check if node_modules directory exists
	if _, err := os.Stat(nodeModulesPath); err != nil {
		if os.IsNotExist(err) {
			return ErrHashStorageUnavailable
		}
		return fmt.Errorf("failed to stat node_modules: %w", err)
	}

	hashFilePath := filepath.Join(nodeModulesPath, DepsHashFile)

	// Write hash with trailing newline for better file handling
	content := hash + "\n"

	return os.WriteFile(hashFilePath, []byte(content), 0644)
}

// ReadStoredDenoDepsHash loads the stored Deno imports hash from the project root.
// If the file does not exist, an empty string is returned.
func ReadStoredDenoDepsHash(cwd string) (string, error) {
	hashFilePath := filepath.Join(cwd, DenoDepsHashFile)

	data, err := os.ReadFile(hashFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// WriteStoredDenoDepsHash persists the Deno imports hash in the project root.
func WriteStoredDenoDepsHash(cwd, hash string) error {
	hashFilePath := filepath.Join(cwd, DenoDepsHashFile)
	content := hash + "\n"
	return os.WriteFile(hashFilePath, []byte(content), 0644)
}
