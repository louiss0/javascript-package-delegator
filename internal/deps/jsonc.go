// Package deps provides functionality for dependency management and detection
// across different JavaScript package managers and runtime environments.
package deps

import (
	"regexp"
	"strings"
)

// NormalizeJSONCToJSON removes comments and trailing commas from JSONC content
// to make it valid JSON for parsing.
// This is a simplified implementation that handles most common JSONC features.
func NormalizeJSONCToJSON(content []byte) []byte {
	text := string(content)

	// Remove single-line comments (//)
	singleLineCommentRegex := regexp.MustCompile(`//.*$`)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = singleLineCommentRegex.ReplaceAllString(line, "")
	}
	text = strings.Join(lines, "\n")

	// Remove multi-line comments (/* ... */)
	multiLineCommentRegex := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	text = multiLineCommentRegex.ReplaceAllString(text, "")

	// Remove trailing commas before closing brackets/braces
	// This handles cases like: "key": "value", } or "key": "value", ]
	trailingCommaRegex := regexp.MustCompile(`,(\s*[}\]])`)
	text = trailingCommaRegex.ReplaceAllString(text, "$1")

	return []byte(text)
}
