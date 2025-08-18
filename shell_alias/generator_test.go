package shell_alias

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Shell Alias Generator", func() {
	var (
		generator    Generator
		testAliasMap map[string][]string
	)

	BeforeEach(func() {
		generator = NewGenerator()
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
