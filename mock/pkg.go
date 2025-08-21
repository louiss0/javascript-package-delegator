// Package mock provides mock implementations for testing the javascript-package-delegator.
package mock

import (
	// standard library
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	// internal
	"github.com/louiss0/javascript-package-delegator/cmd"
	"github.com/louiss0/javascript-package-delegator/services"

	// external
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
	"github.com/stretchr/testify/mock"
	"golang.org/x/exp/rand"
)

// MockDebugExecutor implements the cmd.DebugExecutor interface for testing purposes
type MockDebugExecutor struct {
	mock.Mock
}

// ExecuteIfDebugIsTrue records the call to this method.
// In tests, you can set expectations using `On("ExecuteIfDebugIsTrue")`.
// If the `cb` itself needs to be verified or controlled, the test can pass a mockable function.
func (m *MockDebugExecutor) ExecuteIfDebugIsTrue(cb func()) {
	m.Called(cb)
}

// LogDebugMessageIfDebugIsTrue records the call to this method along with its arguments.
// In tests, you can set expectations and verify arguments using `On("LogDebugMessageIfDebugIsTrue", ...)`
// and `AssertCalled(...)`.
func (m *MockDebugExecutor) LogDebugMessageIfDebugIsTrue(msg string, keyvals ...interface{}) {
	// Append all arguments (msg and keyvals) into a single slice for m.Called()
	args := []interface{}{msg}
	args = append(args, keyvals...)
	m.Called(args...)
}

func (m *MockDebugExecutor) LogJSCommandIfDebugIsTrue(cmd string, args ...string) {
	// Build the expected arguments slice for m.Called()
	callArgs := []interface{}{cmd}
	for _, arg := range args {
		callArgs = append(callArgs, arg)
	}
	m.Called(callArgs...)
}

// MockCommandRunner implements the cmd.CommandRunner interface for testing purposes
// This ensures no real commands are executed during tests - it uses testify/mock for expectations
type MockCommandRunner struct {
	mock.Mock
	// Fields for backward compatibility with existing tests
	HasBeenCalled   bool
	CommandCall     CommandCall
	InvalidCommands []string
	WorkingDir      string
}

// CommandCall represents a single command call with its name and arguments
type CommandCall struct {
	Name string
	Args []string
}

// NewMockCommandRunner creates a new instance of MockCommandRunner
func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		HasBeenCalled:   false,
		CommandCall:     CommandCall{},
		InvalidCommands: []string{},
		WorkingDir:      "",
	}
}

// hasExpectation returns true if there is an explicit expectation set for the given method name.
func (m *MockCommandRunner) hasExpectation(method string) bool {
	for _, c := range m.ExpectedCalls {
		if c.Method == method {
			return true
		}
	}
	return false
}

// hasExpectationWithArgLen returns true if there is an explicit expectation set for the given
// method name AND the number of expected arguments matches the provided length. This helps avoid
// testify panics when a different-arity expectation exists for the same method.
func (m *MockCommandRunner) hasExpectationWithArgLen(method string, argLen int) bool {
	for _, c := range m.ExpectedCalls {
		if c.Method == method {
			if len(c.Arguments) == argLen {
				return true
			}
		}
	}
	return false
}

// Command records the command that would be executed
func (m *MockCommandRunner) Command(name string, args ...string) {
	// Build arguments for mock call
	callArgs := []interface{}{name}
	for _, arg := range args {
		callArgs = append(callArgs, arg)
	}
	// Only invoke testify's Called if an expectation for Command is set with matching arity
	if m.hasExpectationWithArgLen("Command", len(callArgs)) {
		m.Called(callArgs...)
	}

	// Store the command for backward compatibility
	// Also reset the execution flag to avoid bleed-over between tests
	m.HasBeenCalled = false
	m.CommandCall = CommandCall{
		Name: name,
		Args: args,
	}
}

// SetTargetDir sets the target directory for command execution
func (m *MockCommandRunner) SetTargetDir(dir string) error {
	// Simulate real behavior: validate directory exists
	if fileInfo, err := os.Stat(dir); err != nil || !fileInfo.IsDir() {
		return fmt.Errorf("stat %s: no such file or directory", dir)
	}
	// Only invoke testify's Called if an expectation with matching arity is set
	if m.hasExpectationWithArgLen("SetTargetDir", 1) {
		_ = m.Called(dir)
	}
	m.WorkingDir = dir
	return nil
}

