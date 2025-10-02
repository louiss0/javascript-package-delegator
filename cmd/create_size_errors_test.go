package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/mock"
	"github.com/louiss0/javascript-package-delegator/testutil"
)

func TestCreate_InvalidSizeValue_ReturnsError(t *testing.T) {
	mockRunner := mock.NewMockCommandRunner()
	factory := testutil.NewRootCommandFactory(mockRunner)
	factory.SetupBasicCommandRunnerExpectations()
	factory.ResetDebugExecutor()
	testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
	factory.SetupBasicDebugExecutorExpectations()

	root := factory.CreateNpmAsDefault(nil)
	testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

	// --search enables search mode; --size with non-integer should error
	_, err := executeCmd(root, "create", "--search", "--size", "abc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid size value: abc")
}

func TestCreate_SizeMissingValue_ReturnsError(t *testing.T) {
	mockRunner := mock.NewMockCommandRunner()
	factory := testutil.NewRootCommandFactory(mockRunner)
	factory.SetupBasicCommandRunnerExpectations()
	factory.ResetDebugExecutor()
	testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
	factory.SetupBasicDebugExecutorExpectations()

	root := factory.CreateNpmAsDefault(nil)
	testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

	// --size without following value should error
	_, err := executeCmd(root, "create", "--search", "--size")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--size requires a value")
}
