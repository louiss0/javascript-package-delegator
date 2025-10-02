package cmd_test

import (
	"os"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/detect"
	integrations "github.com/louiss0/javascript-package-delegator/internal"
	"github.com/louiss0/javascript-package-delegator/mock"
	"github.com/louiss0/javascript-package-delegator/testutil"
)

var _ = Describe("Completion (Ginkgo conversion)", func() {
	It("shows supported shells in help", func() {
		mockRunner := mock.NewMockCommandRunner()
		factory := testutil.NewRootCommandFactory(mockRunner)
		factory.SetupBasicCommandRunnerExpectations()
		factory.ResetDebugExecutor()
		testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
		factory.SetupBasicDebugExecutorExpectations()

		root := factory.CreateNpmAsDefault(nil)
		testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

		out, err := executeCmd(root, "completion", "--help")
		assert.NoError(GinkgoT(), err)
		assert.Contains(GinkgoT(), out, "Generate shell completion scripts")
		assert.Contains(GinkgoT(), out, "bash")
		assert.Contains(GinkgoT(), out, "zsh")
		assert.Contains(GinkgoT(), out, "fish")
		assert.Contains(GinkgoT(), out, "powershell")
		assert.Contains(GinkgoT(), out, "nushell")
	})

	It("generates per-shell scripts", func() {
		mockRunner := mock.NewMockCommandRunner()
		factory := testutil.NewRootCommandFactory(mockRunner)
		factory.SetupBasicCommandRunnerExpectations()
		factory.ResetDebugExecutor()
		testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
		factory.SetupBasicDebugExecutorExpectations()

		root := factory.CreateNpmAsDefault(nil)
		testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
		out, err := executeCmd(root, "completion", "bash")
		assert.NoError(GinkgoT(), err)
		assert.NotEmpty(GinkgoT(), out)
		assert.Contains(GinkgoT(), out, "__start_completion")

		root = factory.CreateNpmAsDefault(nil)
		testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
		out, err = executeCmd(root, "completion", "zsh")
		assert.NoError(GinkgoT(), err)
		assert.NotEmpty(GinkgoT(), out)
		assert.Contains(GinkgoT(), out, "compdef")

		root = factory.CreateNpmAsDefault(nil)
		testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
		out, err = executeCmd(root, "completion", "fish")
		assert.NoError(GinkgoT(), err)
		assert.NotEmpty(GinkgoT(), out)
		assert.Contains(GinkgoT(), out, "complete -c completion")

		root = factory.CreateNpmAsDefault(nil)
		testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
		out, err = executeCmd(root, "completion", "powershell")
		assert.NoError(GinkgoT(), err)
		assert.NotEmpty(GinkgoT(), out)
		assert.Contains(GinkgoT(), out, "Register-ArgumentCompleter")

		root = factory.CreateNpmAsDefault(nil)
		testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
		out, err = executeCmd(root, "completion", "nushell")
		assert.NoError(GinkgoT(), err)
		assert.NotEmpty(GinkgoT(), out)
		assert.Contains(GinkgoT(), out, "export extern \"jpd\"")
	})

	It("writes to output file", func() {
		mockRunner := mock.NewMockCommandRunner()
		factory := testutil.NewRootCommandFactory(mockRunner)
		factory.SetupBasicCommandRunnerExpectations()
		factory.ResetDebugExecutor()
		testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
		factory.SetupBasicDebugExecutorExpectations()

		root := factory.CreateNpmAsDefault(nil)
		testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

		tmp := GinkgoT().TempDir()
		outFile := filepath.Join(tmp, "jpd_completion.zsh")
		_, err := executeCmd(root, "completion", "zsh", "--output", outFile)
		assert.NoError(GinkgoT(), err)

		data, readErr := os.ReadFile(outFile)
		assert.NoError(GinkgoT(), readErr)
		assert.NotEmpty(GinkgoT(), string(data))
		assert.Contains(GinkgoT(), string(data), "compdef")
	})

	It("writes to nested output file path creating parents", func() {
		mockRunner := mock.NewMockCommandRunner()
		factory := testutil.NewRootCommandFactory(mockRunner)
		factory.SetupBasicCommandRunnerExpectations()
		factory.ResetDebugExecutor()
		testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
		factory.SetupBasicDebugExecutorExpectations()

		root := factory.CreateNpmAsDefault(nil)
		testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

		tmp := GinkgoT().TempDir()
		outFile := filepath.Join(tmp, "nested", "deep", "completion.ps1")
		_, err := executeCmd(root, "completion", "powershell", "--output", outFile)
		assert.NoError(GinkgoT(), err)

		data, readErr := os.ReadFile(outFile)
		assert.NoError(GinkgoT(), readErr)
		assert.NotEmpty(GinkgoT(), string(data))
		assert.Contains(GinkgoT(), string(data), "Register-ArgumentCompleter")
	})

	It("errors for invalid shell and missing arg", func() {
		mockRunner := mock.NewMockCommandRunner()
		factory := testutil.NewRootCommandFactory(mockRunner)
		factory.SetupBasicCommandRunnerExpectations()
		factory.ResetDebugExecutor()
		testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
		factory.SetupBasicDebugExecutorExpectations()

		root := factory.CreateNpmAsDefault(nil)
		testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
		_, err := executeCmd(root, "completion", "noshell")
		assert.Error(GinkgoT(), err)

		root = factory.CreateNpmAsDefault(nil)
		testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
		_, err = executeCmd(root, "completion")
		assert.Error(GinkgoT(), err)
	})

	Describe("with shorthands", func() {
		It("appends alias block (bash)", func() {
			mockRunner := mock.NewMockCommandRunner()
			factory := testutil.NewRootCommandFactory(mockRunner)
			factory.SetupBasicCommandRunnerExpectations()
			factory.ResetDebugExecutor()
			testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
			factory.SetupBasicDebugExecutorExpectations()

			root := factory.CreateNpmAsDefault(nil)
			testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
			out, err := executeCmd(root, "completion", "bash", "--with-shorthands")
			assert.NoError(GinkgoT(), err)
			assert.NotEmpty(GinkgoT(), out)
			assert.Contains(GinkgoT(), out, "function jpi")
		})

		It("appends alias block (zsh)", func() {
			mockRunner := mock.NewMockCommandRunner()
			factory := testutil.NewRootCommandFactory(mockRunner)
			factory.SetupBasicCommandRunnerExpectations()
			factory.ResetDebugExecutor()
			testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
			factory.SetupBasicDebugExecutorExpectations()

			root := factory.CreateNpmAsDefault(nil)
			testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
			out, err := executeCmd(root, "completion", "zsh", "--with-shorthands")
			assert.NoError(GinkgoT(), err)
			assert.NotEmpty(GinkgoT(), out)
			assert.Contains(GinkgoT(), out, "jpi()")
		})

		It("appends alias block (powershell)", func() {
			mockRunner := mock.NewMockCommandRunner()
			factory := testutil.NewRootCommandFactory(mockRunner)
			factory.SetupBasicCommandRunnerExpectations()
			factory.ResetDebugExecutor()
			testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
			factory.SetupBasicDebugExecutorExpectations()

			root := factory.CreateNpmAsDefault(nil)
			testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
			out, err := executeCmd(root, "completion", "powershell", "--with-shorthands")
			assert.NoError(GinkgoT(), err)
			assert.NotEmpty(GinkgoT(), out)
			assert.Contains(GinkgoT(), out, "function jpi {")
		})

		It("appends alias block (nushell)", func() {
			mockRunner := mock.NewMockCommandRunner()
			factory := testutil.NewRootCommandFactory(mockRunner)
			factory.SetupBasicCommandRunnerExpectations()
			factory.ResetDebugExecutor()
			testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
			factory.SetupBasicDebugExecutorExpectations()

			root := factory.CreateNpmAsDefault(nil)
			testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
			out, err := executeCmd(root, "completion", "nushell", "--with-shorthands")
			assert.NoError(GinkgoT(), err)
			assert.NotEmpty(GinkgoT(), out)
			assert.Contains(GinkgoT(), out, "export extern \"jpi\"")
		})

		It("appends alias block (fish)", func() {
			mockRunner := mock.NewMockCommandRunner()
			factory := testutil.NewRootCommandFactory(mockRunner)
			factory.SetupBasicCommandRunnerExpectations()
			factory.ResetDebugExecutor()
			testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
			factory.SetupBasicDebugExecutorExpectations()

			root := factory.CreateNpmAsDefault(nil)
			testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)
			out, err := executeCmd(root, "completion", "fish", "--with-shorthands")
			assert.NoError(GinkgoT(), err)
			assert.NotEmpty(GinkgoT(), out)
			assert.Contains(GinkgoT(), out, "function jpi")
		})
	})
})

