package cmd

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidInstallCommandStringRegex(t *testing.T) {
	regex := regexp.MustCompile(VALID_INSTALL_COMMAND_STRING_RE)

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
			"npm install",     // only two words
			"install yarn",    // only two words
			"deno",           // single word
			"nix profile",    // only two words
			"yarn",           // single word
			"pnpm",           // single word
			"brew install",   // only two words
			"sudo apt-get",   // only two words
			"",               // empty string
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
