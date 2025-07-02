package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/louiss0/javascript-package-delegator/cmd"
	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/services"
	. "github.com/onsi/ginkgo/v2"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
)

// MockCommandRunner implements the cmd.CommandRunner interface for testing purposes
type MockCommandRunner struct {
	// CommandCall stores the single command that has been called
	CommandCall CommandCall
	// InvalidCommands is a list of commands that should return an error when Run() is called
	InvalidCommands []string
	// HasBeenCalled indicates if a command has been set for this run
	HasBeenCalled bool
	isDebug       bool
	WorkingDir    string
}

// CommandCall represents a single command call with its name and arguments
type CommandCall struct {
	Name string
	Args []string
}

// NewMockCommandRunner creates a new instance of MockCommandRunner
func NewMockCommandRunner(isDebug bool) *MockCommandRunner {
	return &MockCommandRunner{
		CommandCall:     CommandCall{},
		InvalidCommands: []string{},
		HasBeenCalled:   false,
		isDebug:         isDebug,
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

func (m MockCommandRunner) IsDebug() bool {
	return m.isDebug
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

type MockYarnCommandVersionOutputer struct {
	version string
}

func (my MockYarnCommandVersionOutputer) Output() (string, error) {

	match, error := regexp.MatchString(`\d\.\d\.\d`, my.version)

	if error != nil {

		return "", error
	}

	if !match {

		return "", fmt.Errorf("invalid version format you must use semver versioning")
	}

	return my.version, nil

}

// This function executes a cobra command with the given arguments and returns the output and error.
// It sets the output and error buffers for the command, sets the arguments, and executes the command.
// If there is an error, it returns an error with the error message from the error buffer.
// If there is no error, it returns the output from the output buffer.
// It's used to test the cobra commands.
// When you use this function, make sure to pass the root command and any arguments you want to test.
// The first argument after the rootCmd is any sub command or flag you want to test.
// This function now properly preserves the command context with CommandRunner.

func executeCmd(cmd *cobra.Command, args ...string) (string, error) {
	// Save the original context to restore it later
	originalCtx := cmd.Context()

	// Create buffers for output
	buf := new(bytes.Buffer)
	errBuff := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(errBuff)
	cmd.SetArgs(args)

	// Execute the command
	err := cmd.Execute()

	// Restore the original context
	if originalCtx != nil {
		cmd.SetContext(originalCtx)
	}

	if errBuff.Len() > 0 {
		return "", fmt.Errorf("command failed: %s", errBuff.String())
	}

	return buf.String(), err
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

func newMockCommandTextUI() *MockCommandTextUI {

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

func NewMockPackageMultiSelectUI(packages []services.PackageInfo) *MockPackageMultiSelectUI {

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

// It ensures that each command has access to the package manager name and CommandRunner

var _ = Describe("JPD Commands", func() {

	assert := assert.New(GinkgoT())

	var rootCmd *cobra.Command
	mockRunner := NewMockCommandRunner(false)

	getSubCommandWithName := func(cmd *cobra.Command, name string) (*cobra.Command, bool) {

		return lo.Find(
			cmd.Commands(),
			func(item *cobra.Command) bool {
				return item.Name() == name
			})
	}

	generateRootCommandWithPackageManagerDetector := func(mockRunner *MockCommandRunner, packageManager string, err error) *cobra.Command {
		return cmd.NewRootCmd(
			cmd.Dependencies{
				CommandRunnerGetter: func(b bool) cmd.CommandRunner {
					return mockRunner
				},
				DetectLockfile: func() (lockfile string, error error) {

					return detect.PACKAGE_LOCK_JSON, nil
				},
				DetectJSPacakgeManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {

					return packageManager, err
				},
				DetectVolta: func() bool {

					return false
				},
			})
	}

	generateRootCommandWithPackageManagerDetectedAndVoltaIsDetected := func(mockRunner *MockCommandRunner, packageManager string) *cobra.Command {
		return cmd.NewRootCmd(
			cmd.Dependencies{
				CommandRunnerGetter: func(b bool) cmd.CommandRunner {
					return mockRunner
				},
				DetectLockfile: func() (lockfile string, error error) {

					return detect.PACKAGE_LOCK_JSON, nil
				},
				DetectJSPacakgeManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {

					return packageManager, nil
				},
				DetectVolta: func() bool {

					return true
				},
			})
	}
	createRootCommandWithBunAsDefault := func(mockRunner *MockCommandRunner, err error) *cobra.Command {
		return generateRootCommandWithPackageManagerDetector(mockRunner, "bun", err)
	}
	createRootCommandWithDenoAsDefault := func(mockRunner *MockCommandRunner, err error) *cobra.Command {
		return generateRootCommandWithPackageManagerDetector(mockRunner, "deno", err)
	}

	createRootCommandWithYarnTwoAsDefault := func(mockRunner *MockCommandRunner, err error) *cobra.Command {

		return cmd.NewRootCmd(
			cmd.Dependencies{
				CommandRunnerGetter: func(b bool) cmd.CommandRunner {
					return mockRunner
				},
				DetectLockfile: func() (lockfile string, error error) {

					return "", nil
				},
				DetectJSPacakgeManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {

					return "yarn", err
				},
				YarnCommandVersionOutputter: MockYarnCommandVersionOutputer{
					version: "2.0.0",
				},

				DetectVolta: func() bool {

					return false
				},
			})
	}

	createRootCommandWithYarnOneAsDefault := func(mockRunner *MockCommandRunner, err error) *cobra.Command {
		return cmd.NewRootCmd(
			cmd.Dependencies{
				CommandRunnerGetter: func(b bool) cmd.CommandRunner {
					return mockRunner
				},
				DetectLockfile: func() (lockfile string, error error) {

					return "", nil
				},
				DetectJSPacakgeManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {

					return "yarn", err
				},
				YarnCommandVersionOutputter: MockYarnCommandVersionOutputer{
					version: "1.0.0",
				},

				DetectVolta: func() bool {

					return false
				},
			})
	}

	createRootCommandWithNoYarnVersion := func(mockRunner *MockCommandRunner, err error) *cobra.Command {
		return cmd.NewRootCmd(
			cmd.Dependencies{
				CommandRunnerGetter: func(b bool) cmd.CommandRunner {
					return mockRunner
				},
				DetectLockfile: func() (lockfile string, error error) {

					return "", nil
				},
				DetectJSPacakgeManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {

					return "yarn", err
				},

				DetectVolta: func() bool {

					return false
				},
				YarnCommandVersionOutputter: MockYarnCommandVersionOutputer{},
			})
	}

	createRootCommandWithPnpmAsDefault := func(mockRunner *MockCommandRunner, err error) *cobra.Command {
		return generateRootCommandWithPackageManagerDetector(mockRunner, "pnpm", err)
	}
	createRootCommandWithNpmAsDefault := func(mockRunner *MockCommandRunner, err error) *cobra.Command {
		return generateRootCommandWithPackageManagerDetector(mockRunner, "npm", err)
	}

	JustBeforeEach(func() {
		rootCmd = createRootCommandWithNpmAsDefault(mockRunner, nil)
		// This needs to be set because Ginkgo will pass a --test.timeout flag to the root command
		// The test.timeout flag will get in the way
		// If the args are empty before they are set by executeCommand the right args can be passed
		rootCmd.SetArgs([]string{})

	})

	AfterEach(func() {

		mockRunner.Reset()
	})

	Describe("Root Command", func() {

		It("should be able to run", func() {
			_, err := executeCmd(rootCmd, "")
			assert.NoError(err)
		})

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "--help")
			assert.NoError(err)
			assert.Contains(output, "JavaScript Package Delegator")
			assert.Contains(output, "jpd")
		})

		Context("CWD Flag (-C)", func() {
			var currentRootCmd *cobra.Command
			var mockRunner *MockCommandRunner

			var originalCwd string

			BeforeEach(func() {
				// Save original CWD
				var err error
				originalCwd, err = os.Getwd()
				assert.NoError(err)

				currentRootCmd = cmd.NewRootCmd(cmd.Dependencies{
					CommandRunnerGetter: func(isDebug bool) cmd.CommandRunner {
						mockRunner = NewMockCommandRunner(isDebug) // MockCommandRunner initialized with the target CWD

						return mockRunner
					},
					DetectJSPacakgeManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {
						return "npm", nil
					},
					DetectLockfile: func() (lockfile string, error error) {
						return "", nil
					},
					YarnCommandVersionOutputter: detect.NewRealYarnCommandVersionRunner(),
					CommandUITexter:             newMockCommandTextUI(),
					DetectVolta:                 func() bool { return false },
					NewTaskSelectorUI:           NewMockTaskSelectUI,
					NewDependencyMultiSelectUI:  NewMockDependencySelectUI,
				})

				// Make sure mocks are properly set for the command
				// Since we're using a custom CommandRunnerGetter, we need to extract the mockRunner from it
				// and ensure it's the one we expect.
				// This is a bit of a workaround for the mock being internal to the getter func.
				// A better pattern might be to pass the mockRunner directly, or have a way to retrieve it.
				// For now, we assume the getter always returns our mockRunner and can set the internal one.
			})

			AfterEach(func() {
				// Clean up the temporary directory if it was created
				// In a real scenario, you'd want to remove tempDir here.
				// For now, we'll rely on the OS to clean up temp files eventually.

				// Restore original CWD, though not strictly necessary if test doesn't change it.
				os.Chdir(originalCwd)
			})

			It("should reject a --cwd flag value that does not end with '/'", func() {
				invalidPath := "/tmp/my-project" // Missing trailing slash
				_, err := executeCmd(currentRootCmd, "--agent", "npm", "--cwd", invalidPath)
				assert.Error(err)
				assert.Contains(err.Error(), "is not a valid POSIX/UNIX folder path (must end with '/' unless it's just '/')")
				assert.Contains(err.Error(), "cwd")       // Check that the flag name is mentioned
				assert.Contains(err.Error(), invalidPath) // Check that the invalid path is mentioned
			})

			It("should reject a --cwd flag value that is a filename", func() {
				invalidPath := "my-file.txt" // A file-like path
				_, err := executeCmd(currentRootCmd, "--agent", "npm", "-C", invalidPath)
				assert.Error(err)
				assert.Contains(err.Error(), "is not a valid POSIX/UNIX folder path (must end with '/' unless it's just '/')")
				assert.Contains(err.Error(), "cwd")
				assert.Contains(err.Error(), invalidPath)
			})

			DescribeTable(
				"should reject invalid --cwd flag values",
				func(invalidPath string, expectedErrors ...string) {
					_, err := executeCmd(currentRootCmd, "--agent", "npm", "--cwd", invalidPath)
					assert.Error(err)
					for _, expectedErr := range expectedErrors {
						assert.Contains(err.Error(), expectedErr)
					}
				},
				Entry("an empty string", "", "The cwd flag cannot be empty or contain only whitespace"),
				Entry("a string with only whitespace", "   ", "The cwd flag cannot be empty or contain only whitespace"),
				Entry("a path with invalid characters", "/path/with:colon/", "is not a valid POSIX/UNIX folder path", "cwd", "/path/with:colon/"),
			)

			DescribeTable(
				"should accept valid folder paths for --cwd",
				func(validPath string) {
					_, err := executeCmd(currentRootCmd, "--agent", "npm", "--cwd", validPath)
					assert.NoError(err)
				},
				Entry("a valid root path '/'", "/"),
				Entry("a valid relative folder path './'", "./"),
				Entry("a valid relative parent folder path '../'", "../"),
			)

			It("should run a command in the specified directory using -C", func() {
				tempDir, err := os.MkdirTemp("", "jpd-cwd-test-1-*")
				assert.NoError(err)
				tempDir = fmt.Sprintf("%s/", tempDir)
				defer os.RemoveAll(tempDir) // Clean up temp directory

				// Execute a command with -C flag
				_, err = executeCmd(currentRootCmd, "install", "--agent", "npm", "-C", tempDir)
				assert.NoError(err)

				// Verify the CommandRunner received the correct working directory
				// We need to retrieve the actual mockRunner instance used by the command.
				// This requires a slight adjustment to how we get the mockRunner,
				// or make it globally accessible in the test suite for simplicity, but that's less ideal.
				// For now, we assume the CommandRunnerGetter passes our mock.
				assert.Equal("npm", mockRunner.CommandCall.Name)
				assert.Contains(mockRunner.CommandCall.Args, "install")
				assert.Equal(tempDir, mockRunner.WorkingDir)
			})

			It("should run a command in the specified directory using --cwd", func() {
				tempDir, err := os.MkdirTemp("", "jpd-cwd-test-2-*")
				assert.NoError(err)
				tempDir = fmt.Sprintf("%s/", tempDir)

				defer os.RemoveAll(tempDir) // Clean up temp directory

				_, err = executeCmd(currentRootCmd, "run", "dev", "--agent", "yarn", "--cwd", tempDir)
				assert.NoError(err)

				assert.Equal("yarn", mockRunner.CommandCall.Name)
				assert.Contains(mockRunner.CommandCall.Args, "run")
				assert.Contains(mockRunner.CommandCall.Args, "dev")
				assert.Equal(tempDir, mockRunner.WorkingDir)
			})

			It("should not set a working directory if -C is not provided", func() {
				_, err := executeCmd(currentRootCmd, "agent", "--agent", "pnpm")
				assert.NoError(err)

				assert.Equal("pnpm", mockRunner.CommandCall.Name)
				// When Dir is not explicitly set, it defaults to the current working directory of the process.
				// An empty string for WorkingDir in our mock implies it wasn't explicitly overridden by -C.
				// We should verify that `e.cmd.Dir` remains unset (empty string) in the mock,
				// which indicates the default behavior of exec.Command().
				assert.Empty(mockRunner.WorkingDir) // Assert that it's empty, implying default behavior
			})

			It("should handle a non-existent directory gracefully (likely fail at exec.Command level)", func() {
				nonExistentDir := "/non/existent/path/for/jpd/test" // A path that should not exist
				_, err := executeCmd(currentRootCmd, "install", "--agent", "npm", "-C", fmt.Sprintf("%s/", nonExistentDir))
				// Expect an error because the directory doesn't exist. The error will come from os/exec.Command.
				assert.Error(err)
				assert.Contains(err.Error(), "no such file or directory") // Specific error message for non-existent path
			})
		})

		Context("How it responds to Package Detection Failure", func() {
			var commandTextUI *MockCommandTextUI
			var currentCommand *cobra.Command
			generateRootCommandUsingPreSelectedValues := func() *cobra.Command {
				commandTextUI = newMockCommandTextUI()

				return cmd.NewRootCmd(
					cmd.Dependencies{
						CommandRunnerGetter: func(b bool) cmd.CommandRunner {
							return mockRunner
						},
						DetectLockfile: func() (lockfile string, error error) {
							return "", nil
						},
						DetectJSPacakgeManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {

							return "", detect.ErrNoPackageManager
						},
						CommandUITexter: commandTextUI,
					},
				)

			}

			BeforeEach(func() {

				currentCommand = generateRootCommandUsingPreSelectedValues()
				currentCommand.SetContext(context.Background())
				currentCommand.ParseFlags([]string{})

			})

			It(
				`propmts the user for which command they would like to use to install package manager
					If the user refuses an error is produced.
				`,
				func() {

					err := currentCommand.PersistentPreRunE(currentCommand, []string{})
					assert.Error(err)

				},
			)

			It(
				"executes the command given when the user gives a correct command",
				func() {

					commandString := "winget install pnpm.pnpm"

					commandTextUI.SetValue(commandString)

					error := currentCommand.PersistentPreRunE(currentCommand, []string{})

					assert.NoError(error)

					re := regexp.MustCompile(`\s+`)
					splitCommandString := re.Split(commandString, -1)

					assert.True(mockRunner.HasBeenCalled)
					assert.Equal(mockRunner.CommandCall.Name, splitCommandString[0])
					assert.Equal(mockRunner.CommandCall.Args, splitCommandString[1:])

				},
			)

			DescribeTable(
				"executes the command based typical instaltion commands",
				func(inputCommand string, expectedCommandName string, expectedCommandArgs []string) {

					commandTextUI.SetValue(inputCommand)

					err := currentCommand.PersistentPreRunE(currentCommand, []string{})

					assert.NoError(err)

					assert.Equal(expectedCommandName, mockRunner.CommandCall.Name)

					assert.Equal(expectedCommandArgs, mockRunner.CommandCall.Args)

				},
				Entry("Using npm to install npm globally", "npm install -g npm", "npm", []string{"install", "-g", "npm"}),
				Entry("Using yarn to add yarn globally", "yarn global add yarn", "yarn", []string{"global", "add", "yarn"}),
				Entry("Using pnpm to add pnpm globally", "pnpm add -g pnpm", "pnpm", []string{"add", "-g", "pnpm"}),
				Entry("Using bun to install bun globally", "bun install -g bun", "bun", []string{"install", "-g", "bun"}),
				Entry("Using deno to install deno CLI tool", "deno install --allow-net --allow-read deno", "deno", []string{"install", "--allow-net", "--allow-read", "deno"}),
				Entry("Using apt-get to install nodejs", "sudo apt-get install nodejs", "sudo", []string{"apt-get", "install", "nodejs"}),
				Entry("Using brew to install pnpm", "brew install pnpm", "brew", []string{"install", "pnpm"}),
				Entry("Using choco to install nodejs", "choco install nodejs", "choco", []string{"install", "nodejs"}),
				Entry("Using winget to install VSCode", "winget install Microsoft.VisualStudioCode", "winget", []string{"install", "Microsoft.VisualStudioCode"}),
				Entry("Using pacman to install git", "sudo pacman -S git", "sudo", []string{"pacman", "-S", "git"}),
				Entry("Using dnf to install yarn", "sudo dnf install yarn", "sudo", []string{"dnf", "install", "yarn"}),
				Entry("Using yum to install nodejs", "sudo yum install nodejs", "sudo", []string{"yum", "install", "nodejs"}),
				Entry("Using zypper to install pnpm", "sudo zypper install pnpm", "sudo", []string{"zypper", "install", "pnpm"}),
				Entry("Using apk to add deno", "sudo apk add deno", "sudo", []string{"apk", "add", "deno"}),
				Entry("Using nix-env to install nodejs", "nix-env -iA nixpkgs.nodejs", "nix-env", []string{"-iA", "nixpkgs.nodejs"}),
				Entry("Using nix profile to install yarn", "nix profile install nixpkgs#yarn", "nix", []string{"profile", "install", "nixpkgs#yarn"}),
			)
		})

		Context("It processes the JPD_AGENT variable properly", func() {

			var currentRootCmd *cobra.Command // Renamed to avoid shadowing and clarify intent
			var mockCommandUITexter *MockCommandTextUI
			var mockYarnVersionOutputter *MockYarnCommandVersionOutputer

			BeforeEach(func() {
				mockCommandUITexter = newMockCommandTextUI()
				mockYarnVersionOutputter = &MockYarnCommandVersionOutputer{version: "1.0.0"} // Default mock for yarn version

				// Create the root command with *all* necessary dependencies
				currentRootCmd = cmd.NewRootCmd(cmd.Dependencies{
					CommandRunnerGetter: func(b bool) cmd.CommandRunner {
						return mockRunner
					},
					DetectLockfile: func() (lockfile string, error error) {
						return "", nil
					},
					// Make sure detector returns an error so JPD_AGENT logic in root.go is hit
					DetectJSPacakgeManagerBasedOnLockFile: func(detectedLockFile string) (string, error) { return "", fmt.Errorf("not detected") },
					YarnCommandVersionOutputter:           mockYarnVersionOutputter,
					CommandUITexter:                       mockCommandUITexter,
				})
				// Must set context because the background isn't activated.
				currentRootCmd.SetContext(context.Background())
				// No need to SetArgs here if we are directly calling PersistentPreRunE

				// This must be set so that the debug flag can be used!
				currentRootCmd.ParseFlags([]string{})

			})

			AfterEach(func() {
				// Ensure environment variable is always cleaned up after each test
				os.Unsetenv(cmd.JPD_AGENT_ENV_VAR)
			})

			It("shows an error when the env JPD_AGENT is set but it's not one of the supported JS package managers", func() {
				os.Setenv(cmd.JPD_AGENT_ENV_VAR, "boo baa")

				// Directly call PersistentPreRunE and capture the error
				err := currentRootCmd.PersistentPreRunE(currentRootCmd, []string{})
				assert.Error(err)
				assert.Contains(err.Error(), fmt.Sprintf("The %s variable is set the wrong way", cmd.JPD_AGENT_ENV_VAR))
				// Verify that the command runner was not called for installation since an invalid agent was set
				assert.False(mockRunner.HasBeenCalled)
			})

			It("sets the package name when the agent is a valid value", func() {
				const expected = "deno"
				os.Setenv(cmd.JPD_AGENT_ENV_VAR, expected)

				// Directly call PersistentPreRunE and capture the error
				err := currentRootCmd.PersistentPreRunE(currentRootCmd, []string{})
				assert.NoError(err)

				// Now, the context of currentRootCmd should have the value set by PersistentPreRunE
				pm, error := currentRootCmd.Flags().GetString(cmd.AGENT_FLAG)
				assert.NoError(error, "The package name was not found in context")
				assert.Equal(expected, pm)

				// Verify that no commands were executed by the mock runner because JPD_AGENT was set
				assert.False(mockRunner.HasBeenCalled)
			})
		})

	})

	Describe("Install Command", func() {

		createRootCommandWithPackageManagerAndMultiSelectUI := func(mockRunner *MockCommandRunner) *cobra.Command {

			return cmd.NewRootCmd(
				cmd.Dependencies{
					CommandRunnerGetter: func(b bool) cmd.CommandRunner {
						return mockRunner
					},
					DetectLockfile: func() (lockfile string, error error) {
						return "", nil
					},
					DetectJSPacakgeManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {
						return "npm", nil
					},
					NewPackageMultiSelectUI: func(pi []services.PackageInfo) cmd.MultiUISelecter {

						return NewMockPackageMultiSelectUI(pi)

					},
					DetectVolta: func() bool {
						return false
					},
				})
		}

		Context("Works with the search flag", func() {

			var rootCmd *cobra.Command

			BeforeEach(func() {
				rootCmd = createRootCommandWithPackageManagerAndMultiSelectUI(mockRunner)
			})

			It("returns err when value isn't passed to the flag ", func() {

				_, err := executeCmd(rootCmd, "install", "--search")

				assert.Error(err)

			})

			It("returns err when an argument is passed when the flag is passed", func() {

				_, err := executeCmd(rootCmd, "install", "vue", "--search", "foo")

				assert.Error(err)
				assert.ErrorContains(err, "No arguments must be passed while the search flag is used")

			})

			It("Returns an error if no packages are found", func() {

				_, err := executeCmd(rootCmd, "install", "--search", "89ispsnsnis")

				assert.Error(err)
				assert.ErrorContains(err, "Your query has failed 89ispsnsnis")

			})

			It("works", func() {

				const expected = "angular"
				_, err := executeCmd(rootCmd, "install", "--search", expected)

				assert.NoError(err)

				assert.Equal("npm", mockRunner.CommandCall.Name)

				assert.Contains(mockRunner.CommandCall.Args, "install")
				assert.NotContains(mockRunner.CommandCall.Args, "--search")
				assert.Conditionf(func() bool {
					return lo.SomeBy(mockRunner.CommandCall.Args, func(item string) bool {
						return strings.Contains(item, expected)
					})
				},
					"The args are supposed to contain a word that is a part of this word %s",
					expected,
				)

			})

		})

		var installCmd *cobra.Command
		BeforeEach(func() {
			installCmd, _ = getSubCommandWithName(rootCmd, "install")
		})

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "install", "--help")
			assert.NoError(err)
			assert.Contains(output, "Install packages")
			assert.Contains(output, "jpd install")
		})

		It("should have correct aliases", func() {
			assert.Contains(installCmd.Aliases, "i")
			assert.Contains(installCmd.Aliases, "add")
		})

		It("should have dev flag", func() {
			flag := installCmd.Flag("dev")
			assert.NotNil(flag)
			assert.Equal("D", flag.Shorthand)
		})

		It("should have global flag", func() {
			flag := installCmd.Flag("global")
			assert.NotNil(flag)
			assert.Equal("g", flag.Shorthand)
		})

		It("should have production flag", func() {
			flag := installCmd.Flag("production")
			assert.NotNil(flag)
			assert.Equal("P", flag.Shorthand)
		})

		It("should have frozen flag", func() {
			flag := installCmd.Flag("frozen")
			assert.NotNil(flag)
		})

		Context("Volta", func() {
			DescribeTable(
				"Appends volta run when a node package manager is the agent",
				func(packageManager string) {

					rootCommmand := generateRootCommandWithPackageManagerDetectedAndVoltaIsDetected(mockRunner, packageManager)

					output, error := executeCmd(rootCommmand, "install")

					assert.NoError(error)
					assert.Empty(output)

					assert.Equal("volta", mockRunner.CommandCall.Name)

					assert.Equal([]string{"run", packageManager, "install"}, mockRunner.CommandCall.Args)

				},
				EntryDescription("Volta run was appended to %s"),
				Entry(nil, detect.NPM),
				Entry(nil, detect.YARN),
				Entry(nil, detect.PNPM),
			)

			DescribeTable(
				"Doesn't append volta run when a non-node package manager is the agent",
				func(packageManager string) {

					rootCommmand := generateRootCommandWithPackageManagerDetectedAndVoltaIsDetected(mockRunner, packageManager)

					var (
						output string
						error  error
					)

					if packageManager == detect.DENO {

						output, error = executeCmd(rootCommmand, "install", "npm:cn-efs")
						assert.NoError(error)
						assert.Empty(output)

						assert.Equal(packageManager, mockRunner.CommandCall.Name)

						assert.Equal([]string{"add", "npm:cn-efs"}, mockRunner.CommandCall.Args)

					} else {

						output, error = executeCmd(rootCommmand, "install")

						assert.NoError(error)
						assert.Empty(output)

						assert.Equal(mockRunner.CommandCall.Name, packageManager)

						assert.Equal(mockRunner.CommandCall.Args, []string{"install"})

					}

				},
				EntryDescription("Volta run was't appended to %s"),
				Entry(nil, detect.DENO),
				Entry(nil, detect.BUN),
			)

			It("rejects volta usage if the --no-volta flag is passed", func() {

				rootCommmand := generateRootCommandWithPackageManagerDetectedAndVoltaIsDetected(mockRunner, "npm")

				output, error := executeCmd(rootCommmand, "install", "--no-volta")

				assert.NoError(error)
				assert.Empty(output)

				assert.Equal("npm", mockRunner.CommandCall.Name)

				assert.Equal([]string{"install"}, mockRunner.CommandCall.Args)
			})

		})

		Context("npm", func() {
			It("should run npm install with no args", func() {
				_, err := executeCmd(rootCmd, "install")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "npm")
				assert.Contains(mockRunner.CommandCall.Args, "install")
			})

			It("should run npm install with package names", func() {
				_, err := executeCmd(rootCmd, "install", "lodash", "express")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "npm")
				assert.Contains(mockRunner.CommandCall.Args, "install")
				assert.Contains(mockRunner.CommandCall.Args, "lodash")
				assert.Contains(mockRunner.CommandCall.Args, "express")
			})

			It("should run npm install with dev flag", func() {
				_, err := executeCmd(rootCmd, "install", "--dev", "typescript")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "npm")
				assert.Contains(mockRunner.CommandCall.Args, "install")
				assert.Contains(mockRunner.CommandCall.Args, "typescript")
				assert.Contains(mockRunner.CommandCall.Args, "--save-dev")
			})

			It("should run npm install with global flag", func() {
				_, err := executeCmd(rootCmd, "install", "--global", "typescript")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "npm")
				assert.Contains(mockRunner.CommandCall.Args, "install")
				assert.Contains(mockRunner.CommandCall.Args, "typescript")
				assert.Contains(mockRunner.CommandCall.Args, "--global")
			})

			It("should run npm install with production flag", func() {
				_, err := executeCmd(rootCmd, "install", "--production")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "npm")
				assert.Contains(mockRunner.CommandCall.Args, "install")
				assert.Contains(mockRunner.CommandCall.Args, "--omit=dev")
			})

			It("should handle frozen flag with npm", func() {
				_, err := executeCmd(rootCmd, "install", "--frozen")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "npm")
				assert.Contains(mockRunner.CommandCall.Args, "install")
				assert.Contains(mockRunner.CommandCall.Args, "--package-lock-only")
			})
		})

		Context("yarn", func() {
			var yarnRootCmd *cobra.Command

			BeforeEach(func() {

				yarnRootCmd = createRootCommandWithYarnTwoAsDefault(mockRunner, nil)

			})

			It("should run yarn add with dev flag", func() {
				_, err := executeCmd(yarnRootCmd, "install", "--dev", "typescript")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "yarn")
				assert.Contains(mockRunner.CommandCall.Args, "add")
				assert.Contains(mockRunner.CommandCall.Args, "--dev")
				assert.Contains(mockRunner.CommandCall.Args, "typescript")
			})

			It("should handle global flag with yarn", func() {
				_, err := executeCmd(yarnRootCmd, "install", "--global", "typescript")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "yarn")
				assert.Contains(mockRunner.CommandCall.Args, "add")
				assert.Contains(mockRunner.CommandCall.Args, "--global")
				assert.Contains(mockRunner.CommandCall.Args, "typescript")
			})

			It("should handle production flag with yarn", func() {
				_, err := executeCmd(yarnRootCmd, "install", "--production")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "yarn")
				assert.Contains(mockRunner.CommandCall.Args, "install")
				assert.Contains(mockRunner.CommandCall.Args, "--production")
			})

			It("should handle frozen flag with yarn", func() {
				_, err := executeCmd(yarnRootCmd, "install", "--frozen")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "yarn")
				assert.Contains(mockRunner.CommandCall.Args, "install")
				assert.Contains(mockRunner.CommandCall.Args, "--frozen-lockfile")
			})

			It("should handle yarn classic with dependencies", func() {
				_, err := executeCmd(yarnRootCmd, "install", "lodash")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "yarn")
				assert.Contains(mockRunner.CommandCall.Args, "add")
				assert.Contains(mockRunner.CommandCall.Args, "lodash")
			})

			It("should handle yarn modern with dependencies", func() {
				// Test yarn version 2+ path
				_, err := executeCmd(yarnRootCmd, "install", "--dev", "typescript")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "yarn")
				assert.Contains(mockRunner.CommandCall.Args, "add")
				assert.Contains(mockRunner.CommandCall.Args, "--dev")
				assert.Contains(mockRunner.CommandCall.Args, "typescript")
			})
		})

		Context("pnpm", func() {
			var pnpmRootCmd *cobra.Command

			BeforeEach(func() {
				pnpmRootCmd = createRootCommandWithPnpmAsDefault(mockRunner, nil)
			})

			It("should run pnpm add with dev flag", func() {
				_, err := executeCmd(pnpmRootCmd, "install", "--dev", "typescript")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "pnpm")
				assert.Contains(mockRunner.CommandCall.Args, "add")
				assert.Contains(mockRunner.CommandCall.Args, "--save-dev")
				assert.Contains(mockRunner.CommandCall.Args, "typescript")
			})

			It("should handle global flag with pnpm", func() {
				_, err := executeCmd(pnpmRootCmd, "install", "--global", "typescript")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "pnpm")
				assert.Contains(mockRunner.CommandCall.Args, "add")
				assert.Contains(mockRunner.CommandCall.Args, "typescript")
				assert.Contains(mockRunner.CommandCall.Args, "--global")
			})

			It("should handle production flag with pnpm", func() {
				_, err := executeCmd(pnpmRootCmd, "install", "--production")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "pnpm")
				assert.Contains(mockRunner.CommandCall.Args, "install")
				assert.Contains(mockRunner.CommandCall.Args, "--prod")
			})

			It("should handle frozen flag with pnpm", func() {
				_, err := executeCmd(pnpmRootCmd, "install", "--frozen")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "pnpm")
				assert.Contains(mockRunner.CommandCall.Args, "install")
				assert.Contains(mockRunner.CommandCall.Args, "--frozen-lockfile")
			})

			It("should handle pnpm with dev dependencies", func() {
				_, err := executeCmd(pnpmRootCmd, "install", "-D", "jest")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "pnpm")
				assert.Contains(mockRunner.CommandCall.Args, "add")
				assert.Contains(mockRunner.CommandCall.Args, "jest")
				assert.Contains(mockRunner.CommandCall.Args, "--save-dev")
			})
		})

		Context("bun", func() {
			var bunRootCmd *cobra.Command

			BeforeEach(func() {
				bunRootCmd = createRootCommandWithBunAsDefault(mockRunner, nil)
			})

			It("should handle bun dev flag", func() {
				_, err := executeCmd(bunRootCmd, "install", "--dev", "typescript")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "bun")
				assert.Contains(mockRunner.CommandCall.Args, "add")
				assert.Contains(mockRunner.CommandCall.Args, "typescript")
				assert.Contains(mockRunner.CommandCall.Args, "--development")
			})

			It("should handle global flag with bun", func() {
				_, err := executeCmd(bunRootCmd, "install", "--global", "typescript")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "bun")
				assert.Contains(mockRunner.CommandCall.Args, "add")
				assert.Contains(mockRunner.CommandCall.Args, "typescript")
				assert.Contains(mockRunner.CommandCall.Args, "--global")
			})

			It("should handle production flag with bun", func() {
				_, err := executeCmd(bunRootCmd, "install", "--production")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "bun")
				assert.Contains(mockRunner.CommandCall.Args, "install")
				assert.Contains(mockRunner.CommandCall.Args, "--production")
			})

			It("should handle bun with dependencies", func() {
				_, err := executeCmd(bunRootCmd, "install", "react")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "bun")
				assert.Contains(mockRunner.CommandCall.Args, "add")
				assert.Contains(mockRunner.CommandCall.Args, "react")
			})
		})

		Context("deno", func() {
			var denoRootCmd *cobra.Command

			BeforeEach(func() {
				denoRootCmd = createRootCommandWithDenoAsDefault(mockRunner, nil)
			})

			It("should return an error if no packages are provided for a non-global install", func() {
				_, err := executeCmd(denoRootCmd, "install")
				assert.Error(err)
				assert.Contains(err.Error(), "For deno one or more packages is required")
				assert.False(mockRunner.HasBeenCalled)
			})

			It("should return an error if --global flag is used without packages", func() {
				// This case should still trigger the "one or more packages" error first
				_, err := executeCmd(denoRootCmd, "install", "--global")
				assert.Error(err)
				assert.Contains(err.Error(), "For deno one or more packages is required")
				assert.False(mockRunner.HasBeenCalled)
			})

			It("should execute deno install with --global flag and packages", func() {
				_, err := executeCmd(denoRootCmd, "install", "--global", "my-global-tool")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("deno", "install", "my-global-tool"))
			})

			It("should return an error when --production flag is used", func() {
				// Pass a package to ensure it bypasses the "no packages" error and hits the "production" error
				_, err := executeCmd(denoRootCmd, "install", "--production", "my-package")
				assert.Error(err)
				assert.Contains(err.Error(), "Deno doesn't support prod!")
				assert.False(mockRunner.HasBeenCalled)
			})

			It("should execute deno add with no flags and packages", func() {
				_, err := executeCmd(denoRootCmd, "install", "my-package")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("deno", "add", "my-package"))
			})

			It("should execute deno add with --dev flag and packages", func() {
				_, err := executeCmd(denoRootCmd, "install", "--dev", "my-dev-dep")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("deno", "add", "my-dev-dep", "--dev"))
			})

			It("should return an error if --dev flag is used without packages", func() {
				// This case should still trigger the "one or more packages" error first
				_, err := executeCmd(denoRootCmd, "install", "--dev")
				assert.Error(err)
				assert.Contains(err.Error(), "For deno one or more packages is required")
				assert.False(mockRunner.HasBeenCalled)
			})
		})

		Context("Error Handling", func() {
			It("should return error for unsupported package manager", func() {
				rootCmd := generateRootCommandWithPackageManagerDetector(mockRunner, "unknown", nil)
				_, err := executeCmd(rootCmd, "install", "lodash")
				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager: unknown")
			})

			It("should return error when command runner fails", func() {
				rootCmd := createRootCommandWithNpmAsDefault(mockRunner, nil)
				mockRunner.InvalidCommands = []string{"npm"}
				_, err := executeCmd(rootCmd, "install")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})
		})
	})

	Describe("Run Command", func() {

		var runCmd *cobra.Command

		BeforeEach(func() {
			runCmd, _ = getSubCommandWithName(rootCmd, "run")
		})

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "run", "--help")
			assert.NoError(err)
			assert.Contains(output, "Run package.json scripts")
			assert.Contains(output, "jpd run", "No jpd run")
		})

		It("should have correct aliases", func() {
			assert.Contains(runCmd.Aliases, "r")
		})

		It("should have if-present flag", func() {
			flag := runCmd.Flag("if-present")
			assert.NotNil(flag)
		})

		Context(
			"How it responds if there are no arguments",
			func() {

				createRootCommandWithTaskSelectorUI := func(mockRunner *MockCommandRunner, packageManager string) *cobra.Command {
					return cmd.NewRootCmd(
						cmd.Dependencies{
							CommandRunnerGetter: func(b bool) cmd.CommandRunner {
								return mockRunner
							},
							DetectLockfile: func() (lockfile string, error error) {
								return "", nil
							},
							DetectJSPacakgeManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {
								return packageManager, nil // Or "deno" depending on the test context
							},
							NewTaskSelectorUI: NewMockTaskSelectUI,
							DetectVolta: func() bool {
								return false
							},
						})
				}

				var (
					tempDir string
					error   error
					cwd     string
					rootCmd *cobra.Command
				)
				BeforeEach(func() {

					cwd, error = os.Getwd()
					assert.NoError(error)

					rootCmd = createRootCommandWithTaskSelectorUI(mockRunner, "npm")
					tempDir, error = os.MkdirTemp("", "jpd-test-*")

					assert.NoError(error)
					assert.DirExists(tempDir)

					os.Chdir(tempDir)

				})

				AfterEach(func() {
					os.Chdir(cwd)
					os.RemoveAll(tempDir)
				})

				It("Should output an indicator saying there are no tasks in deno for deno.json", func() {

					rootCmdWithDenoAsDefault := createRootCommandWithTaskSelectorUI(mockRunner, "deno")

					err := os.WriteFile("deno.json", []byte(
						`{
							"tasks": {

								}
									}
						`),
						os.ModePerm,
					)

					assert.NoError(err)

					_, err = executeCmd(rootCmdWithDenoAsDefault, "run")

					assert.Error(err)
					assert.ErrorContains(err, "No tasks found in deno.json")
				})

				It(
					"prompts the user to select a task from deno.json if pkg is deno",
					func() {

						tasks := map[string]string{
							"dev":   "deno run -A --watch main.ts",
							"build": "deno compile --output my_app main.ts",
							"test":  "deno test",
						}

						result, error := json.Marshal(tasks)
						assert.NoError(error)

						formattedString := fmt.Sprintf(
							`{"tasks": %s }`,
							string(result),
						)

						err := os.WriteFile(
							"deno.json",
							[]byte(formattedString),
							os.ModePerm,
						)

						assert.NoError(err)

						rootCmdWithDenoAsDefault := createRootCommandWithTaskSelectorUI(mockRunner, "deno")

						_, err = executeCmd(rootCmdWithDenoAsDefault, "run")

						assert.NoError(err)

						assert.Equal("deno", mockRunner.CommandCall.Name)

						taskNames := lo.Keys(tasks)

						assert.True(
							lo.Contains(taskNames, mockRunner.CommandCall.Args[1]),
							fmt.Sprintf("The task name isn't one of those tasks %v", taskNames),
						)

					},
				)

				It(
					"returns an error If there is no tasks avaliable",
					func() {

						err := os.WriteFile("package.json", []byte(
							`{
								"scripts": {

									}
										}
							`),
							os.ModePerm,
						)

						assert.NoError(err)

						_, err = executeCmd(rootCmd, "run")

						assert.Error(err)
						assert.Contains(err.Error(), "No scripts found in package.json")
					},
				)

				It(
					"prompts the user to select a task from package .json",
					func() {

						tasks := map[string]string{
							"dev":   "vite",
							"build": "vite build",
							"test":  "vitest",
						}

						result, error := json.Marshal(tasks)
						assert.NoError(error)

						formattedString := fmt.Sprintf(
							`{"scripts": %s }`,
							string(result),
						)

						err := os.WriteFile(
							"package.json",
							[]byte(formattedString),
							os.ModePerm,
						)

						assert.NoError(err)

						_, err = executeCmd(rootCmd, "run")

						assert.NoError(err)

						assert.Equal("npm", mockRunner.CommandCall.Name)

						taskNames := lo.Keys(tasks)

						assert.True(
							lo.Contains(taskNames, mockRunner.CommandCall.Args[1]),
							fmt.Sprintf("The task name isn't one of those tasks %v", taskNames),
						)

					},
				)

			},
		)

		Context("npm", func() {

			It("should run npm run with script name", func() {
				originalDir, _ := os.Getwd()
				path, _ := os.MkdirTemp("", "jpd-test*")
				os.Chdir(path)
				defer os.Chdir(originalDir)
				defer os.RemoveAll(path)

				content := `{ "scripts": { "test": "echo 'test'" } }`
				os.WriteFile("package.json", []byte(content), 0644)
				os.WriteFile(".env", []byte("GO_ENV=development"), 0644)

				_, err := executeCmd(rootCmd, "run", "test")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("npm", "run", "test"))
			})

			It("should run npm run with script args", func() {
				_, err := executeCmd(rootCmd, "run", "test", "--", "--watch")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("npm", "run", "test", "--", "--watch"))
			})

			It("should run npm run with if-present flag", func() {
				originalDir, _ := os.Getwd()
				path, _ := os.MkdirTemp("", "jpd-test*")
				os.Chdir(path)
				defer os.Chdir(originalDir)
				defer os.RemoveAll(path)

				content := `{ "scripts": { "test": "echo 'test'" } }`
				os.WriteFile("package.json", []byte(content), 0644)
				os.WriteFile(".env", []byte("GO_MODE=development"), 0644)

				_, err := executeCmd(rootCmd, "run", "--if-present", "test")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("npm", "run", "--if-present", "test"))
			})

			It("should handle if-present flag with non-existent script", func() {
				err := os.WriteFile("package.json", []byte(`{"name": "test", "scripts": {}}`), 0644)
				assert.NoError(err)
				defer os.Remove("package.json")

				_, err = executeCmd(rootCmd, "run", "--if-present", "nonexistent")
				assert.NoError(err) // Should not error with --if-present
			})

			It("should handle missing package.json with if-present", func() {
				os.Remove("package.json") // Ensure no package.json exists

				_, err := executeCmd(rootCmd, "run", "--if-present", "test")
				assert.Error(err) // Should error with --if-present when no package.json
			})

			It("should handle script not found without if-present", func() {
				err := os.WriteFile("package.json", []byte(`{"name": "test", "scripts": {"build": "echo building"}}`), 0644)
				assert.NoError(err)
				defer os.Remove("package.json")

				_, err = executeCmd(rootCmd, "run", "nonexistent")
				assert.NoError(err) // This behavior might be unexpected but matches original code.
			})
		})

		Context("yarn", func() {
			var yarnRootCmd *cobra.Command

			BeforeEach(func() {
				yarnRootCmd = createRootCommandWithYarnTwoAsDefault(mockRunner, nil)
			})

			It("should run yarn run with script name", func() {
				_, err := executeCmd(yarnRootCmd, "run", "test")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "run", "test"))
			})
		})

		Context("pnpm", func() {
			var pnpmRootCmd *cobra.Command

			BeforeEach(func() {
				pnpmRootCmd = createRootCommandWithPnpmAsDefault(mockRunner, nil)
			})

			It("should run pnpm script using the if-present flag", func() {
				cwd, _ := os.Getwd()
				jpdDir, _ := os.MkdirTemp(cwd, "jpd-test")
				os.Chdir(jpdDir)
				os.WriteFile("package.json", []byte(`{"scripts": {"test": "echo 'test'"}}`), 0644)
				defer os.Chdir(cwd)
				defer os.RemoveAll(jpdDir)

				_, err := executeCmd(pnpmRootCmd, "run", "--if-present", "test")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("pnpm", "run", "--if-present", "test"))
			})

			It("should run pnpm run with script args", func() {
				_, err := executeCmd(pnpmRootCmd, "run", "test", "--", "--watch")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("pnpm", "run", "test", "--", "--watch"))
			})
		})

		Context("bun", func() {
			var bunRootCmd *cobra.Command

			BeforeEach(func() {
				bunRootCmd = createRootCommandWithBunAsDefault(mockRunner, nil)
			})

			It("should handle bun run command", func() {
				_, err := executeCmd(bunRootCmd, "run", "test")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("bun", "run", "test"))
			})
		})

		Context("deno", func() {
			var denoRootCmd *cobra.Command

			BeforeEach(func() {
				denoRootCmd = createRootCommandWithDenoAsDefault(mockRunner, nil)
			})

			It("should return an error if deno is the package manager and the eval flag is passed", func() {
				cwd, _ := os.Getwd()
				jpdDir, _ := os.MkdirTemp(cwd, "jpd-test")
				os.Chdir(jpdDir)
				os.WriteFile("deno.json", []byte(`{"tasks": {"test": "vitest"}}`), 0644)
				defer os.Chdir(cwd)
				defer os.RemoveAll(jpdDir)

				_, err := executeCmd(denoRootCmd, "run", "--", "test", "--eval")
				assert.Error(err)
				assert.Contains(err.Error(), fmt.Sprintf("Don't pass %s here use the exec command instead", "--eval"))
			})

			It("should run deno task with script name", func() {
				_, err := executeCmd(denoRootCmd, "run", "test")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("deno", "task", "test"))
			})
		})

		Context("Error Handling", func() {
			It("should return error when command runner fails", func() {
				rootCmd := createRootCommandWithNpmAsDefault(mockRunner, nil)
				mockRunner.InvalidCommands = []string{"npm"}
				_, err := executeCmd(rootCmd, "run", "test")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})

			It("should handle package.json reading error", func() {
				err := os.WriteFile("package.json", []byte("invalid json"), 0644)
				assert.NoError(err)
				defer os.Remove("package.json")

				_, err = executeCmd(rootCmd, "run", "--if-present", "test")
				assert.Error(err)
			})

			It("should return error for unsupported package manager", func() {
				rootCmd := generateRootCommandWithPackageManagerDetector(mockRunner, "unknown", nil)
				_, err := executeCmd(rootCmd, "run", "test")
				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager: unknown")
			})
		})
	})

	Describe("Exec Command", func() {
		var execCmd *cobra.Command
		BeforeEach(func() {
			execCmd, _ = getSubCommandWithName(rootCmd, "exec")
		})

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "exec", "--help")
			assert.NoError(err)
			assert.Contains(output, "Execute packages")
			assert.Contains(output, "jpd exec")
		})

		It("should have correct aliases", func() {
			assert.Contains(execCmd.Aliases, "x")
		})

		It("should require at least one argument", func() {
			_, err := executeCmd(rootCmd, "exec")
			assert.Error(err)
		})

		It("should handle --help in arguments", func() {
			output, err := executeCmd(rootCmd, "exec", "some-package", "--help")
			assert.NoError(err)
			assert.Contains(output, "Execute packages")
		})

		Context("npm", func() {
			It("should execute npx with package name", func() {
				_, err := executeCmd(rootCmd, "exec", "create-react-app")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("npx", "create-react-app"))
			})

			It("should execute npx with package name and args", func() {
				_, err := executeCmd(rootCmd, "exec", "create-react-app", "my-app")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("npx", "create-react-app", "my-app"))
			})
		})

		Context("yarn", func() {
			var yarnRootCmd *cobra.Command

			BeforeEach(func() {

				yarnRootCmd = createRootCommandWithYarnTwoAsDefault(mockRunner, nil)
			})

			It("should execute yarn with package name (v2+)", func() {
				_, err := executeCmd(yarnRootCmd, "exec", "create-react-app")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "dlx", "create-react-app"))
			})

			It("should handle yarn version detection error (fallback to v1)", func() {
				rootYarnCommandWhereVersionRunnerErrors := createRootCommandWithNoYarnVersion(mockRunner, nil)
				_, err := executeCmd(rootYarnCommandWhereVersionRunnerErrors, "exec", "create-react-app")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "create-react-app"))
			})

			It("should handle yarn version one", func() {
				rootYarnCommandWhereVersionRunnerErrors := createRootCommandWithYarnOneAsDefault(mockRunner, nil)
				_, err := executeCmd(rootYarnCommandWhereVersionRunnerErrors, "exec", "fooo")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "fooo"))
			})
		})

		Context("pnpm", func() {
			var pnpmRootCmd *cobra.Command

			BeforeEach(func() {
				pnpmRootCmd = createRootCommandWithPnpmAsDefault(mockRunner, nil)
			})

			It("should execute pnpm dlx with package name", func() {
				_, err := executeCmd(pnpmRootCmd, "exec", "create-react-app", "my-app")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("pnpm", "dlx", "create-react-app", "my-app"))
			})
		})

		Context("bun", func() {
			var bunRootCmd *cobra.Command

			BeforeEach(func() {
				bunRootCmd = createRootCommandWithBunAsDefault(mockRunner, nil)
			})

			It("should execute bunx with package name", func() {
				_, err := executeCmd(bunRootCmd, "exec", "create-react-app", "my-app")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("bunx", "create-react-app", "my-app"))
			})
		})

		Context("deno", func() {
			var denoRootCmd *cobra.Command

			BeforeEach(func() {
				denoRootCmd = createRootCommandWithDenoAsDefault(mockRunner, nil)
			})

			It("should handle deno exec error", func() {
				_, err := executeCmd(denoRootCmd, "exec", "some-package")
				assert.Error(err)
				assert.Contains(err.Error(), "Deno doesn't have a dlx or x like the others")
			})
		})

		Context("Error Handling", func() {
			It("should handle help flag correctly (no command executed)", func() {
				rootCmd := createRootCommandWithNpmAsDefault(mockRunner, nil)
				rootCmd.SetArgs([]string{})
				_, err := executeCmd(rootCmd, "exec", "--help")
				assert.NoError(err)
				assert.False(mockRunner.HasBeenCalled) // No command should be executed if --help is present
			})

			It("should return error when command runner fails", func() {
				rootCmd := createRootCommandWithNpmAsDefault(mockRunner, nil)
				mockRunner.InvalidCommands = []string{"npx"}
				_, err := executeCmd(rootCmd, "exec", "test-command")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npx' is configured to fail")
			})

			It("should return error for unsupported package manager", func() {
				rootCmd := generateRootCommandWithPackageManagerDetector(mockRunner, "unknown", nil)
				_, err := executeCmd(rootCmd, "exec", "lodash")
				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager: unknown")
			})
		})
	})

	Describe("Update Command", func() {

		var updateCmd *cobra.Command

		BeforeEach(func() {
			updateCmd, _ = getSubCommandWithName(rootCmd, "update")
		})

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "update", "--help")
			assert.NoError(err)
			assert.Contains(output, "Update packages")
			assert.Contains(output, "jpd update")
		})

		It("should have correct aliases", func() {
			assert.Contains(updateCmd.Aliases, "u")
			assert.Contains(updateCmd.Aliases, "up")
			assert.Contains(updateCmd.Aliases, "upgrade")
		})

		It("should have interactive flag", func() {
			flag := updateCmd.Flag("interactive")
			assert.NotNil(flag)
			assert.Equal("i", flag.Shorthand)
		})

		It("should have global flag", func() {
			flag := updateCmd.Flag("global")
			assert.NotNil(flag)
			assert.Equal("g", flag.Shorthand)
		})

		It("should have latest flag", func() {
			flag := updateCmd.Flag("latest")
			assert.NotNil(flag)
		})

		Context("npm", func() {
			It("should error on npm with interactive flag", func() {
				_, err := executeCmd(rootCmd, "update", "--interactive")
				assert.Error(err)
				assert.Contains(err.Error(), "npm does not support interactive updates")
			})

			It("should run npm update with no args", func() {
				_, err := executeCmd(rootCmd, "update")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("npm", "update"))
			})

			It("should run npm update with package names", func() {
				_, err := executeCmd(rootCmd, "update", "lodash", "express")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("npm", "update", "lodash", "express"))
			})

			It("should run npm update with global flag", func() {
				_, err := executeCmd(rootCmd, "update", "--global", "typescript")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("npm", "update", "typescript", "--global"))
			})

			It("should handle latest flag for npm", func() {
				_, err := executeCmd(rootCmd, "update", "--latest", "lodash")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("npm", "install", "lodash@latest"))
			})

			It("should handle latest flag with global for npm", func() {
				_, err := executeCmd(rootCmd, "update", "--latest", "--global", "lodash")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("npm", "install", "lodash@latest", "--global"))
			})
		})

		Context("pnpm", func() {
			var pnpmRootCmd *cobra.Command

			BeforeEach(func() {
				pnpmRootCmd = createRootCommandWithPnpmAsDefault(mockRunner, nil)
			})

			It("should handle pnpm update with interactive flag", func() {
				_, err := executeCmd(pnpmRootCmd, "update", "--interactive")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("pnpm", "update", "--interactive"))
			})

			It("should handle interactive flag with pnpm with args", func() {
				_, err := executeCmd(pnpmRootCmd, "update", "--interactive", "astro")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("pnpm", "update", "--interactive", "astro"))
			})

			It("should handle pnpm update", func() {
				_, err := executeCmd(pnpmRootCmd, "update")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("pnpm", "update"))
			})

			It("should handle pnpm update with multiple args", func() {
				_, err := executeCmd(pnpmRootCmd, "update", "react")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("pnpm", "update", "react"))
			})

			It("should handle pnpm update with --global", func() {
				_, err := executeCmd(pnpmRootCmd, "update", "--global")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("pnpm", "update", "--global"))
			})

			It("should handle pnpm update with --latest", func() {
				_, err := executeCmd(pnpmRootCmd, "update", "--latest")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("pnpm", "update", "--latest"))
			})
		})

		Context("yarn", func() {
			var yarnRootCmd *cobra.Command

			BeforeEach(func() {

				yarnRootCmd = createRootCommandWithYarnTwoAsDefault(mockRunner, nil)
			})

			It("should handle yarn with specific packages", func() {
				_, err := executeCmd(yarnRootCmd, "update", "lodash")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "upgrade", "lodash"))
			})

			It("should handle interactive flag with yarn", func() {
				_, err := executeCmd(yarnRootCmd, "update", "--interactive")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "upgrade-interactive"))
			})

			It("should handle interactive flag with yarn with args", func() {
				_, err := executeCmd(yarnRootCmd, "update", "--interactive", "test")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "upgrade-interactive", "test"))
			})

			It("should handle latest flag with yarn", func() {
				_, err := executeCmd(yarnRootCmd, "update", "--latest")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "upgrade", "--latest"))
			})

			It("should handle yarn with global flag", func() {
				_, err := executeCmd(yarnRootCmd, "update", "--global")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "upgrade", "--global"))
			})

			It("should handle yarn with both interactive and latest flags", func() {
				_, err := executeCmd(yarnRootCmd, "update", "--interactive", "--latest")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "upgrade-interactive", "--latest"))
			})
		})

		Context("deno", func() {

			var denoRootCmd *cobra.Command

			BeforeEach(func() {

				denoRootCmd = createRootCommandWithDenoAsDefault(mockRunner, nil)
			})

			It("should handle deno update --interactive", func() {
				_, err := executeCmd(denoRootCmd, "update", "--interactive")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("deno", "outdated", "-i"))
			})

			It("should handle deno update", func() {
				_, err := executeCmd(denoRootCmd, "update")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("deno", "outdated"))
			})

			It("should handle deno update with multiple args using --latest", func() {
				_, err := executeCmd(denoRootCmd, "update", "react", "--latest")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("deno", "outdated", "--latest", "react"))
			})

			It("should handle deno update with --global", func() {
				_, err := executeCmd(denoRootCmd, "update", "--global")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("deno", "outdated", "--global"))
			})

			It("should handle deno update with --latest", func() {
				_, err := executeCmd(denoRootCmd, "update", "--latest")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("deno", "outdated", "--latest"))
			})

			It("should handle deno update with --latest and arguments", func() {
				_, err := executeCmd(denoRootCmd, "update", "--latest", "react")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("deno", "outdated", "--latest", "react"))
			})
		})

		Context("bun", func() {
			var bunRootCmd *cobra.Command

			BeforeEach(func() {

				bunRootCmd = createRootCommandWithBunAsDefault(mockRunner, nil)
			})

			It("should give an error update with interactive flag", func() {
				_, err := executeCmd(bunRootCmd, "update", "--interactive")
				assert.Error(err)
				assert.ErrorContains(err, "bun does not support interactive updates")
			})

			It("should handle bun update", func() {
				_, err := executeCmd(bunRootCmd, "update")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("bun", "update"))
			})

			It("should handle bun update with multiple args", func() {
				_, err := executeCmd(bunRootCmd, "update", "react")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("bun", "update", "react"))
			})

			It("should handle bun update with --global", func() {
				_, err := executeCmd(bunRootCmd, "update", "--global")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("bun", "update", "--global"))
			})

			It("should handle bun update with --latest", func() {
				_, err := executeCmd(bunRootCmd, "update", "--latest")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("bun", "update", "--latest"))
			})
		})

		Context("Error Handling", func() {
			It("should return error when command runner fails", func() {
				rootCmd := createRootCommandWithNpmAsDefault(mockRunner, nil)
				mockRunner.InvalidCommands = []string{"npm"}
				_, err := executeCmd(rootCmd, "update")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})

			It("should return error for unsupported package manager", func() {
				rootCmd := generateRootCommandWithPackageManagerDetector(mockRunner, "unknown", nil)
				_, err := executeCmd(rootCmd, "update")
				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager: unknown")
			})
		})
	})

	Describe("Uninstall Command", func() {

		var uninstallCmd *cobra.Command

		BeforeEach(func() {
			uninstallCmd, _ = getSubCommandWithName(rootCmd, "uninstall")
		})

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "uninstall", "--help")
			assert.NoError(err)
			assert.Contains(output, "Uninstall packages")
			assert.Contains(output, "jpd uninstall")
		})

		It("should have correct aliases", func() {
			assert.Contains(uninstallCmd.Aliases, "un")
			assert.Contains(uninstallCmd.Aliases, "remove")
			assert.Contains(uninstallCmd.Aliases, "rm")
		})

		It("should have global flag", func() {
			flag := uninstallCmd.Flag("global")
			assert.NotNil(flag)
			assert.Equal("g", flag.Shorthand)
		})

		It("should require at least one argument", func() {
			_, error := executeCmd(rootCmd, "uninstall")
			assert.Error(error)
		})

		Context(
			"Interactive Uninstall",
			func() {

				var (
					tempDir string
					err     error
					cwd     string
				)

				BeforeEach(func() {

					cwd, err = os.Getwd()
					assert.NoError(err)

					tempDir, err = os.MkdirTemp("", "jpd-test-*")
					assert.NoError(err)
					assert.DirExists(tempDir)

					os.Chdir(tempDir)

				})

				AfterEach(func() {
					os.Chdir(cwd)
					os.RemoveAll(tempDir)
				})

				It("should return an error if no packages are found for interactive uninstall", func() {
					//Create a package.json with no dependencies/devDependencies
					err := os.WriteFile("package.json", []byte(
						`{
							"name": "test-project",
							"version": "1.0.0",
							"dependencies": {},
							"devDependencies": {}
									}`),
						os.ModePerm,
					)
					assert.NoError(err)

					// Override the root command to ensure the MultiSelectUI is mocked properly
					// for the interactive uninstall where no packages will be returned by the detector.
					rootCmdForNoPackages := cmd.NewRootCmd(
						cmd.Dependencies{
							CommandRunnerGetter: func(b bool) cmd.CommandRunner {
								return mockRunner
							},
							DetectLockfile: func() (lockfile string, error error) {
								return "", nil
							},
							DetectJSPacakgeManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {
								return "npm", nil // Assume npm for the test
							},
						})

					_, cmdErr := executeCmd(rootCmdForNoPackages, "uninstall", "--interactive")

					assert.Error(cmdErr)
					assert.Contains(cmdErr.Error(), "No packages found for interactive uninstall.")
					assert.False(mockRunner.HasBeenCalled) // No command should be run
				})

				It(
					"should uninstall selected packages when user selects multiple packages from package.json",
					func() {
						// Create a package.json with these dependencies

						dependencies := map[string]string{
							"react":         "18.2.0",
							"react-dom":     "18.2.0",
							"react-scripts": "5.0.1",
							"lodash":        "4.17.21",
							"express":       "4.18.2",
							"vue":           "3.3.4",
							"angular":       "16.2.0",
						}

						devDependencies := map[string]string{
							"jest":       "29.7.0",
							"typescript": "5.2.2",
							"webpack":    "5.88.2",
						}

						marshalledDependencies, err := json.Marshal(devDependencies)
						assert.NoError(err)
						marshalledDevDependencies, err := json.Marshal(dependencies)
						assert.NoError(err)

						pkgJsonContent := fmt.Sprintf(
							`{
						"name": "test-project",
						"version": "1.0.0",
						"dependencies": %s,
						"devDependencies": %s
					}`,
							marshalledDependencies,
							marshalledDevDependencies,
						)

						err = os.WriteFile("package.json", []byte(pkgJsonContent), os.ModePerm)
						assert.NoError(err)

						// Override the root command to inject our custom mock UI
						rootCmdForSelection := cmd.NewRootCmd(
							cmd.Dependencies{
								CommandRunnerGetter: func(b bool) cmd.CommandRunner {
									return mockRunner
								},
								DetectLockfile: func() (lockfile string, error error) {
									return "", nil
								},
								DetectJSPacakgeManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {
									return "npm", nil // Assume npm for the test
								},
								NewDependencyMultiSelectUI: NewMockDependencySelectUI,
								DetectVolta: func() bool {
									return false
								},
							})

						_, cmdErr := executeCmd(rootCmdForSelection, "uninstall", "--interactive")

						assert.NoError(cmdErr)
						assert.True(mockRunner.HasBeenCalled)

						prodAndDevDependencies := lo.Map(
							lo.Entries(lo.Assign(dependencies, devDependencies)),
							func(entry lo.Entry[string, string], _ int) string {
								return entry.Key + "@" + entry.Value
							})

						assert.True(
							lo.Some(
								prodAndDevDependencies,
								mockRunner.CommandCall.Args,
							),
						)

					})

				It(
					"should uninstall selected packages when user selects multiple packages from deno.json",
					func() {
						// Create a package.json with these dependencies

						imports := map[string]string{
							"uuid":     "jsr:@std/uuid@^1.0.0",
							"path":     "jsr:@std/path@^1.0.0",
							"hono":     "jsr:@hono/hono@^3.12.0",
							"zod":      "jsr:@zod/zod@^3.23.8",
							"supabase": "jsr:@supabase/supabase-js@^2.43.4",
							"faker":    "jsr:@faker-js/faker@^8.4.1",
							"dotenv":   "jsr:@deno-core/dotenv@^0.5.0",
						}

						marshalledImports, err := json.Marshal(imports)
						assert.NoError(err)

						pkgJsonContent := fmt.Sprintf(
							`{
						"name": "test-project",
						"version": "1.0.0",
						"imports": %s
					}`,
							marshalledImports,
						)

						err = os.WriteFile("deno.json", []byte(pkgJsonContent), os.ModePerm)
						assert.NoError(err)

						// Override the root command to inject our custom mock UI
						rootCmdForSelection := cmd.NewRootCmd(
							cmd.Dependencies{
								CommandRunnerGetter: func(b bool) cmd.CommandRunner {
									return mockRunner
								},
								DetectLockfile: func() (lockfile string, error error) {
									return "", nil
								},
								DetectJSPacakgeManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {
									return "deno", nil // Assume npm for the test
								},
								NewDependencyMultiSelectUI: NewMockDependencySelectUI,
								DetectVolta: func() bool {
									return false
								},
							})

						_, cmdErr := executeCmd(rootCmdForSelection, "uninstall", "--interactive")

						assert.NoError(cmdErr)
						assert.True(mockRunner.HasBeenCalled)

						importsValues := lo.Values(imports)

						assert.True(
							lo.Some(
								importsValues,
								mockRunner.CommandCall.Args,
							),
						)

					})

			},
		)

		Context("deno", func() {

			var denoRootCmd *cobra.Command

			BeforeEach(func() {
				denoRootCmd = createRootCommandWithDenoAsDefault(mockRunner, nil)
			})

			It("should execute deno remove with package name", func() {
				_, err := executeCmd(denoRootCmd, "uninstall", "my_module")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("deno", "remove", "my_module"))
			})

			It("should execute deno uninstall with global flag and package name", func() {
				_, err := executeCmd(denoRootCmd, "uninstall", "--global", "my-global-tool")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("deno", "uninstall", "my-global-tool"))
			})

			It("should return an error if no packages are provided for non-global uninstall", func() {
				// The cobra.MinimumNArgs(1) check should catch this.
				_, err := executeCmd(denoRootCmd, "uninstall")
				assert.Error(err)
				assert.Contains(err.Error(), "requires at least 1 arg(s)")
				assert.False(mockRunner.HasBeenCalled)
			})

			It("should return an error if no packages are provided for global uninstall", func() {
				// The cobra.MinimumNArgs(1) check should catch this, as --global doesn't negate it.
				_, err := executeCmd(denoRootCmd, "uninstall", "--global")
				assert.Error(err)
				assert.Contains(err.Error(), "requires at least 1 arg(s)")
				assert.False(mockRunner.HasBeenCalled)
			})

			It("should return an error when both global and interactive flags are used", func() {
				_, err := executeCmd(denoRootCmd, "uninstall", "--global", "--interactive")
				assert.Error(err)
				assert.Contains(err.Error(), "if any flags in the group [global interactive] are set none of the others can be")
				assert.False(mockRunner.HasBeenCalled)
			})

		})

		Context("npm", func() {
			It("should run npm uninstall with package name", func() {
				_, err := executeCmd(rootCmd, "uninstall", "lodash")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("npm", "uninstall", "lodash"))
			})

			It("should run npm uninstall with multiple package names", func() {
				_, err := executeCmd(rootCmd, "uninstall", "lodash", "express")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("npm", "uninstall", "lodash", "express"))
			})

			It("should run npm uninstall with global flag", func() {
				_, err := executeCmd(rootCmd, "uninstall", "--global", "typescript")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("npm", "uninstall", "typescript", "--global"))
			})
		})

		Context("yarn", func() {
			var yarnRootCmd *cobra.Command

			BeforeEach(func() {

				yarnRootCmd = createRootCommandWithYarnTwoAsDefault(mockRunner, nil)
			})

			It("should handle yarn uninstall", func() {
				_, err := executeCmd(yarnRootCmd, "uninstall", "lodash")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "remove", "lodash"))
			})

			It("should handle yarn uninstall with global flag", func() {
				_, err := executeCmd(yarnRootCmd, "uninstall", "--global", "lodash")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "remove", "lodash", "--global"))
			})

			It("should run yarn remove with package name", func() {
				_, err := executeCmd(yarnRootCmd, "uninstall", "lodash")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "remove", "lodash"))
			})
		})

		Context("pnpm", func() {
			var pnpmRootCmd *cobra.Command

			BeforeEach(func() {

				pnpmRootCmd = createRootCommandWithPnpmAsDefault(mockRunner, nil)
			})

			It("should run pnpm remove with global flag", func() {
				_, err := executeCmd(pnpmRootCmd, "uninstall", "--global", "typescript")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("pnpm", "remove", "typescript", "--global"))
			})
		})

		Context("bun", func() {
			var bunRootCmd *cobra.Command

			BeforeEach(func() {

				bunRootCmd = createRootCommandWithBunAsDefault(mockRunner, nil)
			})

			It("should run bun remove with multiple packages", func() {
				_, err := executeCmd(bunRootCmd, "uninstall", "react", "react-dom")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("bun", "remove", "react", "react-dom"))
			})

			It("should handle bun uninstall with global flag", func() {
				_, err := executeCmd(bunRootCmd, "uninstall", "--global", "lodash")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("bun", "remove", "lodash", "--global"))
			})
		})

		Context("Error Handling", func() {
			It("should return error when command runner fails", func() {
				rootCmd := createRootCommandWithNpmAsDefault(mockRunner, nil)
				mockRunner.InvalidCommands = []string{"npm"}
				_, err := executeCmd(rootCmd, "uninstall", "lodash")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})

			It("should return error for unsupported package manager", func() {
				rootCmd := generateRootCommandWithPackageManagerDetector(mockRunner, "unknown", nil)
				_, err := executeCmd(rootCmd, "uninstall", "lodash")
				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager: unknown")
			})
		})
	})

	Describe("Clean Install Command", func() {

		var cleanInstallCmd *cobra.Command
		BeforeEach(func() {

			cleanInstallCmd, _ = getSubCommandWithName(rootCmd, "clean-install")
		})
		Context("Volta", func() {

			DescribeTable(
				"Appends volta run when a node package manager is the agent",
				func(packageManager string) {

					rootCommmand := generateRootCommandWithPackageManagerDetectedAndVoltaIsDetected(mockRunner, packageManager)

					output, error := executeCmd(rootCommmand, "install")

					assert.NoError(error)
					assert.Empty(output)

					assert.Equal("volta", mockRunner.CommandCall.Name)

					assert.Equal([]string{"run", packageManager, "install"}, mockRunner.CommandCall.Args)

				},
				EntryDescription("Volta run was appended to %s"),
				Entry(nil, detect.NPM),
				Entry(nil, detect.YARN),
				Entry(nil, detect.PNPM),
			)

			DescribeTable(
				"Doesn't append volta run when a non-node package manager is the agent",
				func(packageManager string) {

					rootCommmand := generateRootCommandWithPackageManagerDetectedAndVoltaIsDetected(mockRunner, packageManager)

					var (
						output string
						error  error
					)

					if packageManager == detect.DENO {

						output, error = executeCmd(rootCommmand, "install", "npm:cn-efs")
						assert.NoError(error)
						assert.Empty(output)

						assert.Equal(packageManager, mockRunner.CommandCall.Name)

						assert.Equal([]string{"add", "npm:cn-efs"}, mockRunner.CommandCall.Args)

					} else {

						output, error = executeCmd(rootCommmand, "install")

						assert.NoError(error)
						assert.Empty(output)

						assert.Equal(mockRunner.CommandCall.Name, packageManager)

						assert.Equal(mockRunner.CommandCall.Args, []string{"install"})

					}

				},
				EntryDescription("Volta run was't appended to %s"),
				Entry(nil, detect.DENO),
				Entry(nil, detect.BUN),
			)

			It("rejects volta usage if the --no-volta flag is passed", func() {

				rootCommmand := generateRootCommandWithPackageManagerDetectedAndVoltaIsDetected(mockRunner, "npm")

				output, error := executeCmd(rootCommmand, "install", "--no-volta")

				assert.NoError(error)
				assert.Empty(output)

				assert.Equal("npm", mockRunner.CommandCall.Name)

				assert.Equal([]string{"install"}, mockRunner.CommandCall.Args)
			})

		})

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "clean-install", "--help")
			assert.NoError(err)
			assert.Contains(output, "Clean install")
			assert.Contains(output, "jpd clean-install")
		})

		It("should have correct aliases", func() {
			assert.Contains(cleanInstallCmd.Aliases, "ci")
		})

		Context("npm", func() {
			It("should run npm ci", func() {
				_, err := executeCmd(rootCmd, "ci")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("npm", "ci"))
			})
		})

		Context("yarn", func() {
			var yarnRootCmd *cobra.Command

			BeforeEach(func() {
				mockRunner = NewMockCommandRunner(false)
				yarnRootCmd = createRootCommandWithYarnTwoAsDefault(mockRunner, nil)
			})

			It("should run yarn install with frozen lockfile (v1)", func() {
				yarnRootCmd = createRootCommandWithYarnOneAsDefault(mockRunner, nil)
				_, err := executeCmd(yarnRootCmd, "clean-install")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "install", "--frozen-lockfile"))
			})

			It("should handle yarn v2+ with immutable flag", func() {
				_, err := executeCmd(yarnRootCmd, "clean-install")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "install", "--immutable"))
			})
		})

		Context("pnpm", func() {
			var pnpmRootCmd *cobra.Command

			BeforeEach(func() {
				mockRunner = NewMockCommandRunner(false)
				pnpmRootCmd = createRootCommandWithPnpmAsDefault(mockRunner, nil)
			})

			It("should run pnpm install with frozen lockfile", func() {
				_, err := executeCmd(pnpmRootCmd, "clean-install")
				assert.NoError(err)
				assert.Equal([]string{"install", "--frozen-lockfile"}, mockRunner.CommandCall.Args)
			})
		})

		Context("bun", func() {
			var bunRootCmd *cobra.Command

			BeforeEach(func() {
				mockRunner = NewMockCommandRunner(false)
				bunRootCmd = createRootCommandWithBunAsDefault(mockRunner, nil)
			})

			It("should run bun install with frozen lockfile", func() {
				_, err := executeCmd(bunRootCmd, "clean-install")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("bun", "install", "--frozen-lockfile"))
			})
		})

		Context("deno", func() {
			var denoRootCmd *cobra.Command

			BeforeEach(func() {
				mockRunner = NewMockCommandRunner(false)
				denoRootCmd = createRootCommandWithDenoAsDefault(mockRunner, nil)
			})

			It("should return error for deno", func() {
				_, err := executeCmd(denoRootCmd, "clean-install")
				assert.Error(err)
				assert.Contains(err.Error(), "deno doesn't support this command")
			})
		})

		Context("Error Handling", func() {
			It("should return error for unsupported package manager", func() {
				rootCmd := generateRootCommandWithPackageManagerDetector(mockRunner, "foo", nil)
				_, err := executeCmd(rootCmd, "ci", "lodash")
				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager: foo")
			})

			It("should return error when command runner fails", func() {
				rootCmd := createRootCommandWithNpmAsDefault(mockRunner, nil)
				mockRunner.InvalidCommands = []string{"npm"}
				_, err := executeCmd(rootCmd, "clean-install")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})
		})
	})

	Describe("Agent Command", func() {

		var agentCmd *cobra.Command
		BeforeEach(func() {
			agentCmd, _ = getSubCommandWithName(rootCmd, "agent")
		})

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "agent", "--help")
			assert.NoError(err)
			assert.Contains(output, "Show information about the detected package manager")
			assert.Contains(output, "jpd agent")
		})

		It("should have correct aliases", func() {
			assert.Contains(agentCmd.Aliases, "a")
		})

		Context("General", func() {
			It("should execute detected package manager", func() {
				_, err := executeCmd(rootCmd, "agent")
				assert.NoError(err)
				assert.Contains(mockRunner.CommandCall.Name, "npm")
				assert.Equal([]string{}, mockRunner.CommandCall.Args)
			})

			It("should pass arguments to package manager", func() {
				_, err := executeCmd(rootCmd, "agent", "--", "--version")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("npm", "--version"))
			})
		})

		Context("yarn", func() {
			var yarnRootCmd *cobra.Command

			BeforeEach(func() {
				mockRunner = NewMockCommandRunner(false)
				yarnRootCmd = createRootCommandWithYarnTwoAsDefault(mockRunner, nil)
			})

			It("should execute yarn with arguments", func() {
				_, err := executeCmd(yarnRootCmd, "agent", "--", "--version")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("yarn", "--version"))
			})
		})

		Context("pnpm", func() {
			var pnpmRootCmd *cobra.Command

			BeforeEach(func() {
				mockRunner = NewMockCommandRunner(false)
				pnpmRootCmd = createRootCommandWithPnpmAsDefault(mockRunner, nil)
			})

			It("should execute pnpm with arguments", func() {
				_, err := executeCmd(pnpmRootCmd, "agent", "info")
				assert.NoError(err)
				assert.True(mockRunner.HasCommand("pnpm", "info"))
			})
		})

		Context("Error Handling", func() {
			It("should fail when command execution fails", func() {
				rootCmd := createRootCommandWithNpmAsDefault(mockRunner, nil)
				mockRunner.InvalidCommands = []string{"npm"}
				_, err := executeCmd(rootCmd, "agent")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})
		})
	})

	Describe("Command Integration", func() {
		It("should have all commands registered", func() {
			commands := rootCmd.Commands()
			commandNames := make([]string, len(commands))
			for i, cmd := range commands {
				commandNames[i] = cmd.Name()
			}

			assert.Contains(commandNames, "install")
			assert.Contains(commandNames, "run")
			assert.Contains(commandNames, "exec")
			assert.Contains(commandNames, "update")
			assert.Contains(commandNames, "uninstall")
			assert.Contains(commandNames, "clean-install")
			assert.Contains(commandNames, "agent")
		})

		It("should maintain command count", func() {
			commands := rootCmd.Commands()
			userCommands := 0
			for _, cmd := range commands {
				if cmd.Name() != "help" && cmd.Name() != "completion" {
					userCommands++
				}
			}
			assert.Equal(7, userCommands)
		})
	})

	Describe("MockCommandRunner Interface (Single Command Expected)", func() {
		It("should properly record a single command", func() {
			testRunner := NewMockCommandRunner(false)
			testRunner.Command("npm", "install", "lodash")
			err := testRunner.Run()
			assert.NoError(err)
			assert.True(testRunner.HasBeenCalled)
			assert.True(testRunner.HasCommand("npm", "install", "lodash"))
		})

		It("should return error if no command is set before run", func() {
			testRunner := NewMockCommandRunner(false)
			err := testRunner.Run()
			assert.Error(err)
			assert.Contains(err.Error(), "no command set to run")
		})

		It("should return errors for invalid commands", func() {
			testRunner := NewMockCommandRunner(false)
			testRunner.InvalidCommands = []string{"npm"}
			testRunner.Command("npm", "install")
			err := testRunner.Run()
			assert.Error(err)
			assert.Contains(err.Error(), "configured to fail")
		})

		It("should correctly check for command execution", func() {
			testRunner := NewMockCommandRunner(false)
			testRunner.Command("npm", "install", "lodash")
			testRunner.Run()

			assert.True(testRunner.HasCommand("npm", "install", "lodash"))
			assert.False(testRunner.HasCommand("yarn", "add", "react")) // Should be false as only one command is stored
		})

		It("should properly reset all state", func() {
			testRunner := NewMockCommandRunner(false)
			testRunner.Command("npm", "install", "lodash")
			testRunner.Run()
			testRunner.InvalidCommands = []string{"yarn"}

			testRunner.Reset()

			assert.False(testRunner.HasBeenCalled)
			assert.Equal(0, len(testRunner.InvalidCommands))
			assert.Equal(CommandCall{}, testRunner.CommandCall)
		})

		It("should return the last executed command", func() {
			testRunner := NewMockCommandRunner(false)

			cmdCall, exists := testRunner.LastCommand()
			assert.False(exists)
			assert.Equal(CommandCall{}, cmdCall)

			testRunner.Command("npm", "install")
			testRunner.Run()

			cmdCall, exists = testRunner.LastCommand()
			assert.True(exists)
			assert.Equal("npm", cmdCall.Name)
			assert.Equal([]string{"install"}, cmdCall.Args)

			// Running another command overwrites the previous one
			testRunner.Command("yarn", "add", "react")
			testRunner.Run()

			cmdCall, exists = testRunner.LastCommand()
			assert.True(exists)
			assert.Equal("yarn", cmdCall.Name)
			assert.Equal([]string{"add", "react"}, cmdCall.Args)
		})
	})

})
