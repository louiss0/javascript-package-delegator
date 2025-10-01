package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
)


var _ = Describe("Run Command Helper Functions", Label("fast", "unit"), func() {

	Context("parsePackageNames function", func() {
		It("should extract package names from dependency@version strings", func() {
			packages := parsePackageNames([]string{"react@18.2.0", "lodash@4.17.21"})
			assert.Equal(GinkgoT(), []string{"react", "lodash"}, packages)
		})

		It("should handle scoped packages with versions", func() {
			packages := parsePackageNames([]string{"@types/node@20.0.0", "@typescript-eslint/parser@6.0.0"})
			assert.Equal(GinkgoT(), []string{"@types/node", "@typescript-eslint/parser"}, packages)
		})

		It("should handle packages without versions", func() {
			packages := parsePackageNames([]string{"react", "@types/node"})
			assert.Equal(GinkgoT(), []string{"react", "@types/node"}, packages)
		})

		It("should handle empty input", func() {
			packages := parsePackageNames([]string{})
			assert.Empty(GinkgoT(), packages)
		})

		It("should handle mixed versioned and unversioned packages", func() {
			packages := parsePackageNames([]string{"react@18.2.0", "lodash", "@types/node@20.0.0", "@babel/core"})
			assert.Equal(GinkgoT(), []string{"react", "lodash", "@types/node", "@babel/core"}, packages)
		})
	})

	Context("isYarnPnpProject function", func() {
		It("should return true when .pnp.cjs exists", func() {
			// Create a temporary directory to test the actual function
			tempDir := GinkgoT().TempDir()
			pnpFile := filepath.Join(tempDir, ".pnp.cjs")
			err := os.WriteFile(pnpFile, []byte("// PnP file"), 0644)
			assert.NoError(GinkgoT(), err)
			
			result := isYarnPnpProject(tempDir)
			assert.True(GinkgoT(), result)
		})

		It("should return true when .pnp.data.json exists", func() {
			tempDir := GinkgoT().TempDir()
			pnpDataFile := filepath.Join(tempDir, ".pnp.data.json")
			err := os.WriteFile(pnpDataFile, []byte("{}"), 0644)
			assert.NoError(GinkgoT(), err)
			
			result := isYarnPnpProject(tempDir)
			assert.True(GinkgoT(), result)
		})

		It("should return false when neither .pnp file exists", func() {
			tempDir := GinkgoT().TempDir()
			
			result := isYarnPnpProject(tempDir)
			assert.False(GinkgoT(), result)
		})
	})

	Context("missingNodePackages function", func() {
		It("should return missing packages when node_modules directories don't exist", func() {
			tempDir := GinkgoT().TempDir()
			
			// Test with actual filesystem - all packages missing
			missing := missingNodePackages(tempDir, []string{"react", "lodash", "typescript"})
			assert.Equal(GinkgoT(), []string{"react", "lodash", "typescript"}, missing)
		})

		It("should return only actually missing packages", func() {
			tempDir := GinkgoT().TempDir()
			nodeModulesPath := filepath.Join(tempDir, "node_modules")
			err := os.MkdirAll(nodeModulesPath, 0755)
			assert.NoError(GinkgoT(), err)
			
			// Create some packages
			reactPath := filepath.Join(nodeModulesPath, "react")
			err = os.MkdirAll(reactPath, 0755)
			assert.NoError(GinkgoT(), err)
			
			typescriptPath := filepath.Join(nodeModulesPath, "typescript")
			err = os.MkdirAll(typescriptPath, 0755)
			assert.NoError(GinkgoT(), err)
			
			// lodash is missing
			missing := missingNodePackages(tempDir, []string{"react", "lodash", "typescript"})
			assert.Equal(GinkgoT(), []string{"lodash"}, missing)
		})

		It("should return empty slice when all packages exist", func() {
			tempDir := GinkgoT().TempDir()
			nodeModulesPath := filepath.Join(tempDir, "node_modules")
			err := os.MkdirAll(nodeModulesPath, 0755)
			assert.NoError(GinkgoT(), err)
			
			// Create all packages
			for _, pkg := range []string{"react", "lodash"} {
				pkgPath := filepath.Join(nodeModulesPath, pkg)
				err = os.MkdirAll(pkgPath, 0755)
				assert.NoError(GinkgoT(), err)
			}
			
			missing := missingNodePackages(tempDir, []string{"react", "lodash"})
			assert.Empty(GinkgoT(), missing)
		})

		It("should handle scoped packages correctly", func() {
			tempDir := GinkgoT().TempDir()
			nodeModulesPath := filepath.Join(tempDir, "node_modules")
			err := os.MkdirAll(nodeModulesPath, 0755)
			assert.NoError(GinkgoT(), err)
			
			// @types/node should be missing
			missing := missingNodePackages(tempDir, []string{"@types/node"})
			assert.Equal(GinkgoT(), []string{"@types/node"}, missing)
			
			// Create the scoped package
			scopedPath := filepath.Join(nodeModulesPath, "@types", "node")
			err = os.MkdirAll(scopedPath, 0755)
			assert.NoError(GinkgoT(), err)
			
			// Now it should exist
			missing = missingNodePackages(tempDir, []string{"@types/node"})
			assert.Empty(GinkgoT(), missing)
		})

		It("should respect maxMissing limit", func() {
			tempDir := GinkgoT().TempDir()
			
			// Create 12 missing packages, should only return first 10
			manyPackages := make([]string, 12)
			for i := 0; i < 12; i++ {
				manyPackages[i] = fmt.Sprintf("package%d", i)
			}
			
			missing := missingNodePackages(tempDir, manyPackages)
			assert.Len(GinkgoT(), missing, 10, "Should respect maxMissing limit of 10")
			assert.Equal(GinkgoT(), manyPackages[:10], missing)
		})
	})
})

func TestRunHelperFunctions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Run Command Helper Functions Suite")
}
