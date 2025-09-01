package cmd

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/services"
)

func TestRunCWDIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Run CWD Integration Suite")
}

// Test doubles for mocking dependencies
type fakeRunner struct {
	name string
	args []string
	ran  bool
	dir  string
}

func (f *fakeRunner) Command(name string, args ...string) {
	f.name = name
	f.args = append([]string{}, args...)
}

func (f *fakeRunner) Run() error {
	f.ran = true
	return nil
}

func (f *fakeRunner) SetTargetDir(dir string) error {
	f.dir = dir
	return nil
}

type selectingTaskSelector struct {
	gotOptions []string
	toSelect   string
	selected   string
}

func (s *selectingTaskSelector) Value() string {
	return s.selected
}

func (s *selectingTaskSelector) Run() error {
	s.selected = s.toSelect
	return nil
}

type fakeDependencyMultiSelector struct {
	values []string
}

func (f *fakeDependencyMultiSelector) Values() []string {
	return f.values
}

func (f *fakeDependencyMultiSelector) Run() error {
	return nil
}

func newStubDependencies(runner *fakeRunner, taskSelector *selectingTaskSelector) Dependencies {
	return Dependencies{
		CommandRunnerGetter: func() CommandRunner {
			return runner
		},
		DetectJSPackageManagerBasedOnLockFile: func(detectedLockFile string) (string, error) {
			// Not used when --agent is provided
			return "", nil
		},
		YarnCommandVersionOutputter: detect.NewRealYarnCommandVersionRunner(),
		NewCommandTextUI: func(lockfile string) CommandUITexter {
			// Not used in these tests
			return nil
		},
		DetectLockfile: func() (lockfile string, err error) {
			// Not used when --agent is provided
			return "", detect.ErrNoPackageManager
		},
		DetectJSPackageManager: func() (string, error) {
			// Not used when --agent is provided
			return "", detect.ErrNoPackageManager
		},
		DetectVolta: func() bool {
			return false
		},
		NewPackageMultiSelectUI: func([]services.PackageInfo) MultiUISelecter {
			// Not used in these tests
			return nil
		},
		NewTaskSelectorUI: func(options []string) TaskUISelector {
			taskSelector.gotOptions = append([]string{}, options...)
			return taskSelector
		},
		NewDependencyMultiSelectUI: func(options []string) DependencyUIMultiSelector {
			return &fakeDependencyMultiSelector{values: options}
		},
		NewDebugExecutor: newDebugExecutor,
	}
}

