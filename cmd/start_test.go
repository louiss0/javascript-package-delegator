package cmd_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/cmd"
	"github.com/louiss0/javascript-package-delegator/mock"
	"github.com/louiss0/javascript-package-delegator/testutil"
)

var _ = Describe("Start Command", func() {
	var (
		rootCmd           *cobra.Command
		mockCommandRunner *mock.MockCommandRunner
		factory           *testutil.RootCommandFactory
		tempDir           string
	)

	assert := assert.New(GinkgoT())

	writeFile := func(name, content string) {
		assert.NoError(os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0644))
	}

	writeNodeProject := func(lockfile string) {
		writeFile("package.json", `{
  "scripts": {
    "dev": "echo dev"
  }
}`)
		writeFile(lockfile, "{}")
	}

	BeforeEach(func() {
		mockCommandRunner = mock.NewMockCommandRunner()
		factory = testutil.NewRootCommandFactory(mockCommandRunner)
		factory.SetupBasicCommandRunnerExpectations()
		factory.ResetDebugExecutor()
		factory.SetupBasicDebugExecutorExpectations()

		rootCmd = factory.CreateNpmAsDefault(nil)
		rootCmd.AddCommand(cmd.NewStartCmd())

		var err error
		tempDir, err = os.MkdirTemp("", "start-cmd-test")
		assert.NoError(err)
	})

	AfterEach(func() {
		assert.NoError(os.RemoveAll(tempDir))
		mockCommandRunner.AssertExpectations(GinkgoT())
		factory.DebugExecutor().AssertExpectations(GinkgoT())
	})

	Context("script resolution for npm projects", func() {
		It("prefers the dev script when both dev and start exist", func() {
			writeFile("package.json", `{
    "scripts": {
      "dev": "echo dev",
      "start": "echo start"
    }
  }`)
			writeFile("package-lock.json", "{}")

			_, err := executeCmd(rootCmd, "start", "--cwd", fmt.Sprintf("%s/", tempDir))
			assert.NoError(err)

			assert.True(mockCommandRunner.HasCommand("npm", "run", "dev"))
		})

		It("falls back to the start script when dev is missing", func() {
			writeFile("package.json", `{
    "scripts": {
      "start": "echo start"
    }
  }`)
			writeFile("package-lock.json", "{}")

			_, err := executeCmd(rootCmd, "start", "--cwd", fmt.Sprintf("%s/", tempDir))
			assert.NoError(err)

			assert.True(mockCommandRunner.HasCommand("npm", "run", "start"))
		})

		It("allows selecting a custom script via --script", func() {
			writeFile("package.json", `{
    "scripts": {
      "dev": "echo dev",
      "preview": "echo preview"
    }
  }`)
			writeFile("package-lock.json", "{}")

			_, err := executeCmd(rootCmd, "start", "--cwd", fmt.Sprintf("%s/", tempDir), "--script", "preview")
			assert.NoError(err)

			assert.True(mockCommandRunner.HasCommand("npm", "run", "preview"))
		})

		It("errors when --script points to a missing script", func() {
			writeFile("package.json", `{
    "scripts": {
      "dev": "echo dev"
    }
  }`)
			writeFile("package-lock.json", "{}")

			_, err := executeCmd(rootCmd, "start", "--cwd", fmt.Sprintf("%s/", tempDir), "--script", "preview")
			assert.Error(err)
		})

		It("matches scripts containing the words dev or start", func() {
			writeFile("package.json", `{
    "scripts": {
      "dev-server": "echo dev server"
    }
  }`)
			writeFile("package-lock.json", "{}")

			_, err := executeCmd(rootCmd, "start", "--cwd", fmt.Sprintf("%s/", tempDir))
			assert.NoError(err)

			assert.True(mockCommandRunner.HasCommand("npm", "run", "dev-server"))
		})
	})

	Context("package manager specific behavior", func() {
		It("inserts an argument separator for pnpm", func() {
			rootCmd = factory.CreatePnpmAsDefault(nil)
			rootCmd.AddCommand(cmd.NewStartCmd())
			writeNodeProject("pnpm-lock.yaml")

			_, err := executeCmd(rootCmd, "start", "--cwd", fmt.Sprintf("%s/", tempDir), "--", "--host", "0.0.0.0", "--port", "4321")
			assert.NoError(err)

			assert.True(
				mockCommandRunner.HasCommand("pnpm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "4321"),
			)
		})

		It("passes args directly for yarn without inserting --", func() {
			rootCmd = factory.CreateYarnTwoAsDefault(nil)
			rootCmd.AddCommand(cmd.NewStartCmd())
			writeNodeProject("yarn.lock")

			_, err := executeCmd(rootCmd, "start", "--cwd", fmt.Sprintf("%s/", tempDir), "--", "--host", "0.0.0.0")
			assert.NoError(err)

			assert.True(
				mockCommandRunner.HasCommand("yarn", "run", "dev", "--host", "0.0.0.0"),
			)
		})

		It("runs bun scripts with positional args", func() {
			rootCmd = factory.CreateBunAsDefault(nil)
			rootCmd.AddCommand(cmd.NewStartCmd())
			writeNodeProject("bun.lockb")

			_, err := executeCmd(rootCmd, "start", "--cwd", fmt.Sprintf("%s/", tempDir), "--", "--inspect")
			assert.NoError(err)

			assert.True(
				mockCommandRunner.HasCommand("bun", "run", "dev", "--inspect"),
			)
		})
	})

	Context("script resolution for deno projects", func() {
		BeforeEach(func() {
			rootCmd = factory.CreateDenoAsDefault(nil)
			rootCmd.AddCommand(cmd.NewStartCmd())
		})

		It("runs the dev task by default", func() {
			writeFile("deno.json", `{
    "tasks": {
      "dev": "deno task dev",
      "start": "deno task start"
    }
  }`)

			_, err := executeCmd(rootCmd, "start", "--cwd", fmt.Sprintf("%s/", tempDir))
			assert.NoError(err)
			assert.True(mockCommandRunner.HasCommand("deno", "task", "dev"))
		})

		It("allows selecting a task via --script", func() {
			writeFile("deno.json", `{
    "tasks": {
      "dev": "deno task dev",
      "serve": "deno task serve"
    }
  }`)

			_, err := executeCmd(rootCmd, "start", "--cwd", fmt.Sprintf("%s/", tempDir), "--script", "serve")
			assert.NoError(err)
			assert.True(mockCommandRunner.HasCommand("deno", "task", "serve"))
		})

		It("rejects --eval usage for deno", func() {
			writeFile("deno.json", `{
    "tasks": {
      "dev": "deno task dev"
    }
  }`)

			_, err := executeCmd(rootCmd, "start", "--cwd", fmt.Sprintf("%s/", tempDir), "--", "--eval")
			assert.Error(err)
			assert.Contains(err.Error(), "don't pass --eval")
		})
	})
})
