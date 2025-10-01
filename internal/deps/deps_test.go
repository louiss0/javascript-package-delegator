package deps_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/internal/deps"
)


var _ = Describe("Deps Package", Label("integration", "unit"), func() {
	assert := assert.New(GinkgoT())

	Context("JSONC Normalization", func() {
		DescribeTable("NormalizeJSONCToJSON should handle various JSONC features",
			func(input, expected string) {
				result := deps.NormalizeJSONCToJSON([]byte(input))
				assert.Equal(expected, string(result))
			},
			Entry("removes single-line comments",
				`{
					"name": "test", // this is a comment
					"version": "1.0.0"
				}`,
				`{
					"name": "test", 
					"version": "1.0.0"
				}`),
			Entry("removes multi-line comments",
				`{
					/* this is a 
					   multi-line comment */
					"name": "test",
					"version": "1.0.0"
				}`,
				`{
					
					"name": "test",
					"version": "1.0.0"
				}`),
			Entry("removes trailing commas",
				`{
					"name": "test",
					"deps": {
						"lodash": "1.0.0",
					},
				}`,
				`{
					"name": "test",
					"deps": {
						"lodash": "1.0.0"
					}
				}`),
			Entry("handles mixed JSONC features with URLs",
				`{
					// Main config
					"imports": {
						"lodash": "npm:lodash@4.17.21", // utility lib
						/* React for UI components */
						"react": "npm:react@18.2.0",
					}, // end imports
				}`,
				`{
					
					"imports": {
						"lodash": "npm:lodash@4.17.21", 
						
						"react": "npm:react@18.2.0"
					} 
				}`),
		)
	})

	Context("Hash Computation", func() {
		Context("Node.js Dependencies Hashing", func() {
			It("should compute consistent hash for package.json dependencies", func() {
				tempDir := GinkgoT().TempDir()
				packageJSON := `{
					"dependencies": {
						"react": "^18.2.0",
						"lodash": "^4.17.21"
					},
					"devDependencies": {
						"typescript": "^5.0.0",
						"@types/node": "^20.0.0"
					}
				}`
				
				packageJSONPath := filepath.Join(tempDir, "package.json")
				err := os.WriteFile(packageJSONPath, []byte(packageJSON), 0644)
				assert.NoError(err)
				
				hash1, err := deps.ComputeNodeDepsHash(tempDir)
				assert.NoError(err)
				assert.NotEmpty(hash1)
				assert.Len(hash1, 64) // SHA256 produces 64 hex characters
			})

			It("should produce same hash regardless of dependency order in package.json", func() {
				tempDir1 := GinkgoT().TempDir()
				tempDir2 := GinkgoT().TempDir()
				
				packageJSON1 := `{
					"dependencies": {
						"react": "^18.2.0",
						"lodash": "^4.17.21"
					},
					"devDependencies": {
						"typescript": "^5.0.0"
					}
				}`

				packageJSON2 := `{
					"dependencies": {
						"lodash": "^4.17.21",
						"react": "^18.2.0"
					},
					"devDependencies": {
						"typescript": "^5.0.0"
					}
				}`

				// Create package.json files with different ordering
				err := os.WriteFile(filepath.Join(tempDir1, "package.json"), []byte(packageJSON1), 0644)
				assert.NoError(err)
				err = os.WriteFile(filepath.Join(tempDir2, "package.json"), []byte(packageJSON2), 0644)
				assert.NoError(err)
				
				// Both should produce the same hash due to sorted keys
				hash1, err1 := deps.ComputeNodeDepsHash(tempDir1)
				hash2, err2 := deps.ComputeNodeDepsHash(tempDir2)
				
				assert.NoError(err1)
				assert.NoError(err2)
				assert.Equal(hash1, hash2, "Hashes should be identical regardless of JSON key order")
			})

			It("should return error for missing package.json", func() {
				_, err := deps.ComputeNodeDepsHash("/nonexistent/dir")
				assert.Error(err)
				assert.Contains(err.Error(), "failed to read package.json")
			})

			It("should return error for invalid JSON", func() {
				tempDir := GinkgoT().TempDir()
				invalidJSON := `{ "dependencies": { "react": }`
				packageJSONPath := filepath.Join(tempDir, "package.json")
				err := os.WriteFile(packageJSONPath, []byte(invalidJSON), 0644)
				assert.NoError(err)
				
				_, err = deps.ComputeNodeDepsHash(tempDir)
				assert.Error(err)
				assert.Contains(err.Error(), "failed to parse package.json")
			})
		})

		Context("Deno Dependencies Hashing", func() {
			It("should compute hash for deno.json imports", func() {
				tempDir := GinkgoT().TempDir()
				denoJSON := `{
					"imports": {
						"lodash": "https://deno.land/x/lodash@4.17.21/mod.ts",
						"react": "https://esm.sh/react@18.2.0"
					}
				}`
				
				denoJSONPath := filepath.Join(tempDir, "deno.json")
				err := os.WriteFile(denoJSONPath, []byte(denoJSON), 0644)
				assert.NoError(err)
				
				hash, err := deps.ComputeDenoImportsHash(tempDir)
				assert.NoError(err)
				assert.NotEmpty(hash)
				assert.Len(hash, 64) // SHA256 produces 64 hex characters
			})

			It("should handle deno.jsonc files with comments", func() {
				tempDir := GinkgoT().TempDir()
				denoJSONC := `{
					/* Import map for dependencies */
					"imports": {
						"lodash": "npm:lodash@4.17.21",
						/* React for UI components */
						"react": "npm:react@18.2.0",
					}
				}`
				
				denoJSONCPath := filepath.Join(tempDir, "deno.jsonc")
				err := os.WriteFile(denoJSONCPath, []byte(denoJSONC), 0644)
				assert.NoError(err)
				
				hash, err := deps.ComputeDenoImportsHash(tempDir)
				assert.NoError(err)
				assert.NotEmpty(hash)
				assert.Len(hash, 64)
			})

			It("should prefer deno.json over deno.jsonc when both exist", func() {
				tempDir := GinkgoT().TempDir()
				denoJSON := `{
					"imports": {
						"lodash": "https://deno.land/x/lodash@4.17.21/mod.ts"
					}
				}`
				
				denoJSONC := `{
					"imports": {
						"react": "https://esm.sh/react@18.2.0"
					}
				}`
				
				// Create both files
				err := os.WriteFile(filepath.Join(tempDir, "deno.json"), []byte(denoJSON), 0644)
				assert.NoError(err)
				err = os.WriteFile(filepath.Join(tempDir, "deno.jsonc"), []byte(denoJSONC), 0644)
				assert.NoError(err)
				
				// Should use deno.json (which has lodash), not deno.jsonc (which has react)
				hash, err := deps.ComputeDenoImportsHash(tempDir)
				assert.NoError(err)
				assert.NotEmpty(hash)
				
				// Create another temp dir with only the deno.json content to compare
				tempDir2 := GinkgoT().TempDir()
				err = os.WriteFile(filepath.Join(tempDir2, "deno.json"), []byte(denoJSON), 0644)
				assert.NoError(err)
				
				expectedHash, err := deps.ComputeDenoImportsHash(tempDir2)
				assert.NoError(err)
				assert.Equal(expectedHash, hash, "Should prefer deno.json over deno.jsonc")
			})

			It("should return error when neither deno.json nor deno.jsonc exist", func() {
				tempDir := GinkgoT().TempDir()
				_, err := deps.ComputeDenoImportsHash(tempDir)
				assert.Error(err)
				assert.Contains(err.Error(), "failed to find deno.json or deno.jsonc")
			})

			It("should produce consistent hash regardless of import order", func() {
				tempDir1 := GinkgoT().TempDir()
				tempDir2 := GinkgoT().TempDir()
				
				denoJSON1 := `{
					"imports": {
						"react": "https://esm.sh/react@18.2.0",
						"lodash": "https://deno.land/x/lodash@4.17.21/mod.ts"
					}
				}`

				denoJSON2 := `{
					"imports": {
						"lodash": "https://deno.land/x/lodash@4.17.21/mod.ts",
						"react": "https://esm.sh/react@18.2.0"
					}
				}`

				err := os.WriteFile(filepath.Join(tempDir1, "deno.json"), []byte(denoJSON1), 0644)
				assert.NoError(err)
				err = os.WriteFile(filepath.Join(tempDir2, "deno.json"), []byte(denoJSON2), 0644)
				assert.NoError(err)
				
				hash1, err1 := deps.ComputeDenoImportsHash(tempDir1)
				hash2, err2 := deps.ComputeDenoImportsHash(tempDir2)
				
				assert.NoError(err1)
				assert.NoError(err2)
				assert.Equal(hash1, hash2, "Hashes should be identical regardless of import order")
			})
		})
	})

	Context("Hash Storage", func() {
		Context("Reading stored hash", func() {
			It("should return empty string when hash file doesn't exist", func() {
				tempDir := GinkgoT().TempDir()
				hash, err := deps.ReadStoredDepsHash(tempDir)
				assert.NoError(err)
				assert.Empty(hash)
			})

			It("should read hash from existing file in node_modules", func() {
				tempDir := GinkgoT().TempDir()
				expectedHash := "abc123def456"
				hashContent := expectedHash + "\n"
				
				// Create node_modules directory and hash file
				nodeModulesPath := filepath.Join(tempDir, "node_modules")
				err := os.MkdirAll(nodeModulesPath, 0755)
				assert.NoError(err)
				
				hashFilePath := filepath.Join(nodeModulesPath, deps.DepsHashFile)
				err = os.WriteFile(hashFilePath, []byte(hashContent), 0644)
				assert.NoError(err)
				
				hash, err := deps.ReadStoredDepsHash(tempDir)
				assert.NoError(err)
				assert.Equal(expectedHash, hash)
			})

			It("should trim whitespace from stored hash", func() {
				tempDir := GinkgoT().TempDir()
				hashWithWhitespace := "  abc123def456  \n\t  "
				
				// Create node_modules directory and hash file
				nodeModulesPath := filepath.Join(tempDir, "node_modules")
				err := os.MkdirAll(nodeModulesPath, 0755)
				assert.NoError(err)
				
				hashFilePath := filepath.Join(nodeModulesPath, deps.DepsHashFile)
				err = os.WriteFile(hashFilePath, []byte(hashWithWhitespace), 0644)
				assert.NoError(err)
				
				hash, err := deps.ReadStoredDepsHash(tempDir)
				assert.NoError(err)
				assert.Equal("abc123def456", hash)
			})
		})

		Context("Writing stored hash", func() {
			It("should write hash with trailing newline and read it back correctly", func() {
				tempDir := GinkgoT().TempDir()
				testHash := "abc123def456"
				
				// Write the hash
				err := deps.WriteStoredDepsHash(tempDir, testHash)
				assert.NoError(err)
				
				// Read it back
				hash, err := deps.ReadStoredDepsHash(tempDir)
				assert.NoError(err)
				assert.Equal(testHash, hash)
				
				// Verify the file was created with proper content in node_modules
				nodeModulesPath := filepath.Join(tempDir, "node_modules")
				hashFilePath := filepath.Join(nodeModulesPath, deps.DepsHashFile)
				content, err := os.ReadFile(hashFilePath)
				assert.NoError(err)
				assert.Equal(testHash+"\n", string(content))
			})

			It("should handle write errors gracefully", func() {
				// Try to write to a path that can't be created
				testHash := "abc123def456"
				
				// Use a path with invalid characters to force an error
				// On Windows, these characters are invalid: < > : " | ? * 
				invalidPath := "invalid<>path"
				err := deps.WriteStoredDepsHash(invalidPath, testHash)
				assert.Error(err)
			})
		})

		It("should verify DepsHashFile constant", func() {
			assert.Equal(".jpd-deps-hash", deps.DepsHashFile)
		})
	})

	Context("Integration Tests", func() {
		It("should handle complete Node.js dependency workflow", func() {
			tempDir := GinkgoT().TempDir()
			packageJSON := `{
				"dependencies": {
					"react": "^18.2.0",
					"lodash": "^4.17.21"
				},
				"devDependencies": {
					"typescript": "^5.0.0"
				}
			}`
			
			// Create package.json
			packageJSONPath := filepath.Join(tempDir, "package.json")
			err := os.WriteFile(packageJSONPath, []byte(packageJSON), 0644)
			assert.NoError(err)
			
			// Compute and store hash
			hash1, err := deps.ComputeNodeDepsHash(tempDir)
			assert.NoError(err)
			assert.NotEmpty(hash1)
			
			// Store the hash
			err = deps.WriteStoredDepsHash(tempDir, hash1)
			assert.NoError(err)
			
			// Read it back and verify
			storedHash, err := deps.ReadStoredDepsHash(tempDir)
			assert.NoError(err)
			assert.Equal(hash1, storedHash)
			
			// Compute hash again, should be identical
			hash2, err := deps.ComputeNodeDepsHash(tempDir)
			assert.NoError(err)
			assert.Equal(hash1, hash2, "Hash should be consistent")
		})

		It("should handle complete Deno import workflow", func() {
			tempDir := GinkgoT().TempDir()
			denoJSON := `{
				"imports": {
					"lodash": "https://deno.land/x/lodash@4.17.21/mod.ts",
					"react": "https://esm.sh/react@18.2.0"
				}
			}`
			
			// Create deno.json
			denoJSONPath := filepath.Join(tempDir, "deno.json")
			err := os.WriteFile(denoJSONPath, []byte(denoJSON), 0644)
			assert.NoError(err)
			
			// Compute and store hash
			hash1, err := deps.ComputeDenoImportsHash(tempDir)
			assert.NoError(err)
			assert.NotEmpty(hash1)
			
			// Store the hash
			err = deps.WriteStoredDepsHash(tempDir, hash1)
			assert.NoError(err)
			
			// Read it back and verify
			storedHash, err := deps.ReadStoredDepsHash(tempDir)
			assert.NoError(err)
			assert.Equal(hash1, storedHash)
			
			// Compute hash again, should be identical
			hash2, err := deps.ComputeDenoImportsHash(tempDir)
			assert.NoError(err)
			assert.Equal(hash1, hash2, "Hash should be consistent")
		})
	})
})

func TestDeps(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deps Suite")
}