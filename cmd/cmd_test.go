package cmd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/build_info"
	"github.com/louiss0/javascript-package-delegator/cmd"
	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/env"
	"github.com/louiss0/javascript-package-delegator/mock" // Import the mock package
	"github.com/louiss0/javascript-package-delegator/services"
	"github.com/louiss0/javascript-package-delegator/testutil"
)

// Helper function to write content to file (copied from integrate.go for testing)
func writeToFile(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			if err == nil {
				err = cerr
			} else {
				err = fmt.Errorf("%w; %w", err, cerr)
			}
		}
	}()

	_, err = file.WriteString(content)
	return err
}

// Command Mapping Tests - Pure unit tests for command building logic
// These tests directly test the command builder functions without executing commands

// Test strategy and mapping rules:
//
// EXEC (local dependencies): Run local deps using package manager's exec feature
// | Tool           | Command                                        |
// | -------------- | ---------------------------------------------- |
// | npm            | npm exec <bin> -- <args>                      |
// | pnpm           | pnpm exec <bin> <args>                        |
// | yarn (any)     | yarn <bin> <args>                             |
// | bun            | bun x <bin> <args>                            |
// | deno           | deno run <script> <args>                      |
//
// DLX (temporary packages): Run packages temporarily without install
// | Tool           | Command                                        |
// | -------------- | ---------------------------------------------- |
// | npm            | npm dlx <package> <args>                      |
// | pnpm           | pnpm dlx <package> <args>                     |
// | yarn v1        | yarn <package> <args> (no dlx)               |
// | yarn v2+       | yarn dlx <package> <args>                     |
// | bun            | bunx <package> <args>                         |
// | deno           | deno run <url> <args> (requires URL)          |

// This function executes a cobra command with the given arguments and returns the output and error.
// It sets the output and error buffers for the command, sets the arguments, and executes the command.
// If there is an error, it returns an error with the error message from the error buffer.
// If there is no error, it returns the output from the output buffer.
// It's used to test the cobra commands.
// When you use this function, make sure to pass the root command and any arguments you want to test.
// The first argument after the rootCmd is any sub command or flag you want to test.
// This function now properly preserves the command context with CommandRunner.

// Test helper functions for standard Go tests (originally from testhelpers_test.go)

// makeTempDir creates a temporary directory for the test

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

// It ensures that each command has access to the package manager name and CommandRunner

