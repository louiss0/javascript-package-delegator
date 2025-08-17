package testutil

import (
	"fmt"
	"os"
	"strings"

	"github.com/louiss0/javascript-package-delegator/cmd"
	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/mock"
	"github.com/louiss0/javascript-package-delegator/services"
	"github.com/spf13/cobra"
	tmock "github.com/stretchr/testify/mock"
)

type debugExecutorExpectationManager struct {
	DebugExecutor *mock.MockDebugExecutor
}

var DebugExecutorExpectationManager debugExecutorExpectationManager

// Debug executor expectation helpers (DRY)
func (m *debugExecutorExpectationManager) ExpectNoLockfile() {
	m.DebugExecutor.On("LogDebugMessageIfDebugIsTrue", "Lock file is not detected").Return()
}

func (m *debugExecutorExpectationManager) ExpectLockfileDetected(lf string) {
	m.DebugExecutor.On("LogDebugMessageIfDebugIsTrue", "Lock file is detected", "lockfile", lf).Return()
}

func (m *debugExecutorExpectationManager) ExpectPMDetectedFromLockfile(pm string) {
	m.DebugExecutor.On("LogDebugMessageIfDebugIsTrue", "Package manager is detected based on lock file", "pm", pm).Return()
}

func (m *debugExecutorExpectationManager) ExpectPMDetectedFromPath(pm string) {
	m.DebugExecutor.On("LogDebugMessageIfDebugIsTrue", "Package manager detected from path", "pm", pm).Return()
}

func (m *debugExecutorExpectationManager) ExpectNoPMFromPath() {
	m.DebugExecutor.On("LogDebugMessageIfDebugIsTrue", "Package manager is not detected from path").Return()
}

func (m *debugExecutorExpectationManager) ExpectJPDAgentSet(agent string) {
	m.DebugExecutor.On("LogDebugMessageIfDebugIsTrue", "JPD_AGENT environment variable detected setting agent", "agent", agent).Return()
}

func (m *debugExecutorExpectationManager) ExpectAgentFlagSet(agent string) {
	m.DebugExecutor.On("LogDebugMessageIfDebugIsTrue", "Agent flag is set", "agent", agent).Return()
}

func (m *debugExecutorExpectationManager) ExpectJSCommandLog(pm string, args ...string) {
	m.DebugExecutor.On("LogJSCommandIfDebugIsTrue", "Executing command:", "command", strings.Join(append([]string{pm}, args...), " ")).Return()
}

func (m *debugExecutorExpectationManager) ExpectJSCommandRandomLog() {
	m.DebugExecutor.On("LogJSCommandIfDebugIsTrue", "Executing command:", "command", tmock.AnythingOfType("string")).Return()
}

// RootCommandFactory is a helper struct for creating cobra.Command instances
// with various mocked dependencies for testing purposes.
type RootCommandFactory struct {
	mockRunner    *mock.MockCommandRunner
	debugExecutor *mock.MockDebugExecutor
}

// NewRootCommandFactory creates a new RootCommandFactory with the given mock runner.
func NewRootCommandFactory(mockRunner *mock.MockCommandRunner) *RootCommandFactory {
	return &RootCommandFactory{
		mockRunner:    mockRunner,
		debugExecutor: &mock.MockDebugExecutor{},
	}
}

func (f *RootCommandFactory) MockCommandRunner() *mock.MockCommandRunner {
	return f.mockRunner
}

func (f *RootCommandFactory) DebugExecutor() *mock.MockDebugExecutor {
	return f.debugExecutor
}

func (f *RootCommandFactory) ResetDebugExecutor() {
	f.debugExecutor = &mock.MockDebugExecutor{}
}

// baseDependencies returns a set of common mocked dependencies that can be overridden.
func (f *RootCommandFactory) baseDependencies() cmd.Dependencies {
	return cmd.Dependencies{
		CommandRunnerGetter: func() cmd.CommandRunner {
			return f.MockCommandRunner()
		},
		NewDebugExecutor: func(bool) cmd.DebugExecutor {
			return f.debugExecutor
		},
		DetectVolta:                 func() bool { return false },               // Default to no Volta detected
		YarnCommandVersionOutputter: mock.NewMockYarnCommandVersionOutputer(""), // Default to no specific yarn version
		NewCommandTextUI:            mock.NewMockCommandTextUI,
		NewPackageMultiSelectUI:     mock.NewMockPackageMultiSelectUI,
		NewTaskSelectorUI:           mock.NewMockTaskSelectUI,
		NewDependencyMultiSelectUI:  mock.NewMockDependencySelectUI,
	}
}

