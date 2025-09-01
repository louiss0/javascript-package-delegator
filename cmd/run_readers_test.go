package cmd

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
)

func TestRunReaders(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Run Readers Suite")
}

var _ = Describe("Directory-aware readers", func() {
	var originalWD string

	BeforeEach(func() {
		wd, _ := os.Getwd()
		originalWD = wd
	})

	AfterEach(func() {
		_ = os.Chdir(originalWD)
	})

	Describe("readPackageJSONAndUnmarshalScriptsFrom", func() {
		It("reads package.json from provided dir not from os.Getwd", func() {
			// Arrange: two temp dirs with different package.json contents
			target := GinkgoT().TempDir()
			other := GinkgoT().TempDir()

			err := os.WriteFile(filepath.Join(target, "package.json"), []byte(`{"scripts":{"from-target":"echo hi"}}`), 0644)
			assert.NoError(GinkgoT(), err)

			err = os.WriteFile(filepath.Join(other, "package.json"), []byte(`{"scripts":{"from-other":"echo nope"}}`), 0644)
			assert.NoError(GinkgoT(), err)

			// Simulate misleading current working directory
			err = os.Chdir(other)
			assert.NoError(GinkgoT(), err)

			// Act
			pkg, err := readPackageJSONAndUnmarshalScriptsFrom(target)

			// Assert
			assert.NoError(GinkgoT(), err)
			assert.Contains(GinkgoT(), pkg.Scripts, "from-target")
			assert.NotContains(GinkgoT(), pkg.Scripts, "from-other")
		})

		It("returns error when package.json not found in base directory", func() {
			nonExistent := GinkgoT().TempDir()
			// Don't create package.json in this directory

			pkg, err := readPackageJSONAndUnmarshalScriptsFrom(nonExistent)

			assert.Error(GinkgoT(), err)
			assert.Nil(GinkgoT(), pkg)
			assert.Contains(GinkgoT(), err.Error(), "failed to read package.json")
		})

		It("returns error when package.json has invalid JSON", func() {
			target := GinkgoT().TempDir()

			err := os.WriteFile(filepath.Join(target, "package.json"), []byte(`{invalid json}`), 0644)
			assert.NoError(GinkgoT(), err)

			pkg, err := readPackageJSONAndUnmarshalScriptsFrom(target)

			assert.Error(GinkgoT(), err)
			assert.Nil(GinkgoT(), pkg)
			assert.Contains(GinkgoT(), err.Error(), "failed to parse package.json")
		})
	})

	Describe("readDenoJSONFrom", func() {
		It("reads deno.json from provided dir not from os.Getwd", func() {
			target := GinkgoT().TempDir()
			other := GinkgoT().TempDir()

			err := os.WriteFile(filepath.Join(target, "deno.json"), []byte(`{"tasks":{"from-target":"deno run foo.ts"}}`), 0644)
			assert.NoError(GinkgoT(), err)

			err = os.WriteFile(filepath.Join(other, "deno.json"), []byte(`{"tasks":{"from-other":"deno run bar.ts"}}`), 0644)
			assert.NoError(GinkgoT(), err)

			// Simulate misleading current working directory
			err = os.Chdir(other)
			assert.NoError(GinkgoT(), err)

			// Act
			pkg, err := readDenoJSONFrom(target)

			// Assert
			assert.NoError(GinkgoT(), err)
			assert.Contains(GinkgoT(), pkg.Tasks, "from-target")
			assert.NotContains(GinkgoT(), pkg.Tasks, "from-other")
		})

		It("returns error when deno.json not found in base directory", func() {
			nonExistent := GinkgoT().TempDir()
			// Don't create deno.json in this directory

			pkg, err := readDenoJSONFrom(nonExistent)

			assert.Error(GinkgoT(), err)
			assert.Nil(GinkgoT(), pkg)
			assert.Contains(GinkgoT(), err.Error(), "failed to read deno.json")
		})

		It("returns error when deno.json has invalid JSON", func() {
			target := GinkgoT().TempDir()

			err := os.WriteFile(filepath.Join(target, "deno.json"), []byte(`{invalid json}`), 0644)
			assert.NoError(GinkgoT(), err)

			pkg, err := readDenoJSONFrom(target)

			assert.Error(GinkgoT(), err)
			assert.Nil(GinkgoT(), pkg)
			assert.Contains(GinkgoT(), err.Error(), "failed to parse deno.json")
		})
	})
})