var _ = Describe("JPD Commands", func() {

	assert := assert.New(GinkgoT())

	var rootCmd *cobra.Command
	mockCommandRunner := mock.NewMockCommandRunner()
	factory := testutil.NewRootCommandFactory(mockCommandRunner) // Initialize the factory
	var DebugExecutorExpectationManager = testutil.DebugExecutorExpectationManager

	getSubCommandWithName := func(cmd *cobra.Command, name string) (*cobra.Command, bool) {

		return lo.Find(
			cmd.Commands(),
			func(item *cobra.Command) bool {
				return item.Name() == name
			})
	}

	BeforeEach(func() {

		// Clear any state from previous tests to prevent cross-contamination
		mockCommandRunner.InvalidCommands = []string{}
		mockCommandRunner.ResetHasBeenCalled()
		// Set up basic mock expectations before each test
		factory.SetupBasicCommandRunnerExpectations()
		// Reset first, then setup debug expectations
		factory.ResetDebugExecutor()

		DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
		factory.SetupBasicDebugExecutorExpectations()
		rootCmd = factory.CreateNpmAsDefault(nil)
		// This needs to be set because Ginkgo will pass a --test.timeout flag to the root command
		// The test.timeout flag will get in the way
		// If the args are empty before they are set by executeCommand the right args can be passed
		rootCmd.SetArgs([]string{})

	})

	AfterEach(func() {

		// Assert that all expectations were met
		mockCommandRunner.AssertExpectations(GinkgoT())
		factory.DebugExecutor().AssertExpectations(GinkgoT())
	})

	const DebugFlagOnSubCommands = "Debug flag on sub commands"
	Describe(DebugFlagOnSubCommands, func() {

		It("should be able to run", func() {
			// Default rootCmd uses lockfile-based npm detection
			DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
			DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
			DebugExecutorExpectationManager.ExpectJSCommandLog(detect.NPM)

			_, err := executeCmd(rootCmd, "agent", "--debug")
			assert.NoError(err)
		})

		It("logs a debug message about the agent flag being set", func() {

			DebugExecutorExpectationManager.ExpectAgentFlagSet("pnpm")
			// Subcommand should also log its start with the chosen agent
			DebugExecutorExpectationManager.ExpectJSCommandLog(detect.PNPM)

			_, err := executeCmd(rootCmd, "agent", "--debug", "--agent", "pnpm")
			assert.NoError(err)

		})

		It("logs a message about the agent flag being set when the JPD variable is set", func() {

			_ = os.Setenv("JPD_AGENT", "pnpm")
			defer func() { _ = os.Unsetenv("JPD_AGENT") }()

			DebugExecutorExpectationManager.ExpectJPDAgentSet("pnpm")
			// Subcommand should also log its start with the env-provided agent
			DebugExecutorExpectationManager.ExpectJSCommandLog(detect.PNPM)

			_, err := executeCmd(rootCmd, "agent", "--debug")
			assert.NoError(err)

		})

		Context("lock file not detected", func() {

			It("logs a message about the package manager not being detected based on path", func() {

				rootCmd := factory.GenerateNoDetectionAtAll("")

				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectNoPMFromPath()

				_, err := executeCmd(rootCmd, "agent", "--debug")
				assert.Error(err)

			})

			It("logs a message about the package manager being detected based on path", func() {

				rootCmd := factory.CreateRootCmdWithPathDetected(detect.PNPM, nil, false)

				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.PNPM)
				// Subcommand should also log its start with the detected PM
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.PNPM)

				_, err := executeCmd(rootCmd, "agent", "--debug")
				assert.NoError(err)
			})

		})

		Context("lock file detected", func() {

			It("logs a message about the package manager being detected based on lockfile", func() {

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				// Subcommand should also log its start with the detected PM
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.YARN)

				_, err := executeCmd(factory.CreateRootCmdWithLockfileDetected(detect.YARN, detect.YARN_LOCK, nil, false), "agent", "--debug")
				assert.NoError(err)

			})

		})

	})

	const RootCommand = "Root Command"
	Describe(RootCommand, func() {

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

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "--help")
			assert.NoError(err)
			assert.Contains(output, "JavaScript Package Delegator")
			assert.Contains(output, "jpd")
		})

		Context("How it responds when no lockfile or global PM is detected", func() {

			It("should prompt user for install command and return error if input is invalid", func() {

				// Simulate no lockfile and no globally detected PM, leading to a prompt for an install command.
				// An empty string for commandTextUIValue will cause MockCommandTextUI.Run() to fail its validation.
				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectNoPMFromPath()
				currentRootCmd := factory.GenerateNoDetectionAtAll("")
				// Set context and parse flags before calling PersistentPreRunE directly
				currentRootCmd.SetContext(context.Background())
				_ = currentRootCmd.ParseFlags([]string{})

				err := currentRootCmd.PersistentPreRunE(currentRootCmd, []string{})
				assert.Error(err)
				assert.Contains(err.Error(), "A command for installing a package is at least three words")
				// Verify that Run was not called since the command was invalid
				mockCommandRunner.AssertNotCalled(GinkgoT(), "Run")
			})

			It("should prompt user for install command and execute it if input is valid", func() {

				const validInstallCommand = "npm install -g npm"
				re := regexp.MustCompile(`\s+`)
				splitCommandString := re.Split(validInstallCommand, -1)
				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectNoPMFromPath()
				DebugExecutorExpectationManager.ExpectJSCommandLog(splitCommandString[0], splitCommandString[1:]...) // Add this line
				currentRootCmd := factory.GenerateNoDetectionAtAll(validInstallCommand)
				// Set context and parse flags before calling PersistentPreRunE directly
				currentRootCmd.SetContext(context.Background())
				_ = currentRootCmd.ParseFlags([]string{})

				err := currentRootCmd.PersistentPreRunE(currentRootCmd, []string{})
				assert.NoError(err)
				// Verify that Run was called with the correct command
				mockCommandRunner.AssertCalled(GinkgoT(), "Run", splitCommandString[0], splitCommandString[1:], "")
			})

			It("should return an error if the user-provided install command fails to execute", func() {

				const validInstallCommand = "npm install -g npm"
				splitCommandString := strings.Fields(validInstallCommand)
				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectNoPMFromPath()
				DebugExecutorExpectationManager.ExpectJSCommandLog(splitCommandString[0], splitCommandString[1:]...) // Add expectation for JS command log

				// Configure the mock runner to make "npm" command fail using InvalidCommands approach
				mockCommandRunner.InvalidCommands = []string{"npm"}

				currentRootCmd := factory.GenerateNoDetectionAtAll(validInstallCommand)
				// Set context and parse flags before calling PersistentPreRunE directly
				currentRootCmd.SetContext(context.Background())
				_ = currentRootCmd.ParseFlags([]string{})

				err := currentRootCmd.PersistentPreRunE(currentRootCmd, []string{})
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})
		})

		Context("CWD Flag (-C)", func() {
			var currentRootCmd *cobra.Command
			var mockCommandRunner *mock.MockCommandRunner // This shadows the global mockRunner for this specific context

			var originalCwd string

			BeforeEach(func() {
				// Save original CWD
				// Path-based detection in this context

				var err error
				originalCwd, err = os.Getwd()
				assert.NoError(err)

				// This section uses a direct cmd.NewRootCmd because it explicitly
				// re-assigns the local 'mockRunner' variable within the CommandRunnerGetter,
				// which is a specific testing pattern for this context.
				currentRootCmd = factory.CreateYarnOneAsDefault(err)
				mockCommandRunner = factory.MockCommandRunner()
				DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
			})

			AfterEach(func() {
				// Restore original CWD
				if originalCwd != "" {
					err := os.Chdir(originalCwd)
					// Log error but don't fail test if we can't restore
					if err != nil {
						GinkgoWriter.Printf("Warning: Failed to restore original working directory: %v\n", err)
					}
				}
				// TempDir is automatically cleaned up by Ginkgo
			})

			It("should reject a --cwd flag value that does not end with '/'", func() {
				// Use platform-appropriate invalid path
				var invalidPath string
				if runtime.GOOS == "windows" {
					// On Windows, use POSIX-style path which should be invalid
					invalidPath = "/tmp/my-project"
				} else {
					// On POSIX, use path without trailing slash
					invalidPath = "/tmp/my-project"
				}

				_, err := executeCmd(currentRootCmd, "--cwd", invalidPath)
				if build_info.InCI() {
					// In CI mode: paths without trailing slash are valid on POSIX (relaxed validation)
					if runtime.GOOS == "windows" {
						// On Windows, POSIX paths should still be invalid even in CI
						assert.Error(err)
						assert.Contains(err.Error(), "is not a valid Windows folder path")
						assert.Contains(err.Error(), "cwd")
						assert.Contains(err.Error(), invalidPath)
					} else {
						assert.NoError(err)
					}
				} else {
					// In non-CI mode: should cause validation errors
					assert.Error(err)
					if runtime.GOOS == "windows" {
						assert.Contains(err.Error(), "is not a valid Windows folder path")
					} else {
						assert.Contains(err.Error(), "is not a valid POSIX/UNIX folder path (must end with '/' unless it's just '/')")
					}
					assert.Contains(err.Error(), "cwd")       // Check that the flag name is mentioned
					assert.Contains(err.Error(), invalidPath) // Check that the invalid path is mentioned
				}
			})

			It("should reject a --cwd flag value that is a filename", func() {
				invalidPath := "my-file.txt" // A file-like path
				_, err := executeCmd(currentRootCmd, "-C", invalidPath)
				// File-like paths should always be invalid regardless of CI mode or platform
				assert.Error(err)
				if runtime.GOOS == "windows" {
					assert.Contains(err.Error(), "is not a valid Windows folder path")
				} else {
					if build_info.InCI() {
						assert.Contains(err.Error(), "is not a valid POSIX/UNIX folder path")
					} else {
						assert.Contains(err.Error(), "is not a valid POSIX/UNIX folder path (must end with '/' unless it's just '/')")
					}
				}
				assert.Contains(err.Error(), "cwd")
				assert.Contains(err.Error(), invalidPath)
			})

			DescribeTable(
				"should reject invalid --cwd flag values",
				func(invalidPath string, expectedErrors ...string) {
					_, err := executeCmd(currentRootCmd, "--cwd", invalidPath)
					assert.Error(err)
					for _, expectedErr := range expectedErrors {
						assert.Contains(err.Error(), expectedErr)
					}
				},
				Entry("an empty string", "", "the cwd flag cannot be empty or contain only whitespace"),
				Entry("a string with only whitespace", "   ", "the cwd flag cannot be empty or contain only whitespace"),
			)

			It("should reject a path with invalid characters", func() {
				// Use platform-appropriate invalid path with illegal characters
				var invalidPath string
				if runtime.GOOS == "windows" {
					// On Windows, POSIX paths with colons are invalid
					invalidPath = "/path/with:colon/"
				} else {
					// On POSIX, paths with certain characters might be invalid
					invalidPath = "/path/with:colon/"
				}

				_, err := executeCmd(currentRootCmd, "--cwd", invalidPath)
				assert.Error(err)
				if runtime.GOOS == "windows" {
					assert.Contains(err.Error(), "is not a valid Windows folder path")
				} else {
					assert.Contains(err.Error(), "is not a valid POSIX/UNIX folder path")
				}
				assert.Contains(err.Error(), "cwd")
				assert.Contains(err.Error(), invalidPath)
			})

			Describe("should accept valid folder paths for --cwd", func() {
				It("accepts platform-appropriate valid paths", func() {
					var validPaths []string
					if runtime.GOOS == "windows" {
						// Use Windows-style valid paths
						validPaths = []string{".", ".\\", "..\\", "C:", "C:/", "C:/Windows/"}
					} else {
						validPaths = []string{"/", "./", "../"}
					}
					for _, p := range validPaths {
						_, err := executeCmd(currentRootCmd, "--cwd", p)
						assert.NoError(err)
					}
				})
			})

			It("should run a command in the specified directory using -C", func() {
				tempDir := GinkgoT().TempDir()
				tempDir = fmt.Sprintf("%s/", tempDir)

				// Execute a command with -C flag
				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "install")
				_, err := executeCmd(currentRootCmd, "install", "-C", tempDir)
				assert.NoError(err)

				// Verify the CommandRunner received the correct working directory
				mockCommandRunner.AssertCalled(GinkgoT(), "Run", "yarn", []string{"install"}, tempDir)
			})

			It("should run a command in the specified directory using --cwd", func() {
				tempDir := GinkgoT().TempDir()
				tempDir = fmt.Sprintf("%s/", tempDir)

				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "run", "dev")
				_, err := executeCmd(currentRootCmd, "run", "dev", "--cwd", tempDir)
				assert.NoError(err)

				mockCommandRunner.AssertCalled(GinkgoT(), "Run", "yarn", []string{"run", "dev"}, tempDir)
			})

			It("should not set a working directory if -C is not provided", func() {
				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn")
				_, err := executeCmd(currentRootCmd, "agent")
				assert.NoError(err)

				// Verify that the yarn agent command was called with no working directory override (empty string)
				mockCommandRunner.AssertCalled(GinkgoT(), "Run", "yarn", []string{}, "")
			})

			It("should handle a non-existent directory gracefully (likely fail at exec.Command level)", func() {
				// Use platform-appropriate non-existent path
				var nonExistentDir string
				if runtime.GOOS == "windows" {
					// On Windows, POSIX paths should be rejected at validation level
					nonExistentDir = "/non/existent/path/for/jpd/test"
				} else {
					nonExistentDir = "/non/existent/path/for/jpd/test"
				}

				_, err := executeCmd(currentRootCmd, "install", "--agent", "npm", "-C", fmt.Sprintf("%s/", nonExistentDir))
				assert.Error(err)
				if runtime.GOOS == "windows" {
					// On Windows, should fail at path validation level
					assert.Contains(err.Error(), "is not a valid Windows folder path")
				} else {
					// On POSIX, should fail at exec.Command level for non-existent directory
					assert.Contains(err.Error(), "no such file or directory")
				}
			})
		})

		Context("How it responds to Package Detection Failure", func() {

			generateRootCommandWithCommandRunnerHavingSetValue := func(value string) *cobra.Command {

				return cmd.NewRootCmdForTesting(
					cmd.Dependencies{
						CommandRunnerGetter: func() cmd.CommandRunner {
							return factory.MockCommandRunner()
						},
						DetectLockfile: func(targetDir string) (lockfile string, err error) {
							return "", nil
						},
						NewDebugExecutor: func(bool) cmd.DebugExecutor {
							return factory.DebugExecutor()
						},
						DetectJSPackageManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {

							return "", detect.ErrNoPackageManager
						},
						DetectJSPackageManager: func() (string, error) {
							return "", fmt.Errorf("format string")
						},
						NewCommandTextUI: func(lockfile string) cmd.CommandUITexter {

							commandTextUI := mock.NewMockCommandTextUI(lockfile).(*mock.MockCommandTextUI)

							commandTextUI.SetValue(value)

							return commandTextUI
						},
						YarnCommandVersionOutputter: mock.NewMockYarnCommandVersionOutputer("1.0.0"),
						NewPackageMultiSelectUI:     mock.NewMockPackageMultiSelectUI,
						NewTaskSelectorUI:           mock.NewMockTaskSelectUI,
						NewDependencyMultiSelectUI:  mock.NewMockDependencySelectUI,
					},
				)

			}

			It(
				`propmts the user for which command they would like to use to install package manager
								If the user refuses an error is produced.
							`,
				func() {

					// The generateRootCommandWithCommandRunnerHavingSetValue function returns ("", nil) from DetectLockfile,
					// so it goes through the "lock file detected" path with an empty string
					DebugExecutorExpectationManager.ExpectLockfileDetected("") // Empty lockfile detected
					currentCommand := generateRootCommandWithCommandRunnerHavingSetValue("")
					currentCommand.SetContext(context.Background())
					_ = currentCommand.ParseFlags([]string{})

					err := currentCommand.PersistentPreRunE(currentCommand, []string{})
					assert.Error(err)

				},
			)

			It(
				"executes the command given when the user gives a correct command",
				func() {

					commandString := "winget install pnpm.pnpm"

					// The generateRootCommandWithCommandRunnerHavingSetValue function returns ("", nil) from DetectLockfile,
					// so it goes through the "lock file detected" path with an empty string
					DebugExecutorExpectationManager.ExpectLockfileDetected("") // Empty lockfile detected
					currentCommand := generateRootCommandWithCommandRunnerHavingSetValue(commandString)
					currentCommand.SetContext(context.Background())
					_ = currentCommand.ParseFlags([]string{})

					splitCommandString := strings.Fields(commandString)
					DebugExecutorExpectationManager.ExpectJSCommandLog(splitCommandString[0], splitCommandString[1:]...) // Add this line

					err := currentCommand.PersistentPreRunE(currentCommand, []string{})

					assert.NoError(err)

					assert.True(mockCommandRunner.HasBeenCalled)
					assert.Equal(mockCommandRunner.CommandCall.Name, splitCommandString[0])
					assert.Equal(mockCommandRunner.CommandCall.Args, splitCommandString[1:])

				},
			)

			DescribeTable(
				"executes the command based typical instaltion commands",
				func(inputCommand string, expectedCommandName string, expectedCommandArgs []string) {

					// The generateRootCommandWithCommandRunnerHavingSetValue function returns ("", nil) from DetectLockfile,
					// so it goes through the "lock file detected" path with an empty string
					DebugExecutorExpectationManager.ExpectLockfileDetected("")                                      // Empty lockfile detected
					DebugExecutorExpectationManager.ExpectJSCommandLog(expectedCommandName, expectedCommandArgs...) // Add this line
					currentCommand := generateRootCommandWithCommandRunnerHavingSetValue(inputCommand)
					currentCommand.SetContext(context.Background())
					_ = currentCommand.ParseFlags([]string{})

					err := currentCommand.PersistentPreRunE(currentCommand, []string{})

					assert.NoError(err)

					assert.Equal(expectedCommandName, mockCommandRunner.CommandCall.Name)

					assert.Equal(expectedCommandArgs, mockCommandRunner.CommandCall.Args)

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

			BeforeEach(func() { // Default mock for yarn version

				// Create the root command with *all* necessary dependencies
				currentRootCmd = cmd.NewRootCmdForTesting(cmd.Dependencies{
					CommandRunnerGetter: func() cmd.CommandRunner {
						return mockCommandRunner
					},
					DetectLockfile: func(targetDir string) (lockfile string, err error) {
						return "", nil
					},
					DetectJSPackageManager: func() (string, error) {
						return "", fmt.Errorf("not detected")
					},
					DetectVolta: func() bool { return false },
					NewDebugExecutor: func(bool) cmd.DebugExecutor {
						return factory.DebugExecutor()
					},
					// Make sure detector returns an error so JPD_AGENT logic in root.go is hit
					DetectJSPackageManagerBasedOnLockFile: func(detectedLockFile string) (string, error) { return "", detect.ErrNoPackageManager },
					DetectJSPackageManager:                func() (string, error) { return "", detect.ErrNoPackageManager },
					YarnCommandVersionOutputter:           mock.NewMockYarnCommandVersionOutputer("1.0.0"),
					NewCommandTextUI:                      mock.NewMockCommandTextUI,
					NewPackageMultiSelectUI:               mock.NewMockPackageMultiSelectUI,
					NewTaskSelectorUI:                     mock.NewMockTaskSelectUI,
					NewDependencyMultiSelectUI:            mock.NewMockDependencySelectUI,
				})
				// Must set context because the background isn't activated.
				currentRootCmd.SetContext(context.Background())
				// No need to SetArgs here if we are directly calling PersistentPreRunE

				// This must be set so that the debug flag can be used!
				_ = currentRootCmd.ParseFlags([]string{})

			})

			AfterEach(func() {
				// Ensure environment variable is always cleaned up after each test
				_ = os.Unsetenv(cmd.JPD_AGENT_ENV_VAR)
			})

			It("shows an error when the env JPD_AGENT is set but it's not one of the supported JS package managers", func() {
				_ = os.Setenv(cmd.JPD_AGENT_ENV_VAR, "boo baa")

				// Directly call PersistentPreRunE and capture the error
				err := currentRootCmd.PersistentPreRunE(currentRootCmd, []string{})
				assert.Error(err)
				assert.Contains(err.Error(), "the JPD_AGENT variable is set the wrong way")
				// Verify that the command runner was not called for installation since an invalid agent was set
				assert.False(mockCommandRunner.HasBeenCalled)
			})

			It("sets the package name when the agent is a valid value", func() {
				// First, the root command will check for lockfile detection
				DebugExecutorExpectationManager.ExpectLockfileDetected("")
				// Then debug log the JPD_AGENT being set
				DebugExecutorExpectationManager.ExpectJPDAgentSet("deno")
				const expected = "deno"
				_ = os.Setenv(cmd.JPD_AGENT_ENV_VAR, expected)

				// Directly call PersistentPreRunE and capture the error
				err := currentRootCmd.PersistentPreRunE(currentRootCmd, []string{})
				assert.NoError(err)

				// Now, the context of currentRootCmd should have the value set by PersistentPreRunE
				pm, err := currentRootCmd.Flags().GetString(cmd.AGENT_FLAG)
				assert.NoError(err, "The package name was not found in context")
				assert.Equal(expected, pm)

				// Verify that no commands were executed by the mock runner because JPD_AGENT was set
				assert.False(mockCommandRunner.HasBeenCalled)
			})
		})
	})

	// Merged tests from root_cwd_integration_test.go
	Context("--cwd Integration Tests", func() {
		var (
			tempDir    string
			subDir     string
			mockFS     *MockFileSystemCwd
			mockLookup *MockPathLookupCwd
			fakeRunner *FakeCommandRunnerCwd
			deps       cmd.Dependencies
		)

		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "jpd-root-cwd-test")
			assert.NoError(err)

			subDir = filepath.Join(tempDir, "project")
			err = os.MkdirAll(subDir, 0755)
			assert.NoError(err)

			mockFS = &MockFileSystemCwd{
				files: make(map[string]bool),
			}
			mockLookup = &MockPathLookupCwd{
				paths: map[string]bool{
					"npm": true,
				},
			}
			fakeRunner = &FakeCommandRunnerCwd{}

			deps = cmd.Dependencies{
				CommandRunnerGetter: func() cmd.CommandRunner {
					return fakeRunner
				},
				DetectJSPackageManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {
					return detect.DetectJSPackageManagerBasedOnLockFile(detectedLockFile, mockLookup)
				},
				YarnCommandVersionOutputter: &MockYarnVersionOutputterCwd{version: "1.22.19"},
				NewCommandTextUI:            newMockCommandTextUICwd,
				DetectLockfile: func(targetDir string) (string, error) {
					return detect.DetectLockfileIn(targetDir, mockFS)
				},
				DetectJSPackageManager: func() (string, error) {
					return detect.DetectJSPackageManager(mockLookup)
				},
				DetectVolta: func() bool {
					return detect.DetectVolta(mockLookup)
				},
				NewPackageMultiSelectUI:    newMockPackageMultiSelectUICwd,
				NewTaskSelectorUI:          newMockTaskSelectorUICwd,
				NewDependencyMultiSelectUI: newMockDependencyMultiSelectUICwd,
				NewDebugExecutor:           newMockDebugExecutorCwd,
			}
		})

		AfterEach(func() {
			_ = os.RemoveAll(tempDir)
		})

		It("should detect lockfile in the specified directory not current directory", func() {
			// Setup: place package-lock.json in subDir only
			packageLockPath := filepath.Join(subDir, "package-lock.json")
			mockFS.files[packageLockPath] = true

			// Create root command with dependencies
			rootCmd := cmd.NewRootCmd(deps)

			// Set the arguments to simulate --cwd flag
			rootCmd.SetArgs([]string{
				"--cwd", subDir + "/", // Add trailing slash for POSIX validation
				"agent", // Run the agent command to trigger lockfile detection
			})

			err := rootCmd.Execute()
			assert.NoError(err)

			// Verify that the agent flag was set to npm (from package-lock.json detection)
			agentFlag, err := rootCmd.PersistentFlags().GetString("agent")
			assert.NoError(err)
			assert.Equal("npm", agentFlag)
		})

		It("should not detect lockfile in current directory when --cwd points elsewhere", func() {
			// Setup: place package-lock.json only in current directory (tempDir)
			// but point --cwd to subDir which has no lockfiles
			currentDirLockFile := filepath.Join(tempDir, "package-lock.json")
			mockFS.files[currentDirLockFile] = true

			// Mock fs.Getwd to return tempDir as current directory
			mockFS.cwd = tempDir

			// Create root command with dependencies
			rootCmd := cmd.NewRootCmd(deps)

			// Set the arguments to simulate --cwd flag pointing to subDir
			rootCmd.SetArgs([]string{
				"--cwd", subDir + "/", // Add trailing slash for POSIX validation
				"agent",
			})

			err := rootCmd.Execute()
			assert.NoError(err)

			// Since no lockfile in subDir, it should fallback to detecting npm from PATH
			agentFlag, err := rootCmd.PersistentFlags().GetString("agent")
			assert.NoError(err)
			assert.Equal("npm", agentFlag) // Should still be npm from PATH detection
		})

		It("should fallback to current directory when --cwd is not provided", func() {
			// Setup: place yarn.lock in tempDir (current directory)
			yarnLockPath := filepath.Join(tempDir, "yarn.lock")
			mockFS.files[yarnLockPath] = true
			mockLookup.paths["yarn"] = true

			// Mock fs.Getwd to return tempDir as current directory
			mockFS.cwd = tempDir

			// Create root command with dependencies
			rootCmd := cmd.NewRootCmd(deps)

			// Don't set --cwd flag, should use current directory
			rootCmd.SetArgs([]string{"agent"})

			err := rootCmd.Execute()
			assert.NoError(err)

			// Verify that yarn was detected from the lockfile in current directory
			agentFlag, err := rootCmd.PersistentFlags().GetString("agent")
			assert.NoError(err)
			assert.Equal("yarn", agentFlag)
		})
	})

	const DLXCommand = "DLX Command"
	Describe(DLXCommand, func() {

		It("should show help", func() {
			// Help doesn't need any additional expectations since it doesn't execute the business logic
			output, err := executeCmd(rootCmd, "dlx", "--help")
			assert.NoError(err)
			assert.Contains(output, "Execute packages temporarily without installing them first")
			assert.Contains(output, "jpd dlx")
		})

		It("should require at least one argument", func() {
			// This test also doesn't execute the full business logic, just validates args
			_, err := executeCmd(rootCmd, "dlx")
			assert.Error(err)
			assert.Contains(err.Error(), "requires at least 1 arg(s)")
		})

		Context("npm", func() {

			It("should execute npx with package name", func() {
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npx", "create-react-app")
				_, err := executeCmd(rootCmd, "dlx", "create-react-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npx", "create-react-app"))
			})

			It("should execute npx with package name and args", func() {
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npx", "create-react-app", "my-app")
				_, err := executeCmd(rootCmd, "dlx", "create-react-app", "--", "my-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npx", "create-react-app", "my-app"))

			})
		})

		Context("yarn", func() {
			It("should execute yarn with package name for yarn v1", func() {
				yarnRootCmd := factory.CreateYarnOneAsDefault(nil)
				DebugExecutorExpectationManager.ExpectCommonPathDetectionFlow(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "create-react-app")
				_, err := executeCmd(yarnRootCmd, "dlx", "create-react-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "create-react-app"))
			})

			It("should execute yarn dlx with package name for yarn v2+", func() {
				yarnRootCmd := factory.CreateYarnTwoAsDefault(nil)
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.YARN, detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "dlx", "create-react-app")
				_, err := executeCmd(yarnRootCmd, "dlx", "create-react-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "dlx", "create-react-app"))
			})

			It("should handle yarn version detection error (fallback to v1)", func() {
				yarnRootCmd := factory.CreateNoYarnVersion(nil)
				DebugExecutorExpectationManager.ExpectCommonPathDetectionFlow(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "create-react-app")
				_, err := executeCmd(yarnRootCmd, "dlx", "create-react-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "create-react-app"))
			})
		})

		Context("pnpm", func() {
			It("should execute pnpm dlx with package name", func() {
				pnpmRootCmd := factory.CreatePnpmAsDefault(nil)
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.PNPM, detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "dlx", "create-react-app")
				_, err := executeCmd(pnpmRootCmd, "dlx", "create-react-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("pnpm", "dlx", "create-react-app"))
			})

			It("should execute pnpm dlx with package name and args", func() {
				pnpmRootCmd := factory.CreatePnpmAsDefault(nil)
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.PNPM, detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "dlx", "create-react-app", "my-app")
				_, err := executeCmd(pnpmRootCmd, "dlx", "create-react-app", "--", "my-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("pnpm", "dlx", "create-react-app", "my-app"))
			})
		})

		Context("bun", func() {
			It("should execute bunx with package name", func() {
				bunRootCmd := factory.CreateBunAsDefault(nil)
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.BUN, detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectJSCommandLog("bunx", "create-react-app")
				_, err := executeCmd(bunRootCmd, "dlx", "create-react-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("bunx", "create-react-app"))
			})

			It("should execute bunx with package name and args", func() {
				bunRootCmd := factory.CreateBunAsDefault(nil)
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.BUN, detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectJSCommandLog("bunx", "create-react-app", "my-app")
				_, err := executeCmd(bunRootCmd, "dlx", "create-react-app", "--", "my-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("bunx", "create-react-app", "my-app"))
			})
		})

		Context("Error Handling", func() {
			It("should return error when command runner fails", func() {
				// Set up expectations for npm lockfile detection
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)

				rootCmd := factory.CreateNpmAsDefault(nil)
				mockCommandRunner.InvalidCommands = []string{"npx"}

				DebugExecutorExpectationManager.ExpectJSCommandLog("npx", "test-command")

				_, err := executeCmd(rootCmd, "dlx", "test-command")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npx' is configured to fail")
			})

			It("should return error for unsupported package manager", func() {
				// Set up expectations for unknown package manager lockfile detection
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile("unknown")

				rootCmd := factory.GenerateWithPackageManagerDetector("unknown", nil)
				// Note: ExpectJSCommandLog is NOT called because the dlx command returns an error
				// before reaching the LogJSCommandIfDebugIsTrue call for unsupported package managers

				_, err := executeCmd(rootCmd, "dlx", "some-package")
				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager: unknown")
			})
		})
	})

	const InstallCommand = "Install Command"
	Describe(InstallCommand, func() {

		Context("Works with the search flag", func() {

			var rootCmd *cobra.Command

			BeforeEach(func() {
				// Path-based detection to npm for this helper
				rootCmd = factory.CreateWithPackageManagerAndMultiSelectUI()
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
				// Set up proper debug expectations for this test
				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.NPM)

				_, err := executeCmd(rootCmd, "install", "--search", "89ispsnsnis")

				assert.Error(err)
				assert.ErrorContains(err, "search failed for \"89ispsnsnis\"")

			})

			It("works", func() {

				const expected = "angular"
				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.NPM)
				// The search functionality returns a long list of packages, so we need to match the actual expected packages
				DebugExecutorExpectationManager.ExpectJSCommandRandomLog()
				_, err := executeCmd(rootCmd, "install", "--search", expected)
				assert.NoError(err)

				assert.Equal("npm", mockCommandRunner.CommandCall.Name)

				assert.Contains(mockCommandRunner.CommandCall.Args, "install")
				assert.NotContains(mockCommandRunner.CommandCall.Args, "--search")
				assert.Conditionf(func() bool {
					return lo.SomeBy(mockCommandRunner.CommandCall.Args, func(item string) bool {
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

					// Set expectations to match the dynamic packageManager
					DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
					DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(packageManager)

					rootCommmand := factory.GenerateWithPackageManagerDetectedAndVolta(packageManager)

					DebugExecutorExpectationManager.ExpectJSCommandLog("volta", "run", packageManager, "install") // REMOVED
					output, err := executeCmd(rootCommmand, "install")

					assert.NoError(err)
					assert.Empty(output)

					assert.Equal("volta", mockCommandRunner.CommandCall.Name)

					assert.Equal([]string{"run", packageManager, "install"}, mockCommandRunner.CommandCall.Args)

				},
				EntryDescription("Volta run was appended to %s"),
				Entry(nil, detect.NPM),
				Entry(nil, detect.YARN),
				Entry(nil, detect.PNPM),
			)

			DescribeTable(
				"Doesn't append volta run when a non-node package manager is the agent",
				func(packageManager string) {

					rootCommmand := factory.GenerateWithPackageManagerDetectedAndVolta(packageManager)

					var (
						output string
						error  error
					)

					switch packageManager {
					case detect.DENO:
						// Set expectations for deno lockfile-based detection
						DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
						DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)

						DebugExecutorExpectationManager.ExpectJSCommandLog(packageManager, "add", "npm:cn-efs") // REMOVED
						output, error = executeCmd(rootCommmand, "install", "npm:cn-efs")
						assert.NoError(error)
						assert.Empty(output)

						assert.Equal(packageManager, mockCommandRunner.CommandCall.Name)

						assert.Equal([]string{"add", "npm:cn-efs"}, mockCommandRunner.CommandCall.Args)

					case detect.BUN:
						// Set expectations for bun lockfile-based detection
						DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
						DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)

						DebugExecutorExpectationManager.ExpectJSCommandLog(packageManager, "install") // REMOVED
						output, error = executeCmd(rootCommmand, "install")
						assert.NoError(error)
						assert.Empty(output)

						assert.Equal(mockCommandRunner.CommandCall.Name, packageManager)

						assert.Equal(mockCommandRunner.CommandCall.Args, []string{"install"})

					default:
						DebugExecutorExpectationManager.ExpectJSCommandLog(packageManager, "install") // REMOVED
						output, error = executeCmd(rootCommmand, "install")

						assert.NoError(error)
						assert.Empty(output)

						assert.Equal(mockCommandRunner.CommandCall.Name, packageManager)

						assert.Equal(mockCommandRunner.CommandCall.Args, []string{"install"})
					}

				},
				EntryDescription("Volta run was't appended to %s"),
				Entry(nil, detect.DENO),
				Entry(nil, detect.BUN),
			)

			It("rejects volta usage if the --no-volta flag is passed", func() {

				rootCommmand := factory.GenerateWithPackageManagerDetectedAndVolta("npm")

				// Add missing lockfile detection expectation
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile("npm")
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "install")
				output, err := executeCmd(rootCommmand, "install", "--no-volta")

				assert.NoError(err)
				assert.Empty(output)

				assert.Equal("npm", mockCommandRunner.CommandCall.Name)

				assert.Equal([]string{"install"}, mockCommandRunner.CommandCall.Args)
			})

		})

		Context("npm", func() {
			It("should run npm install with no args", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "install")
				_, err := executeCmd(rootCmd, "install")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "npm")
				assert.Contains(mockCommandRunner.CommandCall.Args, "install")
			})

			It("should run npm install with package names", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "install", "lodash", "express")
				_, err := executeCmd(rootCmd, "install", "lodash", "express")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "npm")
				assert.Contains(mockCommandRunner.CommandCall.Args, "install")
				assert.Contains(mockCommandRunner.CommandCall.Args, "lodash")
				assert.Contains(mockCommandRunner.CommandCall.Args, "express")
			})

			It("should run npm install with dev flag", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "install", "typescript", "--save-dev")
				_, err := executeCmd(rootCmd, "install", "--dev", "typescript")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "npm")
				assert.Contains(mockCommandRunner.CommandCall.Args, "install")
				assert.Contains(mockCommandRunner.CommandCall.Args, "typescript")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--save-dev")
			})

			It("should run npm install with global flag", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "install", "typescript", "--global")
				_, err := executeCmd(rootCmd, "install", "--global", "typescript")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "npm")
				assert.Contains(mockCommandRunner.CommandCall.Args, "install")
				assert.Contains(mockCommandRunner.CommandCall.Args, "typescript")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--global")
			})

			It("should run npm install with production flag", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "install", "--omit=dev")
				_, err := executeCmd(rootCmd, "install", "--production")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "npm")
				assert.Contains(mockCommandRunner.CommandCall.Args, "install")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--omit=dev")
			})

			It("should handle frozen flag with npm", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "install", "--package-lock-only")
				_, err := executeCmd(rootCmd, "install", "--frozen")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "npm")
				assert.Contains(mockCommandRunner.CommandCall.Args, "install")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--package-lock-only")
			})
		})

		Context("yarn", func() {
			var yarnRootCmd *cobra.Command

			BeforeEach(func() {
				yarnRootCmd = factory.CreateYarnTwoAsDefault(nil)
			})

			It("should run yarn add with dev flag", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "add", "typescript", "--dev")
				_, err := executeCmd(yarnRootCmd, "install", "--dev", "typescript")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "yarn")
				assert.Contains(mockCommandRunner.CommandCall.Args, "add")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--dev")
				assert.Contains(mockCommandRunner.CommandCall.Args, "typescript")
			})

			It("should handle global flag with yarn", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "add", "typescript", "--global")
				_, err := executeCmd(yarnRootCmd, "install", "--global", "typescript")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "yarn")
				assert.Contains(mockCommandRunner.CommandCall.Args, "add")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--global")
				assert.Contains(mockCommandRunner.CommandCall.Args, "typescript")
			})

			It("should handle production flag with yarn", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "install", "--production")
				_, err := executeCmd(yarnRootCmd, "install", "--production")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "yarn")
				assert.Contains(mockCommandRunner.CommandCall.Args, "install")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--production")
			})

			It("should handle frozen flag with yarn", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "install", "--frozen-lockfile")
				_, err := executeCmd(yarnRootCmd, "install", "--frozen")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "yarn")
				assert.Contains(mockCommandRunner.CommandCall.Args, "install")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--frozen-lockfile")
			})

			It("should handle yarn classic with dependencies", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "add", "lodash")
				_, err := executeCmd(yarnRootCmd, "install", "lodash")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "yarn")
				assert.Contains(mockCommandRunner.CommandCall.Args, "add")
				assert.Contains(mockCommandRunner.CommandCall.Args, "lodash")
			})

			It("should handle yarn modern with dependencies", func() {
				// Test yarn version 2+ path
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "add", "typescript", "--dev")
				_, err := executeCmd(yarnRootCmd, "install", "--dev", "typescript")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "yarn")
				assert.Contains(mockCommandRunner.CommandCall.Args, "add")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--dev")
				assert.Contains(mockCommandRunner.CommandCall.Args, "typescript")
			})
		})

		Context("pnpm", func() {
			var pnpmRootCmd *cobra.Command

			BeforeEach(func() {
				pnpmRootCmd = factory.CreatePnpmAsDefault(nil)
			})

			It("should run pnpm add with dev flag", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "add", "typescript", "--save-dev")
				_, err := executeCmd(pnpmRootCmd, "install", "--dev", "typescript")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "pnpm")
				assert.Contains(mockCommandRunner.CommandCall.Args, "add")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--save-dev")
				assert.Contains(mockCommandRunner.CommandCall.Args, "typescript")
			})

			It("should handle global flag with pnpm", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "add", "typescript", "--global")
				_, err := executeCmd(pnpmRootCmd, "install", "--global", "typescript")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "pnpm")
				assert.Contains(mockCommandRunner.CommandCall.Args, "add")
				assert.Contains(mockCommandRunner.CommandCall.Args, "typescript")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--global")
			})

			It("should handle production flag with pnpm", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "install", "--prod")
				_, err := executeCmd(pnpmRootCmd, "install", "--production")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "pnpm")
				assert.Contains(mockCommandRunner.CommandCall.Args, "install")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--prod")
			})

			It("should handle frozen flag with pnpm", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "install", "--frozen-lockfile")
				_, err := executeCmd(pnpmRootCmd, "install", "--frozen")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "pnpm")
				assert.Contains(mockCommandRunner.CommandCall.Args, "install")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--frozen-lockfile")
			})

			It("should handle pnpm with dev dependencies", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "add", "jest", "--save-dev")
				_, err := executeCmd(pnpmRootCmd, "install", "-D", "jest")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "pnpm")
				assert.Contains(mockCommandRunner.CommandCall.Args, "add")
				assert.Contains(mockCommandRunner.CommandCall.Args, "jest")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--save-dev")
			})
		})

		Context("bun", func() {
			var bunRootCmd *cobra.Command

			BeforeEach(func() {
				bunRootCmd = factory.CreateBunAsDefault(nil)
			})

			It("should handle bun dev flag", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("bun", "add", "typescript", "--development")
				_, err := executeCmd(bunRootCmd, "install", "--dev", "typescript")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "bun")
				assert.Contains(mockCommandRunner.CommandCall.Args, "add")
				assert.Contains(mockCommandRunner.CommandCall.Args, "typescript")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--development")
			})

			It("should handle global flag with bun", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("bun", "add", "typescript", "--global")
				_, err := executeCmd(bunRootCmd, "install", "--global", "typescript")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "bun")
				assert.Contains(mockCommandRunner.CommandCall.Args, "add")
				assert.Contains(mockCommandRunner.CommandCall.Args, "typescript")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--global")
			})

			It("should handle production flag with bun", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("bun", "install", "--production")
				_, err := executeCmd(bunRootCmd, "install", "--production")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "bun")
				assert.Contains(mockCommandRunner.CommandCall.Args, "install")
				assert.Contains(mockCommandRunner.CommandCall.Args, "--production")
			})

			It("should handle bun with dependencies", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("bun", "add", "react")
				_, err := executeCmd(bunRootCmd, "install", "react")
				assert.NoError(err)
				assert.Contains(mockCommandRunner.CommandCall.Name, "bun")
				assert.Contains(mockCommandRunner.CommandCall.Args, "add")
				assert.Contains(mockCommandRunner.CommandCall.Args, "react")
			})
		})

		Context("deno", func() {
			var denoRootCmd *cobra.Command

			BeforeEach(func() {
				denoRootCmd = factory.CreateDenoAsDefault(nil)
			})

			It("should return an error if no packages are provided for a non-global install", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				_, err := executeCmd(denoRootCmd, "install")
				assert.Error(err)
				assert.Contains(err.Error(), "for deno one or more packages is required")
				assert.False(mockCommandRunner.HasBeenCalled)
			})

			It("should return an error if --global flag is used without packages", func() {
				// This case should still trigger the "one or more packages" error first
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				_, err := executeCmd(denoRootCmd, "install", "--global")
				assert.Error(err)
				assert.Contains(err.Error(), "for deno one or more packages is required")
				assert.False(mockCommandRunner.HasBeenCalled)
			})

			It("should execute deno install with --global flag and packages", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				DebugExecutorExpectationManager.ExpectJSCommandLog("deno", "install", "my-global-tool")
				_, err := executeCmd(denoRootCmd, "install", "--global", "my-global-tool")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("deno", "install", "my-global-tool"))
			})

			It("should return an error when --production flag is used", func() {
				// Pass a package to ensure it bypasses the "no packages" error and hits the "production" error
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				_, err := executeCmd(denoRootCmd, "install", "--production", "my-package")
				assert.Error(err)
				assert.Contains(err.Error(), "deno doesn't support prod")
				assert.False(mockCommandRunner.HasBeenCalled)
			})

			It("should execute deno add with no flags and packages", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				DebugExecutorExpectationManager.ExpectJSCommandLog("deno", "add", "my-package")
				_, err := executeCmd(denoRootCmd, "install", "my-package")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("deno", "add", "my-package"))
			})

			It("should execute deno add with --dev flag and packages", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				DebugExecutorExpectationManager.ExpectJSCommandLog("deno", "add", "my-dev-dep", "--dev")
				_, err := executeCmd(denoRootCmd, "install", "--dev", "my-dev-dep")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("deno", "add", "my-dev-dep", "--dev"))
			})

			It("should return an error if --dev flag is used without packages", func() {
				// This case should still trigger the "one or more packages" error first
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				_, err := executeCmd(denoRootCmd, "install", "--dev")
				assert.Error(err)
				assert.Contains(err.Error(), "for deno one or more packages is required")
				assert.False(mockCommandRunner.HasBeenCalled)
			})
		})

		Context("Error Handling", func() {
			It("should return error for unsupported package manager", func() {
				// Align debug expectations to unknown PM for this scenario

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile("unknown")
				rootCmd := factory.GenerateWithPackageManagerDetector("unknown", nil)
				_, err := executeCmd(rootCmd, "install", "lodash")
				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager: unknown")
			})

			It("should return error when command runner fails", func() {
				rootCmd := factory.CreateNpmAsDefault(nil)
				mockCommandRunner.InvalidCommands = []string{"npm"}

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.NPM, "install")
				_, err := executeCmd(rootCmd, "install")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})
		})

		It("jpd install -D tsup -C <dir> sets working directory and builds the correct command", func() {

			root := factory.CreateNpmAsDefault(nil)
			tmpDir := fmt.Sprintf("%s/", GinkgoT().TempDir())

			// Add missing debug expectations for lockfile detection
			DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
			DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
			DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "install", "tsup", "--save-dev") // REMOVED
			_, err := executeCmd(root, "install", "-D", "tsup", "-C", tmpDir)
			assert.NoError(err)
			assert.True(factory.MockCommandRunner().HasCommand("npm", "install", "tsup", "--save-dev"))
			assert.Equal(tmpDir, factory.MockCommandRunner().WorkingDir)
		})
	})

	const CreateCommand = "Create Command"
	Describe(CreateCommand, func() {

		var createCmd *cobra.Command
		BeforeEach(func() {
			createCmd, _ = getSubCommandWithName(rootCmd, "create")
		})

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "create", "--help")
			assert.NoError(err)
			assert.Contains(output, "Scaffold a new project using the appropriate package manager's create command")
			assert.Contains(output, "jpd create")
			// Create-specific flags are visible in help
			assert.Contains(output, "--search")
			assert.Contains(output, "-s")
			assert.Contains(output, "--size")
			// Guidance about npm separator normalization
			assert.Contains(output, "JPD automatically inserts the -- separator")
		})

		It("should have correct aliases", func() {
			assert.Contains(createCmd.Aliases, "c")
		})

		It("should require at least one argument", func() {
			_, err := executeCmd(rootCmd, "create")
			assert.Error(err)
			assert.Contains(err.Error(), "requires at least 1 arg(s)")
		})

		// BuildCreateCommand Unit Tests
		Describe("BuildCreateCommand function tests", func() {
			Describe("npm create command", func() {
				It("should build npm exec create-react-app command", func() {
					program, args, err := cmd.BuildCreateCommand("npm", "", "react-app", []string{"my-app"})
					assert.NoError(err)
					assert.Equal("npm", program)
					assert.Equal([]string{"exec", "create-react-app", "--", "my-app"}, args)
				})

				It("should handle package already prefixed with create-", func() {
					program, args, err := cmd.BuildCreateCommand("npm", "", "create-react-app", []string{"my-app"})
					assert.NoError(err)
					assert.Equal("npm", program)
					assert.Equal([]string{"exec", "create-react-app", "--", "my-app"}, args)
				})

				It("should handle version specifiers", func() {
					program, args, err := cmd.BuildCreateCommand("npm", "", "vite@latest", []string{"my-app", "--template", "react"})
					assert.NoError(err)
					assert.Equal("npm", program)
					assert.Equal([]string{"exec", "create-vite@latest", "--", "my-app", "--template", "react"}, args)
				})

				It("should handle scoped packages", func() {
					program, args, err := cmd.BuildCreateCommand("npm", "", "@org/starter", []string{"my-app"})
					assert.NoError(err)
					assert.Equal("npm", program)
					assert.Equal([]string{"exec", "@org/starter", "--", "my-app"}, args)
				})

				It("should return error when no name provided", func() {
					_, _, err := cmd.BuildCreateCommand("npm", "", "", []string{})
					assert.Error(err)
					assert.Contains(err.Error(), "package name is required")
				})

				It("should reject URLs", func() {
					_, _, err := cmd.BuildCreateCommand("npm", "", "https://example.com/create.js", []string{})
					assert.Error(err)
					assert.Contains(err.Error(), "URLs are not supported for npm")
				})
			})

			Describe("pnpm create command", func() {
				It("should build pnpm exec command", func() {
					program, args, err := cmd.BuildCreateCommand("pnpm", "", "react-app", []string{"my-app"})
					assert.NoError(err)
					assert.Equal("pnpm", program)
					assert.Equal([]string{"dlx", "create-react-app", "my-app"}, args)
				})

				It("should handle version specifiers", func() {
					program, args, err := cmd.BuildCreateCommand("pnpm", "", "vite@4", []string{"my-app", "--template", "vue"})
					assert.NoError(err)
					assert.Equal("pnpm", program)
					assert.Equal([]string{"dlx", "create-vite@4", "my-app", "--template", "vue"}, args)
				})
			})

			Describe("yarn create command", func() {
				It("should use npx for yarn v1", func() {
					program, args, err := cmd.BuildCreateCommand("yarn", "1.22.0", "react-app", []string{"my-app"})
					assert.NoError(err)
					assert.Equal("npx", program)
					assert.Equal([]string{"create-react-app", "my-app"}, args)
				})

				It("should use yarn dlx for yarn v2+", func() {
					program, args, err := cmd.BuildCreateCommand("yarn", "2.0.0", "react-app", []string{"my-app"})
					assert.NoError(err)
					assert.Equal("yarn", program)
					assert.Equal([]string{"dlx", "create-react-app", "my-app"}, args)
				})

				It("should use yarn dlx for yarn berry", func() {
					program, args, err := cmd.BuildCreateCommand("yarn", "berry-3.1.0", "react-app", []string{"my-app"})
					assert.NoError(err)
					assert.Equal("yarn", program)
					assert.Equal([]string{"dlx", "create-react-app", "my-app"}, args)
				})
			})

			Describe("bun create command", func() {
				It("should use bunx", func() {
					program, args, err := cmd.BuildCreateCommand("bun", "", "react-app", []string{"my-app"})
					assert.NoError(err)
					assert.Equal("bunx", program)
					assert.Equal([]string{"create-react-app", "my-app"}, args)
				})
			})

			Describe("deno create command", func() {
				It("should use deno run with valid URL", func() {
					program, args, err := cmd.BuildCreateCommand("deno", "", "https://deno.land/x/fresh/init.ts", []string{"my-app"})
					assert.NoError(err)
					assert.Equal("deno", program)
					assert.Equal([]string{"run", "https://deno.land/x/fresh/init.ts", "my-app"}, args)
				})

				It("should reject invalid URL", func() {
					_, _, err := cmd.BuildCreateCommand("deno", "", "not-a-url", []string{})
					assert.Error(err)
					assert.Contains(err.Error(), "deno create requires a valid URL")
				})

				It("should reject empty URL", func() {
					_, _, err := cmd.BuildCreateCommand("deno", "", "", []string{})
					assert.Error(err)
					assert.Contains(err.Error(), "deno create requires a URL")
				})
			})

			Describe("error cases", func() {
				It("should reject unsupported package manager", func() {
					_, _, err := cmd.BuildCreateCommand("unsupported", "", "react-app", []string{})
					assert.Error(err)
					assert.Contains(err.Error(), "unsupported package manager")
				})
			})
		})

		Context("npm", func() {
			It("should execute npm exec create-react-app", func() {
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "exec", "create-react-app", "--", "my-app")
				_, err := executeCmd(rootCmd, "create", "react-app", "my-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "exec", "create-react-app", "--", "my-app"))
			})

			It("should execute npm exec create-react-app with additional args", func() {
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "exec", "create-react-app", "--", "my-app", "--template", "typescript")
				_, err := executeCmd(rootCmd, "create", "react-app", "my-app", "--template", "typescript")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "exec", "create-react-app", "--", "my-app", "--template", "typescript"))
			})

			It("should handle create prefix already present", func() {
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "exec", "create-react-app", "--", "my-app")
				_, err := executeCmd(rootCmd, "create", "create-react-app", "my-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "exec", "create-react-app", "--", "my-app"))
			})

			It("should handle version specifiers", func() {
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "exec", "create-vite@latest", "--", "my-app")
				_, err := executeCmd(rootCmd, "create", "vite@latest", "my-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "exec", "create-vite@latest", "--", "my-app"))
			})

			It("normalizes an extra user-provided -- for npm", func() {
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
				// We only assert the executed command; allow any debug JS command log
				DebugExecutorExpectationManager.ExpectJSCommandRandomLog()
				_, err := executeCmd(rootCmd, "create", "vite@latest", "my-app", "--", "--template", "react")
				assert.NoError(err)
				// Ensure only a single -- is present in the executed command
				assert.True(mockCommandRunner.HasCommand("npm", "exec", "create-vite@latest", "--", "my-app", "--template", "react"))
			})
		})

		Context("pnpm", func() {
			var pnpmRootCmd *cobra.Command

			BeforeEach(func() {
				pnpmRootCmd = factory.CreatePnpmAsDefault(nil)
			})

			It("should execute pnpm exec create-react-app", func() {
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.PNPM, detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "dlx", "create-react-app", "my-app")
				_, err := executeCmd(pnpmRootCmd, "create", "react-app", "my-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("pnpm", "dlx", "create-react-app", "my-app"))
			})

			It("does not strip a user-provided -- for pnpm", func() {
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.PNPM, detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "dlx", "create-vite@latest", "my-app", "--", "--template", "react")
				_, err := executeCmd(pnpmRootCmd, "create", "vite@latest", "my-app", "--", "--template", "react")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("pnpm", "dlx", "create-vite@latest", "my-app", "--", "--template", "react"))
			})
		})

		Context("yarn", func() {
			It("should execute npx create-react-app for yarn v1", func() {
				yarnRootCmd := factory.CreateYarnOneAsDefault(nil)
				DebugExecutorExpectationManager.ExpectCommonPathDetectionFlow(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npx", "create-react-app", "my-app")
				_, err := executeCmd(yarnRootCmd, "create", "react-app", "my-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npx", "create-react-app", "my-app"))
			})

			It("should execute yarn dlx create-react-app for yarn v2+", func() {
				yarnRootCmd := factory.CreateYarnTwoAsDefault(nil)
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.YARN, detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "dlx", "create-react-app", "my-app")
				_, err := executeCmd(yarnRootCmd, "create", "react-app", "my-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "dlx", "create-react-app", "my-app"))
			})
		})

		Context("bun", func() {
			var bunRootCmd *cobra.Command

			BeforeEach(func() {
				bunRootCmd = factory.CreateBunAsDefault(nil)
			})

			It("should execute bunx create-react-app", func() {
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.BUN, detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectJSCommandLog("bunx", "create-react-app", "my-app")
				_, err := executeCmd(bunRootCmd, "create", "react-app", "my-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("bunx", "create-react-app", "my-app"))
			})
		})

		Context("deno", func() {
			var denoRootCmd *cobra.Command

			BeforeEach(func() {
				denoRootCmd = factory.CreateDenoAsDefault(nil)
			})

			It("should execute deno run with URL", func() {
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.DENO, detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectJSCommandLog("deno", "run", "https://deno.land/x/fresh/init.ts", "my-app")
				_, err := executeCmd(denoRootCmd, "create", "https://deno.land/x/fresh/init.ts", "my-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("deno", "run", "https://deno.land/x/fresh/init.ts", "my-app"))
			})

			It("should reject non-URL for deno", func() {
				DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.DENO, detect.DENO_JSON)
				_, err := executeCmd(denoRootCmd, "create", "react-app", "my-app")
				assert.Error(err)
				assert.Contains(err.Error(), "deno create requires a valid URL")
			})
		})

		Context("Error Handling", func() {
			It("should return error when command runner fails", func() {
				rootCmd := factory.CreateNpmAsDefault(nil)
				mockCommandRunner.InvalidCommands = []string{"npm"}

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "exec", "create-react-app", "--", "my-app")
				_, err := executeCmd(rootCmd, "create", "react-app", "my-app")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})

			It("should return error for unsupported package manager", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile("unknown")
				rootCmd := factory.GenerateWithPackageManagerDetector("unknown", nil)
				_, err := executeCmd(rootCmd, "create", "react-app", "my-app")
				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager: unknown")
			})
		})

		Context("Search functionality", func() {
			var rootCmdWithSearch *cobra.Command

			BeforeEach(func() {
				// Use the factory with search UI capability
				rootCmdWithSearch = factory.CreateWithPackageManagerAndMultiSelectUI()
			})

			It("should search for packages when --search flag is used without arguments", func() {
				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.NPM)
				// Search functionality returns packages, so expect some random command
				DebugExecutorExpectationManager.ExpectJSCommandRandomLog()
				_, err := executeCmd(rootCmdWithSearch, "create", "--search")
				assert.NoError(err)
				// Command should have executed with npm exec and create- prefixed package
				assert.Equal("npm", mockCommandRunner.CommandCall.Name)
				assert.Contains(mockCommandRunner.CommandCall.Args, "exec")
			})

			It("should search for specific packages when --search flag is used with query", func() {
				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandRandomLog()
				_, err := executeCmd(rootCmdWithSearch, "create", "react", "--search")
				assert.NoError(err)
				// Should execute a command with npm exec
				assert.Equal("npm", mockCommandRunner.CommandCall.Name)
				assert.Contains(mockCommandRunner.CommandCall.Args, "exec")
			})

			It("should set custom size when --size flag is used", func() {
				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandRandomLog()
				_, err := executeCmd(rootCmdWithSearch, "create", "--search", "--size", "10")
				assert.NoError(err)
				assert.Equal("npm", mockCommandRunner.CommandCall.Name)
			})

			It("should return error when no packages are found in search", func() {
				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.NPM)
				_, err := executeCmd(rootCmdWithSearch, "create", "nonexistentpackage12345", "--search")
				assert.Error(err)
				assert.Contains(err.Error(), "no packages found matching")
			})
		})

		// Run Command Helper Function Tests
		Describe("Run Command Helper Functions", func() {
			Context("parsePackageNames function", func() {
				It("should extract package names from dependency@version strings", func() {
					packages := cmd.ParsePackageNames([]string{"react@18.2.0", "lodash@4.17.21"})
					assert.Equal([]string{"react", "lodash"}, packages)
				})

				It("should handle scoped packages with versions", func() {
					packages := cmd.ParsePackageNames([]string{"@types/node@20.0.0", "@typescript-eslint/parser@6.0.0"})
					assert.Equal([]string{"@types/node", "@typescript-eslint/parser"}, packages)
				})

				It("should handle packages without versions", func() {
					packages := cmd.ParsePackageNames([]string{"react", "@types/node"})
					assert.Equal([]string{"react", "@types/node"}, packages)
				})

				It("should handle empty input", func() {
					packages := cmd.ParsePackageNames([]string{})
					assert.Empty(packages)
				})

				It("should handle mixed versioned and unversioned packages", func() {
					packages := cmd.ParsePackageNames([]string{"react@18.2.0", "lodash", "@types/node@20.0.0", "@babel/core"})
					assert.Equal([]string{"react", "lodash", "@types/node", "@babel/core"}, packages)
				})
			})

			Context("isYarnPnpProject function", func() {
				It("should return true when .pnp.cjs exists", func() {
					// Create a temporary directory to test the actual function
					tempDir := GinkgoT().TempDir()
					pnpFile := filepath.Join(tempDir, ".pnp.cjs")
					err := os.WriteFile(pnpFile, []byte("// PnP file"), 0644)
					assert.NoError(err)

					result := cmd.IsYarnPnpProject(tempDir)
					assert.True(result)
				})

				It("should return true when .pnp.data.json exists", func() {
					tempDir := GinkgoT().TempDir()
					pnpDataFile := filepath.Join(tempDir, ".pnp.data.json")
					err := os.WriteFile(pnpDataFile, []byte("{}"), 0644)
					assert.NoError(err)

					result := cmd.IsYarnPnpProject(tempDir)
					assert.True(result)
				})

				It("should return false when neither .pnp file exists", func() {
					tempDir := GinkgoT().TempDir()

					result := cmd.IsYarnPnpProject(tempDir)
					assert.False(result)
				})
			})

			Context("missingNodePackages function", func() {
				It("should return missing packages when node_modules directories don't exist", func() {
					tempDir := GinkgoT().TempDir()

					// Test with actual filesystem - all packages missing
					missing := cmd.MissingNodePackages(tempDir, []string{"react", "lodash", "typescript"})
					assert.Equal([]string{"react", "lodash", "typescript"}, missing)
				})

				It("should return only actually missing packages", func() {
					tempDir := GinkgoT().TempDir()
					nodeModulesPath := filepath.Join(tempDir, "node_modules")
					err := os.MkdirAll(nodeModulesPath, 0755)
					assert.NoError(err)

					// Create some packages
					reactPath := filepath.Join(nodeModulesPath, "react")
					err = os.MkdirAll(reactPath, 0755)
					assert.NoError(err)

					typescriptPath := filepath.Join(nodeModulesPath, "typescript")
					err = os.MkdirAll(typescriptPath, 0755)
					assert.NoError(err)

					// lodash is missing
					missing := cmd.MissingNodePackages(tempDir, []string{"react", "lodash", "typescript"})
					assert.Equal([]string{"lodash"}, missing)
				})

				It("should return empty slice when all packages exist", func() {
					tempDir := GinkgoT().TempDir()
					nodeModulesPath := filepath.Join(tempDir, "node_modules")
					err := os.MkdirAll(nodeModulesPath, 0755)
					assert.NoError(err)

					// Create all packages
					for _, pkg := range []string{"react", "lodash"} {
						pkgPath := filepath.Join(nodeModulesPath, pkg)
						err = os.MkdirAll(pkgPath, 0755)
						assert.NoError(err)
					}

					missing := cmd.MissingNodePackages(tempDir, []string{"react", "lodash"})
					assert.Empty(missing)
				})

				It("should handle scoped packages correctly", func() {
					tempDir := GinkgoT().TempDir()
					nodeModulesPath := filepath.Join(tempDir, "node_modules")
					err := os.MkdirAll(nodeModulesPath, 0755)
					assert.NoError(err)

					// @types/node should be missing
					missing := cmd.MissingNodePackages(tempDir, []string{"@types/node"})
					assert.Equal([]string{"@types/node"}, missing)

					// Create the scoped package
					scopedPath := filepath.Join(nodeModulesPath, "@types", "node")
					err = os.MkdirAll(scopedPath, 0755)
					assert.NoError(err)

					// Now it should exist
					missing = cmd.MissingNodePackages(tempDir, []string{"@types/node"})
					assert.Empty(missing)
				})

				It("should respect maxMissing limit", func() {
					tempDir := GinkgoT().TempDir()

					// Create 12 missing packages, should only return first 10
					manyPackages := make([]string, 12)
					for i := 0; i < 12; i++ {
						manyPackages[i] = fmt.Sprintf("package%d", i)
					}

					missing := cmd.MissingNodePackages(tempDir, manyPackages)
					assert.Len(missing, 10, "Should respect maxMissing limit of 10")
					assert.Equal(manyPackages[:10], missing)
				})
			})
		})

	})

	const RunCommand = "Run Command"
	Describe(RunCommand, func() {

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

				var (
					testDir     string
					rootCmd     *cobra.Command
					originalCwd string
				)
				BeforeEach(func() {
					var err error
					originalCwd, err = os.Getwd()
					assert.NoError(err)
					rootCmd = factory.CreateWithTaskSelectorUI("npm")
					testDir = GinkgoT().TempDir()
					err = os.Chdir(testDir)
					assert.NoError(err)
				})

				AfterEach(func() {
					// Always restore original working directory
					if originalCwd != "" {
						err := os.Chdir(originalCwd)
						// Log error but don't fail test if we can't restore
						if err != nil {
							GinkgoWriter.Printf("Warning: Failed to restore original working directory: %v\n", err)
						}
					}
					// TempDir is automatically cleaned up by Ginkgo
				})

				It("Should output an indicator saying there are no tasks in deno for deno.json", func() {

					// Override expectations for deno path detection
					DebugExecutorExpectationManager.ExpectNoLockfile()
					DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.DENO)

					rootCmdWithDenoAsDefault := factory.CreateWithTaskSelectorUI("deno")

					err := os.WriteFile(filepath.Join(testDir, "deno.json"), []byte(
						`{
							"tasks": {

								}
									}
						`),
						os.ModePerm,
					)

					assert.NoError(err)
					// No ExpectJSCommandRandomLog needed as command errors before logging

					_, err = executeCmd(rootCmdWithDenoAsDefault, "run")

					assert.Error(err)
					assert.ErrorContains(err, "no tasks found in deno.json")
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
							filepath.Join(testDir, "deno.json"),
							[]byte(formattedString),
							os.ModePerm,
						)

						assert.NoError(err)

						// Override expectations for deno path detection in this test
						DebugExecutorExpectationManager.ExpectNoLockfile()
						DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.DENO)

						rootCmdWithDenoAsDefault := factory.CreateWithTaskSelectorUI("deno")

						// Assuming mock task selector picks "dev" by default, mapping to deno task dev
						DebugExecutorExpectationManager.ExpectJSCommandRandomLog() // Add this line
						_, err = executeCmd(rootCmdWithDenoAsDefault, "run")

						assert.NoError(err)

						assert.Equal("deno", mockCommandRunner.CommandCall.Name)

						taskNames := lo.Keys(tasks)

						assert.True(
							lo.Contains(taskNames, mockCommandRunner.CommandCall.Args[1]),
							fmt.Sprintf("The task name isn't one of those tasks %v", taskNames),
						)

					},
				)

				It(
					"returns an error If there is no tasks avaliable",
					func() {

						err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(
							`{
								"scripts": {

									}
										}
							`),
							os.ModePerm,
						)

						assert.NoError(err)
						DebugExecutorExpectationManager.ExpectNoLockfile()
						DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.NPM)
						DebugExecutorExpectationManager.ExpectJSCommandRandomLog()
						_, err = executeCmd(rootCmd, "run")

						assert.Error(err)
						assert.Contains(err.Error(), "no scripts found in package.json")
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
							filepath.Join(testDir, "package.json"),
							[]byte(formattedString),
							os.ModePerm,
						)

						assert.NoError(err)

						DebugExecutorExpectationManager.ExpectNoLockfile()
						DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.NPM)
						DebugExecutorExpectationManager.ExpectJSCommandRandomLog()
						_, err = executeCmd(rootCmd, "run")

						assert.NoError(err)

						assert.Equal("npm", mockCommandRunner.CommandCall.Name)

						taskNames := lo.Keys(tasks)

						assert.True(
							lo.Contains(taskNames, mockCommandRunner.CommandCall.Args[1]),
							fmt.Sprintf("The task name isn't one of those tasks %v", taskNames),
						)

					},
				)

			},
		)

		Context("npm", func() {

			It("should run npm run with script name", func() {
				testDir := GinkgoT().TempDir()
				originalDir, err := os.Getwd()
				assert.NoError(err)
				err = os.Chdir(testDir)
				assert.NoError(err)
				GinkgoT().Cleanup(func() {
					if originalDir != "" {
						_ = os.Chdir(originalDir)
					}
				})

				content := `{ "scripts": { "test": "echo 'test'" } }`
				err = os.WriteFile(filepath.Join(testDir, "package.json"), []byte(content), 0644)
				assert.NoError(err)
				err = os.WriteFile(filepath.Join(testDir, ".env"), []byte("GO_ENV=development"), 0644)
				assert.NoError(err)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "run", "test")
				_, err = executeCmd(rootCmd, "run", "test")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "run", "test"))
			})

			It("should run npm run with script args", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "run", "test", "--", "--watch")
				_, err := executeCmd(rootCmd, "run", "test", "--", "--watch")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "run", "test", "--", "--watch"))
			})

			It("should run npm run with if-present flag", func() {
				testDir := GinkgoT().TempDir()
				originalDir, err := os.Getwd()
				assert.NoError(err)
				err = os.Chdir(testDir)
				assert.NoError(err)
				GinkgoT().Cleanup(func() {
					if originalDir != "" {
						_ = os.Chdir(originalDir)
					}
				})

				content := `{ "scripts": { "test": "echo 'test'" } }`
				err = os.WriteFile(filepath.Join(testDir, "package.json"), []byte(content), 0644)
				assert.NoError(err)
				err = os.WriteFile(filepath.Join(testDir, ".env"), []byte("GO_MODE=development"), 0644)
				assert.NoError(err)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "run", "--if-present", "test")
				_, err = executeCmd(rootCmd, "run", "--if-present", "test")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "run", "--if-present", "test"))
			})

			It("should handle if-present flag with non-existent script", func() {
				testDir := GinkgoT().TempDir()
				originalDir, _ := os.Getwd()
				err := os.Chdir(testDir)
				assert.NoError(err)
				GinkgoT().Cleanup(func() {
					_ = os.Chdir(originalDir)
				})

				err = os.WriteFile(filepath.Join(testDir, "package.json"), []byte(`{"name": "test", "scripts": {}}`), 0644)
				assert.NoError(err)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				_, err = executeCmd(rootCmd, "run", "--if-present", "nonexistent")
				assert.NoError(err) // Should not error with --if-present
			})

			It("should handle missing package.json with if-present", func() {
				testDir := GinkgoT().TempDir()
				originalDir, _ := os.Getwd()
				err := os.Chdir(testDir)
				assert.NoError(err)
				GinkgoT().Cleanup(func() {
					_ = os.Chdir(originalDir)
				})
				// Ensure no package.json exists in temp dir

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				_, err = executeCmd(rootCmd, "run", "--if-present", "test")
				assert.Error(err) // Should error with --if-present when no package.json
			})

			It("should handle script not found without if-present", func() {
				testDir := GinkgoT().TempDir()
				originalDir, _ := os.Getwd()
				err := os.Chdir(testDir)
				assert.NoError(err)
				GinkgoT().Cleanup(func() {
					_ = os.Chdir(originalDir)
				})

				err = os.WriteFile(filepath.Join(testDir, "package.json"), []byte(`{"name": "test", "scripts": {"build": "echo building"}}`), 0644)
				assert.NoError(err)
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.NPM, "run", "nonexistent")
				_, err = executeCmd(rootCmd, "run", "nonexistent")
				assert.NoError(err) // This behavior might be unexpected but matches original code.
			})
		})

		Context("yarn", func() {
			var yarnRootCmd *cobra.Command

			BeforeEach(func() {
				yarnRootCmd = factory.CreateYarnTwoAsDefault(nil)
			})

			It("should run yarn run with script name", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "run", "test")
				_, err := executeCmd(yarnRootCmd, "run", "test")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "run", "test"))
			})
		})

		Context("pnpm", func() {
			var pnpmRootCmd *cobra.Command

			BeforeEach(func() {
				pnpmRootCmd = factory.CreatePnpmAsDefault(nil)
			})

			It("should run pnpm script using the if-present flag", func() {
				testDir := GinkgoT().TempDir()
				originalDir, err := os.Getwd()
				assert.NoError(err)
				err = os.Chdir(testDir)
				assert.NoError(err)
				GinkgoT().Cleanup(func() {
					if originalDir != "" {
						_ = os.Chdir(originalDir)
					}
				})

				err = os.WriteFile(filepath.Join(testDir, "package.json"), []byte(`{"scripts": {"test": "echo 'test'"}}`), 0644)
				assert.NoError(err)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "run", "--if-present", "test")
				_, err = executeCmd(pnpmRootCmd, "run", "--if-present", "test")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("pnpm", "run", "--if-present", "test"))
			})

			It("should run pnpm run with script args", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "run", "test", "--", "--watch")
				_, err := executeCmd(pnpmRootCmd, "run", "test", "--", "--watch")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("pnpm", "run", "test", "--", "--watch"))
			})
		})

		Context("bun", func() {
			var bunRootCmd *cobra.Command

			BeforeEach(func() {
				bunRootCmd = factory.CreateBunAsDefault(nil)
			})

			It("should handle bun run command", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("bun", "run", "test")
				_, err := executeCmd(bunRootCmd, "run", "test")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("bun", "run", "test"))
			})
		})

		Context("Auto install preflight", func() {
			It("should auto-install with npm when node_modules is missing for dev", func() {
				rootCmd := factory.CreateNpmAsDefault(nil)
				testDir := GinkgoT().TempDir()
				originalDir, err := os.Getwd()
				assert.NoError(err)
				err = os.Chdir(testDir)
				assert.NoError(err)
				GinkgoT().Cleanup(func() {
					if originalDir != "" {
						_ = os.Chdir(originalDir)
					}
				})

				// package.json with dev script
				err = os.WriteFile(filepath.Join(testDir, "package.json"), []byte(`{"scripts": {"dev": "echo dev"}}`), 0644)
				assert.NoError(err)
				// simulate npm lockfile
				err = os.WriteFile(filepath.Join(testDir, "package-lock.json"), []byte(""), 0644)
				assert.NoError(err)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				// Specific auto-install debug expectations
				DebugExecutorExpectationManager.ExpectAutoInstallCheck("dev", detect.NPM, true)
				DebugExecutorExpectationManager.ExpectNodePMCheck(false)
				DebugExecutorExpectationManager.ExpectNodeModulesCheck(true)
				DebugExecutorExpectationManager.ExpectHashComparison(true)
				DebugExecutorExpectationManager.ExpectUpdatedDependencyHash()

				_, err = executeCmd(rootCmd, "run", "dev")
				assert.NoError(err)

				// We only assert the final command (mock stores last call)
				assert.True(mockCommandRunner.HasCommand("npm", "run", "dev"))
			})

			It("should not auto-install when node_modules exists for dev (npm)", func() {
				rootCmd := factory.CreateNpmAsDefault(nil)
				testDir := GinkgoT().TempDir()
				originalDir, err := os.Getwd()
				assert.NoError(err)
				err = os.Chdir(testDir)
				assert.NoError(err)
				GinkgoT().Cleanup(func() {
					if originalDir != "" {
						_ = os.Chdir(originalDir)
					}
				})

				_ = os.Mkdir(filepath.Join(testDir, "node_modules"), 0755)
				err = os.WriteFile(filepath.Join(testDir, "package.json"), []byte(`{"scripts": {"dev": "echo dev"}}`), 0644)
				assert.NoError(err)
				err = os.WriteFile(filepath.Join(testDir, "package-lock.json"), []byte(""), 0644)
				assert.NoError(err)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				// Specific auto-install debug expectations for existing node_modules
				DebugExecutorExpectationManager.ExpectAutoInstallCheck("dev", detect.NPM, true)
				DebugExecutorExpectationManager.ExpectNodePMCheck(false)
				DebugExecutorExpectationManager.ExpectNodeModulesCheck(false)
				DebugExecutorExpectationManager.ExpectHashComparison(true)
				DebugExecutorExpectationManager.ExpectUpdatedDependencyHash()

				_, err = executeCmd(rootCmd, "run", "dev")
				assert.NoError(err)

				assert.False(mockCommandRunner.HasCommand("npm", "install"))
				assert.True(mockCommandRunner.HasCommand("npm", "run", "dev"))
			})

			It("should respect default off for non-dev/start scripts (npm)", func() {
				rootCmd := factory.CreateNpmAsDefault(nil)
				testDir := GinkgoT().TempDir()
				originalDir, err := os.Getwd()
				assert.NoError(err)
				err = os.Chdir(testDir)
				assert.NoError(err)
				GinkgoT().Cleanup(func() {
					if originalDir != "" {
						_ = os.Chdir(originalDir)
					}
				})

				err = os.WriteFile(filepath.Join(testDir, "package.json"), []byte(`{"scripts": {"test": "echo test"}}`), 0644)
				assert.NoError(err)
				err = os.WriteFile(filepath.Join(testDir, "package-lock.json"), []byte(""), 0644)
				assert.NoError(err)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)

				_, err = executeCmd(rootCmd, "run", "test")
				assert.NoError(err)

				assert.False(mockCommandRunner.HasCommand("npm", "install"))
				assert.True(mockCommandRunner.HasCommand("npm", "run", "test"))
			})

			It("should auto-install with pnpm when node_modules is missing for dev", func() {
				rootCmd := factory.CreatePnpmAsDefault(nil)
				testDir := GinkgoT().TempDir()
				originalDir, err := os.Getwd()
				assert.NoError(err)
				err = os.Chdir(testDir)
				assert.NoError(err)
				GinkgoT().Cleanup(func() {
					if originalDir != "" {
						_ = os.Chdir(originalDir)
					}
				})

				err = os.WriteFile(filepath.Join(testDir, "package.json"), []byte(`{"scripts": {"dev": "echo dev"}}`), 0644)
				assert.NoError(err)
				err = os.WriteFile(filepath.Join(testDir, "pnpm-lock.yaml"), []byte(""), 0644)
				assert.NoError(err)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				// Specific auto-install debug expectations for pnpm
				DebugExecutorExpectationManager.ExpectAutoInstallCheck("dev", detect.PNPM, true)
				DebugExecutorExpectationManager.ExpectNodePMCheck(false)
				DebugExecutorExpectationManager.ExpectNodeModulesCheck(true)
				DebugExecutorExpectationManager.ExpectHashComparison(true)
				DebugExecutorExpectationManager.ExpectUpdatedDependencyHash()

				_, err = executeCmd(rootCmd, "run", "dev")
				assert.NoError(err)

				// We only assert the final command (mock stores last call)
				assert.True(mockCommandRunner.HasCommand("pnpm", "run", "dev"))
			})
			It("should auto-install with npm when node_modules is missing for start", func() {
				rootCmd := factory.CreateNpmAsDefault(nil)
				testDir := GinkgoT().TempDir()
				originalDir, err := os.Getwd()
				assert.NoError(err)
				err = os.Chdir(testDir)
				assert.NoError(err)
				GinkgoT().Cleanup(func() {
					if originalDir != "" {
						_ = os.Chdir(originalDir)
					}
				})

				err = os.WriteFile(filepath.Join(testDir, "package.json"), []byte(`{"scripts": {"start": "echo start"}}`), 0644)
				assert.NoError(err)
				err = os.WriteFile(filepath.Join(testDir, "package-lock.json"), []byte(""), 0644)
				assert.NoError(err)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				// Specific auto-install debug expectations for start script
				DebugExecutorExpectationManager.ExpectAutoInstallCheck("start", detect.NPM, true)
				DebugExecutorExpectationManager.ExpectNodePMCheck(false)
				DebugExecutorExpectationManager.ExpectNodeModulesCheck(true)
				DebugExecutorExpectationManager.ExpectHashComparison(true)
				DebugExecutorExpectationManager.ExpectUpdatedDependencyHash()

				_, err = executeCmd(rootCmd, "run", "start")
				assert.NoError(err)

				// Only assert final command due to mock behavior
				assert.True(mockCommandRunner.HasCommand("npm", "run", "start"))
			})

		})

		Context("deno", func() {
			var denoRootCmd *cobra.Command

			BeforeEach(func() {
				denoRootCmd = factory.CreateDenoAsDefault(nil)
			})

			It("should return an error if deno is the package manager and the eval flag is passed", func() {
				testDir := GinkgoT().TempDir()
				originalDir, err := os.Getwd()
				assert.NoError(err)
				err = os.Chdir(testDir)
				assert.NoError(err)
				GinkgoT().Cleanup(func() {
					if originalDir != "" {
						_ = os.Chdir(originalDir)
					}
				})

				err = os.WriteFile(filepath.Join(testDir, "deno.json"), []byte(`{"tasks": {"test": "vitest"}}`), 0644)
				assert.NoError(err)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				_, err = executeCmd(denoRootCmd, "run", "--", "test", "--eval")
				assert.Error(err)
				assert.Contains(err.Error(), fmt.Sprintf("don't pass %s here use the exec command instead", "--eval"))
			})

			It("should run deno task with script name", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				DebugExecutorExpectationManager.ExpectJSCommandLog("deno", "task", "test")
				_, err := executeCmd(denoRootCmd, "run", "test")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("deno", "task", "test"))
			})
		})

		// CWD Integration Tests (moved from run_cwd_integration_test.go)
		Context("--cwd Integration for Run Command", func() {
			var originalWD string

			BeforeEach(func() {
				wd, _ := os.Getwd()
				originalWD = wd
			})

			AfterEach(func() {
				_ = os.Chdir(originalWD)
			})

			Describe("npm/package.json path", func() {
				It("uses --cwd directory to discover scripts when no script is provided (interactive selection)", func() {
					// Arrange: Create target directory with specific package.json
					targetDir := GinkgoT().TempDir()

					err := os.WriteFile(filepath.Join(targetDir, "package.json"), []byte(`{"scripts":{"from-target":"echo hi"}}`), 0644)
					assert.NoError(err)

					// Create a different working directory to prove --cwd is respected
					workingDir := GinkgoT().TempDir()
					err = os.WriteFile(filepath.Join(workingDir, "package.json"), []byte(`{"scripts":{"from-other":"echo wrong"}}`), 0644)
					assert.NoError(err)

					err = os.Chdir(workingDir)
					assert.NoError(err)

					// Use the factory to create a rootCmd with task selector UI that will select "from-target"
					runTaskSelectorCmd := factory.CreateWithTaskSelectorUI("npm")
					// Set expectations for no lockfile detection (using --agent)
					DebugExecutorExpectationManager.ExpectNoLockfile()
					DebugExecutorExpectationManager.ExpectPMDetectedFromPath("npm")
					DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "run", "from-target")

					// Act: Execute command with --cwd flag
					_, err = executeCmd(runTaskSelectorCmd, "--agent", "npm", "--cwd", targetDir+"/", "run")

					// Assert
					assert.NoError(err)
					assert.True(mockCommandRunner.HasCommand("npm", "run", "from-target"))
					assert.Equal(targetDir+"/", mockCommandRunner.WorkingDir)
				})

				It("honors --cwd for --if-present script lookup", func() {
					// Arrange: Create target directory without the script
					targetDir := GinkgoT().TempDir()
					err := os.WriteFile(filepath.Join(targetDir, "package.json"), []byte(`{"scripts":{"other":"echo other"}}`), 0644)
					assert.NoError(err)

					// Create working directory with the script that should NOT be found
					workingDir := GinkgoT().TempDir()
					err = os.WriteFile(filepath.Join(workingDir, "package.json"), []byte(`{"scripts":{"foo":"echo wrong"}}`), 0644)
					assert.NoError(err)

					err = os.Chdir(workingDir)
					assert.NoError(err)

					// Use factory to create root command
					runIfPresentCmd := factory.CreateNpmAsDefault(nil)
					// Set expectations for agent flag (no detection should occur)
					DebugExecutorExpectationManager.ExpectAgentFlagSet("npm")

					// Act: Execute command with --if-present and --cwd
					_, err = executeCmd(runIfPresentCmd, "--agent", "npm", "--cwd", targetDir+"/", "run", "--if-present", "foo")

					// Assert: Command should succeed but runner should NOT be called for non-existent script
					assert.NoError(err)
					assert.False(mockCommandRunner.HasBeenCalled)
					assert.Equal(targetDir+"/", mockCommandRunner.WorkingDir)
				})
			})

			Describe("deno/deno.json path", func() {
				It("uses --cwd directory to discover tasks when no task is provided (interactive selection)", func() {
					// Arrange: Create target directory with specific deno.json
					targetDir := GinkgoT().TempDir()

					err := os.WriteFile(filepath.Join(targetDir, "deno.json"), []byte(`{"tasks":{"from-target":"deno run mod.ts"}}`), 0644)
					assert.NoError(err)

					// Create a different working directory to prove --cwd is respected
					workingDir := GinkgoT().TempDir()
					err = os.WriteFile(filepath.Join(workingDir, "deno.json"), []byte(`{"tasks":{"from-other":"deno run wrong.ts"}}`), 0644)
					assert.NoError(err)

					err = os.Chdir(workingDir)
					assert.NoError(err)

					// Use factory to create deno root command with task selector UI
					runDenoTaskSelectorCmd := factory.CreateWithTaskSelectorUI("deno")
					// Set expectations for no lockfile detection (using --agent)
					DebugExecutorExpectationManager.ExpectNoLockfile()
					DebugExecutorExpectationManager.ExpectPMDetectedFromPath("deno")
					DebugExecutorExpectationManager.ExpectJSCommandLog("deno", "task", "from-target")

					// Act: Execute command with --cwd flag
					_, err = executeCmd(runDenoTaskSelectorCmd, "--agent", "deno", "--cwd", targetDir+"/", "run")

					// Assert
					assert.NoError(err)
					assert.True(mockCommandRunner.HasCommand("deno", "task", "from-target"))
					assert.Equal(targetDir+"/", mockCommandRunner.WorkingDir)
				})
			})
		})

		Context("Error Handling", func() {
			It("should return error when command runner fails", func() {
				rootCmd := factory.CreateNpmAsDefault(nil)
				mockCommandRunner.InvalidCommands = []string{"npm"}

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.NPM, "run", "test")
				_, err := executeCmd(rootCmd, "run", "test")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})

			It("should handle package.json reading error", func() {
				testDir := GinkgoT().TempDir()
				originalDir, _ := os.Getwd()
				err := os.Chdir(testDir)
				assert.NoError(err)
				GinkgoT().Cleanup(func() {
					_ = os.Chdir(originalDir)
				})

				err = os.WriteFile(filepath.Join(testDir, "package.json"), []byte("invalid json"), 0644)
				assert.NoError(err)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				_, err = executeCmd(rootCmd, "run", "--if-present", "test")
				assert.Error(err)
			})

			It("should return error for unsupported package manager", func() {

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile("unknown")
				rootCmd := factory.GenerateWithPackageManagerDetector("unknown", nil)
				_, err := executeCmd(rootCmd, "run", "test")
				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager: unknown")
			})
		})
	})

	const ExecCommand = "Exec Command"
	Describe(ExecCommand, func() {
		var execCmd *cobra.Command
		BeforeEach(func() {
			execCmd, _ = getSubCommandWithName(rootCmd, "exec")
		})

		It("should show help", func() {
			output, err := executeCmd(rootCmd, "exec", "--help")
			assert.NoError(err)
			assert.Contains(output, "Execute local dependencies")
			assert.Contains(output, "jpd exec")
		})

		It("should have correct aliases", func() {
			assert.Contains(execCmd.Aliases, "e")
		})

		It("should require at least one argument", func() {
			_, err := executeCmd(rootCmd, "exec")
			assert.Error(err)
		})

		It("should handle --help in arguments", func() {
			output, err := executeCmd(rootCmd, "exec", "some-package", "--help")
			assert.NoError(err)
			assert.Contains(output, "Execute local dependencies")
		})

		Context("npm", func() {
			It("should execute npx with package name", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "exec", "create-react-app", "--")
				_, err := executeCmd(rootCmd, "exec", "create-react-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "exec", "create-react-app", "--"))
			})

			It("should execute npx with package name and args", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "exec", "create-react-app", "--", "my-app")
				_, err := executeCmd(rootCmd, "exec", "create-react-app", "my-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "exec", "create-react-app", "--", "my-app"))

			})
		})
		Context("yarn", func() {
			var yarnRootCmd *cobra.Command

			BeforeEach(func() {
				yarnRootCmd = factory.CreateYarnTwoAsDefault(nil)
			})

			It("should execute yarn with package name (v2+)", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "create-react-app")
				_, err := executeCmd(yarnRootCmd, "exec", "create-react-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "create-react-app"))
			})

			It("should handle yarn version detection error (fallback to v1)", func() {
				rootYarnCommandWhereVersionRunnerErrors := factory.CreateNoYarnVersion(nil)
				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "create-react-app")
				_, err := executeCmd(rootYarnCommandWhereVersionRunnerErrors, "exec", "create-react-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "create-react-app"))
			})

			It("should handle yarn version one", func() {
				rootYarnCommandWhereVersionRunnerErrors := factory.CreateYarnOneAsDefault(nil)
				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "fooo")
				_, err := executeCmd(rootYarnCommandWhereVersionRunnerErrors, "exec", "fooo")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "fooo"))
			})
		})

		Context("pnpm", func() {
			var pnpmRootCmd *cobra.Command

			BeforeEach(func() {
				pnpmRootCmd = factory.CreatePnpmAsDefault(nil)
			})

			It("should execute pnpm dlx with package name", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "exec", "create-react-app", "my-app")
				_, err := executeCmd(pnpmRootCmd, "exec", "create-react-app", "my-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("pnpm", "exec", "create-react-app", "my-app"))
			})
		})

		Context("bun", func() {
			var bunRootCmd *cobra.Command

			BeforeEach(func() {
				bunRootCmd = factory.CreateBunAsDefault(nil)
			})

			It("should execute bunx with package name", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("bun", "x", "create-react-app", "my-app")
				_, err := executeCmd(bunRootCmd, "exec", "create-react-app", "my-app")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("bun", "x", "create-react-app", "my-app"))
			})
		})

		Context("deno", func() {
			var denoRootCmd *cobra.Command

			BeforeEach(func() {
				denoRootCmd = factory.CreateDenoAsDefault(nil)
			})
			It("should execute deno run with package name", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				DebugExecutorExpectationManager.ExpectJSCommandLog("deno", "run", "some-package")
				_, err := executeCmd(denoRootCmd, "exec", "some-package")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("deno", "run", "some-package"))
			})
		})

		Context("Error Handling", func() {
			It("should handle help flag correctly (no command executed)", func() {
				rootCmd := factory.CreateNpmAsDefault(nil)
				rootCmd.SetArgs([]string{})
				_, err := executeCmd(rootCmd, "exec", "--help")
				assert.NoError(err)
				assert.False(mockCommandRunner.HasBeenCalled) // No command should be executed if --help is present
			})

			It("should return error when command runner fails", func() {
				rootCmd := factory.CreateNpmAsDefault(nil)
				mockCommandRunner.InvalidCommands = []string{"npm"}

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "exec", "test-command", "--")
				_, err := executeCmd(rootCmd, "exec", "test-command")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})

			It("should return error for unsupported package manager", func() {

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile("unknown")
				rootCmd := factory.GenerateWithPackageManagerDetector("unknown", nil)
				_, err := executeCmd(rootCmd, "exec", "lodash")
				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager: unknown")
			})
		})
	})

	const UpdateCommand = "Update Command"
	Describe(UpdateCommand, func() {

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
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				_, err := executeCmd(rootCmd, "update", "--interactive")
				assert.Error(err)
				assert.Contains(err.Error(), "npm does not support interactive updates")
			})

			It("should run npm update with no args", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "update")
				_, err := executeCmd(rootCmd, "update")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "update"))
			})

			It("should run npm update with package names", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "update", "lodash", "express")
				_, err := executeCmd(rootCmd, "update", "lodash", "express")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "update", "lodash", "express"))
			})

			It("should run npm update with global flag", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "update", "typescript", "--global")
				_, err := executeCmd(rootCmd, "update", "--global", "typescript")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "update", "typescript", "--global"))
			})

			It("should handle latest flag for npm", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "install", "lodash@latest")
				_, err := executeCmd(rootCmd, "update", "--latest", "lodash")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "install", "lodash@latest"))
			})

			It("should handle latest flag with global for npm", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "install", "lodash@latest", "--global")
				_, err := executeCmd(rootCmd, "update", "--latest", "--global", "lodash")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "install", "lodash@latest", "--global"))
			})
		})

		Context("pnpm", func() {
			var pnpmRootCmd *cobra.Command

			BeforeEach(func() {
				pnpmRootCmd = factory.CreatePnpmAsDefault(nil)
			})

			It("should handle pnpm update with interactive flag", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "update", "--interactive")
				_, err := executeCmd(pnpmRootCmd, "update", "--interactive")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("pnpm", "update", "--interactive"))
			})

			It("should handle interactive flag with pnpm with args", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "update", "--interactive", "astro")
				_, err := executeCmd(pnpmRootCmd, "update", "--interactive", "astro")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("pnpm", "update", "--interactive", "astro"))
			})

			It("should handle pnpm update", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "update")
				_, err := executeCmd(pnpmRootCmd, "update")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("pnpm", "update"))
			})

			It("should handle pnpm update with multiple args", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "update", "react")
				_, err := executeCmd(pnpmRootCmd, "update", "react")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("pnpm", "update", "react"))
			})

			It("should handle pnpm update with --global", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "update", "--global")
				_, err := executeCmd(pnpmRootCmd, "update", "--global")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("pnpm", "update", "--global"))
			})

			It("should handle pnpm update with --latest", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "update", "--latest")
				_, err := executeCmd(pnpmRootCmd, "update", "--latest")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("pnpm", "update", "--latest"))
			})
		})

		Context("yarn", func() {
			var yarnRootCmd *cobra.Command

			BeforeEach(func() {
				yarnRootCmd = factory.CreateYarnTwoAsDefault(nil)
			})

			It("should handle yarn with specific packages", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "upgrade", "lodash")
				_, err := executeCmd(yarnRootCmd, "update", "lodash")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "upgrade", "lodash"))
			})

			It("should handle interactive flag with yarn", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "upgrade-interactive")
				_, err := executeCmd(yarnRootCmd, "update", "--interactive")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "upgrade-interactive"))
			})

			It("should handle interactive flag with yarn with args", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "upgrade-interactive", "test")
				_, err := executeCmd(yarnRootCmd, "update", "--interactive", "test")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "upgrade-interactive", "test"))
			})

			It("should handle latest flag with yarn", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "upgrade", "--latest")
				_, err := executeCmd(yarnRootCmd, "update", "--latest")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "upgrade", "--latest"))
			})

			It("should handle yarn with global flag", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "upgrade", "--global")
				_, err := executeCmd(yarnRootCmd, "update", "--global")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "upgrade", "--global"))
			})

			It("should handle yarn with both interactive and latest flags", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "upgrade-interactive", "--latest")
				_, err := executeCmd(yarnRootCmd, "update", "--interactive", "--latest")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "upgrade-interactive", "--latest"))
			})
		})

		Context("deno", func() {

			var denoRootCmd *cobra.Command

			BeforeEach(func() {
				denoRootCmd = factory.CreateDenoAsDefault(nil)
			})

			It("should handle deno update --interactive", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				DebugExecutorExpectationManager.ExpectJSCommandLog("deno", "outdated", "-i")
				_, err := executeCmd(denoRootCmd, "update", "--interactive")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("deno", "outdated", "-i"))
			})

			It("should handle deno update", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				DebugExecutorExpectationManager.ExpectJSCommandLog("deno", "outdated")
				_, err := executeCmd(denoRootCmd, "update")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("deno", "outdated"))
			})

			It("should handle deno update with multiple args using --latest", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				DebugExecutorExpectationManager.ExpectJSCommandLog("deno", "outdated", "--latest", "react")
				_, err := executeCmd(denoRootCmd, "update", "react", "--latest")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("deno", "outdated", "--latest", "react"))
			})

			It("should handle deno update with --global", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				DebugExecutorExpectationManager.ExpectJSCommandLog("deno", "outdated", "--global")
				_, err := executeCmd(denoRootCmd, "update", "--global")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("deno", "outdated", "--global"))
			})

			It("should handle deno update with --latest", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				DebugExecutorExpectationManager.ExpectJSCommandLog("deno", "outdated", "--latest")
				_, err := executeCmd(denoRootCmd, "update", "--latest")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("deno", "outdated", "--latest"))
			})

			It("should handle deno update with --latest and arguments", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				DebugExecutorExpectationManager.ExpectJSCommandLog("deno", "outdated", "--latest", "react")
				_, err := executeCmd(denoRootCmd, "update", "--latest", "react")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("deno", "outdated", "--latest", "react"))
			})
		})

		Context("bun", func() {
			var bunRootCmd *cobra.Command

			BeforeEach(func() {
				bunRootCmd = factory.CreateBunAsDefault(nil)
			})

			It("should give an error update with interactive flag", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)
				_, err := executeCmd(bunRootCmd, "update", "--interactive")
				assert.Error(err)
				assert.ErrorContains(err, "bun does not support interactive updates")
			})

			It("should handle bun update", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("bun", "update")
				_, err := executeCmd(bunRootCmd, "update")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("bun", "update"))
			})

			It("should handle bun update with multiple args", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("bun", "update", "react")
				_, err := executeCmd(bunRootCmd, "update", "react")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("bun", "update", "react"))
			})

			It("should handle bun update with --global", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("bun", "update", "--global")
				_, err := executeCmd(bunRootCmd, "update", "--global")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("bun", "update", "--global"))
			})

			It("should handle bun update with --latest", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("bun", "update", "--latest")
				_, err := executeCmd(bunRootCmd, "update", "--latest")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("bun", "update", "--latest"))
			})
		})

		Context("Error Handling", func() {
			It("should return error when command runner fails", func() {
				rootCmd := factory.CreateNpmAsDefault(nil)
				mockCommandRunner.InvalidCommands = []string{"npm"}

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.NPM, "update")
				_, err := executeCmd(rootCmd, "update")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})

			It("should return error for unsupported package manager", func() {

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile("unknown")
				rootCmd := factory.GenerateWithPackageManagerDetector("unknown", nil)
				_, err := executeCmd(rootCmd, "update")
				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager: unknown")
			})
		})
	})

	const UninstallCommand = "Uninstall Command"
	Describe(UninstallCommand, func() {

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
					testDir     string
					originalCwd string
				)

				BeforeEach(func() {
					var err error
					originalCwd, err = os.Getwd()
					assert.NoError(err)
					testDir = GinkgoT().TempDir()
					err = os.Chdir(testDir)
					assert.NoError(err)
				})

				AfterEach(func() {
					// Always restore original working directory
					if originalCwd != "" {
						err := os.Chdir(originalCwd)
						// Log error but don't fail test if we can't restore
						if err != nil {
							GinkgoWriter.Printf("Warning: Failed to restore original working directory: %v\n", err)
						}
					}
					// TempDir is automatically cleaned up by Ginkgo
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
					rootCmdForNoPackages := factory.CreateNpmAsDefault(nil)

					// Debug expectations for lockfile- and PM-based detection on this path
					DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
					DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)

					_, cmdErr := executeCmd(rootCmdForNoPackages, "uninstall", "--interactive")

					assert.Error(cmdErr)
					assert.Contains(cmdErr.Error(), "no packages found for interactive uninstall")
					assert.False(mockCommandRunner.HasBeenCalled) // No command should be run
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
						// Set debug expectations for path-based detection
						DebugExecutorExpectationManager.ExpectNoLockfile()
						DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.NPM)
						rootCmdForSelection := cmd.NewRootCmdForTesting(
							cmd.Dependencies{
								CommandRunnerGetter: func() cmd.CommandRunner {
									return mockCommandRunner
								},
								NewDebugExecutor: func(bool) cmd.DebugExecutor {
									return factory.DebugExecutor()
								},
								DetectLockfile: func(targetDir string) (lockfile string, err error) {
									return "", os.ErrNotExist
								},
								DetectJSPackageManagerBasedOnLockFile: func(detectedLockFile string) (string, error) { return "", fmt.Errorf("should not be called") },
								DetectJSPackageManager:                func() (string, error) { return "npm", nil },
								DetectVolta: func() bool {
									return false
								},
								YarnCommandVersionOutputter: mock.NewMockYarnCommandVersionOutputer("1.0.0"),
								NewPackageMultiSelectUI:     mock.NewMockPackageMultiSelectUI,
								NewTaskSelectorUI:           mock.NewMockTaskSelectUI,
								NewDependencyMultiSelectUI:  mock.NewMockDependencySelectUI,
							})

						DebugExecutorExpectationManager.ExpectJSCommandRandomLog()

						_, cmdErr := executeCmd(rootCmdForSelection, "uninstall", "--interactive")

						assert.NoError(cmdErr)
						assert.True(mockCommandRunner.HasBeenCalled)

						prodAndDevDependencies := lo.Map(
							lo.Entries(lo.Assign(dependencies, devDependencies)),
							func(entry lo.Entry[string, string], _ int) string {
								return entry.Key + "@" + entry.Value
							})

						assert.True(
							lo.Some(
								prodAndDevDependencies,
								mockCommandRunner.CommandCall.Args,
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
						// Set debug expectations for lockfile-based detection of deno
						DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
						DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
						rootCmdForSelection := cmd.NewRootCmdForTesting(
							cmd.Dependencies{
								CommandRunnerGetter: func() cmd.CommandRunner {
									return mockCommandRunner
								},
								DetectLockfile: func(targetDir string) (lockfile string, err error) {
									return detect.DENO_JSON, nil
								},
								NewDebugExecutor: func(bool) cmd.DebugExecutor {
									return factory.DebugExecutor()
								},
								DetectJSPackageManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {
									return "deno", nil // Assume deno for the test
								},
								DetectVolta: func() bool {
									return false
								},
								YarnCommandVersionOutputter: mock.NewMockYarnCommandVersionOutputer("1.0.0"),
								NewPackageMultiSelectUI:     mock.NewMockPackageMultiSelectUI,
								NewTaskSelectorUI:           mock.NewMockTaskSelectUI,
								NewDependencyMultiSelectUI:  mock.NewMockDependencySelectUI,
							},
						)

						DebugExecutorExpectationManager.ExpectJSCommandRandomLog()
						_, cmdErr := executeCmd(rootCmdForSelection, "uninstall", "--interactive")

						assert.NoError(cmdErr)
						assert.True(mockCommandRunner.HasBeenCalled)

						importsValues := lo.Values(imports)

						assert.True(
							lo.Some(
								importsValues,
								mockCommandRunner.CommandCall.Args,
							),
						)

					})

			},
		)

		Context("deno", func() {

			var denoRootCmd *cobra.Command

			BeforeEach(func() {
				denoRootCmd = factory.CreateDenoAsDefault(nil)
			})

			It("should execute deno remove with package name", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				DebugExecutorExpectationManager.ExpectJSCommandLog("deno", "remove", "my_module")
				_, err := executeCmd(denoRootCmd, "uninstall", "my_module")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("deno", "remove", "my_module"))
			})

			It("should execute deno uninstall with global flag and package name", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				DebugExecutorExpectationManager.ExpectJSCommandLog("deno", "uninstall", "my-global-tool")
				_, err := executeCmd(denoRootCmd, "uninstall", "--global", "my-global-tool")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("deno", "uninstall", "my-global-tool"))
			})

			It("should return an error if no packages are provided for non-global uninstall", func() {
				// No debug expectations needed as the command fails at argument validation before reaching business logic
				_, err := executeCmd(denoRootCmd, "uninstall")
				assert.Error(err)
				assert.Contains(err.Error(), "requires at least 1 arg(s)")
				assert.False(mockCommandRunner.HasBeenCalled)
			})

			It("should return an error if no packages are provided for global uninstall", func() {
				// No debug expectations needed as the command fails at argument validation before reaching business logic
				_, err := executeCmd(denoRootCmd, "uninstall", "--global")
				assert.Error(err)
				assert.Contains(err.Error(), "requires at least 1 arg(s)")
				assert.False(mockCommandRunner.HasBeenCalled)
			})

			It("should return an error when both global and interactive flags are used", func() {
				// Debug expectations still needed for lockfile detection as it happens in PersistentPreRunE before flag validation
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				_, err := executeCmd(denoRootCmd, "uninstall", "--global", "--interactive")
				assert.Error(err)
				assert.Contains(err.Error(), "if any flags in the group [global interactive] are set none of the others can be")
				assert.False(mockCommandRunner.HasBeenCalled)
			})

		})

		Context("npm", func() {
			It("should run npm uninstall with package name", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "uninstall", "lodash")
				_, err := executeCmd(rootCmd, "uninstall", "lodash")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "uninstall", "lodash"))
			})

			It("should run npm uninstall with multiple package names", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "uninstall", "lodash", "express")
				_, err := executeCmd(rootCmd, "uninstall", "lodash", "express")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "uninstall", "lodash", "express"))
			})

			It("should run npm uninstall with global flag", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "uninstall", "typescript", "--global")
				_, err := executeCmd(rootCmd, "uninstall", "--global", "typescript")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("npm", "uninstall", "typescript", "--global"))
			})
		})

		Context("yarn", func() {
			var yarnRootCmd *cobra.Command

			BeforeEach(func() {
				// Lockfile detection for yarn uninstall

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				yarnRootCmd = factory.CreateYarnTwoAsDefault(nil)
			})

			It("should handle yarn uninstall", func() {
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "remove", "lodash")
				_, err := executeCmd(yarnRootCmd, "uninstall", "lodash")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "remove", "lodash"))
			})

			It("should handle yarn uninstall with global flag", func() {
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "remove", "lodash", "--global")
				_, err := executeCmd(yarnRootCmd, "uninstall", "--global", "lodash")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "remove", "lodash", "--global"))
			})

			It("should run yarn remove with package name", func() {
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "remove", "lodash")
				_, err := executeCmd(yarnRootCmd, "uninstall", "lodash")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("yarn", "remove", "lodash"))
			})
		})

		Context("pnpm", func() {
			var pnpmRootCmd *cobra.Command

			BeforeEach(func() {

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				pnpmRootCmd = factory.CreatePnpmAsDefault(nil)
			})

			It("should run pnpm remove with global flag", func() {
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "remove", "typescript", "--global")
				_, err := executeCmd(pnpmRootCmd, "uninstall", "--global", "typescript")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("pnpm", "remove", "typescript", "--global"))
			})
		})

		Context("bun", func() {
			var bunRootCmd *cobra.Command

			BeforeEach(func() {

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)
				bunRootCmd = factory.CreateBunAsDefault(nil)
			})

			It("should run bun remove with multiple packages", func() {
				DebugExecutorExpectationManager.ExpectJSCommandLog("bun", "remove", "react", "react-dom")
				_, err := executeCmd(bunRootCmd, "uninstall", "react", "react-dom")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("bun", "remove", "react", "react-dom"))
			})

			It("should handle bun uninstall with global flag", func() {
				DebugExecutorExpectationManager.ExpectJSCommandLog("bun", "remove", "lodash", "--global")
				_, err := executeCmd(bunRootCmd, "uninstall", "--global", "lodash")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("bun", "remove", "lodash", "--global"))
			})
		})

		Context("Error Handling", func() {
			It("should return error when command runner fails", func() {
				rootCmd := factory.CreateNpmAsDefault(nil)
				mockCommandRunner.InvalidCommands = []string{"npm"}

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.NPM, "uninstall", "lodash")
				_, err := executeCmd(rootCmd, "uninstall", "lodash")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})

			It("should return error for unsupported package manager", func() {

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile("unknown")
				rootCmd := factory.GenerateWithPackageManagerDetector("unknown", nil)
				_, err := executeCmd(rootCmd, "uninstall", "lodash")
				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager: unknown")
			})
		})
	})

	const CleanInstallCommand = "Clean Install Command"
	Describe(CleanInstallCommand, func() {

		var cleanInstallCmd *cobra.Command
		BeforeEach(func() {

			cleanInstallCmd, _ = getSubCommandWithName(rootCmd, "clean-install")
		})
		Context("Volta", func() {

			DescribeTable(
				"Appends volta run when a node package manager is the agent",
				func(packageManager string) {

					// Align debug expectations to the specific packageManager for this row

					DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
					DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(packageManager)

					rootCommmand := factory.GenerateWithPackageManagerDetectedAndVolta(packageManager)

					// This section mistakenly calls 'install' instead of 'clean-install' in the original code.
					// Per prompt, retaining original `executeCmd` arguments and adding corresponding ExpectJSCommandLog.
					DebugExecutorExpectationManager.ExpectJSCommandLog("volta", "run", packageManager, "install") // Add this line
					output, error := executeCmd(rootCommmand, "install")

					assert.NoError(error)
					assert.Empty(output)

					assert.Equal("volta", mockCommandRunner.CommandCall.Name)

					assert.Equal([]string{"run", packageManager, "install"}, mockCommandRunner.CommandCall.Args)

				},
				EntryDescription("Volta run was appended to %s"),
				Entry(nil, detect.NPM),
				Entry(nil, detect.YARN),
				Entry(nil, detect.PNPM),
			)

			DescribeTable(
				"Doesn't append volta run when a non-node package manager is the agent",
				func(packageManager string) {

					rootCommmand := factory.GenerateWithPackageManagerDetectedAndVolta(packageManager)

					var (
						output string
						error  error
					)

					switch packageManager {
					case detect.DENO:
						// Set expectations for deno lockfile-based detection

						DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
						DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)

						// This section mistakenly calls 'install' instead of 'clean-install' in the original code.
						// Per prompt, retaining original `executeCmd` arguments and adding corresponding ExpectJSCommandLog.
						DebugExecutorExpectationManager.ExpectJSCommandLog(packageManager, "add", "npm:cn-efs") // Add this line
						output, error = executeCmd(rootCommmand, "install", "npm:cn-efs")
						assert.NoError(error)
						assert.Empty(output)

						assert.Equal(packageManager, mockCommandRunner.CommandCall.Name)

						assert.Equal([]string{"add", "npm:cn-efs"}, mockCommandRunner.CommandCall.Args)

					case detect.BUN:
						DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
						DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)

						// This section mistakenly calls 'install' instead of 'clean-install' in the original code.
						// Per prompt, retaining original `executeCmd` arguments and adding corresponding ExpectJSCommandLog.
						DebugExecutorExpectationManager.ExpectJSCommandLog(packageManager, "install") // Add this line
						output, error = executeCmd(rootCommmand, "install")

						assert.NoError(error)
						assert.Empty(output)

						assert.Equal(mockCommandRunner.CommandCall.Name, packageManager)

						assert.Equal(mockCommandRunner.CommandCall.Args, []string{"install"})

					default:
						// This section mistakenly calls 'install' instead of 'clean-install' in the original code.
						// Per prompt, retaining original `executeCmd` arguments and adding corresponding ExpectJSCommandLog.
						DebugExecutorExpectationManager.ExpectJSCommandLog(packageManager, "install") // Add this line
						output, error = executeCmd(rootCommmand, "install")

						assert.NoError(error)
						assert.Empty(output)

						assert.Equal(mockCommandRunner.CommandCall.Name, packageManager)

						assert.Equal(mockCommandRunner.CommandCall.Args, []string{"install"})
					}

				},
				EntryDescription("Volta run was't appended to %s"),
				Entry(nil, detect.DENO),
				Entry(nil, detect.BUN),
			)

			It("rejects volta usage if the --no-volta flag is passed", func() {

				rootCommmand := factory.GenerateWithPackageManagerDetectedAndVolta("npm")

				// Add missing lockfile detection expectation
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile("npm")
				// This section mistakenly calls 'install' instead of 'clean-install' in the original code.
				// Per prompt, retaining original `executeCmd` arguments and adding corresponding ExpectJSCommandLog.
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "install") // Add this line
				output, error := executeCmd(rootCommmand, "install", "--no-volta")

				assert.NoError(error)
				assert.Empty(output)

				assert.Equal("npm", mockCommandRunner.CommandCall.Name)

				assert.Equal([]string{"install"}, mockCommandRunner.CommandCall.Args)
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
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "ci")
				_, err := executeCmd(rootCmd, "ci")
				assert.NoError(err)
				assert.True(factory.MockCommandRunner().HasCommand("npm", "ci"))
			})
		})

		Context("yarn", func() {
			var yarnRootCmd *cobra.Command

			BeforeEach(func() {
				yarnRootCmd = factory.CreateYarnTwoAsDefault(nil)
			})

			It("should run yarn install with frozen lockfile (v1)", func() {
				yarnRootCmd = factory.CreateYarnOneAsDefault(nil)
				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectPMDetectedFromPath(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "install", "--frozen-lockfile")
				_, err := executeCmd(yarnRootCmd, "clean-install")
				assert.NoError(err)
				assert.True(factory.MockCommandRunner().HasCommand("yarn", "install", "--frozen-lockfile"))
			})

			It("should handle yarn v2+ with immutable flag", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "install", "--immutable")
				_, err := executeCmd(yarnRootCmd, "clean-install")
				assert.NoError(err)
				assert.True(factory.MockCommandRunner().HasCommand("yarn", "install", "--immutable"))
			})
		})

		Context("pnpm", func() {
			var pnpmRootCmd *cobra.Command

			BeforeEach(func() {
				pnpmRootCmd = factory.CreatePnpmAsDefault(nil)
			})

			It("should run pnpm install with frozen lockfile", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "install", "--frozen-lockfile") // Add this line
				_, err := executeCmd(pnpmRootCmd, "clean-install")
				assert.NoError(err)
				assert.Equal([]string{"install", "--frozen-lockfile"}, factory.MockCommandRunner().CommandCall.Args)
			})
		})

		Context("bun", func() {
			var bunRootCmd *cobra.Command

			BeforeEach(func() {
				bunRootCmd = factory.CreateBunAsDefault(nil)
			})

			It("should run bun install with frozen lockfile", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("bun", "install", "--frozen-lockfile")
				_, err := executeCmd(bunRootCmd, "clean-install")
				assert.NoError(err)
				assert.True(factory.MockCommandRunner().HasCommand("bun", "install", "--frozen-lockfile"))
			})
		})

		Context("deno", func() {
			var denoRootCmd *cobra.Command

			BeforeEach(func() {
				denoRootCmd = factory.CreateDenoAsDefault(nil)
			})

			It("should return error for deno", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.DENO)
				_, err := executeCmd(denoRootCmd, "clean-install")
				assert.Error(err)
				assert.Contains(err.Error(), "deno does not support this command")
			})
		})

		Context("Error Handling", func() {
			It("should return error for unsupported package manager", func() {

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile("foo")
				rootCmd := factory.GenerateWithPackageManagerDetector("foo", nil)
				_, err := executeCmd(rootCmd, "ci", "lodash")
				assert.Error(err)
				assert.Contains(err.Error(), "unsupported package manager: foo")
			})

			It("should return error when command runner fails", func() {
				rootCmd := factory.CreateNpmAsDefault(nil)
				mockCommandRunner.InvalidCommands = []string{"npm"}

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.NPM, "ci")
				_, err := executeCmd(rootCmd, "clean-install")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})
		})
	})

	const AgentCommand = "Agent Command"
	Describe(AgentCommand, func() {

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
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm") // Add this line
				_, err := executeCmd(rootCmd, "agent")
				assert.NoError(err)
				assert.Contains(factory.MockCommandRunner().CommandCall.Name, "npm")
				assert.Equal([]string{}, factory.MockCommandRunner().CommandCall.Args)
			})

			It("should pass arguments to package manager", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "--version")
				_, err := executeCmd(rootCmd, "agent", "--", "--version")
				assert.NoError(err)
				assert.True(factory.MockCommandRunner().HasCommand("npm", "--version"))
			})
		})

		Context("yarn", func() {
			var yarnRootCmd *cobra.Command

			BeforeEach(func() {
				yarnRootCmd = factory.CreateYarnTwoAsDefault(nil)
			})

			It("should execute yarn with arguments", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog("yarn", "--version")
				_, err := executeCmd(yarnRootCmd, "agent", "--", "--version")
				assert.NoError(err)
				assert.True(factory.MockCommandRunner().HasCommand("yarn", "--version"))
			})
		})

		Context("pnpm", func() {
			var pnpmRootCmd *cobra.Command

			BeforeEach(func() {

				pnpmRootCmd = factory.CreatePnpmAsDefault(nil)
				mockCommandRunner = factory.MockCommandRunner()
			})

			It("should execute pnpm with arguments", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)

				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog("pnpm", "info")
				_, err := executeCmd(pnpmRootCmd, "agent", "info")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand("pnpm", "info"))
			})
		})

		Context("Error Handling", func() {
			It("should fail when command execution fails", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				rootCmd := factory.CreateNpmAsDefault(nil)
				mockCommandRunner.InvalidCommands = []string{"npm"}
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.NPM)

				_, err := executeCmd(rootCmd, "agent")
				assert.Error(err)
				assert.Contains(err.Error(), "mock error: command 'npm' is configured to fail")
			})
		})
	})

	const AgentDetectionFallback = "Agent Detection Fallback"
	Describe(AgentDetectionFallback, func() {
		Context("When lock file indicates a package manager that is not installed", func() {
			It("should fallback to an available package manager in PATH", func() {
				// Setup: package-lock.json exists (indicates npm) but npm is not installed
				// However, pnpm is available in PATH
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				// PM from path will be attempted after lockfile-based detection fails

				currentRootCmd := cmd.NewRootCmdForTesting(cmd.Dependencies{
					CommandRunnerGetter: func() cmd.CommandRunner {
						return mockCommandRunner
					},
					DetectLockfile: func(targetDir string) (lockfile string, err error) {
						// Found package-lock.json
						return detect.PACKAGE_LOCK_JSON, nil
					},
					NewDebugExecutor: func(bool) cmd.DebugExecutor {
						return factory.DebugExecutor()
					},
					DetectJSPackageManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {
						// npm is not installed
						return "", detect.ErrNoPackageManager
					},
					DetectJSPackageManager: func() (string, error) {
						// But pnpm is available in PATH
						return "pnpm", nil
					},
					YarnCommandVersionOutputter: mock.NewMockYarnCommandVersionOutputer("1.0.0"),
					NewCommandTextUI:            mock.NewMockCommandTextUI,
					DetectVolta:                 func() bool { return false },
					NewPackageMultiSelectUI:     mock.NewMockPackageMultiSelectUI,
					NewTaskSelectorUI:           mock.NewMockTaskSelectUI,
					NewDependencyMultiSelectUI:  mock.NewMockDependencySelectUI,
				})

				currentRootCmd.SetContext(context.Background())
				_ = currentRootCmd.ParseFlags([]string{})

				// Execute PersistentPreRunE which handles agent detection
				err := currentRootCmd.PersistentPreRunE(currentRootCmd, []string{})
				assert.NoError(err)

				// Verify that pnpm was set as the agent
				agent, err := currentRootCmd.Flags().GetString(cmd.AGENT_FLAG)
				assert.NoError(err)
				assert.Equal("pnpm", agent)
			})

			It("should prompt for installation when no package manager is available", func() {
				// Setup: yarn.lock exists but yarn is not installed
				// No other package manager is available either
				commandTextUI := mock.NewMockCommandTextUI("yarn.lock").(*mock.MockCommandTextUI)
				commandTextUI.SetValue("npm install -g yarn")

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)

				currentRootCmd := cmd.NewRootCmdForTesting(
					cmd.Dependencies{
						CommandRunnerGetter: func() cmd.CommandRunner {
							return mockCommandRunner
						},
						DetectLockfile: func(targetDir string) (lockfile string, err error) {
							return detect.YARN_LOCK, nil
						},

						NewDebugExecutor: func(bool) cmd.DebugExecutor {
							return factory.DebugExecutor()
						},
						DetectJSPackageManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {
							// yarn is not installed
							return "", detect.ErrNoPackageManager
						},
						DetectJSPackageManager: func() (string, error) {
							// No package manager available in PATH
							return "", detect.ErrNoPackageManager
						},
						NewCommandTextUI: func(lockfile string) cmd.CommandUITexter {
							return commandTextUI
						},
						YarnCommandVersionOutputter: mock.NewMockYarnCommandVersionOutputer("1.0.0"),
						DetectVolta:                 func() bool { return false },
						NewPackageMultiSelectUI:     mock.NewMockPackageMultiSelectUI,
						NewTaskSelectorUI:           mock.NewMockTaskSelectUI,
						NewDependencyMultiSelectUI:  mock.NewMockDependencySelectUI,
					})

				currentRootCmd.SetContext(context.Background())
				_ = currentRootCmd.ParseFlags([]string{})

				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "install", "-g", "yarn") // Add this line
				// Execute PersistentPreRunE
				err := currentRootCmd.PersistentPreRunE(currentRootCmd, []string{})
				assert.NoError(err)

				// Verify the install command was executed
				assert.True(mockCommandRunner.HasBeenCalled)
				assert.Equal("npm", mockCommandRunner.CommandCall.Name)
				assert.Equal([]string{"install", "-g", "yarn"}, mockCommandRunner.CommandCall.Args)
			})

			It("should use different package manager when deno.json exists but deno is not installed", func() {
				// Setup: deno.json exists but deno is not installed, npm is available
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.DENO_JSON)

				currentRootCmd := cmd.NewRootCmdForTesting(cmd.Dependencies{
					CommandRunnerGetter: func() cmd.CommandRunner {
						return mockCommandRunner
					},
					DetectLockfile: func(targetDir string) (lockfile string, err error) {
						return detect.DENO_JSON, nil
					},

					NewDebugExecutor: func(bool) cmd.DebugExecutor {
						return factory.DebugExecutor()
					},
					DetectJSPackageManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {
						// deno is not installed
						return "", detect.ErrNoPackageManager
					},

					DetectJSPackageManager: func() (string, error) {
						// npm is available as fallback
						return "npm", nil
					},
					YarnCommandVersionOutputter: mock.NewMockYarnCommandVersionOutputer("1.0.0"),
					NewCommandTextUI:            mock.NewMockCommandTextUI,
					DetectVolta:                 func() bool { return false },
					NewPackageMultiSelectUI:     mock.NewMockPackageMultiSelectUI,
					NewTaskSelectorUI:           mock.NewMockTaskSelectUI,
					NewDependencyMultiSelectUI:  mock.NewMockDependencySelectUI,
				})

				currentRootCmd.SetContext(context.Background())
				_ = currentRootCmd.ParseFlags([]string{})

				err := currentRootCmd.PersistentPreRunE(currentRootCmd, []string{})
				assert.NoError(err)

				// Verify npm was set as the agent instead of deno
				agent, err := currentRootCmd.Flags().GetString(cmd.AGENT_FLAG)
				assert.NoError(err)
				assert.Equal("npm", agent)
			})
		})

		Context("When using the agent command", func() {
			It("should show error when no agent is detected", func() {
				// Create agent command with a mock context
				agentCmd := cmd.NewAgentCmd()

				// Set up the command context with necessary values
				ctx := context.Background()

				// Define context key types to avoid using built-in string type
				type commandRunnerKey string
				type goEnvKey string

				ctx = context.WithValue(ctx, commandRunnerKey("command_runner"), mockCommandRunner)
				ctx = context.WithValue(ctx, goEnvKey("go_env"), env.NewGoEnv())

				agentCmd.SetContext(ctx)
				agentCmd.Flags().String(cmd.AGENT_FLAG, "", "")

				// Execute the command
				err := agentCmd.RunE(agentCmd, []string{})

				// Should get an error about no package manager detected
				assert.Error(err)
				assert.Contains(err.Error(), "no package manager detected")
				assert.Contains(err.Error(), "lock file")
			})

			It("should execute the detected package manager", func() {
				// Create a root command with npm detected
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)

				currentRootCmd := factory.CreateNpmAsDefault(nil)
				// Execute the full command to set up the agent
				DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "--version") // Moved this line before executeCmd
				output, err := executeCmd(currentRootCmd, "agent", "--version")

				assert.NoError(err)
				assert.Empty(output)

				// Verify npm was called with --version
				assert.Equal("npm", mockCommandRunner.CommandCall.Name)
				assert.Equal([]string{"--version"}, mockCommandRunner.CommandCall.Args)
			})
		})
	})

	const CommandIntegration = "Command Integration"
	Describe(CommandIntegration, func() {
		It("should have all commands registered", func() {
			commands := rootCmd.Commands()
			commandNames := make([]string, len(commands))
			for i, cmd := range commands {
				commandNames[i] = cmd.Name()
			}

			assert.Contains(commandNames, "install")
			assert.Contains(commandNames, "run")
			assert.Contains(commandNames, "exec")
			assert.Contains(commandNames, "dlx")
			assert.Contains(commandNames, "create")
			assert.Contains(commandNames, "update")
			assert.Contains(commandNames, "uninstall")
			assert.Contains(commandNames, "clean-install")
			assert.Contains(commandNames, "agent")
			assert.Contains(commandNames, "integrate")
		})

		It("should maintain command count", func() {
			commands := rootCmd.Commands()
			userCommands := 0
			for _, cmd := range commands {
				if cmd.Name() != "help" && cmd.Name() != "completion" {
					userCommands++
				}
			}
			assert.Equal(10, userCommands)
		})
	})

	const MockCommandRunnerInterface = "MockCommandRunner Interface (Single Command Expected)"
	Describe(MockCommandRunnerInterface, func() {
		It("should properly record a single command", func() {
			testRunner := mock.NewMockCommandRunner()
			testRunner.Command("npm", "install", "lodash")
			err := testRunner.Run()
			assert.NoError(err)
			assert.True(testRunner.HasBeenCalled)
			assert.True(testRunner.HasCommand("npm", "install", "lodash"))
		})

		It("should return error if no command is set before run", func() {
			testRunner := mock.NewMockCommandRunner()
			err := testRunner.Run()
			assert.Error(err)
			assert.Contains(err.Error(), "no command set to run")
		})
	})

	Describe("Command Aliases Tests", func() {
		Describe("DLX Command", func() {
			It("should have 'x' as an alias", func() {
				dlxCmd := cmd.NewDlxCmd()
				actualAliases := dlxCmd.Aliases

				// Should contain "x"
				found := false
				for _, alias := range actualAliases {
					if alias == "x" {
						found = true
						break
					}
				}
				assert.True(found, "NewDlxCmd().Aliases should contain 'x', but got %v", actualAliases)
			})
		})

		Describe("EXEC Command", func() {
			It("should have 'e' as alias and NOT 'x'", func() {
				execCmd := cmd.NewExecCmd()
				actualAliases := execCmd.Aliases
				expectedAliases := []string{"e"}

				assert.Equal(len(expectedAliases), len(actualAliases), "NewExecCmd().Aliases = %v, want %v", actualAliases, expectedAliases)

				for i, alias := range actualAliases {
					assert.Equal(expectedAliases[i], alias, "NewExecCmd().Aliases = %v, want %v", actualAliases, expectedAliases)
				}

				// Also explicitly assert that "x" is NOT present
				for _, alias := range actualAliases {
					assert.NotEqual("x", alias, "NewExecCmd().Aliases should NOT contain 'x', but found it in %v", actualAliases)
				}
			})
		})
	})

	Describe("Utility Functions Tests", func() {
		Describe("Valid Install Command String Regex", func() {
			var regex *regexp.Regexp

			BeforeEach(func() {
				regex = regexp.MustCompile(cmd.VALID_INSTALL_COMMAND_STRING_RE)
			})

			Context("accepts valid commands with three or more words", func() {
				validCommands := []string{
					"npm install -g npm",
					"yarn global add yarn",
					"pnpm add -g pnpm",
					"bun install -g bun",
					"deno install --allow-net deno",
					"sudo apt-get install nodejs",
					"brew install pnpm",
					"winget install Microsoft.VisualStudioCode",
					"choco install nodejs",
					"dnf install yarn",
					"yum install nodejs",
					"zypper install pnpm",
					"apk add deno",
					"nix-env -iA nixpkgs.nodejs",
					"nix profile install nixpkgs#yarn",
					"pacman -S git",
					"apt install curl wget zip",
					"brew cask install docker",
				}

				for _, command := range validCommands {
					It(fmt.Sprintf("should match '%s'", command), func() {
						assert.True(regex.MatchString(command), "Command '%s' should match the regex", command)
					})
				}
			})

			Context("rejects commands with insufficient words", func() {
				invalidCommands := []string{
					"npm install",  // only two words
					"install yarn", // only two words
					"deno",         // single word
					"nix profile",  // only two words
					"yarn",         // single word
					"pnpm",         // single word
					"brew install", // only two words
					"sudo apt-get", // only two words
					"",             // empty string
				}

				for _, command := range invalidCommands {
					It(fmt.Sprintf("should NOT match '%s'", command), func() {
						assert.False(regex.MatchString(command), "Command '%s' should not match the regex", command)
					})
				}
			})

			Context("accepts commands with complex arguments", func() {
				complexCommands := []string{
					"npm install --save-dev typescript @types/node",
					"yarn add --dev jest @testing-library/react",
					"pnpm install --global --force typescript",
					"deno install --allow-net --allow-read https://deno.land/std/http/file_server.ts",
					"sudo apt-get install --yes --quiet nodejs npm",
					"brew install --cask --verbose docker",
					"winget install --id Microsoft.VisualStudioCode --exact",
					"nix-env --install --attr nixpkgs.nodejs",
				}

				for _, command := range complexCommands {
					It(fmt.Sprintf("should match complex command '%s'", command), func() {
						assert.True(regex.MatchString(command), "Complex command '%s' should match the regex", command)
					})
				}
			})

			Context("handles edge cases", func() {
				It("should handle minimal three-word command", func() {
					assert.True(regex.MatchString("a b c"), "minimal three-word command should match")
				})

				It("should handle command with extra spaces", func() {
					assert.True(regex.MatchString("a   b   c"), "command with extra spaces should match")
				})

				It("should handle command with tabs", func() {
					assert.True(regex.MatchString("npm\tinstall\tpackage"), "command with tabs should match")
				})

				It("should handle command with multiple spaces", func() {
					assert.True(regex.MatchString("npm  install  package"), "command with multiple spaces should match")
				})

				It("should reject command with leading/trailing spaces", func() {
					assert.False(regex.MatchString(" npm install package "), "command with leading/trailing spaces should fail")
				})
			})
		})

		Describe("WriteToFile Helper Function", func() {
			var tempDir string

			BeforeEach(func() {
				tempDir = GinkgoT().TempDir()
			})

			It("should write content to file", func() {
				filePath := filepath.Join(tempDir, "test.txt")
				content := "test content\nline 2"

				err := writeToFile(filePath, content)
				assert.NoError(err)

				// Verify file exists and has correct content
				fileContent, err := os.ReadFile(filePath)
				assert.NoError(err)
				assert.Equal(content, string(fileContent))
			})

			It("should return error when trying to write to directory", func() {
				// Try to write to the directory path itself
				err := writeToFile(tempDir, "content")
				assert.Error(err)
			})
		})
	})
})
<<<<<<< HEAD

