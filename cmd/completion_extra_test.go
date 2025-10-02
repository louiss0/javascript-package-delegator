package cmd_test

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/louiss0/javascript-package-delegator/detect"
"github.com/louiss0/javascript-package-delegator/testutil"
	"github.com/louiss0/javascript-package-delegator/mock"
)

func TestCompletion_Help_ShowsSupportedShells(t *testing.T) {
mockRunner := mock.NewMockCommandRunner()
    factory := testutil.NewRootCommandFactory(mockRunner)
    factory.SetupBasicCommandRunnerExpectations()
    factory.ResetDebugExecutor()
    testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
    factory.SetupBasicDebugExecutorExpectations()

    root := factory.CreateNpmAsDefault(nil)
    // Common detection logs
    testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

out, err := executeCmd(root, "completion", "--help")
    assert.NoError(t, err)
    assert.Contains(t, out, "Generate shell completion scripts")
    assert.Contains(t, out, "bash")
    assert.Contains(t, out, "zsh")
    assert.Contains(t, out, "fish")
    assert.Contains(t, out, "powershell")
    assert.Contains(t, out, "nushell")
}

func TestCompletion_GeneratesPerShellScripts(t *testing.T) {
mockRunner := mock.NewMockCommandRunner()
    factory := testutil.NewRootCommandFactory(mockRunner)
    factory.SetupBasicCommandRunnerExpectations()
    factory.ResetDebugExecutor()
    testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
    factory.SetupBasicDebugExecutorExpectations()

    root := factory.CreateNpmAsDefault(nil)
    testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

    // bash
out, err := executeCmd(root, "completion", "bash")
    assert.NoError(t, err)
    assert.NotEmpty(t, out)
    // Look for a recognizable bash completion marker
assert.Contains(t, out, "__start_completion")

    // zsh
    root = factory.CreateNpmAsDefault(nil)
    testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
out, err = executeCmd(root, "completion", "zsh")
    assert.NoError(t, err)
    assert.NotEmpty(t, out)
    assert.Contains(t, out, "compdef")

    // fish
    root = factory.CreateNpmAsDefault(nil)
    testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
out, err = executeCmd(root, "completion", "fish")
    assert.NoError(t, err)
    assert.NotEmpty(t, out)
assert.Contains(t, out, "complete -c completion")

// powershell
    root = factory.CreateNpmAsDefault(nil)
    testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
    out, err = executeCmd(root, "completion", "powershell")
    assert.NoError(t, err)
    assert.NotEmpty(t, out)
    assert.Contains(t, out, "Register-ArgumentCompleter")

    // nushell
    root = factory.CreateNpmAsDefault(nil)
    testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
    out, err = executeCmd(root, "completion", "nushell")
    assert.NoError(t, err)
    assert.NotEmpty(t, out)
    assert.Contains(t, out, "export extern \"jpd\"")
}

func TestCompletion_WritesToOutputFile(t *testing.T) {
mockRunner := mock.NewMockCommandRunner()
    factory := testutil.NewRootCommandFactory(mockRunner)
    factory.SetupBasicCommandRunnerExpectations()
    factory.ResetDebugExecutor()
    testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
    factory.SetupBasicDebugExecutorExpectations()

    root := factory.CreateNpmAsDefault(nil)
    testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

    tmp := t.TempDir()
    outFile := filepath.Join(tmp, "jpd_completion.zsh")

_, err := executeCmd(root, "completion", "zsh", "--output", outFile)
    assert.NoError(t, err)

    data, readErr := os.ReadFile(outFile)
    assert.NoError(t, readErr)
    assert.NotEmpty(t, string(data))
assert.Contains(t, string(data), "compdef")
}

func TestCompletion_WritesToOutputFile_CreatesParentDirs(t *testing.T) {
    mockRunner := mock.NewMockCommandRunner()
    factory := testutil.NewRootCommandFactory(mockRunner)
    factory.SetupBasicCommandRunnerExpectations()
    factory.ResetDebugExecutor()
    testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
    factory.SetupBasicDebugExecutorExpectations()

    root := factory.CreateNpmAsDefault(nil)
    testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

    tmp := t.TempDir()
    outFile := filepath.Join(tmp, "nested", "deep", "completion.ps1")

    _, err := executeCmd(root, "completion", "powershell", "--output", outFile)
    assert.NoError(t, err)

    data, readErr := os.ReadFile(outFile)
    assert.NoError(t, readErr)
    assert.NotEmpty(t, string(data))
    assert.Contains(t, string(data), "Register-ArgumentCompleter")
}

func TestCompletion_InvalidShellErrors(t *testing.T) {
mockRunner := mock.NewMockCommandRunner()
    factory := testutil.NewRootCommandFactory(mockRunner)
    factory.SetupBasicCommandRunnerExpectations()
    factory.ResetDebugExecutor()
    testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
    factory.SetupBasicDebugExecutorExpectations()

    root := factory.CreateNpmAsDefault(nil)
    testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

_, err := executeCmd(root, "completion", "noshell")
    assert.Error(t, err)
}