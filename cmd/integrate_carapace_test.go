package cmd_test

import (
    "os"
    "path/filepath"
    "runtime"
    "testing"

    "github.com/stretchr/testify/assert"

    integrations "github.com/louiss0/javascript-package-delegator/internal"
    "github.com/louiss0/javascript-package-delegator/detect"
"github.com/louiss0/javascript-package-delegator/testutil"
	"github.com/louiss0/javascript-package-delegator/mock"
)

func TestIntegrateCarapace_OutputFile(t *testing.T) {
mockRunner := mock.NewMockCommandRunner()
    factory := testutil.NewRootCommandFactory(mockRunner)
    factory.SetupBasicCommandRunnerExpectations()
    factory.ResetDebugExecutor()
    testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
    factory.SetupBasicDebugExecutorExpectations()

    root := factory.CreateNpmAsDefault(nil)
    testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

    tmp := t.TempDir()
    outPath := filepath.Join(tmp, "jpd_carapace.yaml")

_, err := executeCmd(root, "integrate", "carapace", "--output", outPath)
    assert.NoError(t, err)

    content, readErr := os.ReadFile(outPath)
    assert.NoError(t, readErr)
    s := string(content)
    assert.NotEmpty(t, s)
    assert.Contains(t, s, "Name: javascript-package-delegator")
}

func TestIntegrateCarapace_Stdout(t *testing.T) {
mockRunner := mock.NewMockCommandRunner()
    factory := testutil.NewRootCommandFactory(mockRunner)
    factory.SetupBasicCommandRunnerExpectations()
    factory.ResetDebugExecutor()
    testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
    factory.SetupBasicDebugExecutorExpectations()

    root := factory.CreateNpmAsDefault(nil)
    testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

out, err := executeCmd(root, "integrate", "carapace", "--stdout")
    assert.NoError(t, err)
    assert.NotEmpty(t, out)
    assert.Contains(t, out, "Name: javascript-package-delegator")
}

func TestIntegrateCarapace_DefaultInstall_RespectsTempDataHome(t *testing.T) {
mockRunner := mock.NewMockCommandRunner()
    factory := testutil.NewRootCommandFactory(mockRunner)
    factory.SetupBasicCommandRunnerExpectations()
    factory.ResetDebugExecutor()
    testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
    factory.SetupBasicDebugExecutorExpectations()

    root := factory.CreateNpmAsDefault(nil)
    testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

    // Ensure we write into a temp location
    tmp := t.TempDir()
    if runtime.GOOS == "windows" {
        t.Setenv("APPDATA", tmp)
    } else {
        t.Setenv("XDG_DATA_HOME", tmp)
    }

    // Determine expected spec path using the same resolver as the app
    specPath, err := integrations.DefaultCarapaceSpecPath()
    assert.NoError(t, err)

_, err = executeCmd(root, "integrate", "carapace")
    assert.NoError(t, err)

    data, readErr := os.ReadFile(specPath)
    assert.NoError(t, readErr)
    assert.NotEmpty(t, data)
}

func TestIntegrateCarapace_OutputFile_ParentMissing_ShouldFail(t *testing.T) {
mockRunner := mock.NewMockCommandRunner()
    factory := testutil.NewRootCommandFactory(mockRunner)
    factory.SetupBasicCommandRunnerExpectations()
    factory.ResetDebugExecutor()
    testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
    factory.SetupBasicDebugExecutorExpectations()

    root := factory.CreateNpmAsDefault(nil)
    testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

    tmp := t.TempDir()
    // Create a nested non-existing path
    outPath := filepath.Join(tmp, "nope", "nested", "carapace.yaml")

_, err := executeCmd(root, "integrate", "carapace", "--output", outPath)
    assert.Error(t, err)
}