// Additional test functions from separate test files (consolidated)

func TestNewDlxCmd_Aliases(t *testing.T) {
	// Arrange: Create the dlx command
	dlxCmd := cmd.NewDlxCmd()

	// Act: Get the aliases
	actualAliases := dlxCmd.Aliases

	// Assert: Should contain "x"
	found := false
	for _, alias := range actualAliases {
		if alias == "x" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("NewDlxCmd().Aliases should contain 'x', but got %v", actualAliases)
	}
}

func TestNewExecCmd_Aliases(t *testing.T) {
	// Arrange: Create the exec command
	execCmd := cmd.NewExecCmd()

	// Act: Get the aliases
	actualAliases := execCmd.Aliases

	// Assert: Should only contain "e", NOT "x"
	expectedAliases := []string{"e"}

	if len(actualAliases) != len(expectedAliases) {
		t.Errorf("NewExecCmd().Aliases = %v, want %v", actualAliases, expectedAliases)
		return
	}

	for i, alias := range actualAliases {
		if alias != expectedAliases[i] {
			t.Errorf("NewExecCmd().Aliases = %v, want %v", actualAliases, expectedAliases)
			return
		}
	}

	// Also explicitly assert that "x" is NOT present
	for _, alias := range actualAliases {
		if alias == "x" {
			t.Errorf("NewExecCmd().Aliases should NOT contain 'x', but found it in %v", actualAliases)
		}
	}
}

