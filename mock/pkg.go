package mock

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/louiss0/javascript-package-delegator/cmd"
	"github.com/louiss0/javascript-package-delegator/services"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
	"github.com/stretchr/testify/mock"
	"golang.org/x/exp/rand"
)

// MockDebugExecutor implements the cmd.DebugExecutor interface for testing purposes
type MockDebugExecutor struct {
	mock.Mock
}

// ExcuteIfDebugIsTrue records the call to this method.
// In tests, you can set expectations using `On("ExcuteIfDebugIsTrue")`.
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

// MockCommandRunner implements the cmd.CommandRunner interface for testing purposes
// This ensures no real commands are executed during tests - it only records what would be run
type MockCommandRunner struct {
	// CommandCall stores the single command that has been called
	CommandCall CommandCall
	// InvalidCommands is a list of commands that should return an error when Run() is called
	InvalidCommands []string
	// HasBeenCalled indicates if a command has been set for this run
	HasBeenCalled bool
	WorkingDir    string
}

// CommandCall represents a single command call with its name and arguments
type CommandCall struct {
	Name string
	Args []string
}

// NewMockCommandRunner creates a new instance of MockCommandRunner
func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		CommandCall:     CommandCall{},
		InvalidCommands: []string{},
		HasBeenCalled:   false,
	}
}

// Command records the command that would be executed
func (m *MockCommandRunner) Command(name string, args ...string) {
	m.CommandCall = CommandCall{
		Name: name,
		Args: args,
	}
	m.HasBeenCalled = true
}

func (m *MockCommandRunner) SetTargetDir(dir string) error {

	fileInfo, err := os.Stat(dir) // Get file information
	if err != nil {

		return err
	}

	// Check if it's a directory
	if !fileInfo.IsDir() {
		return fmt.Errorf("target directory %s is not a directory", dir)
	}

	m.WorkingDir = dir
	return nil
}

// Run simulates running the command and records it.
// Returns an error if the command is in the InvalidCommands list.
func (m *MockCommandRunner) Run() error {
	if !m.HasBeenCalled {
		return fmt.Errorf("no command set to run")
	}

	// Check if this command should fail
	for _, invalidCmd := range m.InvalidCommands {
		if m.CommandCall.Name == invalidCmd {
			return fmt.Errorf("mock error: command '%s' is configured to fail", m.CommandCall.Name)
		}
	}

	return nil
}

// Reset clears the recorded command and invalid commands
func (m *MockCommandRunner) Reset() {
	m.CommandCall = CommandCall{}
	m.InvalidCommands = []string{}
	m.WorkingDir = ""
	m.HasBeenCalled = false
}

// HasCommand checks if the current command matches the given name and args
func (m *MockCommandRunner) HasCommand(name string, args ...string) bool {
	if !m.HasBeenCalled || m.CommandCall.Name != name {
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

// LastCommand returns the most recently executed command
func (m *MockCommandRunner) LastCommand() (CommandCall, bool) {
	if !m.HasBeenCalled {
		return CommandCall{}, false
	}
	return m.CommandCall, true
}

// MockYarnCommandVersionOutputer is a fake implementation that doesn't execute real yarn commands
type mockYarnCommandVersionOutputer struct {
	version string
}

func (my mockYarnCommandVersionOutputer) Output() (string, error) {

	match, error := regexp.MatchString(`\d\.\d\.\d`, my.version)

	if error != nil {

		return "", error
	}

	if !match {

		return "", fmt.Errorf("invalid version format you must use semver versioning")
	}

	return my.version, nil

}

// NewMockYarnCommandVersionOutputer creates a fake yarn version outputer for tests
func NewMockYarnCommandVersionOutputer(version string) mockYarnCommandVersionOutputer {
	return mockYarnCommandVersionOutputer{version: version}
}

type MockCommandTextUI struct {
	value string
}

func (ui MockCommandTextUI) Value() string {

	return ui.value
}

func (ui *MockCommandTextUI) SetValue(value string) string {
	ui.value = value
	return ui.value
}

func NewMockCommandTextUI(string) cmd.CommandUITexter {

	return &MockCommandTextUI{}
}

func (ui *MockCommandTextUI) Run() error {

	match, error := regexp.MatchString(cmd.VALID_INSTALL_COMMAND_STRING_RE, ui.Value())

	if error != nil {
		return error
	}

	if match {

		return nil
	}

	return fmt.Errorf(strings.Join(cmd.INVALID_COMMAND_STRUCTURE_ERROR_MESSAGE_STRUCTURE, "\n"), ui.value)

}

type MockPackageMultiSelectUI struct {
	values []string
}

func (ui MockPackageMultiSelectUI) Values() []string {

	return ui.values
}

func NewMockPackageMultiSelectUI(packages []services.PackageInfo) cmd.MultiUISelecter {

	return &MockPackageMultiSelectUI{
		values: lo.Map(packages, func(item services.PackageInfo, index int) string {

			return item.Name
		}),
	}
}

func (ui *MockPackageMultiSelectUI) Run() error {

	if len(ui.values) == 0 {
		return fmt.Errorf("no packages available")
	}

	min := 1
	max := 20

	// 2. Seed the random number generator with the current time
	// This ensures a different sequence of numbers each time the program runs.
	source := rand.NewSource(uint64(time.Now().UnixNano()))
	rng := rand.New(source)

	// 3. Generate a random number within the range [min, max]
	// rng.Intn(n) generates a number in [0, n).
	// So, to get a number in [min, max], we need a range of (max - min + 1).
	randomNumber := rng.Intn(max-min+1) + min

	mutable.Shuffle(ui.values)

	ui.values = lo.Slice(ui.values, 0, randomNumber)

	return nil
}

type MockTaskSelectUI struct {
	selectedValue string
	options       []string
}

func NewMockTaskSelectUI(options []string) cmd.TaskUISelector {
	return &MockTaskSelectUI{
		options: options,
	}
}

func (t MockTaskSelectUI) Value() string {
	return t.selectedValue
}

func (t *MockTaskSelectUI) Run() error {

	if len(t.options) == 0 {
		return fmt.Errorf("no tasks available for selection")
	}

	// Randomly select one option
	source := rand.NewSource(uint64(time.Now().UnixNano()))
	rng := rand.New(source)
	randomIndex := rng.Intn(len(t.options))
	t.selectedValue = t.options[randomIndex]

	return nil
}

type MockDependencyUISelector struct {
	selectedValues []string
	options        []string
}

func (m MockDependencyUISelector) Values() []string {
	return m.selectedValues
}

func (m *MockDependencyUISelector) Run() error {
	if len(m.options) == 0 {
		return fmt.Errorf("no dependencies available for selection")
	}

	// Randomly select one option
	source := rand.NewSource(uint64(time.Now().UnixNano()))
	rng := rand.New(source)
	randomIndex := rng.Intn(len(m.options))
	m.selectedValues = append(m.selectedValues, m.options[randomIndex])

	return nil
}

func NewMockDependencySelectUI(options []string) cmd.DependencyUIMultiSelector {
	return &MockDependencyUISelector{
		options: options,
	}
}
