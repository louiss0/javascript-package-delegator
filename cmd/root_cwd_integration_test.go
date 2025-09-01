package cmd_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/louiss0/javascript-package-delegator/cmd"
	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/services"
)

var _ = Describe("Root Command --cwd Integration", func() {
	var (
		tempDir     string
		subDir      string
		mockFS      *MockFileSystem
		mockLookup  *MockPathLookup
		fakeRunner  *FakeCommandRunner
		deps        cmd.Dependencies
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "jpd-root-cwd-test")
		Expect(err).ToNot(HaveOccurred())

		subDir = filepath.Join(tempDir, "project")
		err = os.MkdirAll(subDir, 0755)
		Expect(err).ToNot(HaveOccurred())

		mockFS = &MockFileSystem{
			files: make(map[string]bool),
		}
		mockLookup = &MockPathLookup{
			paths: map[string]bool{
				"npm": true,
			},
		}
		fakeRunner = &FakeCommandRunner{}

		deps = cmd.Dependencies{
			CommandRunnerGetter: func() cmd.CommandRunner {
				return fakeRunner
			},
			DetectJSPackageManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {
				return detect.DetectJSPackageManagerBasedOnLockFile(detectedLockFile, mockLookup)
			},
			YarnCommandVersionOutputter: &MockYarnVersionOutputter{version: "1.22.19"},
			NewCommandTextUI:            newMockCommandTextUI,
			DetectLockfile: func(targetDir string) (string, error) {
				return detect.DetectLockfileIn(targetDir, mockFS)
			},
			DetectJSPackageManager: func() (string, error) {
				return detect.DetectJSPackageManager(mockLookup)
			},
			DetectVolta: func() bool {
				return detect.DetectVolta(mockLookup)
			},
			NewPackageMultiSelectUI:    newMockPackageMultiSelectUI,
			NewTaskSelectorUI:          newMockTaskSelectorUI,
			NewDependencyMultiSelectUI: newMockDependencyMultiSelectUI,
			NewDebugExecutor:           newMockDebugExecutor,
		}
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Context("when --cwd flag is provided", func() {
		It("should detect lockfile in the specified directory not current directory", func() {
			// Setup: place package-lock.json in subDir only
			packageLockPath := filepath.Join(subDir, "package-lock.json")
			mockFS.files[packageLockPath] = true

			// Create root command with dependencies
			rootCmd := cmd.NewRootCmd(deps)

			// Set the arguments to simulate --cwd flag
			rootCmd.SetArgs([]string{
				"--cwd", subDir + "/", // Add trailing slash for POSIX validation
				"agent", // Run the agent command to trigger lockfile detection
			})

			err := rootCmd.Execute()
			Expect(err).ToNot(HaveOccurred())

			// Verify that the agent flag was set to npm (from package-lock.json detection)
			agentFlag, err := rootCmd.PersistentFlags().GetString("agent")
			Expect(err).ToNot(HaveOccurred())
			Expect(agentFlag).To(Equal("npm"))
		})

		It("should not detect lockfile in current directory when --cwd points elsewhere", func() {
			// Setup: place package-lock.json only in current directory (tempDir)
			// but point --cwd to subDir which has no lockfiles
			currentDirLockFile := filepath.Join(tempDir, "package-lock.json")
			mockFS.files[currentDirLockFile] = true

			// Mock fs.Getwd to return tempDir as current directory
			mockFS.cwd = tempDir

			// Create root command with dependencies
			rootCmd := cmd.NewRootCmd(deps)

			// Set the arguments to simulate --cwd flag pointing to subDir
			rootCmd.SetArgs([]string{
				"--cwd", subDir + "/", // Add trailing slash for POSIX validation
				"agent",
			})

			err := rootCmd.Execute()
			Expect(err).ToNot(HaveOccurred())

			// Since no lockfile in subDir, it should fallback to detecting npm from PATH
			agentFlag, err := rootCmd.PersistentFlags().GetString("agent")
			Expect(err).ToNot(HaveOccurred())
			Expect(agentFlag).To(Equal("npm")) // Should still be npm from PATH detection
		})

		It("should fallback to current directory when --cwd is not provided", func() {
			// Setup: place yarn.lock in tempDir (current directory)
			yarnLockPath := filepath.Join(tempDir, "yarn.lock")
			mockFS.files[yarnLockPath] = true
			mockLookup.paths["yarn"] = true

			// Mock fs.Getwd to return tempDir as current directory
			mockFS.cwd = tempDir

			// Create root command with dependencies
			rootCmd := cmd.NewRootCmd(deps)

			// Don't set --cwd flag, should use current directory
			rootCmd.SetArgs([]string{"agent"})

			err := rootCmd.Execute()
			Expect(err).ToNot(HaveOccurred())

			// Verify that yarn was detected from the lockfile in current directory
			agentFlag, err := rootCmd.PersistentFlags().GetString("agent")
			Expect(err).ToNot(HaveOccurred())
			Expect(agentFlag).To(Equal("yarn"))
		})
	})
})

