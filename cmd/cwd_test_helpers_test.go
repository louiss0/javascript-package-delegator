package cmd_test

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/louiss0/javascript-package-delegator/cmd"
	"github.com/louiss0/javascript-package-delegator/services"
)

type MockFileSystemCwd struct {
	files map[string]bool
	cwd   string
}

func (m *MockFileSystemCwd) Exists(path string) bool {
	return m.files[path]
}

func (m *MockFileSystemCwd) Getwd() (string, error) {
	if m.cwd != "" {
		return m.cwd, nil
	}
	return os.Getwd()
}

func (m *MockFileSystemCwd) Stat(name string) (os.FileInfo, error) {
	if m.files[name] {
		return &mockFileInfo{name: filepath.Base(name)}, nil
	}
	return nil, os.ErrNotExist
}

type mockFileInfo struct {
	name string
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil }

type MockPathLookupCwd struct {
	paths map[string]bool
}

func (m *MockPathLookupCwd) LookPath(executable string) (string, error) {
	if m.paths[executable] {
		return "/usr/bin/" + executable, nil
	}
	return "", fmt.Errorf("executable file not found in $PATH")
}

type FakeCommandRunnerCwd struct {
	lastCommand string
	lastArgs    []string
	lastWorkDir string
}

func (f *FakeCommandRunnerCwd) Command(name string, arg ...string) {
	f.lastCommand = name
	f.lastArgs = arg
}

func (f *FakeCommandRunnerCwd) SetTargetDir(dir string) error {
	f.lastWorkDir = dir
	return nil
}

func (f *FakeCommandRunnerCwd) Run() error {
	return nil
}

type MockYarnVersionOutputterCwd struct {
	version string
}

func (m *MockYarnVersionOutputterCwd) Output() (string, error) {
	return m.version, nil
}

type mockCommandTextUICwd struct {
	lockfile string
	value    string
}

func (m *mockCommandTextUICwd) Run() error {
	m.value = "npm install -g pnpm"
	return nil
}

func (m *mockCommandTextUICwd) Value() string {
	return m.value
}

func newMockCommandTextUICwd(lockfile string) cmd.CommandUITexter {
	return &mockCommandTextUICwd{lockfile: lockfile}
}

type mockPackageMultiSelectUICwd struct{}

func (m *mockPackageMultiSelectUICwd) Run() error {
	return nil
}

func (m *mockPackageMultiSelectUICwd) Values() []string {
	return []string{"test-package@1.0.0"}
}

func newMockPackageMultiSelectUICwd(packageInfos []services.PackageInfo) cmd.MultiUISelecter {
	return &mockPackageMultiSelectUICwd{}
}

type mockTaskSelectorUICwd struct {
	value string
}

func (m *mockTaskSelectorUICwd) Run() error {
	m.value = "dev"
	return nil
}

func (m *mockTaskSelectorUICwd) Value() string {
	return m.value
}

func newMockTaskSelectorUICwd(options []string) cmd.TaskUISelector {
	return &mockTaskSelectorUICwd{}
}

type mockDependencyMultiSelectUICwd struct {
	values []string
}

func (m *mockDependencyMultiSelectUICwd) Run() error {
	m.values = []string{"lodash@4.17.21"}
	return nil
}

func (m *mockDependencyMultiSelectUICwd) Values() []string {
	return m.values
}

func newMockDependencyMultiSelectUICwd(options []string) cmd.DependencyUIMultiSelector {
	return &mockDependencyMultiSelectUICwd{}
}

type mockDebugExecutorCwd struct{}

func (m *mockDebugExecutorCwd) ExecuteIfDebugIsTrue(cb func())                                  {}
func (m *mockDebugExecutorCwd) LogDebugMessageIfDebugIsTrue(msg string, keyvals ...interface{}) {}
func (m *mockDebugExecutorCwd) LogJSCommandIfDebugIsTrue(name string, args ...string)           {}

func newMockDebugExecutorCwd(debug bool) cmd.DebugExecutor {
	return &mockDebugExecutorCwd{}
}