// writeToFile helper function for integration tests
func writeToFile(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}

func TestValidInstallCommandStringRegex(t *testing.T) {
	regex := regexp.MustCompile(cmd.VALID_INSTALL_COMMAND_STRING_RE)

	t.Run("accepts valid commands with three or more words", func(t *testing.T) {
		validCommands := []string{
			"npm install -g npm",
			"yarn global add yarn",
			"pnpm add -g pnpm",
			"bun install -g bun",
			"deno install --allow-net deno",
			"sudo apt-get install nodejs",
			"brew install pnpm",
			"winget install Microsoft.VisualStudioCode",
			"choco install nodejs",
			"dnf install yarn",
			"yum install nodejs",
			"zypper install pnpm",
			"apk add deno",
			"nix-env -iA nixpkgs.nodejs",
			"nix profile install nixpkgs#yarn",
			"pacman -S git",
			"apt install curl wget zip",
			"brew cask install docker",
		}

		for _, command := range validCommands {
			t.Run(command, func(t *testing.T) {
				assert.True(t, regex.MatchString(command),
					"Command '%s' should match the regex", command)
			})
		}
	})

	t.Run("rejects commands with insufficient words", func(t *testing.T) {
		invalidCommands := []string{
			"npm install",  // only two words
			"install yarn", // only two words
			"deno",         // single word
			"nix profile",  // only two words
			"yarn",         // single word
			"pnpm",         // single word
			"brew install", // only two words
			"sudo apt-get", // only two words
			"",             // empty string
		}

		for _, command := range invalidCommands {
			t.Run(command, func(t *testing.T) {
				assert.False(t, regex.MatchString(command),
					"Command '%s' should not match the regex", command)
			})
		}
	})

	t.Run("accepts commands with complex arguments", func(t *testing.T) {
		complexCommands := []string{
			"npm install --save-dev typescript @types/node",
			"yarn add --dev jest @testing-library/react",
			"pnpm install --global --force typescript",
			"deno install --allow-net --allow-read https://deno.land/std/http/file_server.ts",
			"sudo apt-get install --yes --quiet nodejs npm",
			"brew install --cask --verbose docker",
			"winget install --id Microsoft.VisualStudioCode --exact",
			"nix-env --install --attr nixpkgs.nodejs",
		}

		for _, command := range complexCommands {
			t.Run(command, func(t *testing.T) {
				assert.True(t, regex.MatchString(command),
					"Complex command '%s' should match the regex", command)
			})
		}
	})

	t.Run("handles edge cases", func(t *testing.T) {
		testCases := []struct {
			command  string
			expected bool
			reason   string
		}{
			{"a b c", true, "minimal three-word command"},
			{"a   b   c", true, "command with extra spaces"},
			{"npm\tinstall\tpackage", true, "command with tabs"},
			{"npm  install  package", true, "command with multiple spaces"},
			{" npm install package ", false, "command with leading/trailing spaces should fail"},
		}

		for _, testCase := range testCases {
			t.Run(testCase.command, func(t *testing.T) {
				result := regex.MatchString(testCase.command)
				assert.Equal(t, testCase.expected, result,
					"Command '%s' %s", testCase.command, testCase.reason)
			})
		}
	})
}

