package detect

import (
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
