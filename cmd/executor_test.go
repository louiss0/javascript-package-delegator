package cmd_test

import (
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

var _ = Describe("executor SetTargetDir and Command interplay", func() {
	It("stores dir when cmd is nil and applies it when Command() is called later", func() {
		// Arrange
		e := NewMockExecutor(false)
		tmpDir := GinkgoT().TempDir()

		// Act
		err := e.SetTargetDir(tmpDir)
		Expect(err).NotTo(HaveOccurred())
		e.Command("bash", "-c", "echo ok")

		// Assert
		Expect(e.cmd).NotTo(BeNil())
		Expect(e.cmd.Dir).To(Equal(tmpDir))
	})

	It("sets cmd.Dir immediately when a command already exists", func() {
		// Arrange
		e := NewMockExecutor(false)
		e.Command("bash", "-c", "echo ok")
		tmpDir := GinkgoT().TempDir()

		// Act
		err := e.SetTargetDir(tmpDir)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(e.cmd).NotTo(BeNil())
		Expect(e.cmd.Dir).To(Equal(tmpDir))
	})

	It("rejects non-existent paths", func() {
		// Arrange
		e := NewMockExecutor(false)

		// Act
		err := e.SetTargetDir("/path/that/does/not/exist")

		// Assert
		Expect(err).To(HaveOccurred())
	})

	It("rejects regular files", func() {
		// Arrange
		f, err := os.CreateTemp("", "executor-file-*")
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = os.Remove(f.Name()) }()
		_ = f.Close()
		e := NewMockExecutor(false)

		// Act
		err = e.SetTargetDir(f.Name())

		// Assert
		Expect(err).To(HaveOccurred())
	})

	It("Run returns error if Command() was never called", func() {
		// Arrange
		e := NewMockExecutor(false)

		// Act
		err := e.Run()

		// Assert
		Expect(err).To(HaveOccurred())
	})

	It("IsDebug returns the debug flag set at construction time", func() {
		// Arrange
		e := NewMockExecutor(true)

		// Assert
		Expect(e.IsDebug()).To(BeTrue())
	})

	It("applies the last SetTargetDir() call and persists across subsequent Command() calls", func() {
		e := NewMockExecutor(false)
		tmpDir1 := GinkgoT().TempDir()
		tmpDir2 := GinkgoT().TempDir()

		// First set and command
		Expect(e.SetTargetDir(tmpDir1)).To(Succeed())
		e.Command("bash", "-c", "echo ok")
		Expect(e.cmd.Dir).To(Equal(tmpDir1))

		// Second set should override
		Expect(e.SetTargetDir(tmpDir2)).To(Succeed())
		Expect(e.cmd.Dir).To(Equal(tmpDir2))

		// New command should inherit the latest dir
		e.Command("bash", "-c", "echo again")
		Expect(e.cmd.Dir).To(Equal(tmpDir2))
	})

	It("accepts relative and absolute paths for SetTargetDir", func() {
		e := NewMockExecutor(false)

		// Save and switch CWD to a temp dir to make relative path deterministic
		orig, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = os.Chdir(orig) }()

		base := GinkgoT().TempDir()
		Expect(os.Chdir(base)).To(Succeed())

		// Create a relative subfolder
		rel := "relwork"
		Expect(os.Mkdir(rel, 0o755)).To(Succeed())

		// Relative
		Expect(e.SetTargetDir(rel)).To(Succeed())
		e.Command("bash", "-c", "echo ok")
		Expect(e.cmd.Dir).To(Equal(rel))

		// Absolute
		abs := base
		Expect(e.SetTargetDir(abs)).To(Succeed())
		e.Command("bash", "-c", "echo ok")
		Expect(e.cmd.Dir).To(Equal(abs))
	})
})
