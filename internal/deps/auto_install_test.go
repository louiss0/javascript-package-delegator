package deps_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/louiss0/javascript-package-delegator/internal/deps"
	"github.com/stretchr/testify/assert"
)

// TestAutoInstallLogicIntegration verifies that the auto-install detection works correctly
// after WriteStoredDepsHash no longer creates node_modules directory
func TestAutoInstallLogicIntegration(t *testing.T) {
	t.Run("should detect missing node_modules and require installation", func(t *testing.T) {
		// Create temporary directory
		tempDir := t.TempDir()
		
		// Create package.json with dependencies
		packageJSON := `{
			"dependencies": {
				"react": "^18.2.0",
				"lodash": "^4.17.21"
			},
			"devDependencies": {
				"typescript": "^5.0.0"
			}
		}`
		
		packageJSONPath := filepath.Join(tempDir, "package.json")
		err := os.WriteFile(packageJSONPath, []byte(packageJSON), 0644)
		assert.NoError(t, err)
		
		// Verify node_modules doesn't exist (simulating fresh project state)
		nodeModulesPath := filepath.Join(tempDir, "node_modules")
		_, err = os.Stat(nodeModulesPath)
		assert.True(t, os.IsNotExist(err), "node_modules should not exist initially")
		
		// Compute hash (this should work regardless of node_modules existence)
		currentHash, err := deps.ComputeNodeDepsHash(tempDir)
		assert.NoError(t, err)
		assert.NotEmpty(t, currentHash)
		
		// Try to read stored hash (should return empty string, no error)
		storedHash, err := deps.ReadStoredDepsHash(tempDir)
		assert.NoError(t, err)
		assert.Empty(t, storedHash, "stored hash should be empty when node_modules doesn't exist")
		
		// Verify that trying to write hash fails when node_modules doesn't exist
		err = deps.WriteStoredDepsHash(tempDir, currentHash)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "node_modules directory does not exist")
		
		// This simulates the auto-install logic:
		// 1. node_modules missing -> shouldInstall = true
		// 2. Install runs (creating node_modules) 
		// 3. WriteStoredDepsHash succeeds after installation
		
		// Simulate installation creating node_modules
		err = os.MkdirAll(nodeModulesPath, 0755)
		assert.NoError(t, err)
		
		// Now WriteStoredDepsHash should succeed
		err = deps.WriteStoredDepsHash(tempDir, currentHash)
		assert.NoError(t, err)
		
		// Verify hash was stored correctly
		retrievedHash, err := deps.ReadStoredDepsHash(tempDir)
		assert.NoError(t, err)
		assert.Equal(t, currentHash, retrievedHash)
	})
	
	t.Run("should handle scenario where node_modules exists but hash is missing", func(t *testing.T) {
		// Create temporary directory
		tempDir := t.TempDir()
		
		// Create package.json
		packageJSON := `{
			"dependencies": {
				"express": "^4.18.0"
			}
		}`
		
		packageJSONPath := filepath.Join(tempDir, "package.json")
		err := os.WriteFile(packageJSONPath, []byte(packageJSON), 0644)
		assert.NoError(t, err)
		
		// Create node_modules directory (simulating after installation)
		nodeModulesPath := filepath.Join(tempDir, "node_modules")
		err = os.MkdirAll(nodeModulesPath, 0755)
		assert.NoError(t, err)
		
		// Stored hash should be empty (no hash file yet)
		storedHash, err := deps.ReadStoredDepsHash(tempDir)
		assert.NoError(t, err)
		assert.Empty(t, storedHash)
		
		// Compute and store hash
		currentHash, err := deps.ComputeNodeDepsHash(tempDir)
		assert.NoError(t, err)
		
		// Writing hash should succeed since node_modules exists
		err = deps.WriteStoredDepsHash(tempDir, currentHash)
		assert.NoError(t, err)
		
		// Verify it was stored
		retrievedHash, err := deps.ReadStoredDepsHash(tempDir)
		assert.NoError(t, err)
		assert.Equal(t, currentHash, retrievedHash)
	})
}