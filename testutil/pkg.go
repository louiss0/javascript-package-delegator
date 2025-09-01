package testutil

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	tmock "github.com/stretchr/testify/mock"

	"github.com/louiss0/javascript-package-delegator/cmd"
	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/mock"
	"github.com/louiss0/javascript-package-delegator/services"
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

func (m *debugExecutorExpectationManager) ExpectNoPMFromLockfile() {
	m.DebugExecutor.On("LogDebugMessageIfDebugIsTrue", "Package manager is not detected from lockfile").Return()
}

func (m *debugExecutorExpectationManager) ExpectJPDAgentSet(agent string) {
	m.DebugExecutor.On("LogDebugMessageIfDebugIsTrue", "JPD_AGENT environment variable detected setting agent", "agent", agent).Return()
}

func (m *debugExecutorExpectationManager) ExpectAgentFlagSet(agent string) {
	m.DebugExecutor.On("LogDebugMessageIfDebugIsTrue", "Agent flag is set", "agent", agent).Return()
}

func (m *debugExecutorExpectationManager) ExpectJSCommandLog(pm string, args ...string) {
	// Build expected arguments slice for mock expectation
	expectedArgs := []interface{}{pm}
	for _, arg := range args {
		expectedArgs = append(expectedArgs, arg)
	}
	m.DebugExecutor.On("LogJSCommandIfDebugIsTrue", expectedArgs...).Return()
}

func (m *debugExecutorExpectationManager) ExpectJSCommandRandomLog() {
	// Use mock.Anything multiple times to handle variable arguments (up to 25 args should be enough)
	m.DebugExecutor.On("LogJSCommandIfDebugIsTrue",
		tmock.AnythingOfType("string"), // Package manager
		tmock.Anything,                 // Command
		tmock.Anything,                 // Args...
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
		tmock.Anything,
	).Return().Maybe() // Maybe allows this expectation to match zero or more times
}

// ExpectAnyDebugMessages allows any debug messages to be logged without specific expectations
func (m *debugExecutorExpectationManager) ExpectAnyDebugMessages() {
	// Allow any LogDebugMessageIfDebugIsTrue calls
	m.DebugExecutor.On("LogDebugMessageIfDebugIsTrue", tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything).Return().Maybe()
	m.DebugExecutor.On("LogDebugMessageIfDebugIsTrue", tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything).Return().Maybe()
	m.DebugExecutor.On("LogDebugMessageIfDebugIsTrue", tmock.Anything, tmock.Anything, tmock.Anything).Return().Maybe()
	m.DebugExecutor.On("LogDebugMessageIfDebugIsTrue", tmock.Anything, tmock.Anything).Return().Maybe()
	m.DebugExecutor.On("LogDebugMessageIfDebugIsTrue", tmock.Anything).Return().Maybe()

	// Allow any LogJSCommandIfDebugIsTrue calls
	m.ExpectJSCommandRandomLog()
}

// ExpectCommonPMDetectionFlow expects the most common package manager detection flow based on lockfile
func (m *debugExecutorExpectationManager) ExpectCommonPMDetectionFlow(pm, lockfile string) {
	m.ExpectLockfileDetected(lockfile)
	m.ExpectPMDetectedFromLockfile(pm)
}

