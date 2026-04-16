package integrations

import (
	_ "embed"
)

//go:embed assets/jpd-extern.nu
var nushellCompletionScript string

// NushellCompletionScript returns the embedded Nushell completion script content.
func NushellCompletionScript() string {
	return nushellCompletionScript
}
