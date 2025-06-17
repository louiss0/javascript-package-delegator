package detect

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
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

// Detects one of the packages supported by this library
func SupportedOperatingSystemPackageManager() (string, error) {

	supportedOperatingSystemPackageManagers := []string{
		"winget",
		"nix",
		"scoop",
		"choco",
		"brew",
	}

	detectedPackageManager, ok := lo.Find(supportedOperatingSystemPackageManagers, func(path string) bool {

		_, error := exec.LookPath(path)

		return error != nil

	})

	if !ok {

		return "", fmt.Errorf(
			"You don't have one of the suppoted package managers installed: %s",
			strings.Join(supportedOperatingSystemPackageManagers, " , "),
		)
	}

	return detectedPackageManager, nil

}
