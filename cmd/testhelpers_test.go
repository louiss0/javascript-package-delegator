package cmd

import (
	"io"
	"os"
	"os/exec"
	"testing"
)

// captureStdout captures stdout during function execution
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	
	// Save original stdout
	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()
	
	// Create pipe
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	
	// Replace stdout
	os.Stdout = w
	
	// Channel to capture output
	outputChan := make(chan string)
	go func() {
		defer close(outputChan)
		output, err := io.ReadAll(r)
		if err != nil {
			t.Errorf("Failed to read from pipe: %v", err)
			return
		}
		outputChan <- string(output)
	}()
	
	// Execute function
	fn()
	
	// Close writer and restore stdout
	w.Close()
	os.Stdout = originalStdout
	
	// Get captured output
	return <-outputChan
}

// withEnv temporarily sets an environment variable
func withEnv(t *testing.T, key, value string, fn func()) {
	t.Helper()
	
	original := os.Getenv(key)
	originalSet := os.Getenv(key) != "" || os.Getenv(key) == ""
	
	err := os.Setenv(key, value)
	if err != nil {
		t.Fatalf("Failed to set environment variable %s: %v", key, err)
	}
	
	defer func() {
		if originalSet && original != "" {
			os.Setenv(key, original)
		} else {
			os.Unsetenv(key)
		}
	}()
	
	fn()
}

// makeTempDir creates a temporary directory for the test
func makeTempDir(t *testing.T, fn func(string)) {
	t.Helper()
	
	tempDir := t.TempDir()
	fn(tempDir)
}

// execAvailable checks if a binary is available for execution
func execAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
