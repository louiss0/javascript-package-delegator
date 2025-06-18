package detect

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/samber/lo"
)

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
