package integrations

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
)

var _ = Describe("Warp Workflow Generator", func() {
	var (
		generator WarpGenerator
		tempDir   string
	)

	BeforeEach(func() {
		generator = NewWarpGenerator()

		// Create a temporary directory for testing
		var err error
		tempDir, err = os.MkdirTemp("", "warp-test-*")
		assert.NoError(GinkgoT(), err, "Expected no error creating temp dir")
	})

	AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
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
	})

	Describe("Multi-document format", func() {
		It("should separate workflows with YAML document separators", func() {
			result, err := generator.RenderJPDWorkflowsMultiDoc()

			assert.NoError(GinkgoT(), err, "Expected no error rendering multi-doc YAML")

			// Count document separators
			separatorCount := strings.Count(result, "---")
			// Should have separators between documents (8 workflows = 7 separators + 1 at start)
			assert.True(GinkgoT(), separatorCount >= 7, "Expected at least 7 document separators, got %d", separatorCount)
		})
	})
})