// ExpectCommonPathDetectionFlow expects the most common package manager detection flow based on PATH
func (m *debugExecutorExpectationManager) ExpectCommonPathDetectionFlow(pm string) {
	m.ExpectNoLockfile()
	m.ExpectPMDetectedFromPath(pm)
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

// SetupBasicCommandRunnerExpectations sets up basic expectations for the MockCommandRunner
// This is a helper to avoid repeating common mock setup in tests
func (f *RootCommandFactory) SetupBasicCommandRunnerExpectations() {
	// Allow any commands to be set with variable number of arguments
	f.mockRunner.On("Command", tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything).Return().Maybe()
	f.mockRunner.On("Command", tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything).Return().Maybe()
	f.mockRunner.On("Command", tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything).Return().Maybe()
	f.mockRunner.On("Command", tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything).Return().Maybe()
	f.mockRunner.On("Command", tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything).Return().Maybe()
	f.mockRunner.On("Command", tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything).Return().Maybe()
	f.mockRunner.On("Command", tmock.Anything, tmock.Anything, tmock.Anything, tmock.Anything).Return().Maybe()
	f.mockRunner.On("Command", tmock.Anything, tmock.Anything, tmock.Anything).Return().Maybe()
	f.mockRunner.On("Command", tmock.Anything, tmock.Anything).Return().Maybe()
	f.mockRunner.On("Command", tmock.Anything).Return().Maybe()
	// Allow any target directory to be set
	f.mockRunner.On("SetTargetDir", tmock.Anything).Return(nil).Maybe()
	// Allow commands to run successfully by default
	f.mockRunner.On("Run").Return(nil).Maybe()
	// Allow Run to be asserted with synthesized arguments: name, args slice, and working dir
	f.mockRunner.On("Run", tmock.Anything, tmock.Anything, tmock.Anything).Return(nil).Maybe()
}

// SetupBasicDebugExecutorExpectations sets up permissive expectations for the debug executor
// This allows tests to run without panicking on unexpected debug calls
func (f *RootCommandFactory) SetupBasicDebugExecutorExpectations() {
	// Set up the global debug executor expectation manager
	DebugExecutorExpectationManager.DebugExecutor = f.debugExecutor
	// Allow any debug messages and JS command logging
	DebugExecutorExpectationManager.ExpectAnyDebugMessages()
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
	deps.DetectLockfile = func(targetDir string) (string, error) {
		return lockfile, nil // Lockfile successfully detected and returned
	}
	deps.DetectJSPackageManagerBasedOnLockFile = func(detectedLockFile string) (string, error) {
		// This mock takes the detected lockfile string as input and returns the package manager.
		// The `lockfile` argument passed to this factory method is what `DetectLockfile` will return.
		return pm, pmDetectionErr // PM detected based on the lockfile string
	}
	deps.DetectJSPackageManager = func() (string, error) {
		// This function should not be called if lockfile detection succeeded
		return "", fmt.Errorf("detectJSPackageManager should not be called in lockfile detection scenario")
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
	deps.DetectLockfile = func(targetDir string) (string, error) {
		return "", os.ErrNotExist // No lockfile found, forcing path detection
	}
	deps.DetectJSPackageManagerBasedOnLockFile = func(detectedLockFile string) (string, error) {
		// This function should not be called if lockfile detection failed
		return "", fmt.Errorf("detectJSPackageManagerBasedOnLockFile should not be called when lockfile detection fails")
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
	return f.CreateRootCmdWithLockfileDetected("bun", detect.BUN_LOCKB, err, false)
}

// CreateDenoAsDefault creates a root command with "deno" as the default detected package manager,
// simulating lockfile-based detection.
func (f *RootCommandFactory) CreateDenoAsDefault(err error) *cobra.Command {
	return f.CreateRootCmdWithLockfileDetected("deno", detect.DENO_JSON, err, false)
}

// CreateYarnTwoAsDefault creates a root command with "yarn" (version 2+) as the default detected package manager,
// simulating lockfile-based detection.
func (f *RootCommandFactory) CreateYarnTwoAsDefault(err error) *cobra.Command {
	deps := f.baseDependencies()
	deps.DetectLockfile = func(targetDir string) (string, error) {
		return detect.YARN_LOCK, nil // Lockfile found
	}
	deps.DetectJSPackageManagerBasedOnLockFile = func(detectedLockFile string) (string, error) {
		return "yarn", err // PM detected from lockfile
	}
	deps.DetectJSPackageManager = func() (string, error) {
		// This function should not be called if lockfile detection succeeded
		return "", fmt.Errorf("detectJSPackageManager should not be called in lockfile detection scenario")
	}
	deps.DetectVolta = func() bool {
		return false // Default to no Volta detected
	}
	// Override specific dependency for Yarn version output
	deps.YarnCommandVersionOutputter = mock.NewMockYarnCommandVersionOutputer("2.0.0")
	return cmd.NewRootCmd(deps)
}

// CreateYarnOneAsDefault creates a root command with "yarn" (version 1) as the default detected package manager,
// simulating detection via PATH and specific yarn version output.
func (f *RootCommandFactory) CreateYarnOneAsDefault(err error) *cobra.Command {
	deps := f.baseDependencies()

	deps.DetectLockfile = func(targetDir string) (string, error) {
		return "", os.ErrNotExist // No lockfile found, forcing path detection
	}
	deps.DetectJSPackageManagerBasedOnLockFile = func(detectedLockFile string) (string, error) {
		// This function should not be called if lockfile detection failed
		return "", fmt.Errorf("detectJSPackageManagerBasedOnLockFile should not be called when lockfile detection fails")
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
	return f.CreateRootCmdWithLockfileDetected("pnpm", detect.PNPM_LOCK_YAML, err, false)
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

	deps.DetectLockfile = func(targetDir string) (lockfile string, err error) {
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
		mockUI := mock.NewMockCommandTextUI(commandTextUIValue).(*mock.MockCommandTextUI)
		// Set up specific expectations for the UI behavior
		if commandTextUIValue != "" {
			mockUI.On("SetValue", commandTextUIValue).Return(commandTextUIValue)
		}
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
	deps.DetectLockfile = func(targetDir string) (lockfile string, err error) {
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
	deps.DetectLockfile = func(targetDir string) (lockfile string, err error) {
		return "", os.ErrNotExist
	}
	deps.DetectJSPackageManager = func() (string, error) {
		return packageManager, nil
	}
	deps.NewTaskSelectorUI = mock.NewMockTaskSelectUI
	return cmd.NewRootCmd(deps)
}

// CreateWithDependencySelectUI creates a root command configured for dependency selection UI based on a
// package manager detected via PATH.
func (f *RootCommandFactory) CreateWithDependencySelectUI(packageManager string) *cobra.Command {
	deps := f.baseDependencies()
	deps.DetectLockfile = func(targetDir string) (lockfile string, err error) {
		return "", os.ErrNotExist
	}
	deps.DetectJSPackageManager = func() (string, error) {
		return packageManager, nil
	}
	deps.NewDependencyMultiSelectUI = mock.NewMockDependencySelectUI
	return cmd.NewRootCmd(deps)
}
