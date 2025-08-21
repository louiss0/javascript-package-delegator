package integrations

import (
	"path/filepath"
	"testing"
)

func TestResolveCarapaceSpecsDirFor(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		getenv   func(string) string
		home     func() (string, error)
		expected string
		hasError bool
	}{
		{
			name: "Linux with XDG_DATA_HOME set",
			goos: "linux",
			getenv: func(key string) string {
				if key == "XDG_DATA_HOME" {
					return "/custom/xdg/data"
				}
				return ""
			},
			home:     func() (string, error) { return "/home/user", nil },
			expected: filepath.Join("/custom/xdg/data", "carapace", "specs"),
		},
		{
			name: "Linux without XDG_DATA_HOME (fallback)",
			goos: "linux",
			getenv: func(key string) string {
				return "" // No XDG_DATA_HOME set
			},
			home:     func() (string, error) { return "/home/user", nil },
			expected: filepath.Join("/home/user", ".local", "share", "carapace", "specs"),
		},
		{
			name: "macOS fallback to local share",
			goos: "darwin",
			getenv: func(key string) string {
				return ""
			},
			home:     func() (string, error) { return "/Users/user", nil },
			expected: filepath.Join("/Users/user", ".local", "share", "carapace", "specs"),
		},
		{
			name: "Windows with APPDATA",
			goos: "windows",
			getenv: func(key string) string {
				if key == "APPDATA" {
					return "C:\\Users\\User\\AppData\\Roaming"
				}
				return ""
			},
			home:     func() (string, error) { return "C:\\Users\\User", nil },
			expected: filepath.Join("C:\\Users\\User\\AppData\\Roaming", "carapace", "specs"),
		},
		{
			name: "Windows without APPDATA (fallback)",
			goos: "windows",
			getenv: func(key string) string {
				return ""
			},
			home:     func() (string, error) { return "C:\\Users\\User", nil },
			expected: filepath.Join("C:\\Users\\User", "AppData", "Roaming", "carapace", "specs"),
		},
		{
			name: "Home directory error",
			goos: "linux",
			getenv: func(key string) string {
				return ""
			},
			home:     func() (string, error) { return "", &testError{"home directory not found"} },
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveCarapaceSpecsDirFor(tt.goos, tt.getenv, tt.home)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// testError is a simple error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

func TestCarapaceSpecFileName(t *testing.T) {
	expected := "javascript-package-delegator.yaml"
	if CarapaceSpecFileName != expected {
		t.Errorf("Expected CarapaceSpecFileName to be %q, got %q", expected, CarapaceSpecFileName)
	}
}
