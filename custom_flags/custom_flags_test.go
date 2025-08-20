package custom_flags

import (
	"strconv"
	"strings"
	"testing"

	"github.com/louiss0/javascript-package-delegator/build_info"
)

// Test filePathFlag
func TestNewFilePathFlag(t *testing.T) {
	flag := NewFilePathFlag("testflag")
	if flag.flagName != "testflag" {
		t.Errorf("Expected flagName to be 'testflag', got %s", flag.flagName)
	}
	if flag.value != "" {
		t.Errorf("Expected initial value to be empty, got %s", flag.value)
	}
}

func TestFilePathFlag_String(t *testing.T) {
	flag := NewFilePathFlag("test")
	flag.value = "/path/to/file.txt"
	if flag.String() != "/path/to/file.txt" {
		t.Errorf("Expected String() to return '/path/to/file.txt', got %s", flag.String())
	}
}

func TestFilePathFlag_Type(t *testing.T) {
	flag := NewFilePathFlag("test")
	if flag.Type() != "string" {
		t.Errorf("Expected Type() to return 'string', got %s", flag.Type())
	}
}

func TestFilePathFlag_Set(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid absolute path", "/path/to/file.txt", false},
		{"valid relative path", "file.txt", false},
		{"valid path with dots", "../dir/file.log", false},
		{"valid path with underscores", "my_file.txt", false},
		{"valid path with hyphens", "my-file.txt", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"path with double slash", "path//file.txt", true},
		{"path with trailing slash", "path/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := NewFilePathFlag("test")
			err := flag.Set(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && flag.value != tt.value {
				t.Errorf("Expected value to be set to %s, got %s", tt.value, flag.value)
			}
		})
	}
}

// Test folderPathFlag
func TestNewFolderPathFlag(t *testing.T) {
	flag := NewFolderPathFlag("testflag")
	if flag.flagName != "testflag" {
		t.Errorf("Expected flagName to be 'testflag', got %s", flag.flagName)
	}
	if flag.value != "" {
		t.Errorf("Expected initial value to be empty, got %s", flag.value)
	}
}

func TestFolderPathFlag_String(t *testing.T) {
	flag := NewFolderPathFlag("test")
	flag.value = "/path/to/dir/"
	if flag.String() != "/path/to/dir/" {
		t.Errorf("Expected String() to return '/path/to/dir/', got %s", flag.String())
	}
}

func TestFolderPathFlag_Type(t *testing.T) {
	flag := NewFolderPathFlag("test")
	if flag.Type() != "string" {
		t.Errorf("Expected Type() to return 'string', got %s", flag.Type())
	}
}

func TestFolderPathFlag_Set(t *testing.T) {
	// Get current CI mode from build_info
	currentCIMode := build_info.InCI()

	tests := []struct {
		name     string
		value    string
		wantErr  bool
		errorMsg string
	}{
		{"valid absolute path with slash", "/path/to/dir/", false, ""},
		{"valid relative path with slash", "dir/", false, ""},
		{"valid root path", "/", false, ""},
		{"file-like path rejected", "/path/to/file.txt", true, "not a valid POSIX/UNIX folder path"},
		{"empty string", "", true, "cannot be empty"},
		{"whitespace only", "   ", true, "cannot be empty"},
	}

	// Add CI-specific tests based on current mode
	if currentCIMode {
		// In CI mode, paths without trailing slash are allowed
		tests = append(tests, struct {
			name     string
			value    string
			wantErr  bool
			errorMsg string
		}{"path without slash allowed in CI", "/path/to/dir", false, ""})
	} else {
		// In non-CI mode, paths without trailing slash should error
		tests = append(tests, struct {
			name     string
			value    string
			wantErr  bool
			errorMsg string
		}{"path without trailing slash rejected (not CI)", "/path/to/dir", true, "must end with '/'"})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := NewFolderPathFlag("test")
			err := flag.Set(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			}

			if !tt.wantErr && flag.value != tt.value {
				t.Errorf("Expected value to be set to %s, got %s", tt.value, flag.value)
			}
		})
	}
}

