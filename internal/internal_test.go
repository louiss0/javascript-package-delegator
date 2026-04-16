package integrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	integrations "github.com/louiss0/javascript-package-delegator/internal"
)

func TestIntegrations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Internal Package Suite")
}

// Shell Alias Generator tests
var _ = Describe("Shell Alias Generator", func() {
	var (
		generator    integrations.AliasGenerator
		testAliasMap map[string][]string
	)

	BeforeEach(func() {
		generator = integrations.NewAliasGenerator()
		testAliasMap = map[string][]string{
			"install":       {"jpi", "jpadd", "jpd-install"},
			"run":           {"jpr", "jpd-run"},
			"exec":          {"jpe", "jpd-exec"},
			"dlx":           {"jpx", "jpd-dlx"},
			"update":        {"jpu", "jpup", "jpupgrade", "jpd-update"},
			"uninstall":     {"jpun", "jprm", "jpremove", "jpd-uninstall"},
			"clean-install": {"jpci", "jpd-clean-install"},
			"agent":         {"jpa", "jpd-agent"},
		}
	})

	Describe("GenerateBash", func() {
		It("should generate bash function signatures", func() {
			result := generator.GenerateBash(testAliasMap)

			expectedFunctions := []string{
				"function jpi",
				"function jpadd",
				"function jpd-install",
				"function jpr",
				"function jpd-run",
				"function jpe",
				"function jpd-exec",
				"function jpx",
				"function jpd-dlx",
			}

			for _, expected := range expectedFunctions {
				assert.Contains(GinkgoT(), result, expected, "Expected bash output to contain function signature")
			}
		})

		It("should generate completion wiring", func() {
			result := generator.GenerateBash(testAliasMap)

			expectedCompletions := []string{
				"complete -F __start_jpd jpi",
				"complete -F __start_jpd jpadd",
				"complete -F __start_jpd jpd-install",
				"complete -F __start_jpd jpr",
				"complete -F __start_jpd jpd-run",
			}

			for _, expected := range expectedCompletions {
				assert.Contains(GinkgoT(), result, expected, "Expected bash output to contain completion wiring")
			}
		})

		It("should include guard clause", func() {
			result := generator.GenerateBash(testAliasMap)
			assert.Contains(GinkgoT(), result, "command -v jpd > /dev/null || return 0", "Expected bash output to contain guard clause")
		})
	})

	Describe("GenerateZsh", func() {
		It("should generate zsh function signatures", func() {
			result := generator.GenerateZsh(testAliasMap)

			expectedFunctions := []string{
				"jpi()",
				"jpadd()",
				"jpd-install()",
				"jpr()",
				"jpd-run()",
				"jpe()",
				"jpd-exec()",
				"jpx()",
				"jpd-dlx()",
			}

			for _, expected := range expectedFunctions {
				assert.Contains(GinkgoT(), result, expected, "Expected zsh output to contain function signature")
			}
		})

		It("should generate completion wiring", func() {
			result := generator.GenerateZsh(testAliasMap)

			expectedCompletions := []string{
				"compdef _jpd jpi",
				"compdef _jpd jpadd",
				"compdef _jpd jpd-install",
				"compdef _jpd jpr",
				"compdef _jpd jpd-run",
			}

			for _, expected := range expectedCompletions {
				assert.Contains(GinkgoT(), result, expected, "Expected zsh output to contain completion wiring")
			}
		})

		It("should include guard clause", func() {
			result := generator.GenerateZsh(testAliasMap)
			assert.Contains(GinkgoT(), result, "(( $+commands[jpd] )) || return", "Expected zsh output to contain guard clause")
		})
	})

	Describe("GenerateFish", func() {
		It("should generate fish function signatures", func() {
			result := generator.GenerateFish(testAliasMap)

			expectedFunctions := []string{
				"function jpi",
				"function jpadd",
				"function jpd-install",
				"function jpr",
				"function jpd-run",
				"function jpe",
				"function jpd-exec",
				"function jpx",
				"function jpd-dlx",
			}

			for _, expected := range expectedFunctions {
				assert.Contains(GinkgoT(), result, expected, "Expected fish output to contain function signature")
			}
		})

		It("should generate completion wiring", func() {
			result := generator.GenerateFish(testAliasMap)

			expectedCompletions := []string{
				"complete -c jpi -w jpd",
				"complete -c jpadd -w jpd",
				"complete -c jpd-install -w jpd",
				"complete -c jpr -w jpd",
				"complete -c jpd-run -w jpd",
			}

			for _, expected := range expectedCompletions {
				assert.Contains(GinkgoT(), result, expected, "Expected fish output to contain completion wiring")
			}
		})

		It("should include function end markers", func() {
			result := generator.GenerateFish(testAliasMap)
			assert.Contains(GinkgoT(), result, "end", "Expected fish output to contain function 'end' markers")
		})
	})

	Describe("GenerateNushell", func() {
		It("should generate extern signatures", func() {
			result := generator.GenerateNushell(testAliasMap)

			expectedExterns := []string{
				"export extern \"jpi\"",
				"export extern \"jpadd\"",
				"export extern \"jpd-install\"",
				"export extern \"jpr\"",
				"export extern \"jpd-run\"",
				"export extern \"jpe\"",
				"export extern \"jpd-exec\"",
				"export extern \"jpx\"",
				"export extern \"jpd-dlx\"",
			}

			for _, expected := range expectedExterns {
				assert.Contains(GinkgoT(), result, expected, "Expected nushell output to contain extern signature")
			}
		})

		It("should generate function definitions", func() {
			result := generator.GenerateNushell(testAliasMap)

			expectedDefs := []string{
				"export def jpi",
				"export def jpadd",
				"export def jpd-install",
				"export def jpr",
				"export def jpd-run",
				"export def jpe",
				"export def jpd-exec",
				"export def jpx",
				"export def jpd-dlx",
			}

			for _, expected := range expectedDefs {
				assert.Contains(GinkgoT(), result, expected, "Expected nushell output to contain function definition")
			}
		})

		It("should include rest args pattern", func() {
			result := generator.GenerateNushell(testAliasMap)
			assert.Contains(GinkgoT(), result, "...args: string", "Expected nushell output to contain rest args pattern")
		})
	})

	Describe("GeneratePowerShell", func() {
		It("should generate PowerShell function signatures", func() {
			result := generator.GeneratePowerShell(testAliasMap)

			expectedFunctions := []string{
				"function jpi {",
				"function jpadd {",
				"function jpd-install {",
				"function jpr {",
				"function jpd-run {",
				"function jpe {",
				"function jpd-exec {",
				"function jpx {",
				"function jpd-dlx {",
			}

			for _, expected := range expectedFunctions {
				assert.Contains(GinkgoT(), result, expected, "Expected PowerShell output to contain function signature")
			}
		})

		It("should generate completion registration", func() {
			result := generator.GeneratePowerShell(testAliasMap)

			expectedCompletions := []string{
				"Register-ArgumentCompleter -CommandName 'jpi'",
				"Register-ArgumentCompleter -CommandName 'jpadd'",
				"Register-ArgumentCompleter -CommandName 'jpd-install'",
				"Register-ArgumentCompleter -CommandName 'jpr'",
				"Register-ArgumentCompleter -CommandName 'jpd-run'",
			}

			for _, expected := range expectedCompletions {
				assert.Contains(GinkgoT(), result, expected, "Expected PowerShell output to contain completion registration")
			}
		})

		It("should include guard clause", func() {
			result := generator.GeneratePowerShell(testAliasMap)
			assert.Contains(GinkgoT(), result, "if (-not (Get-Command jpd -ErrorAction SilentlyContinue))", "Expected PowerShell output to contain guard clause")
		})

		It("should use @args for parameter splatting", func() {
			result := generator.GeneratePowerShell(testAliasMap)
			assert.Contains(GinkgoT(), result, "jpd install @args", "Expected PowerShell output to use @args splatting")
			assert.Contains(GinkgoT(), result, "jpd run @args", "Expected PowerShell output to use @args splatting")
		})
	})
})

