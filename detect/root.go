package detect

import (
	"os"
	"path/filepath"
)

// JSPackageManager detects the package manager based on lock files
func JSPackageManager() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	lockFiles := map[string]string{
		"deno.lock":         "deno",
		"deno.json":         "deno",
		"deno.jsonc":        "deno",
		"bun.lockb":         "bun",
		"pnpm-lock.yaml":    "pnpm",
		"yarn.lock":         "yarn",
		"package-lock.json": "npm",
	}

	// Check for lock files and config files in order of preference

	for lockFile, pm := range lockFiles {
		if _, err := os.Stat(filepath.Join(cwd, lockFile)); err == nil {
			return pm, nil
		}
	}

	// Default to npm if no lock file is found
	return "npm", nil
}
