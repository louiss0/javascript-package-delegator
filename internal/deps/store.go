// Package deps provides functionality for dependency management and detection
// across different JavaScript package managers and runtime environments.
package deps

import (
	"os"
	"path/filepath"
	"strings"
)

// DepsHashFile is the filename for storing the computed dependency hash
const DepsHashFile = ".jpd-deps-hash"

// ReadStoredDepsHash reads the stored dependency hash from the project directory.
// If the file does not exist, returns empty string and nil error (not an error condition).
func ReadStoredDepsHash(cwd string) (string, error) {
	hashFilePath := filepath.Join(cwd, DepsHashFile)
	
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

// WriteStoredDepsHash writes the dependency hash to the project directory.
// Creates or overwrites the hash file with the provided hash value.
func WriteStoredDepsHash(cwd, hash string) error {
	hashFilePath := filepath.Join(cwd, DepsHashFile)
	
	// Write hash with trailing newline for better file handling
	content := hash + "\n"
	
	return os.WriteFile(hashFilePath, []byte(content), 0644)
}