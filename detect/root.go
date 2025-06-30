package detect

import (
	"errors" // Import the errors package
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const DENO = "deno"
const BUN = "bun"
const NPM = "npm"
const PNPM = "pnpm"
const YARN = "yarn"

var SupportedJSPackageManagers = [5]string{DENO, BUN, NPM, PNPM, YARN}

// ErrNoPackageManager is returned when no supported JavaScript package manager
// lock file or configuration file is found in the current directory.
var ErrNoPackageManager = errors.New("no supported JavaScript package manager found")

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
		// Pass through the original error from os.Getwd().
		// This error is already descriptive (e.g., *os.PathError).
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Check for lock files and config files in order of preference
	for _, lockFileAndPakageName := range LOCKFILES {

		lockFile := lockFileAndPakageName[0]
		packageName := lockFileAndPakageName[1]

		filePath := filepath.Join(cwd, lockFile)
		_, err := os.Stat(filePath)

		if err == nil {
			// File exists, we found the package manager
			return packageName, nil
		}

		// If the error indicates the file simply does not exist, continue to the next one.
		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		// If it's any other error, it's a deeper issue (e.g., permissions, corrupted file system).
		// Return this error immediately. Wrap it for context.
		return "", fmt.Errorf("failed to stat file %q while detecting package manager: %w", filePath, err)
	}

	// Return our specific error for this condition
	return "", ErrNoPackageManager
}

type YarnCommandVersionOutputter interface {
	Output() (string, error)
}

type RealYarnCommandVersionRunner struct {
	cmd *exec.Cmd
}

func NewRealYarnCommandVersionRunner() RealYarnCommandVersionRunner {

	return RealYarnCommandVersionRunner{
		cmd: exec.Command("yarn", "--version"),
	}
}

func (r RealYarnCommandVersionRunner) Output() (string, error) {

	output, error := r.cmd.Output()

	if error != nil {

		return "", error
	}

	return string(output), nil

}

func DetectYarnVersion(yarnVersionRunner YarnCommandVersionOutputter) (string, error) {

	result, error := yarnVersionRunner.Output()

	if error != nil {

		return "", error
	}

	return result, nil

}

const VOLTA = "volta"

var VOLTA_RUN_COMMNAD = []string{VOLTA, "run"}

func DetectVolta() bool {

	_, err := exec.LookPath(VOLTA)

	if err != nil {
		return false
	}

	return true

}