// CreateRootCmdWithLockfileDetected creates a root command simulating package manager
// detection based on a specific lockfile being found.
//
// `pm` is the package manager expected to be detected from the lockfile.
// `lockfile` specifies the detected lockfile string (e.g., detect.PACKAGE_LOCK_JSON).
// `pmDetectionErr` is an optional error for the PM detection *based on the lockfile*.
// `volta` specifies if Volta should be detected.
func (f *RootCommandFactory) CreateRootCmdWithLockfileDetected(pm string, lockfile string, pmDetectionErr error, volta bool) *cobra.Command {
	deps := f.baseDependencies()
	// As per the prompt, if a package manager is detected based on a lock file,
	// the lock file should be returned by the detector.
	// `DetectLockfile` is the primary detector for the lockfile itself.
	deps.DetectLockfile = func() (string, error) {
		return lockfile, nil // Lockfile successfully detected and returned
	}
	deps.DetectJSPackageManagerBasedOnLockFile = func(detectedLockFile string) (string, error) {
		// This mock takes the detected lockfile string as input and returns the package manager.
		// The `lockfile` argument passed to this factory method is what `DetectLockfile` will return.
		return pm, pmDetectionErr // PM detected based on the lockfile string
	}
	deps.DetectJSPackageManager = func() (string, error) {
		// This function should not be called if lockfile detection succeeded
		return "", fmt.Errorf("DetectJSPackageManager should not be called in lockfile detection scenario")
	}
	deps.DetectVolta = func() bool {
		return volta
	}
	return cmd.NewRootCmd(deps)
}

// CreateRootCmdWithPathDetected creates a root command simulating package manager
// detection by checking the global PATH (no lockfile found).
//
// `pm` is the package manager expected to be detected globally.
// `pmDetectionErr` is an optional error for the global PM detection.
// `volta` specifies if Volta should be detected.
func (f *RootCommandFactory) CreateRootCmdWithPathDetected(pm string, pmDetectionErr error, volta bool) *cobra.Command {
	deps := f.baseDependencies()
	deps.DetectLockfile = func() (string, error) {
		return "", os.ErrNotExist // No lockfile found, forcing path detection
	}
	deps.DetectJSPackageManagerBasedOnLockFile = func(detectedLockFile string) (string, error) {
		// This function should not be called if lockfile detection failed
		return "", fmt.Errorf("DetectJSPackageManagerBasedOnLockFile should not be called when lockfile detection fails")
	}
	deps.DetectJSPackageManager = func() (string, error) {
		return pm, pmDetectionErr // PM detected globally via PATH
	}
	deps.DetectVolta = func() bool {
		return volta
	}
	return cmd.NewRootCmd(deps)
}

// GenerateWithPackageManagerDetector creates a root command with a specific package manager detected,
// and can simulate an error during detection. This simulates lockfile-based detection.
func (f *RootCommandFactory) GenerateWithPackageManagerDetector(packageManager string, err error) *cobra.Command {
	return f.CreateRootCmdWithLockfileDetected(packageManager, detect.PACKAGE_LOCK_JSON, err, false)
}

// GenerateWithPackageManagerDetectedAndVolta creates a root command where Volta is also detected,
// and package manager detection is lockfile-based.
func (f *RootCommandFactory) GenerateWithPackageManagerDetectedAndVolta(packageManager string) *cobra.Command {
	return f.CreateRootCmdWithLockfileDetected(packageManager, detect.PACKAGE_LOCK_JSON, nil, true)
}

// CreateBunAsDefault creates a root command with "bun" as the default detected package manager,
// simulating lockfile-based detection.
func (f *RootCommandFactory) CreateBunAsDefault(err error) *cobra.Command {
	return f.GenerateWithPackageManagerDetector("bun", err)
}

// CreateDenoAsDefault creates a root command with "deno" as the default detected package manager,
// simulating lockfile-based detection.
func (f *RootCommandFactory) CreateDenoAsDefault(err error) *cobra.Command {
	return f.GenerateWithPackageManagerDetector("deno", err)
}

// CreateYarnTwoAsDefault creates a root command with "yarn" (version 2+) as the default detected package manager,
// simulating detection via PATH and specific yarn version output.
func (f *RootCommandFactory) CreateYarnTwoAsDefault(err error) *cobra.Command {
	deps := f.baseDependencies()
	deps.DetectLockfile = func() (string, error) {
		return "", os.ErrNotExist // No lockfile found, forcing path detection
	}
	deps.DetectJSPackageManagerBasedOnLockFile = func(detectedLockFile string) (string, error) {
		// This function should not be called if lockfile detection failed
		return "", fmt.Errorf("DetectJSPackageManagerBasedOnLockFile should not be called when lockfile detection fails")
	}
	deps.DetectJSPackageManager = func() (string, error) {
		return "yarn", err // PM detected globally via PATH, for yarn
	}
	deps.DetectVolta = func() bool {
		return false // Default to no Volta detected
	}
	// Override specific dependency for Yarn version output, as it's part of how yarn is "path-detected"
	deps.YarnCommandVersionOutputter = mock.NewMockYarnCommandVersionOutputer("2.0.0")
	return cmd.NewRootCmd(deps)
}

