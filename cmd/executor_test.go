package cmd

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("executor SetTargetDir and Command interplay", func() {
	It("stores dir when cmd is nil and applies it when Command() is called later", func() {
		// Arrange
		e := newExecutor(exec.Command, false)
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
		e := newExecutor(exec.Command, false)
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
		e := newExecutor(exec.Command, false)

		// Act
		err := e.SetTargetDir("/path/that/does/not/exist")

		// Assert
		Expect(err).To(HaveOccurred())
	})

	It("rejects regular files", func() {
		// Arrange
		f, err := os.CreateTemp("", "executor-file-*")
		Expect(err).NotTo(HaveOccurred())
		defer os.Remove(f.Name())
		f.Close()
		e := newExecutor(exec.Command, false)

		// Act
		err = e.SetTargetDir(f.Name())

		// Assert
		Expect(err).To(HaveOccurred())
	})

	It("Run returns error if Command() was never called", func() {
		// Arrange
		e := newExecutor(exec.Command, false)

		// Act
		err := e.Run()

		// Assert
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no command set to run"))
	})

It("IsDebug returns the debug flag set at construction time", func() {
        // Arrange
        e := newExecutor(exec.Command, true)

        // Assert
        Expect(e.IsDebug()).To(BeTrue())
    })

    It("applies the last SetTargetDir() call and persists across subsequent Command() calls", func() {
        e := newExecutor(exec.Command, false)
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
        e := newExecutor(exec.Command, false)

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
