package cmd_test

import (
	"testing"

	"github.com/louiss0/javascript-package-delegator/cmd"
)

func TestNewDlxCmd_Aliases(t *testing.T) {
	// Arrange: Create the dlx command
	dlxCmd := cmd.NewDlxCmd()

	// Act: Get the aliases
	actualAliases := dlxCmd.Aliases

	// Assert: Should contain "x"
	found := false
	for _, alias := range actualAliases {
		if alias == "x" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("NewDlxCmd().Aliases should contain 'x', but got %v", actualAliases)
	}
}
