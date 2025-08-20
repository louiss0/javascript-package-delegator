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
		})

		It("should include all JPD subcommands", func() {
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
			}

			for _, expected := range expectedCommands {
				assert.Contains(GinkgoT(), result, expected, "Expected YAML to contain command: %s", expected)
			}
		})

		It("should include persistent flags with expected short forms", func() {
			result, err := generator.GenerateYAMLSpec(mockCmd)

			assert.NoError(GinkgoT(), err, "Expected no error generating YAML spec")

			// Check for agent flag with shorthand
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
		})

		It("should include completion hints for relevant commands", func() {
			result, err := generator.GenerateYAMLSpec(mockCmd)

			assert.NoError(GinkgoT(), err, "Expected no error generating YAML spec")

			// Check for package completion hints
			assert.Contains(GinkgoT(), result, "$carapace.packages.npm", "Expected package completion hints")
			// Check for scripts completion hints
			assert.Contains(GinkgoT(), result, "$carapace.scripts.npm", "Expected scripts completion hints")
			// Check for directory completion hints
			assert.Contains(GinkgoT(), result, "$carapace.directories", "Expected directory completion hints")
		})

		It("should include completion command flags", func() {
			result, err := generator.GenerateYAMLSpec(mockCmd)

			assert.NoError(GinkgoT(), err, "Expected no error generating YAML spec")

			// Check for completion command flags
			assert.Contains(GinkgoT(), result, "filename:", "Expected completion command to have filename flag")
			assert.Contains(GinkgoT(), result, "with-shorthand:", "Expected completion command to have with-shorthand flag")
		})

		It("should generate valid YAML structure", func() {
			result, err := generator.GenerateYAMLSpec(mockCmd)

			assert.NoError(GinkgoT(), err, "Expected no error generating YAML spec")

			// Remove header comments for YAML parsing
			yamlContent := strings.Split(result, "---\n")
			if len(yamlContent) >= 2 {
				yamlPart := yamlContent[1]
				// Should not panic or error when parsing as YAML
				assert.NotEmpty(GinkgoT(), yamlPart, "Expected non-empty YAML content after header")
				assert.Contains(GinkgoT(), yamlPart, "name: jpd", "Expected main YAML content to contain name")
			}
		})
	})

	Describe("GetNushellCompletionScript", func() {
		It("should return embedded nushell completion script", func() {
			result := GetNushellCompletionScript()

			assert.NotEmpty(GinkgoT(), result, "Expected non-empty nushell completion script")
			// The embedded file should contain nushell extern declarations
			assert.Contains(GinkgoT(), result, "extern", "Expected nushell script to contain extern declarations")
		})
	})
})
