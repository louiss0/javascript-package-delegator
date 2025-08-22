package integrations

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Carapace Spec Generator", func() {
	var (
		generator CarapaceSpecGenerator
		mockCmd   *cobra.Command
	)

	BeforeEach(func() {
		generator = NewCarapaceSpecGenerator()
		mockCmd = &cobra.Command{
			Use:   "jpd",
			Short: "JavaScript Package Delegator - A universal package manager interface",
		}
	})

	Describe("GenerateYAMLSpec", func() {
		It("should generate valid YAML spec with jpd name", func() {
			result, err := generator.GenerateYAMLSpec(mockCmd)

			assert.NoError(GinkgoT(), err, "Expected no error generating YAML spec")
			assert.Contains(GinkgoT(), result, "name: jpd", "Expected YAML to contain 'name: jpd'")
			assert.Contains(GinkgoT(), result, "# Carapace completion spec for jpd", "Expected YAML to contain header comment")
			assert.Contains(GinkgoT(), result, "description: JavaScript Package Delegator - A universal package manager interface", "Expected YAML to contain correct description")
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
			assert.Contains(GinkgoT(), result, "persistentFlags:", "Expected YAML to contain persistentFlags section")
			assert.Contains(GinkgoT(), result, "  agent:", "Expected YAML to contain agent flag")
			assert.Contains(GinkgoT(), result, "    shorthand: a", "Expected agent flag to have shorthand 'a'")

			// Check for debug flag with shorthand
			assert.Contains(GinkgoT(), result, "  debug:", "Expected YAML to contain debug flag")
			assert.Contains(GinkgoT(), result, "    shorthand: d", "Expected debug flag to have shorthand 'd'")

			// Check for cwd flag with shorthand
			assert.Contains(GinkgoT(), result, "  cwd:", "Expected YAML to contain cwd flag")
			assert.Contains(GinkgoT(), result, "    shorthand: C", "Expected cwd flag to have shorthand 'C'")
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
			assert.Contains(GinkgoT(), result, "enum:\n      - npm", "Expected agent enum to be correctly formatted")
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
			assert.Contains(GinkgoT(), result, "commands:\n    completion:\n      description: Generate shell completion scripts\n      completion: $carapace.shells", "Expected shell completion hints for completion command")

		})

		It("should include completion command flags", func() {
			result, err := generator.GenerateYAMLSpec(mockCmd)

			assert.NoError(GinkgoT(), err, "Expected no error generating YAML spec")

			// Check for completion command flags
			assert.Contains(GinkgoT(), result, "    completion:\n      description: Generate shell completion scripts\n      flags:\n        filename:\n          shorthand: f\n          description: Output completion script to file", "Expected completion command to have filename flag with shorthand")
			assert.Contains(GinkgoT(), result, "        with-shorthand:\n          shorthand: s\n          description: Include shorthand alias functions", "Expected completion command to have with-shorthand flag with shorthand")
		})

		It("should include integrate command and its nested commands with flags", func() {
			result, err := generator.GenerateYAMLSpec(mockCmd)

			assert.NoError(GinkgoT(), err, "Expected no error generating YAML spec")

			// Check for 'integrate' command
			assert.Contains(GinkgoT(), result, "  integrate:\n    description: Generate integration files for external tools", "Expected integrate command")

			// Check for 'integrate warp' nested command and its flags
			assert.Contains(GinkgoT(), result, "      warp:\n        description: Generate Warp terminal workflow files\n        flags:\n          output-dir:\n            shorthand: o\n            description: Output directory for workflow files\n            completion: $carapace.directories", "Expected integrate warp command with output-dir flag")

			// Check for 'integrate carapace' nested command and its flags
			assert.Contains(GinkgoT(), result, "      carapace:\n        description: Generate Carapace completion spec file\n        flags:\n          output:\n            shorthand: o\n            description: Output file for Carapace spec\n            completion: $carapace.files", "Expected integrate carapace command with output flag")
			assert.Contains(GinkgoT(), result, "          stdout:\n            description: Print Carapace spec to stdout instead of installing", "Expected integrate carapace command with stdout flag")
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
			assert.Contains(GinkgoT(), yamlPart, "name: jpd", "Expected main YAML content to contain name")
			assert.Contains(GinkgoT(), yamlPart, "commands:", "Expected YAML to contain 'commands' section")
			assert.Contains(GinkgoT(), yamlPart, "persistentFlags:", "Expected YAML to contain 'persistentFlags' section")
		})
	})

	Describe("NushellCompletionScript", func() {
		It("should return embedded nushell completion script", func() {
			result := NushellCompletionScript()

			assert.NotEmpty(GinkgoT(), result, "Expected non-empty nushell completion script")
			// The embedded file should contain nushell extern declarations
			assert.Contains(GinkgoT(), result, "extern", "Expected nushell script to contain extern declarations")
			assert.Contains(GinkgoT(), result, "jpd", "Expected nushell script to be for jpd")
		})
	})
})
