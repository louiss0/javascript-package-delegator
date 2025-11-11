// Package deps provides functionality for dependency management and detection
// across different JavaScript package managers and runtime environments.
package deps

import "strings"

// NormalizeJSONCToJSON removes comments and trailing commas from JSONC content
// to make it valid JSON for parsing.
// This implementation is careful to leave comment-like sequences that appear
// inside string literals untouched.
func NormalizeJSONCToJSON(content []byte) []byte {
	text := string(content)

	// First pass: strip comments while respecting string literals.
	var b strings.Builder
	b.Grow(len(text))

	inString := false
	escape := false
	inSingleLineComment := false
	inMultiLineComment := false

	for i := 0; i < len(text); i++ {
		ch := text[i]

		if inSingleLineComment {
			if ch == '\n' {
				inSingleLineComment = false
				b.WriteByte(ch)
			}
			continue
		}

		if inMultiLineComment {
			if ch == '*' && i+1 < len(text) && text[i+1] == '/' {
				inMultiLineComment = false
				i++ // Skip the closing '/'
			}
			continue
		}

		if inString {
			b.WriteByte(ch)
			if escape {
				escape = false
				continue
			}
			if ch == '\\' {
				escape = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		if ch == '"' {
			inString = true
			b.WriteByte(ch)
			continue
		}

		if ch == '/' && i+1 < len(text) {
			next := text[i+1]
			if next == '/' {
				inSingleLineComment = true
				i++
				continue
			}
			if next == '*' {
				inMultiLineComment = true
				i++
				continue
			}
		}

		b.WriteByte(ch)
	}

	text = b.String()

	// Second pass: remove trailing commas before closing braces/brackets,
	// again ensuring we do not modify strings.
	var out strings.Builder
	out.Grow(len(text))

	inString = false
	escape = false

	for i := 0; i < len(text); i++ {
		ch := text[i]

		if inString {
			out.WriteByte(ch)
			if escape {
				escape = false
				continue
			}
			if ch == '\\' {
				escape = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		if ch == '"' {
			inString = true
			out.WriteByte(ch)
			continue
		}

		if ch == ',' {
			j := i + 1
			for j < len(text) {
				if text[j] == ' ' || text[j] == '\t' || text[j] == '\r' || text[j] == '\n' {
					j++
					continue
				}
				break
			}
			if j < len(text) && (text[j] == '}' || text[j] == ']') {
				// Skip the comma and continue without writing it.
				continue
			}
		}

		out.WriteByte(ch)
	}

	return []byte(out.String())
}
