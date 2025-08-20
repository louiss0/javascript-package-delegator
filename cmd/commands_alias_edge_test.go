package cmd_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/mock"
	"github.com/louiss0/javascript-package-delegator/testutil"
)

var _ = Describe("Commands Alias and Edge Test Coverage", func() {
	assert := assert.New(GinkgoT())

	var rootCmd *cobra.Command
	mockCommandRunner := mock.NewMockCommandRunner()
	factory := testutil.NewRootCommandFactory(mockCommandRunner)
	var DebugExecutorExpectationManager = testutil.DebugExecutorExpectationManager

	BeforeEach(func() {
		rootCmd = factory.CreateNpmAsDefault(nil)
		rootCmd.SetArgs([]string{})
		factory.ResetDebugExecutor()
		DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
	})

	AfterEach(func() {
		mockCommandRunner.Reset()
		factory.DebugExecutor().AssertExpectations(GinkgoT())
	})

	Describe("Agent aliases", func() {
		Context("jpd a runs and uses detected PM", func() {
			It("should execute with npm when detected", func() {
				// Set expectations for npm detection via lockfile
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.NPM)

				_, err := executeCmd(rootCmd, "a")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand(detect.NPM))
			})

			It("should execute with yarn when detected", func() {
				rootCmd = factory.CreateRootCmdWithLockfileDetected(detect.YARN, detect.YARN_LOCK, nil, false)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.YARN)

				_, err := executeCmd(rootCmd, "a")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand(detect.YARN))
			})

			It("should execute with pnpm when detected", func() {
				rootCmd = factory.CreateRootCmdWithLockfileDetected(detect.PNPM, detect.PNPM_LOCK_YAML, nil, false)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.PNPM)

				_, err := executeCmd(rootCmd, "a")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand(detect.PNPM))
			})
		})
	})

	Describe("Update aliases", func() {
		Context("jpd up and jpd u map to update", func() {
			It("should execute npm update via 'up' alias", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.NPM, "update")

				_, err := executeCmd(rootCmd, "up")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand(detect.NPM, "update"))
			})

			It("should execute yarn upgrade via 'u' alias", func() {
				rootCmd = factory.CreateRootCmdWithLockfileDetected(detect.YARN, detect.YARN_LOCK, nil, false)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.YARN, "upgrade")

				_, err := executeCmd(rootCmd, "u")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand(detect.YARN, "upgrade"))
			})

			It("should execute pnpm update via 'up' alias", func() {
				rootCmd = factory.CreateRootCmdWithLockfileDetected(detect.PNPM, detect.PNPM_LOCK_YAML, nil, false)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.PNPM, "update")

				_, err := executeCmd(rootCmd, "up")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand(detect.PNPM, "update"))
			})
		})

		Context("update command with specific packages", func() {
			It("should handle package names correctly with npm", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.NPM, "update", "lodash", "react")

				_, err := executeCmd(rootCmd, "update", "lodash", "react")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand(detect.NPM, "update", "lodash", "react"))
			})
		})
	})

	Describe("Uninstall aliases", func() {
		Context("jpd rm and jpd remove map correctly", func() {
			It("should execute npm uninstall via 'rm' alias", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.NPM, "uninstall", "lodash")

				_, err := executeCmd(rootCmd, "rm", "lodash")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand(detect.NPM, "uninstall", "lodash"))
			})

			It("should execute yarn remove via 'remove' alias", func() {
				rootCmd = factory.CreateRootCmdWithLockfileDetected(detect.YARN, detect.YARN_LOCK, nil, false)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.YARN_LOCK)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.YARN)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.YARN, "remove", "react")

				_, err := executeCmd(rootCmd, "remove", "react")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand(detect.YARN, "remove", "react"))
			})

			It("should execute pnpm remove via 'rm' alias", func() {
				rootCmd = factory.CreateRootCmdWithLockfileDetected(detect.PNPM, detect.PNPM_LOCK_YAML, nil, false)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.PNPM, "remove", "typescript")

				_, err := executeCmd(rootCmd, "rm", "typescript")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand(detect.PNPM, "remove", "typescript"))
			})

			It("should execute bun remove via 'remove' alias", func() {
				rootCmd = factory.CreateRootCmdWithLockfileDetected(detect.BUN, detect.BUN_LOCKB, nil, false)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.BUN, "remove", "eslint")

				_, err := executeCmd(rootCmd, "remove", "eslint")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand(detect.BUN, "remove", "eslint"))
			})
		})
	})

	Describe("Clean-install alias", func() {
		Context("jpd ci already covered", func() {
			It("should execute npm ci via 'ci' alias", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.NPM, "ci")

				_, err := executeCmd(rootCmd, "ci")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand(detect.NPM, "ci"))
			})

			It("should execute pnpm install --frozen-lockfile via 'ci' alias", func() {
				rootCmd = factory.CreateRootCmdWithLockfileDetected(detect.PNPM, detect.PNPM_LOCK_YAML, nil, false)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PNPM_LOCK_YAML)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.PNPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.PNPM, "install", "--frozen-lockfile")

				_, err := executeCmd(rootCmd, "ci")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand(detect.PNPM, "install", "--frozen-lockfile"))
			})
		})
	})

	Describe("Additional edge flags", func() {
		Context("Mutual exclusivity errors", func() {
			It("should handle conflicting flags on uninstall command", func() {
				// The uninstall command has mutually exclusive --global and --interactive flags
				// This should produce an error, but we still need debug expectations for the root command pre-run
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)

				_, err := executeCmd(rootCmd, "uninstall", "--global", "--interactive")
				assert.Error(err)
				// The actual cobra error message for mutual exclusivity
				assert.Contains(err.Error(), "were all set")
			})
		})

		Context("Update command edge cases", func() {
			It("should handle interactive flag with npm (should error)", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)

				_, err := executeCmd(rootCmd, "update", "--interactive")
				assert.Error(err)
				assert.Contains(err.Error(), "npm does not support interactive updates")
			})

			It("should handle interactive flag with bun (should error)", func() {
				rootCmd = factory.CreateRootCmdWithLockfileDetected(detect.BUN, detect.BUN_LOCKB, nil, false)

				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.BUN_LOCKB)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.BUN)

				_, err := executeCmd(rootCmd, "update", "--interactive")
				assert.Error(err)
				assert.Contains(err.Error(), "bun does not support interactive updates")
			})
		})

		Context("Agent command edge cases", func() {
			It("should handle --version flag passthrough", func() {
				DebugExecutorExpectationManager.ExpectLockfileDetected(detect.PACKAGE_LOCK_JSON)
				DebugExecutorExpectationManager.ExpectPMDetectedFromLockfile(detect.NPM)
				DebugExecutorExpectationManager.ExpectJSCommandLog(detect.NPM, "--version")

				_, err := executeCmd(rootCmd, "agent", "--version")
				assert.NoError(err)
				assert.True(mockCommandRunner.HasCommand(detect.NPM, "--version"))
			})

			It("should error when no package manager is detected", func() {
				rootCmd = factory.GenerateNoDetectionAtAll("")

				DebugExecutorExpectationManager.ExpectNoLockfile()
				DebugExecutorExpectationManager.ExpectNoPMFromPath()

				_, err := executeCmd(rootCmd, "agent")
				assert.Error(err)
				// The actual error is about a command for installing a package when no PM is detected
				assert.Contains(err.Error(), "A command for installing a package is at least three words")
			})
		})
	})
})
