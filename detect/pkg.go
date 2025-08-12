package detect

import (
	"errors"
	"fmt"
	"os"
	"os/exec" // Keep this import for RealPathLookup
	"path/filepath"

	"github.com/samber/lo"
)

// PathLookup interface abstracts the exec.LookPath functionality.
type PathLookup interface {
	LookPath(file string) (string, error)
}

// RealPathLookup is the production implementation of PathLookup.
type RealPathLookup struct{}

// LookPath implements PathLookup using the real exec.LookPath.
func (r RealPathLookup) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// FileSystem interface abstracts file system operations for testability.
type FileSystem interface {
	Stat(name string) (os.FileInfo, error)
	Getwd() (string, error)
}

// RealFileSystem is the production implementation of FileSystem.
type RealFileSystem struct{}

// Stat implements FileSystem using the real os.Stat.
func (r RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Getwd implements FileSystem using the real os.Getwd.
func (r RealFileSystem) Getwd() (string, error) {
	return os.Getwd()
}

const DENO = "deno"
const BUN = "bun"
const NPM = "npm"
const PNPM = "pnpm"
const YARN = "yarn"

// ErrNoPackageManager is returned when no supported JavaScript package manager
// lock file or configuration file is found in the current directory.
var ErrNoPackageManager = errors.New("no supported JavaScript package manager found")

const (
	BUN_LOCKB         = "bun.lockb"
	DENO_JSONC        = "deno.jsonc"
	DENO_JSON         = "deno.json"
	DENO_LOCK         = "deno.lock"
	PNPM_LOCK_YAML    = "pnpm-lock.yaml"
	YARN_LOCK         = "yarn.lock"
	PACKAGE_LOCK_JSON = "package-lock.json"
	YARN_LOCK_JSON    = "yarn.lock.json"
	BUN_LOCK_JSON     = "bun.lock.json"
)

var lockFiles = [9]string{
	DENO_LOCK,
	DENO_JSON,
	DENO_JSONC,
	BUN_LOCKB,
	BUN_LOCK_JSON,
	PNPM_LOCK_YAML,
	YARN_LOCK,
	YARN_LOCK_JSON,
	PACKAGE_LOCK_JSON,
}

func DetectLockfile(fs FileSystem) (lockfile string, error error) {

	cwd, err := fs.Getwd() // Use the injected FileSystem

	if err != nil {
		return "", err
	}

	for _, lockFile := range lockFiles {
		// Use the injected FileSystem's Stat method
		// fmt.Sprintf ensures correct path joining, especially for Windows if needed later.
		_, err := fs.Stat(filepath.Join(cwd, lockFile))

		if err == nil {
			return lockFile, nil // Return the lockFile and nil error if found
		}
		// DO NOT BREAK HERE! The original code had a bug where it would break after the first iteration.
		// We need to check all lock files.
	}

	return "", fmt.Errorf("No lock file found") // Return our specific error if no lockfile is found after checking all
}

// This is a list of the JS package managers
// ! NPM must be last if the user has node on their computer it will be detected before the others
var SupportedJSPackageManagers = [5]string{DENO, BUN, PNPM, YARN, NPM}

func DetectJSPackageManager(pathLookup PathLookup) (string, error) {

	for _, manager := range SupportedJSPackageManagers {
		if _, err := pathLookup.LookPath(manager); err == nil {
			return manager, nil
		}
	}

	return "", ErrNoPackageManager
}

var LockFileToPackageManagerMap = map[string]string{
	DENO_JSON:         DENO,
	DENO_LOCK:         DENO,
	DENO_JSONC:        DENO,
	PACKAGE_LOCK_JSON: NPM,
	PNPM_LOCK_YAML:    PNPM,
	BUN_LOCKB:         BUN,
	BUN_LOCK_JSON:     BUN,
	YARN_LOCK:         YARN,
	YARN_LOCK_JSON:    YARN,
}

func DetectJSPacakgeManagerBasedOnLockFile(detectedLockFile string, pathLookup PathLookup) (packageManager string, error error) {

	if !lo.Contains(lockFiles[:], detectedLockFile) {

		return "", fmt.Errorf("unsupported lockfile %s it must be one of these %v", detectedLockFile, lockFiles)

	}

	packageManagerToFind := LockFileToPackageManagerMap[detectedLockFile]
	// Use the injected pathLookup here
	_, err := pathLookup.LookPath(packageManagerToFind)

	if err != nil {
		// If LookPath returns os.ErrNotExist, return our specific error
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrNoPackageManager
		}
		return "", err // Return other errors as they are
	}

	return packageManagerToFind, nil // Return the actual package manager name
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

// DetectVolta now accepts a PathLookup interface to enable mocking.
func DetectVolta(pathLookup PathLookup) bool {

	_, err := pathLookup.LookPath(VOLTA) // Use the injected pathLookup

	if err != nil {
		return false
	}

	return true
}