// CreateYarnOneAsDefault creates a root command with "yarn" (version 1) as the default detected package manager,
// simulating detection via PATH and specific yarn version output.
func (f *RootCommandFactory) CreateYarnOneAsDefault(err error) *cobra.Command {
	deps := f.baseDependencies()

	deps.DetectLockfile = func() (string, error) {
		return "", os.ErrNotExist // No lockfile found, forcing path detection
	}
	deps.DetectJSPackageManagerBasedOnLockFile = func(detectedLockFile string) (string, error) {
		// This function should not be called if lockfile detection failed
		return "", fmt.Errorf("DetectJSPackageManagerBasedOnLockFile should not be called when lockfile detection fails")
	}
	deps.DetectJSPackageManager = func() (string, error) {
		return "yarn", err // PM detected globally via PATH, for yarn
	}
	deps.DetectVolta = func() bool {
		return false // Default to no Volta detected
	}
	// Override specific dependency for Yarn version output
	deps.YarnCommandVersionOutputter = mock.NewMockYarnCommandVersionOutputer("1.0.0")
	return cmd.NewRootCmd(deps)
}

// CreateNoYarnVersion creates a root command simulating no yarn version detection,
// implying a path-based detection (or lack thereof for yarn version).
func (f *RootCommandFactory) CreateNoYarnVersion(err error) *cobra.Command {
	rootCmd := f.CreateRootCmdWithPathDetected("yarn", err, false)
	// YarnCommandVersionOutputter is default empty in baseDependencies, no need to set again
	return rootCmd
}

// CreatePnpmAsDefault creates a root command with "pnpm" as the default detected package manager,
// simulating lockfile-based detection.
func (f *RootCommandFactory) CreatePnpmAsDefault(err error) *cobra.Command {
	return f.GenerateWithPackageManagerDetector("pnpm", err)
}

// CreateNpmAsDefault creates a root command with "npm" as the default detected package manager,
// simulating lockfile-based detection.
func (f *RootCommandFactory) CreateNpmAsDefault(err error) *cobra.Command {
	return f.GenerateWithPackageManagerDetector("npm", err)
}

// GenerateNoDetectionAtAll creates a root command simulating no lockfile or global PM detection,
// forcing a prompt for an install command. This is a specific "no detection" scenario.
func (f *RootCommandFactory) GenerateNoDetectionAtAll(commandTextUIValue string) *cobra.Command {
	deps := f.baseDependencies()

	deps.DetectLockfile = func() (lockfile string, error error) {
		return "", os.ErrNotExist // No lockfile detected
	}
	deps.DetectJSPackageManagerBasedOnLockFile = func(detectedLockFile string) (string, error) {
		// Should not be called as DetectLockfile returned an error
		return "", nil
	}
	deps.DetectJSPackageManager = func() (string, error) {
		return "", detect.ErrNoPackageManager // No PM found on PATH
	}
	deps.NewCommandTextUI = func(lockfile string) cmd.CommandUITexter {
		mockUI := mock.NewMockCommandTextUI(lockfile).(*mock.MockCommandTextUI)
		mockUI.SetValue(commandTextUIValue)
		return mockUI
	}
	return cmd.NewRootCmd(deps)
}

// CreateWithPackageManagerAndMultiSelectUI creates a root command configured for package manager
// detection via PATH and multi-select UI.
func (f *RootCommandFactory) CreateWithPackageManagerAndMultiSelectUI() *cobra.Command {
	// Original used DetectLockfile: "", nil and DetectJSPackageManagerBasedOnLockFile: "npm", nil.
	// Refactoring to explicitly use PATH detection for non-specific lockfile scenarios as per prompt.
	deps := f.baseDependencies()
	deps.DetectLockfile = func() (lockfile string, error error) {
		return "", os.ErrNotExist
	}
	deps.DetectJSPackageManager = func() (string, error) {
		return "npm", nil
	}
	deps.NewPackageMultiSelectUI = func(pi []services.PackageInfo) cmd.MultiUISelecter {
		return mock.NewMockPackageMultiSelectUI(pi)
	}
	return cmd.NewRootCmd(deps)
}

// CreateWithTaskSelectorUI creates a root command configured for task selection UI based on a
// package manager detected via PATH.
func (f *RootCommandFactory) CreateWithTaskSelectorUI(packageManager string) *cobra.Command {
	// Original used DetectLockfile: "", nil and DetectJSPackageManagerBasedOnLockFile.
	// Refactoring to explicitly use PATH detection for non-specific lockfile scenarios as per prompt.
	deps := f.baseDependencies()
	deps.DetectLockfile = func() (lockfile string, error error) {
		return "", os.ErrNotExist
	}
	deps.DetectJSPackageManager = func() (string, error) {
		return packageManager, nil
	}
	deps.NewTaskSelectorUI = mock.NewMockTaskSelectUI
	return cmd.NewRootCmd(deps)
}