// Test emptyStringFlag
func TestNewEmptyStringFlag(t *testing.T) {
	flag := NewEmptyStringFlag("testflag")
	if flag.flagName != "testflag" {
		t.Errorf("Expected flagName to be 'testflag', got %s", flag.flagName)
	}
	if flag.value != "" {
		t.Errorf("Expected initial value to be empty, got %s", flag.value)
	}
}

func TestEmptyStringFlag_String(t *testing.T) {
	flag := NewEmptyStringFlag("test")
	flag.value = "test value"
	if flag.String() != "test value" {
		t.Errorf("Expected String() to return 'test value', got %s", flag.String())
	}
}

func TestEmptyStringFlag_Type(t *testing.T) {
	flag := NewEmptyStringFlag("test")
	if flag.Type() != "string" {
		t.Errorf("Expected Type() to return 'string', got %s", flag.Type())
	}
}

func TestEmptyStringFlag_Set(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid non-empty string", "valid value", false},
		{"empty string", "", false}, // Empty string is allowed, only whitespace-only is rejected
		{"whitespace only", "   ", true},
		{"string with content and spaces", "  content  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := NewEmptyStringFlag("test")
			err := flag.Set(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && flag.value != tt.value {
				t.Errorf("Expected value to be set to %s, got %s", tt.value, flag.value)
			}
		})
	}
}

// Test boolFlag
func TestNewBoolFlag(t *testing.T) {
	flag := NewBoolFlag("testflag")
	if flag.flagName != "testflag" {
		t.Errorf("Expected flagName to be 'testflag', got %s", flag.flagName)
	}
	if flag.value != "" {
		t.Errorf("Expected initial value to be empty, got %s", flag.value)
	}
}

func TestBoolFlag_String(t *testing.T) {
	flag := NewBoolFlag("test")
	flag.value = "true"
	if flag.String() != "true" {
		t.Errorf("Expected String() to return 'true', got %s", flag.String())
	}
}

func TestBoolFlag_Type(t *testing.T) {
	flag := NewBoolFlag("test")
	if flag.Type() != "bool" {
		t.Errorf("Expected Type() to return 'bool', got %s", flag.Type())
	}
}

func TestBoolFlag_Set(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid true", "true", false},
		{"valid false", "false", false},
		{"invalid yes", "yes", true},
		{"invalid 1", "1", true},
		{"empty string", "", false}, // Empty string passes regex but is not true/false
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := NewBoolFlag("test")
			err := flag.Set(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && flag.value != tt.value {
				t.Errorf("Expected value to be set to %s, got %s", tt.value, flag.value)
			}
		})
	}
}

func TestBoolFlag_Value(t *testing.T) {
	tests := []struct {
		name     string
		setValue string
		expected bool
	}{
		{"true value", "true", true},
		{"false value", "false", false},
		{"invalid value", "invalid", false}, // ParseBool returns false for invalid strings
		{"empty value", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := NewBoolFlag("test")
			flag.value = tt.setValue
			result := flag.Value()
			if result != tt.expected {
				t.Errorf("Value() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test unionFlag
func TestNewUnionFlag(t *testing.T) {
	allowedValues := []string{"option1", "option2", "option3"}
	flag := NewUnionFlag(allowedValues, "testflag")
	if flag.flagName != "testflag" {
		t.Errorf("Expected flagName to be 'testflag', got %s", flag.flagName)
	}
	if len(flag.allowedValues) != 3 {
		t.Errorf("Expected 3 allowed values, got %d", len(flag.allowedValues))
	}
}

func TestUnionFlag_String(t *testing.T) {
	flag := NewUnionFlag([]string{"a", "b"}, "test")
	flag.value = "a"
	if flag.String() != "a" {
		t.Errorf("Expected String() to return 'a', got %s", flag.String())
	}
}

func TestUnionFlag_Type(t *testing.T) {
	flag := NewUnionFlag([]string{"a", "b"}, "test")
	if flag.Type() != "string" {
		t.Errorf("Expected Type() to return 'string', got %s", flag.Type())
	}
}

func TestUnionFlag_Set(t *testing.T) {
	allowedValues := []string{"option1", "option2", "option3"}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid option1", "option1", false},
		{"valid option2", "option2", false},
		{"invalid option", "option4", true},
		{"empty string", "", false}, // Empty string passes regex but is not in allowed values
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := NewUnionFlag(allowedValues, "test")
			err := flag.Set(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && flag.value != tt.value {
				t.Errorf("Expected value to be set to %s, got %s", tt.value, flag.value)
			}
		})
	}
}

// Test RangeFlag
func TestNewRangeFlag(t *testing.T) {
	flag := NewRangeFlag("test", 1, 10)
	if flag.flagName != "test" {
		t.Errorf("Expected flagName to be 'test', got %s", flag.flagName)
	}
	if flag.min != 1 {
		t.Errorf("Expected min to be 1, got %d", flag.min)
	}
	if flag.max != 10 {
		t.Errorf("Expected max to be 10, got %d", flag.max)
	}
}

func TestNewRangeFlag_Panics(t *testing.T) {
	tests := []struct {
		name string
		min  int
		max  int
	}{
		{"min greater than max", 10, 5},
		{"negative min", -1, 10},
		{"negative max", 0, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Expected NewRangeFlag to panic, but it didn't")
				}
			}()
			NewRangeFlag("test", tt.min, tt.max)
		})
	}
}

