package detect

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/samber/lo"
)

func DetectJSPacakgeManager() (string, error) {

	var LOCKFILES = [7][2]string{
		{"deno.lock", "deno"},
		{"deno.json", "deno"},
		{"deno.jsonc", "deno"},
		{"bun.lockb", "bun"},
		{"pnpm-lock.yaml", "pnpm"},
		{"yarn.lock", "yarn"},
		{"package-lock.json", "npm"},
	}

	cwd, err := os.Getwd()

	if err != nil {
		return "", err
	}

	// Check for lock files and config files in order of preference
	for _, lockFileAndPakageName := range LOCKFILES {

		lockFile := lockFileAndPakageName[0]
		packageName := lockFileAndPakageName[1]

		if _, err := os.Stat(filepath.Join(cwd, lockFile)); err == nil {

			log.Infof("Found lock file %s", lockFile)

			return packageName, nil

		}
	}

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

		return error == nil

	})

	if !ok {

		return "", fmt.Errorf(
			"You don't have one of the suppoted package managers installed: %s",
			strings.Join(supportedOperatingSystemPackageManagers, " , "),
		)
	}

	return detectedPackageManager, nil

}