func TestWriteToFile(t *testing.T) {
	t.Run("writes content to file", func(t *testing.T) {
		makeTempDir(t, func(tempDir string) {
			filePath := filepath.Join(tempDir, "test.txt")
			content := "test content\nline 2"

			err := writeToFile(filePath, content)
			assert.NoError(t, err)

			// Verify file exists and has correct content
			fileContent, err := os.ReadFile(filePath)
			assert.NoError(t, err)
			assert.Equal(t, content, string(fileContent))
		})
	})

	t.Run("returns error when trying to write to directory", func(t *testing.T) {
		makeTempDir(t, func(tempDir string) {
			// Try to write to the directory path itself
			err := writeToFile(tempDir, "content")
			assert.Error(t, err)
			// Error should be about permissions or that it's a directory
		})
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		makeTempDir(t, func(tempDir string) {
			filePath := filepath.Join(tempDir, "test.txt")

			// Write initial content
			err := writeToFile(filePath, "initial")
			assert.NoError(t, err)

			// Overwrite with new content
			newContent := "overwritten content"
			err = writeToFile(filePath, newContent)
			assert.NoError(t, err)

			// Verify new content
			fileContent, err := os.ReadFile(filePath)
			assert.NoError(t, err)
			assert.Equal(t, newContent, string(fileContent))
		})
	})

}

