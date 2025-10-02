package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/mock"
	"github.com/louiss0/javascript-package-delegator/testutil"
)

func TestCreate_URLNotSupportedForNpm_ReturnsError(t *testing.T) {
	mockRunner := mock.NewMockCommandRunner()
	factory := testutil.NewRootCommandFactory(mockRunner)
	factory.SetupBasicCommandRunnerExpectations()
	factory.ResetDebugExecutor()
	testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
	factory.SetupBasicDebugExecutorExpectations()

	root := factory.CreateNpmAsDefault(nil)
	testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

	_, err := executeCmd(root, "create", "https://example.com/template.js", "my-app")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "URLs are not supported for npm")
}
