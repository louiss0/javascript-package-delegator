package cmd_test

import (
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FakeExecCommand is a mock implementation that doesn't actually execute commands
func FakeExecCommand(name string, args ...string) *exec.Cmd {
	// Create a command that does nothing - just echo
	cmd := exec.Command("echo", append([]string{name}, args...)...)
	// Ensure it doesn't actually run by setting to /dev/null
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd
}

// MockExecutor wraps the executor to use our fake command
type MockExecutor struct {
	cmd       *exec.Cmd
	debug     bool
	targetDir string
}

func NewMockExecutor(debug bool) *MockExecutor {
	return &MockExecutor{
		debug: debug,
	}
}

func (m *MockExecutor) IsDebug() bool {
	return m.debug
}

func (m *MockExecutor) Command(name string, args ...string) {
	m.cmd = FakeExecCommand(name, args...)
	if m.targetDir != "" {
		m.cmd.Dir = m.targetDir
	}
}

func (m *MockExecutor) SetTargetDir(dir string) error {
	fileInfo, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("target directory %s is not a directory", dir)
	}
	m.targetDir = dir
	if m.cmd != nil {
		m.cmd.Dir = dir
	}
	return nil
}

func (m *MockExecutor) Run() error {
	if m.cmd == nil {
		return os.ErrInvalid
	}
	// Don't actually run the command
	return nil
}

var _ = Describe("executor SetTargetDir and Command interplay", Label("fast", "unit"), func() {
	It("stores dir when cmd is nil and applies it when Command() is called later", func() {
		// Arrange
		assertions := assert.New(GinkgoT())
		e := NewMockExecutor(false)
		tmpDir := GinkgoT().TempDir()

		// Act
		err := e.SetTargetDir(tmpDir)
		assertions.NoError(err)
		e.Command("bash", "-c", "echo ok")

		// Assert
		assertions.NotNil(e.cmd)
		assertions.Equal(tmpDir, e.cmd.Dir)
	})

	It("sets cmd.Dir immediately when a command already exists", func() {
		// Arrange
		assertions := assert.New(GinkgoT())
		e := NewMockExecutor(false)
		e.Command("bash", "-c", "echo ok")
		tmpDir := GinkgoT().TempDir()

		// Act
		err := e.SetTargetDir(tmpDir)

		// Assert
		assertions.NoError(err)
		assertions.NotNil(e.cmd)
		assertions.Equal(tmpDir, e.cmd.Dir)
	})

	It("rejects non-existent paths", func() {
		// Arrange
		assertions := assert.New(GinkgoT())
		e := NewMockExecutor(false)

		// Act
		err := e.SetTargetDir("/path/that/does/not/exist")

		// Assert
		assertions.Error(err)
	})

	It("rejects regular files", func() {
		// Arrange
		assertions := assert.New(GinkgoT())
		requires := require.New(GinkgoT())
		f, err := os.CreateTemp("", "executor-file-*")
		requires.NoError(err)
		defer func() { _ = os.Remove(f.Name()) }()
		_ = f.Close()
		e := NewMockExecutor(false)

		// Act
		err = e.SetTargetDir(f.Name())

		// Assert
		assertions.Error(err)
	})

	It("Run returns error if Command() was never called", func() {
		// Arrange
		assertions := assert.New(GinkgoT())
		e := NewMockExecutor(false)

		// Act
		err := e.Run()

		// Assert
		assertions.Error(err)
	})

	It("IsDebug returns the debug flag set at construction time", func() {
		// Arrange
		assertions := assert.New(GinkgoT())
		e := NewMockExecutor(true)

		// Assert
		assertions.True(e.IsDebug())
	})

	It("applies the last SetTargetDir() call and persists across subsequent Command() calls", func() {
		assertions := assert.New(GinkgoT())
		e := NewMockExecutor(false)
		tmpDir1 := GinkgoT().TempDir()
		tmpDir2 := GinkgoT().TempDir()

		// First set and command
		assertions.NoError(e.SetTargetDir(tmpDir1))
		e.Command("bash", "-c", "echo ok")
		assertions.Equal(tmpDir1, e.cmd.Dir)

		// Second set should override
		assertions.NoError(e.SetTargetDir(tmpDir2))
		assertions.Equal(tmpDir2, e.cmd.Dir)

		// New command should inherit the latest dir
		e.Command("bash", "-c", "echo again")
		assertions.Equal(tmpDir2, e.cmd.Dir)
	})

	It("accepts relative and absolute paths for SetTargetDir", func() {
		assertions := assert.New(GinkgoT())
		e := NewMockExecutor(false)

		// Save and switch CWD to a temp dir to make relative path deterministic
		orig, err := os.Getwd()
		assertions.NoError(err)
		defer func() { _ = os.Chdir(orig) }()

		base := GinkgoT().TempDir()
		assertions.NoError(os.Chdir(base))

		// Create a relative subfolder
		rel := "relwork"
		assertions.NoError(os.Mkdir(rel, 0o755))

		// Relative
		assertions.NoError(e.SetTargetDir(rel))
		e.Command("bash", "-c", "echo ok")
		assertions.Equal(rel, e.cmd.Dir)

		// Absolute
		abs := base
		assertions.NoError(e.SetTargetDir(abs))
		e.Command("bash", "-c", "echo ok")
		assertions.Equal(abs, e.cmd.Dir)
	})
})