// MockFileSystem implements detect.FileSystem for testing
type MockFileSystem struct {
	files map[string]bool
	cwd   string
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.files[name] {
		return &mockFileInfo{}, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Getwd() (string, error) {
	if m.cwd != "" {
		return m.cwd, nil
	}
	return "/default/cwd", nil
}

type mockFileInfo struct{}

func (m *mockFileInfo) Name() string       { return "mockfile" }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil }

// MockPathLookup implements detect.PathLookup for testing
type MockPathLookup struct {
	paths map[string]bool
}

func (m *MockPathLookup) LookPath(file string) (string, error) {
	if m.paths[file] {
		return "/usr/bin/" + file, nil
	}
	return "", os.ErrNotExist
}

// Mock implementations for other UI components
func newMockCommandTextUI(lockfile string) cmd.CommandUITexter {
	return &MockCommandTextUI{value: "npm install -g npm"}
}

type MockCommandTextUI struct {
	value string
}

func (m *MockCommandTextUI) Run() error  { return nil }
func (m *MockCommandTextUI) Value() string { return m.value }

func newMockPackageMultiSelectUI(packages []services.PackageInfo) cmd.MultiUISelecter {
	return &MockMultiUISelecter{}
}

type MockMultiUISelecter struct{}

func (m *MockMultiUISelecter) Values() []string { return []string{} }
func (m *MockMultiUISelecter) Run() error       { return nil }

func newMockTaskSelectorUI(options []string) cmd.TaskUISelector {
	return &MockTaskUISelector{}
}

type MockTaskUISelector struct{}

func (m *MockTaskUISelector) Value() string { return "" }
func (m *MockTaskUISelector) Run() error    { return nil }

func newMockDependencyMultiSelectUI(options []string) cmd.DependencyUIMultiSelector {
	return &MockDependencyUIMultiSelector{}
}

type MockDependencyUIMultiSelector struct{}

func (m *MockDependencyUIMultiSelector) Values() []string { return []string{} }
func (m *MockDependencyUIMultiSelector) Run() error       { return nil }

func newMockDebugExecutor(debugFlag bool) cmd.DebugExecutor {
	return &MockDebugExecutor{}
}

type MockDebugExecutor struct{}

func (m *MockDebugExecutor) ExecuteIfDebugIsTrue(cb func()) {}
func (m *MockDebugExecutor) LogDebugMessageIfDebugIsTrue(msg string, keyvals ...interface{}) {}
func (m *MockDebugExecutor) LogJSCommandIfDebugIsTrue(command string, args ...string) {}

type MockYarnVersionOutputter struct {
	version string
}

func (m *MockYarnVersionOutputter) Output() (string, error) {
	return m.version, nil
}

// FakeCommandRunner for testing command execution
type FakeCommandRunner struct {
	commands [][]string
	targetDir string
}

func (f *FakeCommandRunner) Command(name string, args ...string) {
	f.commands = append(f.commands, append([]string{name}, args...))
}

func (f *FakeCommandRunner) Run() error {
	return nil
}

func (f *FakeCommandRunner) SetTargetDir(dir string) error {
	f.targetDir = dir
	return nil
}
