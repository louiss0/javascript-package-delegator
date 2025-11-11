package integrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
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

// Carapace Spec Generator tests
var _ = Describe("Carapace Spec Generator", func() {
	var (
		generator integrations.CarapaceSpecGenerator
		mockCmd   *cobra.Command
	)

	BeforeEach(func() {
		generator = integrations.NewCarapaceSpecGenerator()
		mockCmd = &cobra.Command{
			Use:   "jpd",
			Short: "JavaScript Package Delegator - A universal package manager interface",
		}
	})

	Describe("GenerateYAMLSpec", func() {
		It("should generate valid YAML spec with jpd name", func() {
			result, err := generator.GenerateYAMLSpec(mockCmd)

			assert.NoError(GinkgoT(), err, "Expected no error generating YAML spec")
			assert.Contains(GinkgoT(), result, "Name: javascript-package-delegator", "Expected YAML to contain 'Name: javascript-package-delegator'")
			assert.Contains(GinkgoT(), result, "# Carapace completion spec for jpd", "Expected YAML to contain header comment")
			assert.Contains(GinkgoT(), result, "Description: JavaScript Package Delegator - A universal package manager interface", "Expected YAML to contain correct description")
		})

		It("should include all JPD top-level commands", func() {
			result, err := generator.GenerateYAMLSpec(mockCmd)

			assert.NoError(GinkgoT(), err, "Expected no error generating YAML spec")

			expectedCommands := []string{
				"install:",
				"run:",
				"exec:",
				"dlx:",
				"update:",
				"uninstall:",
				"clean-install:",
				"agent:",
				"completion:",
				"integrate:", // Added integrate command
			}

			for _, expected := range expectedCommands {
				assert.Contains(GinkgoT(), result, expected, "Expected YAML to contain command: %s", expected)
			}
		})

		It("should include persistent flags with expected short forms", func() {
			result, err := generator.GenerateYAMLSpec(mockCmd)

			assert.NoError(GinkgoT(), err, "Expected no error generating YAML spec")

			// Check for agent flag with shorthand
			assert.Contains(GinkgoT(), result, "PersistentFlags:", "Expected YAML to contain PersistentFlags section")
			assert.Contains(GinkgoT(), result, "agent:", "Expected YAML to contain agent flag")
			assert.Contains(GinkgoT(), result, "shorthand: a", "Expected agent flag to have shorthand 'a'")

			// Check for debug flag with shorthand
			assert.Contains(GinkgoT(), result, "debug:", "Expected YAML to contain debug flag")
			assert.Contains(GinkgoT(), result, "shorthand: d", "Expected debug flag to have shorthand 'd'")

			// Check for cwd flag with shorthand
			assert.Contains(GinkgoT(), result, "cwd:", "Expected YAML to contain cwd flag")
			assert.Contains(GinkgoT(), result, "shorthand: C", "Expected cwd flag to have shorthand 'C'")
		})

		It("should include enumeration for agent flag", func() {
			result, err := generator.GenerateYAMLSpec(mockCmd)

			assert.NoError(GinkgoT(), err, "Expected no error generating YAML spec")

			expectedManagers := []string{
				"npm",
				"yarn",
				"pnpm",
				"bun",
				"deno",
			}

			for _, manager := range expectedManagers {
				assert.Contains(GinkgoT(), result, manager, "Expected agent enum to contain manager: %s", manager)
			}
			assert.Contains(GinkgoT(), result, "enum:\n            - npm", "Expected agent enum to be correctly formatted")
		})

		It("should include completion hints for relevant commands and flags", func() {
			result, err := generator.GenerateYAMLSpec(mockCmd)

			assert.NoError(GinkgoT(), err, "Expected no error generating YAML spec")

			// Check for package completion hints
			assert.Contains(GinkgoT(), result, "completion: $carapace.packages.npm", "Expected package completion hints for install, run, exec, dlx, update, uninstall")
			// Check for scripts completion hints
			assert.Contains(GinkgoT(), result, "completion: $carapace.scripts.npm", "Expected scripts completion hints for run")
			// Check for directory completion hints for persistent 'cwd' flag and 'warp output-dir'
			assert.Contains(GinkgoT(), result, "completion: $carapace.directories", "Expected directory completion hints")
			// Check for file completion hints for 'completion filename' and 'carapace output'
			assert.Contains(GinkgoT(), result, "completion: $carapace.files", "Expected file completion hints")
			// Check for shell completion hints for 'completion' command
			assert.Contains(GinkgoT(), result, "completion:\n        description: Generate shell completion scripts\n        flags:\n            filename:\n                shorthand: f\n                description: Output completion script to file\n                completion: $carapace.files\n            with-shorthand:\n                shorthand: s\n                description: Include shorthand alias functions\n        completion: $carapace.shells", "Expected shell completion hints for completion command")

		})

		It("should include completion command flags", func() {
			result, err := generator.GenerateYAMLSpec(mockCmd)

			assert.NoError(GinkgoT(), err, "Expected no error generating YAML spec")

			// Check for completion command flags
			assert.Contains(GinkgoT(), result, "completion:\n        description: Generate shell completion scripts\n        flags:\n            filename:\n                shorthand: f\n                description: Output completion script to file", "Expected completion command to have filename flag with shorthand")
			assert.Contains(GinkgoT(), result, "with-shorthand:\n                shorthand: s\n                description: Include shorthand alias functions", "Expected completion command to have with-shorthand flag with shorthand")
		})

		It("should include integrate command and its nested commands with flags", func() {
			result, err := generator.GenerateYAMLSpec(mockCmd)

			assert.NoError(GinkgoT(), err, "Expected no error generating YAML spec")

			// Check for 'integrate' command
			assert.Contains(GinkgoT(), result, "integrate:\n        description: Generate integration files for external tools", "Expected integrate command")

			// Check for 'integrate warp' nested command and its flags
			assert.Contains(GinkgoT(), result, "warp:\n                description: Generate Warp terminal workflow files\n                flags:\n                    output-dir:\n                        shorthand: o\n                        description: Output directory for workflow files\n                        completion: $carapace.directories", "Expected integrate warp command with output-dir flag")

			// Check for 'integrate carapace' nested command and its flags
			assert.Contains(GinkgoT(), result, "carapace:\n                description: Generate Carapace completion spec file\n                flags:\n                    output:\n                        shorthand: o\n                        description: Output file for Carapace spec\n                        completion: $carapace.files", "Expected integrate carapace command with output flag")
			assert.Contains(GinkgoT(), result, "stdout:\n                        description: Print Carapace spec to stdout instead of installing", "Expected integrate carapace command with stdout flag")
		})

		It("should generate valid YAML structure", func() {
			result, err := generator.GenerateYAMLSpec(mockCmd)

			assert.NoError(GinkgoT(), err, "Expected no error generating YAML spec")

			// Remove header comments for YAML parsing
			yamlContent := strings.Split(result, "---\n")
			assert.Len(GinkgoT(), yamlContent, 2, "Expected YAML content to be split into header and body")
			yamlPart := yamlContent[1]
			assert.NotEmpty(GinkgoT(), yamlPart, "Expected non-empty YAML content after header")

			// Basic check for structure; deeper parsing could be done with a YAML library
			assert.Contains(GinkgoT(), yamlPart, "Name: javascript-package-delegator", "Expected main YAML content to contain name")
			assert.Contains(GinkgoT(), yamlPart, "Commands:", "Expected YAML to contain 'Commands' section")
			assert.Contains(GinkgoT(), yamlPart, "PersistentFlags:", "Expected YAML to contain 'PersistentFlags' section")
		})
	})

	Describe("NushellCompletionScript", func() {
		It("should return embedded nushell completion script", func() {
			result := integrations.NushellCompletionScript()

			assert.NotEmpty(GinkgoT(), result, "Expected non-empty nushell completion script")
			// The embedded file should contain nushell extern declarations
			assert.Contains(GinkgoT(), result, "extern", "Expected nushell script to contain extern declarations")
			assert.Contains(GinkgoT(), result, "jpd", "Expected nushell script to be for jpd")
		})
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

			if err == nil {
				Skip("GenerateJPDWorkflows unexpectedly succeeded; skipping error assertion in permissive environment")
			}

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
			// Should have separators between documents (9 workflows = 8 separators + 1 at start)
			assert.True(GinkgoT(), separatorCount >= 8, "Expected at least 8 document separators, got %d", separatorCount)
		})
	})
})