var _ = Describe("Nushell completion asset", func() {
	It("returns the embedded script", func() {
		result := integrations.NushellCompletionScript()

		assert.NotEmpty(GinkgoT(), result, "Expected non-empty nushell completion script")
		assert.Contains(GinkgoT(), result, "extern", "Expected nushell script to contain extern declarations")
		assert.Contains(GinkgoT(), result, "jpd", "Expected nushell script to be for jpd")
		assert.NotContains(GinkgoT(), result, "jpd integrate carapace", "Expected no legacy integrate carapace extern")
	})
})

// Warp Workflow Generator tests
var _ = Describe("Warp Workflow Generator", func() {
	var (
		generator integrations.WarpGenerator
		tempDir   string
	)

	BeforeEach(func() {
		generator = integrations.NewWarpGenerator()

		// Create a temporary directory for testing
		var err error
		tempDir, err = os.MkdirTemp("", "warp-test-*")
		assert.NoError(GinkgoT(), err, "Expected no error creating temp dir")
	})

	AfterEach(func() {
		if tempDir != "" {
			if err := os.RemoveAll(tempDir); err != nil {
				assert.NoError(GinkgoT(), err, "Error removing temporary directory")
			}
		}
	})

	Describe("DefaultWarpWorkflowsDir", func() {
		Context("when XDG_DATA_HOME is set", func() {
			It("should return XDG_DATA_HOME/warp-terminal/workflows", func() {
				// Save original env and create temp dir
				originalXDG := os.Getenv("XDG_DATA_HOME")
				defer func() { _ = os.Setenv("XDG_DATA_HOME", originalXDG) }()

				tmpDir, err := os.MkdirTemp("", "xdg-test-*")
				assert.NoError(GinkgoT(), err)
				defer func() { _ = os.RemoveAll(tmpDir) }()

				// Set XDG_DATA_HOME to temp dir
				err = os.Setenv("XDG_DATA_HOME", tmpDir)
				assert.NoError(GinkgoT(), err)

				// Call function
				result, err := integrations.DefaultWarpWorkflowsDir()

				// Assertions
				assert.NoError(GinkgoT(), err)
				expected := filepath.Join(tmpDir, "warp-terminal", "workflows")
				assert.Equal(GinkgoT(), expected, result)
			})
		})

		Context("when XDG_DATA_HOME is not set", func() {
			It("should return HOME/.local/share/warp-terminal/workflows", func() {
				// Save original env vars
				originalXDG := os.Getenv("XDG_DATA_HOME")
				originalHOME := os.Getenv("HOME")
				defer func() {
					_ = os.Setenv("XDG_DATA_HOME", originalXDG)
					_ = os.Setenv("HOME", originalHOME)
				}()

				// Create temp home dir
				tmpHome, err := os.MkdirTemp("", "home-test-*")
				assert.NoError(GinkgoT(), err)
				defer func() { _ = os.RemoveAll(tmpHome) }()

				// Unset XDG_DATA_HOME and set HOME
				err = os.Unsetenv("XDG_DATA_HOME")
				assert.NoError(GinkgoT(), err)
				_ = os.Setenv("HOME", tmpHome) // On Windows, HOME may be ignored by os.UserHomeDir

				// Call function
				result, err := integrations.DefaultWarpWorkflowsDir()

				// Determine expected dataHome per implementation
				var expectedDataHome string
				if runtime.GOOS == "windows" {
					// On Windows, DefaultWarpWorkflowsDir uses os.UserHomeDir() ignoring HOME
					realHome, e := os.UserHomeDir()
					assert.NoError(GinkgoT(), e)
					expectedDataHome = filepath.Join(realHome, ".local", "share")
				} else {
					expectedDataHome = filepath.Join(tmpHome, ".local", "share")
				}

				// Assertions
				assert.NoError(GinkgoT(), err)
				expected := filepath.Join(expectedDataHome, "warp-terminal", "workflows")
				assert.Equal(GinkgoT(), expected, result)
			})
		})
	})

	Describe("RenderJPDWorkflowsMultiDoc", func() {
		It("should render multi-document YAML with all workflows", func() {
			result, err := generator.RenderJPDWorkflowsMultiDoc()

			assert.NoError(GinkgoT(), err, "Expected no error rendering multi-doc YAML")
			assert.NotEmpty(GinkgoT(), result, "Expected non-empty result")

			// Should contain YAML document separators
			assert.Contains(GinkgoT(), result, "---", "Expected YAML multi-doc separators")

			// Should contain workflow name fields
			assert.Contains(GinkgoT(), result, "name:", "Expected workflow name fields")

			// Should contain command fields with multiline format
			assert.Contains(GinkgoT(), result, "command:", "Expected command fields")
		})

		It("should include all JPD subcommand workflows", func() {
			result, err := generator.RenderJPDWorkflowsMultiDoc()

			assert.NoError(GinkgoT(), err, "Expected no error rendering multi-doc YAML")

			expectedWorkflows := []string{
				"JPD Install",
				"JPD Run",
				"JPD Start",
				"JPD Exec",
				"JPD DLX",
				"JPD Create",
				"JPD Update",
				"JPD Uninstall",
				"JPD Clean Install",
				"JPD Agent",
			}

			for _, workflow := range expectedWorkflows {
				assert.Contains(GinkgoT(), result, workflow, "Expected workflow: %s", workflow)
			}
		})

		It("should include required workflow fields", func() {
			result, err := generator.RenderJPDWorkflowsMultiDoc()

			assert.NoError(GinkgoT(), err, "Expected no error rendering multi-doc YAML")

			// Check for author field
			assert.Contains(GinkgoT(), result, "author: the-code-fixer-23", "Expected author field")

			// Check for tags
			assert.Contains(GinkgoT(), result, "jpd", "Expected jpd tag")
			assert.Contains(GinkgoT(), result, "javascript", "Expected javascript tag")
			assert.Contains(GinkgoT(), result, "package-manager", "Expected package-manager tag")

			// Check for shells array (empty means all)
			assert.Contains(GinkgoT(), result, "shells: []", "Expected empty shells array")

			// Check for commands with template variables
			assert.Contains(GinkgoT(), result, "jpd install {{package}}", "Expected install command with template")
			assert.Contains(GinkgoT(), result, "jpd run {{script}}", "Expected run command with template")
		})

		It("should include workflow arguments for commands that need them", func() {
			result, err := generator.RenderJPDWorkflowsMultiDoc()

			assert.NoError(GinkgoT(), err, "Expected no error rendering multi-doc YAML")

			// Check for arguments fields
			assert.Contains(GinkgoT(), result, "arguments:", "Expected arguments fields")
			assert.Contains(GinkgoT(), result, "Package name to install", "Expected install argument description")
			assert.Contains(GinkgoT(), result, "Script name to run", "Expected run argument description")
		})
	})

	Describe("GenerateJPDWorkflows", func() {
		It("should create all workflow files in the specified directory", func() {
			err := generator.GenerateJPDWorkflows(tempDir)

			assert.NoError(GinkgoT(), err, "Expected no error generating workflow files")

			expectedFiles := []string{
				"jpd-install.yaml",
				"jpd-run.yaml",
				"jpd-start.yaml",
				"jpd-exec.yaml",
				"jpd-dlx.yaml",
				"jpd-create.yaml",
				"jpd-update.yaml",
				"jpd-uninstall.yaml",
				"jpd-clean-install.yaml",
				"jpd-agent.yaml",
			}

			for _, filename := range expectedFiles {
				filePath := filepath.Join(tempDir, filename)
				_, err := os.Stat(filePath)
				assert.NoError(GinkgoT(), err, "Expected file to exist: %s", filename)
			}
		})

		It("should create workflow files with valid YAML content", func() {
			err := generator.GenerateJPDWorkflows(tempDir)

			assert.NoError(GinkgoT(), err, "Expected no error generating workflow files")

			// Check content of one workflow file
			installFilePath := filepath.Join(tempDir, "jpd-install.yaml")
			content, err := os.ReadFile(installFilePath)
			assert.NoError(GinkgoT(), err, "Expected no error reading install workflow file")

			contentStr := string(content)
			assert.Contains(GinkgoT(), contentStr, "name: JPD Install", "Expected workflow name")
			assert.Contains(GinkgoT(), contentStr, "command: jpd install {{package}}", "Expected command template")
			assert.Contains(GinkgoT(), contentStr, "author: the-code-fixer-23", "Expected author")
			assert.Contains(GinkgoT(), contentStr, "tags:", "Expected tags field")
			assert.Contains(GinkgoT(), contentStr, "- jpd", "Expected jpd tag")
		})

		It("should create directory if it doesn't exist", func() {
			nestedDir := filepath.Join(tempDir, "nested", "workflows")

			// Ensure directory doesn't exist initially
			_, err := os.Stat(nestedDir)
			assert.True(GinkgoT(), os.IsNotExist(err), "Expected directory to not exist initially")

			// Generate workflows
			err = generator.GenerateJPDWorkflows(nestedDir)
			assert.NoError(GinkgoT(), err, "Expected no error generating workflow files")

			// Directory should now exist
			_, err = os.Stat(nestedDir)
			assert.NoError(GinkgoT(), err, "Expected directory to exist after generation")

			// Files should be created
			installFilePath := filepath.Join(nestedDir, "jpd-install.yaml")
			_, err = os.Stat(installFilePath)
			assert.NoError(GinkgoT(), err, "Expected install workflow file to exist")
		})

		It("should handle commands without arguments correctly", func() {
			err := generator.GenerateJPDWorkflows(tempDir)

			assert.NoError(GinkgoT(), err, "Expected no error generating workflow files")

			// Check agent workflow (no arguments)
			agentFilePath := filepath.Join(tempDir, "jpd-agent.yaml")
			content, err := os.ReadFile(agentFilePath)
			assert.NoError(GinkgoT(), err, "Expected no error reading agent workflow file")

			contentStr := string(content)
			assert.Contains(GinkgoT(), contentStr, "name: JPD Agent", "Expected agent workflow name")
			assert.Contains(GinkgoT(), contentStr, "command: jpd agent", "Expected agent command")
			// Should not contain arguments field when no arguments
			assert.NotContains(GinkgoT(), contentStr, "arguments:", "Expected no arguments field for agent")
		})

		It("should handle commands with multiple arguments correctly", func() {
			err := generator.GenerateJPDWorkflows(tempDir)

			assert.NoError(GinkgoT(), err, "Expected no error generating workflow files")

			// Check exec workflow (multiple arguments)
			execFilePath := filepath.Join(tempDir, "jpd-exec.yaml")
			content, err := os.ReadFile(execFilePath)
			assert.NoError(GinkgoT(), err, "Expected no error reading exec workflow file")

			contentStr := string(content)
			assert.Contains(GinkgoT(), contentStr, "name: JPD Exec", "Expected exec workflow name")
			assert.Contains(GinkgoT(), contentStr, "command: jpd exec {{package}} {{args}}", "Expected exec command with templates")
			assert.Contains(GinkgoT(), contentStr, "arguments:", "Expected arguments field")
			assert.Contains(GinkgoT(), contentStr, "name: package", "Expected package argument")
			assert.Contains(GinkgoT(), contentStr, "name: args", "Expected args argument")
		})

		It("should return error when unable to create output directory", func() {
			var inaccessiblePath string
			if runtime.GOOS == "windows" {
				// Use a path with invalid characters to guarantee failure on Windows
				inaccessiblePath = filepath.Join(os.TempDir(), "bad:name", "inaccessible")
			} else {
				// Create a read-only directory and attempt to write beneath it
				readOnlyDir := filepath.Join(tempDir, "readonly")
				err := os.MkdirAll(readOnlyDir, 0555)
				assert.NoError(GinkgoT(), err, "Expected no error creating read-only directory")
				defer func() { _ = os.Chmod(readOnlyDir, 0755) }()
				inaccessiblePath = filepath.Join(readOnlyDir, "inaccessible")
			}

			err := generator.GenerateJPDWorkflows(inaccessiblePath)

			// Should get an error
			assert.Error(GinkgoT(), err, "Expected error when trying to create directory in an inaccessible location")
			assert.Contains(GinkgoT(), err.Error(), "failed to create output directory", "Expected specific error message")
		})
	})

	Describe("Multi-document format", func() {
		It("should separate workflows with YAML document separators", func() {
			result, err := generator.RenderJPDWorkflowsMultiDoc()

			assert.NoError(GinkgoT(), err, "Expected no error rendering multi-doc YAML")

			// Count document separators
			separatorCount := strings.Count(result, "---")
			// Should have separators between documents (10 workflows = 9 separators + 1 at start)
			assert.True(GinkgoT(), separatorCount >= 9, "Expected at least 9 document separators, got %d", separatorCount)
		})
	})
})
