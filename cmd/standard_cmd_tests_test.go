package cmd_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/detect"
	integrations "github.com/louiss0/javascript-package-delegator/internal"
	"github.com/louiss0/javascript-package-delegator/mock"
	"github.com/louiss0/javascript-package-delegator/testutil"
)

func TestCompletion_Help_ShowsSupportedShells(t *testing.T) {
	mockRunner := mock.NewMockCommandRunner()
	factory := testutil.NewRootCommandFactory(mockRunner)
	factory.SetupBasicCommandRunnerExpectations()
	factory.ResetDebugExecutor()
	testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
	factory.SetupBasicDebugExecutorExpectations()

	root := factory.CreateNpmAsDefault(nil)
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

func TestCompletion_RequiresExactlyOneArg_Error(t *testing.T) {
	mockRunner := mock.NewMockCommandRunner()
	factory := testutil.NewRootCommandFactory(mockRunner)
	factory.SetupBasicCommandRunnerExpectations()
	factory.ResetDebugExecutor()
	testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
	factory.SetupBasicDebugExecutorExpectations()

	root := factory.CreateNpmAsDefault(nil)
	testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

	_, err := executeCmd(root, "completion")
	assert.Error(t, err)
}

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

func TestCompletion_WithShorthands_AppendsAliasBlock_Fish(t *testing.T) {
	mockRunner := mock.NewMockCommandRunner()
	factory := testutil.NewRootCommandFactory(mockRunner)
	factory.SetupBasicCommandRunnerExpectations()
	factory.ResetDebugExecutor()
	testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
	factory.SetupBasicDebugExecutorExpectations()

	root := factory.CreateNpmAsDefault(nil)
	testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

	out, err := executeCmd(root, "completion", "fish", "--with-shorthands")
	assert.NoError(t, err)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "function jpi")
}

func TestCreate_SkipsAgentAndCwdFlags(t *testing.T) {
	mockRunner := mock.NewMockCommandRunner()
	factory := testutil.NewRootCommandFactory(mockRunner)
	factory.SetupBasicCommandRunnerExpectations()
	factory.ResetDebugExecutor()
	testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
	factory.SetupBasicDebugExecutorExpectations()

	root := factory.CreateNpmAsDefault(nil)
	testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
	testutil.DebugExecutorExpectationManager.ExpectJSCommandLog("npm", "exec", "create-react-app", "--", "my-app")

	tmp := t.TempDir()
	out, err := executeCmd(root, "create", "-a", "npm", "-C", filepath.Join(tmp, "."), "react-app", "my-app")
	assert.NoError(t, err)
	assert.Empty(t, out)
	assert.True(t, mockRunner.HasCommand("npm", "exec", "create-react-app", "--", "my-app"))
}

func TestCreate_InvalidSizeValue_ReturnsError(t *testing.T) {
	mockRunner := mock.NewMockCommandRunner()
	factory := testutil.NewRootCommandFactory(mockRunner)
	factory.SetupBasicCommandRunnerExpectations()
	factory.ResetDebugExecutor()
	testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
	factory.SetupBasicDebugExecutorExpectations()

	root := factory.CreateNpmAsDefault(nil)
	testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

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

	_, err := executeCmd(root, "create", "--search", "--size")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--size requires a value")
}

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

	tmp := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", tmp)
	} else {
		t.Setenv("XDG_DATA_HOME", tmp)
	}

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
	outPath := filepath.Join(tmp, "nope", "nested", "carapace.yaml")

	_, err := executeCmd(root, "integrate", "carapace", "--output", outPath)
	assert.Error(t, err)
}
