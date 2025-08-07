package cmd

import (
	"errors"
	"os"
	"os/exec"

	"github.com/louiss0/javascript-package-delegator/services"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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
		// Build a NewRootCmd with stubbed dependencies per spec
		root := NewRootCmd(Dependencies{
			CommandRunnerGetter: func(debug bool) CommandRunner { return newExecutor(exec.Command, false) },
			DetectLockfile: func() (string, error) { return "", errors.New("no lockfile") },
			DetectJSPacakgeManager: func() (string, error) { return "npm", nil },
			DetectJSPacakgeManagerBasedOnLockFile: func(string) (string, error) { return "npm", nil },
			YarnCommandVersionOutputter:          dummyYarnVersionRunner{},
			NewCommandTextUI: func(string) CommandUITexter { return noopCommandTextUI{} },
			DetectVolta: func() bool { return false },
			NewPackageMultiSelectUI: func(_ []services.PackageInfo) MultiUISelecter { return noopMultiSelect{} },
			NewTaskSelectorUI:       func(_ []string) TaskUISelector { return noopTaskSelector{} },
			NewDependencyMultiSelectUI: func(_ []string) DependencyUIMultiSelector { return noopMultiSelect{} },
		})

		// Prepare tmp dir
		tmpDir := GinkgoT().TempDir()
		// custom folder path flag requires trailing slash to be considered valid
		cwdValue := tmpDir + string(os.PathSeparator)

		root.SetArgs([]string{"-C", cwdValue})

Expect(func() { _ = root.Execute() }).NotTo(Panic())
	})

	It("rejects absolute path without trailing slash for -C", func() {
        root := NewRootCmd(Dependencies{
            CommandRunnerGetter: func(debug bool) CommandRunner { return newExecutor(exec.Command, false) },
            DetectLockfile: func() (string, error) { return "", errors.New("no lockfile") },
            DetectJSPacakgeManager: func() (string, error) { return "npm", nil },
            DetectJSPacakgeManagerBasedOnLockFile: func(string) (string, error) { return "npm", nil },
            YarnCommandVersionOutputter:          dummyYarnVersionRunner{},
            NewCommandTextUI: func(string) CommandUITexter { return noopCommandTextUI{} },
            DetectVolta: func() bool { return false },
            NewPackageMultiSelectUI: func(_ []services.PackageInfo) MultiUISelecter { return noopMultiSelect{} },
            NewTaskSelectorUI:       func(_ []string) TaskUISelector { return noopTaskSelector{} },
            NewDependencyMultiSelectUI: func(_ []string) DependencyUIMultiSelector { return noopMultiSelect{} },
        })

        tmpDir := GinkgoT().TempDir()
        // no trailing slash
        root.SetArgs([]string{"-C", tmpDir})
        err := root.Execute()
        Expect(err).To(HaveOccurred())
    })


	It("accepts relative path with trailing slash and rejects without it", func() {
        rootValid := NewRootCmd(Dependencies{
            CommandRunnerGetter: func(debug bool) CommandRunner { return newExecutor(exec.Command, false) },
            DetectLockfile: func() (string, error) { return "", errors.New("no lockfile") },
            DetectJSPacakgeManager: func() (string, error) { return "npm", nil },
            DetectJSPacakgeManagerBasedOnLockFile: func(string) (string, error) { return "npm", nil },
            YarnCommandVersionOutputter:          dummyYarnVersionRunner{},
            NewCommandTextUI: func(string) CommandUITexter { return noopCommandTextUI{} },
            DetectVolta: func() bool { return false },
            NewPackageMultiSelectUI: func(_ []services.PackageInfo) MultiUISelecter { return noopMultiSelect{} },
            NewTaskSelectorUI:       func(_ []string) TaskUISelector { return noopTaskSelector{} },
            NewDependencyMultiSelectUI: func(_ []string) DependencyUIMultiSelector { return noopMultiSelect{} },
        })

        // Create a relative folder in a temp CWD
        orig, _ := os.Getwd()
        defer func() { _ = os.Chdir(orig) }()
        base := GinkgoT().TempDir()
        _ = os.Chdir(base)
        _ = os.Mkdir("rel", 0o755)

        // Valid with trailing slash
        rootValid.SetArgs([]string{"-C", "rel" + string(os.PathSeparator)})
        Expect(rootValid.Execute()).To(Succeed())

        // Invalid without trailing slash
        rootInvalid := NewRootCmd(Dependencies{
            CommandRunnerGetter: func(debug bool) CommandRunner { return newExecutor(exec.Command, false) },
            DetectLockfile: func() (string, error) { return "", errors.New("no lockfile") },
            DetectJSPacakgeManager: func() (string, error) { return "npm", nil },
            DetectJSPacakgeManagerBasedOnLockFile: func(string) (string, error) { return "npm", nil },
            YarnCommandVersionOutputter:          dummyYarnVersionRunner{},
            NewCommandTextUI: func(string) CommandUITexter { return noopCommandTextUI{} },
            DetectVolta: func() bool { return false },
            NewPackageMultiSelectUI: func(_ []services.PackageInfo) MultiUISelecter { return noopMultiSelect{} },
            NewTaskSelectorUI:       func(_ []string) TaskUISelector { return noopTaskSelector{} },
            NewDependencyMultiSelectUI: func(_ []string) DependencyUIMultiSelector { return noopMultiSelect{} },
        })
        rootInvalid.SetArgs([]string{"-C", "rel"})
        Expect(rootInvalid.Execute()).NotTo(Succeed())
    })
})

