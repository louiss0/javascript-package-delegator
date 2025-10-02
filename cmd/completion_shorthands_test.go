package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/mock"
	"github.com/louiss0/javascript-package-delegator/testutil"
)

func TestCompletion_WithShorthands_AppendsAliasBlock_Bash(t *testing.T) {
	mockRunner := mock.NewMockCommandRunner()
	factory := testutil.NewRootCommandFactory(mockRunner)
	factory.SetupBasicCommandRunnerExpectations()
	factory.ResetDebugExecutor()
	testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
	factory.SetupBasicDebugExecutorExpectations()

	root := factory.CreateNpmAsDefault(nil)
	testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

	out, err := executeCmd(root, "completion", "bash", "--with-shorthands")
	assert.NoError(t, err)
	assert.NotEmpty(t, out)
	// Alias block should include bash function signatures like 'function jpi'
	assert.Contains(t, out, "function jpi")
}

func TestCompletion_WithShorthands_AppendsAliasBlock_Zsh(t *testing.T) {
	mockRunner := mock.NewMockCommandRunner()
	factory := testutil.NewRootCommandFactory(mockRunner)
	factory.SetupBasicCommandRunnerExpectations()
	factory.ResetDebugExecutor()
	testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
	factory.SetupBasicDebugExecutorExpectations()

	root := factory.CreateNpmAsDefault(nil)
	testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

	out, err := executeCmd(root, "completion", "zsh", "--with-shorthands")
	assert.NoError(t, err)
	assert.NotEmpty(t, out)
	// Alias block for zsh contains 'jpi()'
	assert.Contains(t, out, "jpi()")
}

func TestCompletion_WithShorthands_AppendsAliasBlock_PowerShell(t *testing.T) {
	mockRunner := mock.NewMockCommandRunner()
	factory := testutil.NewRootCommandFactory(mockRunner)
	factory.SetupBasicCommandRunnerExpectations()
	factory.ResetDebugExecutor()
	testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
	factory.SetupBasicDebugExecutorExpectations()

	root := factory.CreateNpmAsDefault(nil)
	testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

	out, err := executeCmd(root, "completion", "powershell", "--with-shorthands")
	assert.NoError(t, err)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "function jpi {")
}

func TestCompletion_WithShorthands_AppendsAliasBlock_Nushell(t *testing.T) {
	mockRunner := mock.NewMockCommandRunner()
	factory := testutil.NewRootCommandFactory(mockRunner)
	factory.SetupBasicCommandRunnerExpectations()
	factory.ResetDebugExecutor()
	testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
	factory.SetupBasicDebugExecutorExpectations()

	root := factory.CreateNpmAsDefault(nil)
	testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

	out, err := executeCmd(root, "completion", "nushell", "--with-shorthands")
	assert.NoError(t, err)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "export extern \"jpi\"")
}