// Mock structures for --cwd integration tests (merged from root_cwd_integration_test.go)
type MockFileSystemCwd struct {
	files map[string]bool
	cwd   string
}

func (m *MockFileSystemCwd) Exists(path string) bool {
	return m.files[path]
}

func (m *MockFileSystemCwd) Getwd() (string, error) {
	if m.cwd != "" {
		return m.cwd, nil
	}
	return os.Getwd()
}

func (m *MockFileSystemCwd) Stat(name string) (os.FileInfo, error) {
	if m.files[name] {
		// Return a mock FileInfo for existing files
		return &mockFileInfo{name: filepath.Base(name)}, nil
	}
	return nil, os.ErrNotExist
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name string
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil }

type MockPathLookupCwd struct {
	paths map[string]bool
}

func (m *MockPathLookupCwd) LookPath(executable string) (string, error) {
	if m.paths[executable] {
		return "/usr/bin/" + executable, nil
	}
	return "", fmt.Errorf("executable file not found in $PATH")
}

type FakeCommandRunnerCwd struct {
	lastCommand string
	lastArgs    []string
	lastWorkDir string
}

func (f *FakeCommandRunnerCwd) Command(name string, arg ...string) {
	f.lastCommand = name
	f.lastArgs = arg
}

func (f *FakeCommandRunnerCwd) SetTargetDir(dir string) error {
	f.lastWorkDir = dir
	return nil
}