var _ = Describe("jpd run --cwd integration", func() {
	var originalWD string

	BeforeEach(func() {
		wd, _ := os.Getwd()
		originalWD = wd
	})

	AfterEach(func() {
		_ = os.Chdir(originalWD)
	})

	Describe("npm/package.json path", func() {
		It("uses --cwd directory to discover scripts when no script is provided (interactive selection)", func() {
			// Arrange: Create target directory with specific package.json
			targetDir := GinkgoT().TempDir()

			err := os.WriteFile(filepath.Join(targetDir, "package.json"), []byte(`{"scripts":{"from-target":"echo hi"}}`), 0644)
			assert.NoError(GinkgoT(), err)

			// Create a different working directory to prove --cwd is respected
			workingDir := GinkgoT().TempDir()
			err = os.WriteFile(filepath.Join(workingDir, "package.json"), []byte(`{"scripts":{"from-other":"echo wrong"}}`), 0644)
			assert.NoError(GinkgoT(), err)

			err = os.Chdir(workingDir)
			assert.NoError(GinkgoT(), err)

			// Set up test doubles
			runner := &fakeRunner{}
			taskSelector := &selectingTaskSelector{toSelect: "from-target"}
			deps := newStubDependencies(runner, taskSelector)

			// Create root command with test dependencies
			rootCmd := NewRootCmd(deps)

			// Act: Execute command with --cwd flag
			args := []string{"--agent", "npm", "--cwd", targetDir + "/", "run"}
			rootCmd.SetArgs(args)
			err = rootCmd.Execute()

			// Assert
			assert.NoError(GinkgoT(), err)
			assert.Contains(GinkgoT(), taskSelector.gotOptions, "from-target")
			assert.NotContains(GinkgoT(), taskSelector.gotOptions, "from-other")
			assert.Equal(GinkgoT(), "npm", runner.name)
			assert.Equal(GinkgoT(), []string{"run", "from-target"}, runner.args)
			assert.True(GinkgoT(), runner.ran)
			assert.Equal(GinkgoT(), targetDir+"/", runner.dir)
		})

		It("honors --cwd for --if-present script lookup", func() {
			// Arrange: Create target directory without the script
			targetDir := GinkgoT().TempDir()
			err := os.WriteFile(filepath.Join(targetDir, "package.json"), []byte(`{"scripts":{"other":"echo other"}}`), 0644)
			assert.NoError(GinkgoT(), err)

			// Create working directory with the script that should NOT be found
			workingDir := GinkgoT().TempDir()
			err = os.WriteFile(filepath.Join(workingDir, "package.json"), []byte(`{"scripts":{"foo":"echo wrong"}}`), 0644)
			assert.NoError(GinkgoT(), err)

			err = os.Chdir(workingDir)
			assert.NoError(GinkgoT(), err)

			// Set up test doubles
			runner := &fakeRunner{}
			taskSelector := &selectingTaskSelector{}
			deps := newStubDependencies(runner, taskSelector)

			// Create root command with test dependencies
			rootCmd := NewRootCmd(deps)

			// Act: Execute command with --if-present and --cwd
			args := []string{"--agent", "npm", "--cwd", targetDir + "/", "run", "--if-present", "foo"}
			rootCmd.SetArgs(args)
			err = rootCmd.Execute()

			// Assert: Command should succeed but runner should NOT be called
			assert.NoError(GinkgoT(), err)
			assert.False(GinkgoT(), runner.ran, "Runner should not be called when script doesn't exist in target directory")
			assert.Equal(GinkgoT(), targetDir+"/", runner.dir)
		})
	})

	Describe("deno/deno.json path", func() {
		It("uses --cwd directory to discover tasks when no task is provided (interactive selection)", func() {
			// Arrange: Create target directory with specific deno.json
			targetDir := GinkgoT().TempDir()

			err := os.WriteFile(filepath.Join(targetDir, "deno.json"), []byte(`{"tasks":{"from-target":"deno run mod.ts"}}`), 0644)
			assert.NoError(GinkgoT(), err)

			// Create a different working directory to prove --cwd is respected
			workingDir := GinkgoT().TempDir()
			err = os.WriteFile(filepath.Join(workingDir, "deno.json"), []byte(`{"tasks":{"from-other":"deno run wrong.ts"}}`), 0644)
			assert.NoError(GinkgoT(), err)

			err = os.Chdir(workingDir)
			assert.NoError(GinkgoT(), err)

			// Set up test doubles
			runner := &fakeRunner{}
			taskSelector := &selectingTaskSelector{toSelect: "from-target"}
			deps := newStubDependencies(runner, taskSelector)

			// Create root command with test dependencies
			rootCmd := NewRootCmd(deps)

			// Act: Execute command with --cwd flag
			args := []string{"--agent", "deno", "--cwd", targetDir + "/", "run"}
			rootCmd.SetArgs(args)
			err = rootCmd.Execute()

			// Assert
			assert.NoError(GinkgoT(), err)
			assert.Contains(GinkgoT(), taskSelector.gotOptions, "from-target")
			assert.NotContains(GinkgoT(), taskSelector.gotOptions, "from-other")
			assert.Equal(GinkgoT(), "deno", runner.name)
			assert.Equal(GinkgoT(), []string{"task", "from-target"}, runner.args)
			assert.True(GinkgoT(), runner.ran)
			assert.Equal(GinkgoT(), targetDir+"/", runner.dir)
		})
	})
})