// Run simulates running the command
func (m *MockCommandRunner) Run() error {
	// If no command was set, return an error (unless tests override via expectation)
	if m.CommandCall.Name == "" {
		if m.hasExpectation("Run") {
			args := m.Called(m.CommandCall.Name, m.CommandCall.Args, m.WorkingDir)
			return args.Error(0)
		}
		return fmt.Errorf("no command set to run")
	}

	// Mark that a run attempt has been made whenever a command is present
	m.HasBeenCalled = true

	// If an expectation with matching arity exists, obtain its result first
	var expectedErr error
	if m.hasExpectationWithArgLen("Run", 3) {
		args := m.Called(m.CommandCall.Name, m.CommandCall.Args, m.WorkingDir)
		expectedErr = args.Error(0)
	}

	// If this command is configured as invalid and no explicit error was provided by expectations, fail deterministically
	for _, invalidCmd := range m.InvalidCommands {
		if m.CommandCall.Name == invalidCmd {
			if expectedErr != nil {
				return expectedErr
			}
			return fmt.Errorf("mock error: command '%s' is configured to fail", invalidCmd)
		}
	}

	// If we had an expectation and it returned an error (or nil), return it now
	if m.hasExpectation("Run") {
		return expectedErr
	}

	// Otherwise, consider that no actual run was attempted (help path, etc.)
	return nil
}

// HasCommand checks if a specific command with args was called
func (m *MockCommandRunner) HasCommand(name string, args ...string) bool {
	if m.CommandCall.Name != name {
		return false
	}

	if len(m.CommandCall.Args) != len(args) {
		return false
	}

	for i, arg := range args {
		if m.CommandCall.Args[i] != arg {
			return false
		}
	}

	return true
}

// Reset clears all state for reuse
func (m *MockCommandRunner) Reset() {
	m.HasBeenCalled = false
	m.CommandCall = CommandCall{}
	m.InvalidCommands = []string{}
	m.WorkingDir = ""
	m.Mock = mock.Mock{}
}

// ResetHasBeenCalled resets only the HasBeenCalled flag. Useful to ensure clean state between specs
func (m *MockCommandRunner) ResetHasBeenCalled() {
	m.HasBeenCalled = false
}

// LastCommand returns the last executed command
func (m *MockCommandRunner) LastCommand() (CommandCall, bool) {
	if m.CommandCall.Name == "" {
		return CommandCall{}, false
	}
	return m.CommandCall, true
}

// MockYarnCommandVersionOutputer is a testify/mock implementation for yarn version commands
type MockYarnCommandVersionOutputer struct {
	mock.Mock
}

// Output executes the yarn version output with mock expectations
func (my *MockYarnCommandVersionOutputer) Output() (string, error) {
	args := my.Called()
	return args.String(0), args.Error(1)
}

// NewMockYarnCommandVersionOutputer creates a new MockYarnCommandVersionOutputer
func NewMockYarnCommandVersionOutputer(version string) *MockYarnCommandVersionOutputer {
	mockOutputer := &MockYarnCommandVersionOutputer{}
	if version != "" {
		// Pre-configure the mock with the expected version if provided
		match, err := regexp.MatchString(`\d\.\d\.\d`, version)
		if err != nil {
			mockOutputer.On("Output").Return("", err)
		} else if !match {
			mockOutputer.On("Output").Return("", fmt.Errorf("invalid version format you must use semver versioning"))
		} else {
			mockOutputer.On("Output").Return(version, nil)
		}
	} else {
		// When no version is provided, simulate an error to trigger fallback logic in code
		mockOutputer.On("Output").Return("", fmt.Errorf("unable to detect yarn version"))
	}
	return mockOutputer
}

// MockCommandTextUI implements the cmd.CommandUITexter interface using testify/mock
type MockCommandTextUI struct {
	mock.Mock
	storedValue string
}

// hasExpectation returns true if there is an explicit expectation set for the given method name.
func (ui *MockCommandTextUI) hasExpectation(method string) bool {
	for _, c := range ui.ExpectedCalls {
		if c.Method == method {
			return true
		}
	}
	return false
}

// Value returns the current value of the text UI
func (ui *MockCommandTextUI) Value() string {
	if ui.storedValue != "" {
		return ui.storedValue
	}
	if ui.hasExpectation("Value") {
		args := ui.Called()
		return args.String(0)
	}
	return ""
}

// SetValue sets the value of the text UI
func (ui *MockCommandTextUI) SetValue(value string) string {
	ui.storedValue = value
	if ui.hasExpectation("SetValue") {
		_ = ui.Called(value)
	}
	return value
}

// Run executes the text UI
func (ui *MockCommandTextUI) Run() error {
	if ui.hasExpectation("Run") {
		args := ui.Called()
		return args.Error(0)
	}
	return nil
}