// Carapace Directory Resolution tests
var _ = Describe("Carapace Directory Resolution", func() {
	Describe("CarapaceSpecsDir", func() {
		It("should return a valid directory path", func() {
			dir, err := integrations.CarapaceSpecsDir()
			assert.NoError(GinkgoT(), err, "Expected no error getting carapace specs dir")
			assert.NotEmpty(GinkgoT(), dir, "Expected non-empty directory path")
			assert.Contains(GinkgoT(), dir, "carapace", "Expected path to contain 'carapace'")
			assert.Contains(GinkgoT(), dir, "specs", "Expected path to contain 'specs'")
		})
	})

	Describe("DefaultCarapaceSpecPath", func() {
		It("should return the full path for the spec file", func() {
			path, err := integrations.DefaultCarapaceSpecPath()
			assert.NoError(GinkgoT(), err, "Expected no error getting default spec path")
			assert.NotEmpty(GinkgoT(), path, "Expected non-empty spec path")
			assert.Contains(GinkgoT(), path, integrations.CarapaceSpecFileName, "Expected path to contain spec filename")
			assert.Contains(GinkgoT(), path, "carapace", "Expected path to contain 'carapace'")
			assert.Contains(GinkgoT(), path, "specs", "Expected path to contain 'specs'")
		})
	})

	Describe("EnsureDir", func() {
		var tempDir string

		BeforeEach(func() {
			var err error
			tempDir, err = os.MkdirTemp("", "ensure-dir-test-*")
			assert.NoError(GinkgoT(), err, "Expected no error creating temp dir")
			// Remove the dir so we can test creating it
			err = os.RemoveAll(tempDir)
			assert.NoError(GinkgoT(), err, "Expected no error removing temp dir")
		})

		AfterEach(func() {
			if tempDir != "" {
				_ = os.RemoveAll(tempDir) // Ignore cleanup errors
			}
		})

		It("should create directory if it doesn't exist", func() {
			// Directory should not exist initially
			_, err := os.Stat(tempDir)
			assert.True(GinkgoT(), os.IsNotExist(err), "Expected directory to not exist initially")

			// EnsureDir should create it
			err = integrations.EnsureDir(tempDir)
			assert.NoError(GinkgoT(), err, "Expected no error ensuring directory")

			// Directory should now exist
			info, err := os.Stat(tempDir)
			assert.NoError(GinkgoT(), err, "Expected no error stating directory")
			assert.True(GinkgoT(), info.IsDir(), "Expected path to be a directory")
		})

		It("should create nested directories", func() {
			nestedDir := filepath.Join(tempDir, "nested", "deeply", "nested")

			// Nested directory should not exist initially
			_, err := os.Stat(nestedDir)
			assert.True(GinkgoT(), os.IsNotExist(err), "Expected nested directory to not exist initially")

			// EnsureDir should create it
			err = integrations.EnsureDir(nestedDir)
			assert.NoError(GinkgoT(), err, "Expected no error ensuring nested directory")

			// Directory should now exist
			info, err := os.Stat(nestedDir)
			assert.NoError(GinkgoT(), err, "Expected no error stating nested directory")
			assert.True(GinkgoT(), info.IsDir(), "Expected nested path to be a directory")
		})

		It("should not error if directory already exists", func() {
			// Create directory first
			err := os.MkdirAll(tempDir, 0755)
			assert.NoError(GinkgoT(), err, "Expected no error creating directory")

			// EnsureDir should still succeed
			err = integrations.EnsureDir(tempDir)
			assert.NoError(GinkgoT(), err, "Expected no error ensuring existing directory")

			// Directory should still exist
			info, err := os.Stat(tempDir)
			assert.NoError(GinkgoT(), err, "Expected no error stating directory")
			assert.True(GinkgoT(), info.IsDir(), "Expected path to be a directory")
		})
	})
})

func TestCarapaceSpecFileName(t *testing.T) {
	expected := "javascript-package-delegator.yaml"
	if integrations.CarapaceSpecFileName != expected {
		t.Errorf("Expected CarapaceSpecFileName to be %q, got %q", expected, integrations.CarapaceSpecFileName)
	}
}
