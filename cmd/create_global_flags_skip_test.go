package cmd_test

import (
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/louiss0/javascript-package-delegator/detect"
    "github.com/louiss0/javascript-package-delegator/testutil"
    "github.com/louiss0/javascript-package-delegator/mock"
)

func TestCreate_SkipsAgentAndCwdFlags(t *testing.T) {
    mockRunner := mock.NewMockCommandRunner()
    factory := testutil.NewRootCommandFactory(mockRunner)
    factory.SetupBasicCommandRunnerExpectations()
    factory.ResetDebugExecutor()
    testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
    factory.SetupBasicDebugExecutorExpectations()

    root := factory.CreateNpmAsDefault(nil)
    testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
    // Expect command log for npm exec create-react-app -- my-app
    testutil.DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "exec", "create-react-app", "--", "my-app")

    tmp := t.TempDir()
    // Pass global flags -a and -C, create parser should skip them in its own parsing
    out, err := executeCmd(root, "create", "-a", "npm", "-C", filepath.Join(tmp, "."), "react-app", "my-app")
    assert.NoError(t, err)
    assert.Empty(t, out)

    assert.True(t, mockRunner.HasCommand("npm", "exec", "create-react-app", "--", "my-app"))
}