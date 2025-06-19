package detect

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
)

var SupportedJSPackageManagers = [5]string{"deno", "bun", "npm", "pnpm", "yarn"}

func DetectJSPacakgeManager() (string, error) {

	var LOCKFILES = [7][2]string{
		{"deno.lock", SupportedJSPackageManagers[0]},
		{"deno.json", SupportedJSPackageManagers[0]},
		{"deno.jsonc", SupportedJSPackageManagers[0]},
		{"bun.lockb", SupportedJSPackageManagers[1]},
		{"pnpm-lock.yaml", SupportedJSPackageManagers[3]},
		{"yarn.lock", SupportedJSPackageManagers[4]},
		{"package-lock.json", SupportedJSPackageManagers[2]},
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

			return packageName, nil

		}
	}

	return "npm", nil

}

var SupportedOperatingSystemPackageManagers = [5]string{
	"winget",
	"nix",
	"scoop",
	"choco",
	"brew",
}

// Detects one of the packages supported by this library
func SupportedOperatingSystemPackageManager() (string, error) {

	detectedPackageManager, ok := lo.Find(SupportedOperatingSystemPackageManagers[:], func(path string) bool {

		_, error := exec.LookPath(path)

		return error == nil

	})

	if !ok {

		return "", fmt.Errorf(
			"You don't have one of the suppoted package managers installed: %s",
			strings.Join(SupportedOperatingSystemPackageManagers[:], " , "),
		)
	}

	return detectedPackageManager, nil

}