// NewMockCommandTextUI creates a new MockCommandTextUI with default behavior
func NewMockCommandTextUI(defaultValue string) cmd.CommandUITexter {
	mockUI := &MockCommandTextUI{}
	// Seed the stored value so Value() works even without explicit expectations
	mockUI.storedValue = defaultValue
	// Default to success; specific tests can override
	mockUI.On("Run").Return(nil).Maybe()

	if defaultValue != "" {
		// Set up default Run behavior based on validation
		match, err := regexp.MatchString(cmd.VALID_INSTALL_COMMAND_STRING_RE, defaultValue)
		if err != nil {
			mockUI.On("Run").Return(err).Maybe()
		} else if match {
			mockUI.On("Run").Return(nil).Maybe()
		} else {
			mockUI.On("Run").Return(fmt.Errorf(strings.Join(cmd.INVALID_COMMAND_STRUCTURE_ERROR_MESSAGE_STRUCTURE, "\n"), defaultValue)).Maybe()
		}
	}
	return mockUI
}

// MockPackageMultiSelectUI implements the cmd.MultiUISelecter interface using testify/mock
type MockPackageMultiSelectUI struct {
	mock.Mock
}

// Values returns the selected package values
func (ui *MockPackageMultiSelectUI) Values() []string {
	args := ui.Called()
	return args.Get(0).([]string)
}

// Run executes the multi-select UI
func (ui *MockPackageMultiSelectUI) Run() error {
	args := ui.Called()
	return args.Error(0)
}

// NewMockPackageMultiSelectUI creates a new MockPackageMultiSelectUI with default behavior
func NewMockPackageMultiSelectUI(packages []services.PackageInfo) cmd.MultiUISelecter {
	mockUI := &MockPackageMultiSelectUI{}
	packageNames := lo.Map(packages, func(item services.PackageInfo, index int) string {
		return item.Name
	})

	if len(packages) == 0 {
		mockUI.On("Values").Return([]string{}).Maybe()
		mockUI.On("Run").Return(fmt.Errorf("no packages available")).Maybe()
	} else {
		// Set up default behavior with randomized selection
		min := 1
		max := 20
		if len(packageNames) < max {
			max = len(packageNames)
		}

		source := rand.NewSource(uint64(time.Now().UnixNano()))
		rng := rand.New(source)
		randomNumber := rng.Intn(max-min+1) + min
		if randomNumber > len(packageNames) {
			randomNumber = len(packageNames)
		}

		mutable.Shuffle(packageNames)
		selectedPackages := lo.Slice(packageNames, 0, randomNumber)

		mockUI.On("Values").Return(selectedPackages).Maybe()
		mockUI.On("Run").Return(nil).Maybe()
	}
	return mockUI
}

// MockTaskSelectUI implements the cmd.TaskUISelector interface using testify/mock
type MockTaskSelectUI struct {
	mock.Mock
}

// Value returns the selected task value
func (t *MockTaskSelectUI) Value() string {
	args := t.Called()
	return args.String(0)
}

// Run executes the task selection UI
func (t *MockTaskSelectUI) Run() error {
	args := t.Called()
	return args.Error(0)
}

// NewMockTaskSelectUI creates a new MockTaskSelectUI with default behavior
func NewMockTaskSelectUI(options []string) cmd.TaskUISelector {
	mockUI := &MockTaskSelectUI{}

	if len(options) == 0 {
		mockUI.On("Value").Return("").Maybe()
		mockUI.On("Run").Return(fmt.Errorf("no tasks available for selection")).Maybe()
	} else {
		// Randomly select one option for default behavior
		source := rand.NewSource(uint64(time.Now().UnixNano()))
		rng := rand.New(source)
		randomIndex := rng.Intn(len(options))
		selectedValue := options[randomIndex]

		mockUI.On("Value").Return(selectedValue).Maybe()
		mockUI.On("Run").Return(nil).Maybe()
	}
	return mockUI
}

// MockDependencyUISelector implements the cmd.DependencyUIMultiSelector interface using testify/mock
type MockDependencyUISelector struct {
	mock.Mock
}

// Values returns the selected dependency values
func (m *MockDependencyUISelector) Values() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

// Run executes the dependency selection UI
func (m *MockDependencyUISelector) Run() error {
	args := m.Called()
	return args.Error(0)
}

// NewMockDependencySelectUI creates a new MockDependencyUISelector with default behavior
func NewMockDependencySelectUI(options []string) cmd.DependencyUIMultiSelector {
	mockUI := &MockDependencyUISelector{}

	if len(options) == 0 {
		mockUI.On("Values").Return([]string{}).Maybe()
		mockUI.On("Run").Return(fmt.Errorf("no dependencies available for selection")).Maybe()
	} else {
		// Randomly select one option for default behavior
		source := rand.NewSource(uint64(time.Now().UnixNano()))
		rng := rand.New(source)
		randomIndex := rng.Intn(len(options))
		selectedValues := []string{options[randomIndex]}

		mockUI.On("Values").Return(selectedValues).Maybe()
		mockUI.On("Run").Return(nil).Maybe()
	}
	return mockUI
}