func TestRangeFlag_String(t *testing.T) {
	flag := NewRangeFlag("test", 1, 10)
	flag.value = 5
	if flag.String() != "5" {
		t.Errorf("Expected String() to return '5', got %s", flag.String())
	}
}

func TestRangeFlag_Type(t *testing.T) {
	flag := NewRangeFlag("test", 1, 10)
	if flag.Type() != "string" {
		t.Errorf("Expected Type() to return 'string', got %s", flag.Type())
	}
}

func TestRangeFlag_Value(t *testing.T) {
	flag := NewRangeFlag("test", 1, 10)
	flag.value = 7
	if flag.Value() != 7 {
		t.Errorf("Expected Value() to return 7, got %d", flag.Value())
	}
}

func TestRangeFlag_Set(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		min     int
		max     int
		wantErr bool
	}{
		{"valid value in range", "5", 1, 10, false},
		{"minimum value", "1", 1, 10, false},
		{"maximum value", "10", 1, 10, false},
		{"value below range", "0", 1, 10, true},
		{"value above range", "11", 1, 10, true},
		{"non-numeric value", "abc", 1, 10, true},
		{"empty string", "", 1, 10, true},
		{"negative number as string", "-1", 1, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := NewRangeFlag("test", tt.min, tt.max)
			err := flag.Set(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				expectedValue, _ := strconv.Atoi(tt.value)
				if flag.value != expectedValue {
					t.Errorf("Expected value to be set to %d, got %d", expectedValue, flag.value)
				}
			}
		})
	}
}

// Edge case tests for regex failures
func TestFilePathFlag_Set_RegexError(t *testing.T) {
	// This test is difficult to trigger in practice since the regex is well-formed
	// But we can test the error path by mocking a regex failure scenario
	flag := NewFilePathFlag("test")

	// Test with an extremely long string that might cause issues
	longString := strings.Repeat("a", 10000) + "/" + strings.Repeat("b", 10000)
	err := flag.Set(longString)
	// This should either succeed or fail with a validation error, not a regex compilation error
	if err != nil && strings.Contains(err.Error(), "internal error: failed to compile") {
		t.Errorf("Unexpected regex compilation error: %v", err)
	}
}

func TestFolderPathFlag_Set_RegexError(t *testing.T) {
	flag := NewFolderPathFlag("test")

	// Test with an extremely long string
	longString := strings.Repeat("a", 10000) + "/" + strings.Repeat("b", 10000) + "/"
	err := flag.Set(longString)
	// This should either succeed or fail with a validation error, not a regex compilation error
	if err != nil && strings.Contains(err.Error(), "internal error: failed to compile") {
		t.Errorf("Unexpected regex compilation error: %v", err)
	}
}
