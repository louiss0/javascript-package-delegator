package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/louiss0/javascript-package-delegator/services"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert" // Import testify/assert
)

// FakeCommandRunner is a test double that doesn't execute real commands
type FakeCommandRunner struct {
	targetDir  string
	commandSet bool
}

func NewFakeCommandRunner() CommandRunner {
	return &FakeCommandRunner{}
}

func (f *FakeCommandRunner) Command(name string, args ...string) {
	f.commandSet = true
}

func (f *FakeCommandRunner) SetTargetDir(dir string) error {
	fileInfo, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("target directory %s is not a directory", dir)
	}
	f.targetDir = dir
	return nil
}

func (f *FakeCommandRunner) Run() error {
	if !f.commandSet {
		return fmt.Errorf("no command set to run")
	}
	return nil
}

type dummyYarnVersionRunner struct{}

func (d dummyYarnVersionRunner) Output() (string, error) {
	return "1.22.0", nil
}

type noopCommandTextUI struct{}

func (n noopCommandTextUI) Value() string { return "" }
func (n noopCommandTextUI) Run() error    { return nil }

type noopMultiSelect struct{}

func (n noopMultiSelect) Values() []string { return nil }
func (n noopMultiSelect) Run() error       { return nil }

type noopTaskSelector struct{}

func (n noopTaskSelector) Value() string { return "" }
func (n noopTaskSelector) Run() error    { return nil }

var _ = Describe("root -C flag regression", func() {
	It("does not panic when executing root with only the -C flag before any Command() call", func() {
		a := assert.New(GinkgoT()) // Create assert instance

		// Build a NewRootCmd with stubbed dependencies per spec
		root := NewRootCmd(Dependencies{
			CommandRunnerGetter:                   func() CommandRunner { return NewFakeCommandRunner() },
			DetectLockfile:                        func() (string, error) { return "", errors.New("no lockfile") },
			DetectJSPacakgeManager:                func() (string, error) { return "npm", nil },
			DetectJSPacakgeManagerBasedOnLockFile: func(string) (string, error) { return "npm", nil },
			YarnCommandVersionOutputter:           dummyYarnVersionRunner{},
			NewCommandTextUI:                      func(string) CommandUITexter { return noopCommandTextUI{} },
			DetectVolta:                           func() bool { return false },
			NewPackageMultiSelectUI:               func(_ []services.PackageInfo) MultiUISelecter { return noopMultiSelect{} },
			NewTaskSelectorUI:                     func(_ []string) TaskUISelector { return noopTaskSelector{} },
			NewDependencyMultiSelectUI:            func(_ []string) DependencyUIMultiSelector { return noopMultiSelect{} },
		})

		// Prepare tmp dir
		tmpDir := GinkgoT().TempDir()
		// custom folder path flag requires trailing slash to be considered valid
		cwdValue := tmpDir + string(os.PathSeparator)

		root.SetArgs([]string{"-C", cwdValue})

		a.NotPanics(func() { _ = root.Execute() }) // Use assert.NotPanics
	})

	It("rejects absolute path without trailing slash for -C", func() {
		a := assert.New(GinkgoT()) // Create assert instance

		root := NewRootCmd(Dependencies{
			CommandRunnerGetter:                   func() CommandRunner { return NewFakeCommandRunner() },
			DetectLockfile:                        func() (string, error) { return "", errors.New("no lockfile") },
			DetectJSPacakgeManager:                func() (string, error) { return "npm", nil },
			DetectJSPacakgeManagerBasedOnLockFile: func(string) (string, error) { return "npm", nil },
			YarnCommandVersionOutputter:           dummyYarnVersionRunner{},
			NewCommandTextUI:                      func(string) CommandUITexter { return noopCommandTextUI{} },
			DetectVolta:                           func() bool { return false },
			NewPackageMultiSelectUI:               func(_ []services.PackageInfo) MultiUISelecter { return noopMultiSelect{} },
			NewTaskSelectorUI:                     func(_ []string) TaskUISelector { return noopTaskSelector{} },
			NewDependencyMultiSelectUI:            func(_ []string) DependencyUIMultiSelector { return noopMultiSelect{} },
		})

		tmpDir := GinkgoT().TempDir()
		// no trailing slash
		root.SetArgs([]string{"-C", tmpDir})
		err := root.Execute()
		a.Error(err) // Use assert.Error
	})

	It("accepts relative path with trailing slash and rejects without it", func() {
		a := assert.New(GinkgoT()) // Create assert instance

		rootValid := NewRootCmd(Dependencies{
			CommandRunnerGetter:                   func() CommandRunner { return NewFakeCommandRunner() },
			DetectLockfile:                        func() (string, error) { return "", errors.New("no lockfile") },
			DetectJSPacakgeManager:                func() (string, error) { return "npm", nil },
			DetectJSPacakgeManagerBasedOnLockFile: func(string) (string, error) { return "npm", nil },
			YarnCommandVersionOutputter:           dummyYarnVersionRunner{},
			NewCommandTextUI:                      func(string) CommandUITexter { return noopCommandTextUI{} },
			DetectVolta:                           func() bool { return false },
			NewPackageMultiSelectUI:               func(_ []services.PackageInfo) MultiUISelecter { return noopMultiSelect{} },
			NewTaskSelectorUI:                     func(_ []string) TaskUISelector { return noopTaskSelector{} },
			NewDependencyMultiSelectUI:            func(_ []string) DependencyUIMultiSelector { return noopMultiSelect{} },
		})

		// Create a relative folder in a temp CWD
		orig, _ := os.Getwd()
		defer func() { _ = os.Chdir(orig) }()
		base := GinkgoT().TempDir()
		_ = os.Chdir(base)
		_ = os.Mkdir("rel", 0o755)

		// Valid with trailing slash
		rootValid.SetArgs([]string{"-C", "rel" + string(os.PathSeparator)})
		a.NoError(rootValid.Execute()) // Use assert.NoError

		// Invalid without trailing slash
		rootInvalid := NewRootCmd(Dependencies{
			CommandRunnerGetter:                   func() CommandRunner { return NewFakeCommandRunner() },
			DetectLockfile:                        func() (string, error) { return "", errors.New("no lockfile") },
			DetectJSPacakgeManager:                func() (string, error) { return "npm", nil },
			DetectJSPacakgeManagerBasedOnLockFile: func(string) (string, error) { return "npm", nil },
			YarnCommandVersionOutputter:           dummyYarnVersionRunner{},
			NewCommandTextUI:                      func(string) CommandUITexter { return noopCommandTextUI{} },
			DetectVolta:                           func() bool { return false },
			NewPackageMultiSelectUI:               func(_ []services.PackageInfo) MultiUISelecter { return noopMultiSelect{} },
			NewTaskSelectorUI:                     func(_ []string) TaskUISelector { return noopTaskSelector{} },
			NewDependencyMultiSelectUI:            func(_ []string) DependencyUIMultiSelector { return noopMultiSelect{} },
		})
		rootInvalid.SetArgs([]string{"-C", "rel"})
		a.Error(rootInvalid.Execute()) // Use assert.Error
	})
})