var _ = Describe("Integrate carapace (Ginkgo conversion)", func() {
	It("writes to output file", func() {
		mockRunner := mock.NewMockCommandRunner()
		factory := testutil.NewRootCommandFactory(mockRunner)
		factory.SetupBasicCommandRunnerExpectations()
		factory.ResetDebugExecutor()
		testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
		factory.SetupBasicDebugExecutorExpectations()

		root := factory.CreateNpmAsDefault(nil)
		testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

		tmp := GinkgoT().TempDir()
		outPath := filepath.Join(tmp, "jpd_carapace.yaml")
		_, err := executeCmd(root, "integrate", "carapace", "--output", outPath)
		assert.NoError(GinkgoT(), err)

		content, readErr := os.ReadFile(outPath)
		assert.NoError(GinkgoT(), readErr)
		s := string(content)
		assert.NotEmpty(GinkgoT(), s)
		assert.Contains(GinkgoT(), s, "Name: javascript-package-delegator")
	})

	It("prints to stdout", func() {
		mockRunner := mock.NewMockCommandRunner()
		factory := testutil.NewRootCommandFactory(mockRunner)
		factory.SetupBasicCommandRunnerExpectations()
		factory.ResetDebugExecutor()
		testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
		factory.SetupBasicDebugExecutorExpectations()

		root := factory.CreateNpmAsDefault(nil)
		testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

		out, err := executeCmd(root, "integrate", "carapace", "--stdout")
		assert.NoError(GinkgoT(), err)
		assert.NotEmpty(GinkgoT(), out)
		assert.Contains(GinkgoT(), out, "Name: javascript-package-delegator")
	})

	It("default install respects XDG/APPDATA", func() {
		mockRunner := mock.NewMockCommandRunner()
		factory := testutil.NewRootCommandFactory(mockRunner)
		factory.SetupBasicCommandRunnerExpectations()
		factory.ResetDebugExecutor()
		testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
		factory.SetupBasicDebugExecutorExpectations()

		root := factory.CreateNpmAsDefault(nil)
		testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

		tmp := GinkgoT().TempDir()
		if runtime.GOOS == "windows" {
			_ = os.Setenv("APPDATA", tmp)
		} else {
			_ = os.Setenv("XDG_DATA_HOME", tmp)
		}

		specPath, err := integrations.DefaultCarapaceSpecPath()
		assert.NoError(GinkgoT(), err)

		_, err = executeCmd(root, "integrate", "carapace")
		assert.NoError(GinkgoT(), err)

		data, readErr := os.ReadFile(specPath)
		assert.NoError(GinkgoT(), readErr)
		assert.NotEmpty(GinkgoT(), data)
	})

	It("errors when parent directories are missing for output path", func() {
		mockRunner := mock.NewMockCommandRunner()
		factory := testutil.NewRootCommandFactory(mockRunner)
		factory.SetupBasicCommandRunnerExpectations()
		factory.ResetDebugExecutor()
		testutil.DebugExecutorExpectationManager.DebugExecutor = factory.DebugExecutor()
		factory.SetupBasicDebugExecutorExpectations()

		root := factory.CreateNpmAsDefault(nil)
		testutil.DebugExecutorExpectationManager.ExpectCommonPMDetectionFlow(detect.NPM, detect.PACKAGE_LOCK_JSON)

		tmp := GinkgoT().TempDir()
		outPath := filepath.Join(tmp, "nope", "nested", "carapace.yaml")
		_, err := executeCmd(root, "integrate", "carapace", "--output", outPath)
		assert.Error(GinkgoT(), err)
	})
})