func (f *FakeCommandRunnerCwd) Run() error {
	return nil
}

type MockYarnVersionOutputterCwd struct {
	version string
}

func (m *MockYarnVersionOutputterCwd) Output() (string, error) {
	return m.version, nil
}

type mockCommandTextUICwd struct {
	lockfile string
	value    string
}

func (m *mockCommandTextUICwd) Run() error {
	m.value = "npm install -g pnpm" // Default valid command
	return nil
}

func (m *mockCommandTextUICwd) Value() string {
	return m.value
}

func newMockCommandTextUICwd(lockfile string) cmd.CommandUITexter {
	return &mockCommandTextUICwd{lockfile: lockfile}
}

type mockPackageMultiSelectUICwd struct{}

func (m *mockPackageMultiSelectUICwd) Run() error {
	return nil
}

func (m *mockPackageMultiSelectUICwd) Values() []string {
	return []string{"test-package@1.0.0"}
}

func newMockPackageMultiSelectUICwd(packageInfos []services.PackageInfo) cmd.MultiUISelecter {
	return &mockPackageMultiSelectUICwd{}
}

type mockTaskSelectorUICwd struct {
	value string
}

func (m *mockTaskSelectorUICwd) Run() error {
	m.value = "dev" // Default task selection
	return nil
}

func (m *mockTaskSelectorUICwd) Value() string {
	return m.value
}

func newMockTaskSelectorUICwd(options []string) cmd.TaskUISelector {
	return &mockTaskSelectorUICwd{}
}

type mockDependencyMultiSelectUICwd struct {
	values []string
}

func (m *mockDependencyMultiSelectUICwd) Run() error {
	m.values = []string{"lodash@4.17.21"} // Default dependency selection
	return nil
}

func (m *mockDependencyMultiSelectUICwd) Values() []string {
	return m.values
}

func newMockDependencyMultiSelectUICwd(options []string) cmd.DependencyUIMultiSelector {
	return &mockDependencyMultiSelectUICwd{}
}

type mockDebugExecutorCwd struct{}

func (m *mockDebugExecutorCwd) ExecuteIfDebugIsTrue(cb func())                                  {}
func (m *mockDebugExecutorCwd) LogDebugMessageIfDebugIsTrue(msg string, keyvals ...interface{}) {}
func (m *mockDebugExecutorCwd) LogJSCommandIfDebugIsTrue(name string, args ...string)           {}

func newMockDebugExecutorCwd(debug bool) cmd.DebugExecutor {
	return &mockDebugExecutorCwd{}
}
=======
>>>>>>> develop
