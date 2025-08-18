package cmd_test

import (
	"reflect"
	"testing"

	"github.com/louiss0/javascript-package-delegator/cmd"
)

func TestNewExecCmd_Aliases(t *testing.T) {
	// Arrange: Create the exec command
	execCmd := cmd.NewExecCmd()

	// Act: Get the aliases
	actualAliases := execCmd.Aliases

	// Assert: Should only contain "e", NOT "x"
	expectedAliases := []string{"e"}

	if !reflect.DeepEqual(actualAliases, expectedAliases) {
		t.Errorf("NewExecCmd().Aliases = %v, want %v", actualAliases, expectedAliases)
	}

	// Also explicitly assert that "x" is NOT present
	for _, alias := range actualAliases {
		if alias == "x" {
			t.Errorf("NewExecCmd().Aliases should NOT contain 'x', but found it in %v", actualAliases)
		}
	}
}
