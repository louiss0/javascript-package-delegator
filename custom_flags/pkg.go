// Package custom_flags provides custom flag types for command-line argument parsing.
// It implements various flag types that can be used with the cobra CLI framework.
package custom_flags

import (
	"fmt"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/pflag"

	"github.com/louiss0/javascript-package-delegator/build_info"
	"github.com/louiss0/javascript-package-delegator/custom_errors"
)

// Cross-platform path validation utilities

// isWindows returns true if running on Windows
func isWindows() bool {
	return runtime.GOOS == "windows"
}

// validateFilePath validates a file path according to the current platform
func validateFilePath(value string) bool {
	if isWindows() {
		return validateWindowsFilePath(value)
	}
	return validatePosixFilePath(value)
}

// validateFolderPath validates a folder path according to the current platform
func validateFolderPath(value string) bool {
	if isWindows() {
		return validateWindowsFolderPath(value)
	}
	return validatePosixFolderPath(value)
}

// validateWindowsFilePath validates Windows file paths
func validateWindowsFilePath(value string) bool {
	// Windows file path regex - more permissive:
	// - Absolute: C:\path\to\file.ext or C:/path/to/file.ext
	// - Relative: path\to\file.ext, .\file.ext, ..\path\file.ext, file.ext
	// - UNC: \\server\share\file.ext
	// Accepts both forward and backward slashes, must end with filename (no trailing slash)
	windowsFilePathRegex := `^(?:(?:[a-zA-Z]:[/\\]|\\\\[^/\\:*?"<>|]+\\[^/\\:*?"<>|]+[/\\]|\.{1,2}[/\\])(?:[^/\\:*?"<>|]+[/\\])*|(?:[^/\\:*?"<>|]+[/\\])+)?[^/\\:*?"<>|]+$`
	match, _ := regexp.MatchString(windowsFilePathRegex, value)
	return match
}

// validateWindowsFolderPath validates Windows folder paths
func validateWindowsFolderPath(value string) bool {
	// Check if it looks like a file (has an extension)
	// On Windows, we also need to reject file-like paths
	trimmed := strings.TrimRight(value, "/\\")
	if trimmed != "" {
		// Get the last component of the path
		lastComponent := trimmed
		if lastSlash := strings.LastIndexAny(trimmed, "/\\"); lastSlash != -1 {
			lastComponent = trimmed[lastSlash+1:]
		}

		// Check if it has a file extension (and it's not . or ..)
		if strings.Contains(lastComponent, ".") && lastComponent != "." && lastComponent != ".." {
			// It looks like a file, reject it
			return false
		}
	}

	// Windows folder path regex - same pattern for now, CI distinction added at validation level
	// Accepts both forward and backward slashes, with or without trailing separator
	windowsFolderPathRegex := `^(?:[a-zA-Z]:[/\\]?|\\\\[^/\\:*?"<>|]+\\[^/\\:*?"<>|]+[/\\]?|\.{1,2}[/\\]?|[^/\\:*?"<>|]+)(?:[/\\][^/\\:*?"<>|]+)*[/\\]?$`

	match, _ := regexp.MatchString(windowsFolderPathRegex, value)
	if !match {
		return false
	}

	// In CI mode, accept all paths that match the basic regex
	if build_info.InCI() {
		return true
	}

	// In non-CI mode, apply stricter rules for Windows paths without trailing separators
	// Allow special cases: drive roots (C:), current dir (.), parent dir (..)
	if value == "." || value == ".." {
		return true
	}

	// Allow drive roots like "C:" or "D:"
	if matched, _ := regexp.MatchString(`^[a-zA-Z]:$`, value); matched {
		return true
	}

	// If path doesn't end with separator, it's invalid in strict mode
	// unless it's one of the special cases above
	if !strings.HasSuffix(value, "/") && !strings.HasSuffix(value, "\\") {
		return false
	}

	return match
}

// validatePosixFilePath validates POSIX/UNIX file paths
func validatePosixFilePath(value string) bool {
	// Regex for general POSIX/UNIX file paths (relative or absolute)
	posixUnixFilePathRegex := `^(?:/?(?:[a-zA-Z0-9._-]+|\.{1,2})(?:/(?:[a-zA-Z0-9._-]+|\.{1,2}))*)?/?([a-zA-Z0-9._-]+)$`
	match, _ := regexp.MatchString(posixUnixFilePathRegex, value)
	return match
}

// validatePosixFolderPath validates POSIX/UNIX folder paths
func validatePosixFolderPath(value string) bool {
	// Strict mode (default): requires a trailing slash unless it's just "/"
	posixUnixFolderPathStrict := `^(?:/?(?:[a-zA-Z0-9._-]+|\.{1,2})(?:/(?:[a-zA-Z0-9._-]+|\.{1,2}))*/|\/)$`
	// CI-relaxed mode: accepts with or without trailing slash
	posixUnixFolderPathRelaxed := `^(?:/?(?:[a-zA-Z0-9._-]+|\.{1,2})(?:/(?:[a-zA-Z0-9._-]+|\.{1,2}))*/?|\/)$`

	regexToUse := posixUnixFolderPathStrict
	if build_info.InCI() {
		regexToUse = posixUnixFolderPathRelaxed
	}

	match, _ := regexp.MatchString(regexToUse, value)
	return match
}

// Interfaces extending pflag.Value for testability

// FilePathFlag extends pflag.Value for file path flags
type FilePathFlag interface {
	pflag.Value
	FlagName() string
}

// FolderPathFlag extends pflag.Value for folder path flags
type FolderPathFlag interface {
	pflag.Value
	FlagName() string
}

// EmptyStringFlag extends pflag.Value for empty string flags
type EmptyStringFlag interface {
	pflag.Value
	FlagName() string
}

// BoolFlag extends pflag.Value for boolean flags
type BoolFlag interface {
	pflag.Value
	FlagName() string
	Value() bool
}

// UnionFlag extends pflag.Value for union flags
type UnionFlag interface {
	pflag.Value
	FlagName() string
	AllowedValues() []string
}

// RangeFlag extends pflag.Value for range flags
type RangeFlag interface {
	pflag.Value
	FlagName() string
	Value() int
	Min() int
	Max() int
}

// filePathFlag represents a flag that must contain a valid POSIX/UNIX file path
type filePathFlag struct {
	value    string
	flagName string
}

// NewFilePathFlag creates a new FilePathFlag with the given flag name
func NewFilePathFlag(flagName string) FilePathFlag {
	return &filePathFlag{
		flagName: flagName,
	}
}

// String returns the flag's value as a string
func (p filePathFlag) String() string {
	return p.value
}

// Set validates and sets the flag's value, checking for valid file path format
func (p *filePathFlag) Set(value string) error {
	// First, check if the value is empty or just whitespace
	if len(value) == 0 || regexp.MustCompile(`^\s+$`).MatchString(value) {
		return fmt.Errorf("the %s flag cannot be empty or contain only whitespace", p.flagName)
	}

	// Use cross-platform validation
	if !validateFilePath(value) {
		platform := "POSIX/UNIX"
		if isWindows() {
			platform = "Windows"
		}
		return fmt.Errorf("the %s flag value '%s' is not a valid %s file path", p.flagName, value, platform)
	}

	p.value = value
	return nil
}

// Type returns the flag type as a string
func (p filePathFlag) Type() string {
	return "string"
}

// FlagName returns the flag's name for testing
func (p filePathFlag) FlagName() string {
	return p.flagName
}

// folderPathFlag represents a flag that must contain a valid POSIX/UNIX path
type folderPathFlag struct {
	value    string
	flagName string
}

// NewFolderPathFlag creates a new PathFlag with the given flag name
func NewFolderPathFlag(flagName string) FolderPathFlag {
	return &folderPathFlag{
		flagName: flagName,
	}
}

// String returns the flag's value as a string
func (p folderPathFlag) String() string {
	return p.value
}

// Set validates and sets the flag's value, checking for valid path format
// File-like paths are always rejected regardless of CI mode; only trailing slash validation is relaxed in CI
func (p *folderPathFlag) Set(value string) error {
	// First, check if the value is empty or just whitespace
	if len(value) == 0 || regexp.MustCompile(`^\s+$`).MatchString(value) {
		return fmt.Errorf("the %s flag cannot be empty or contain only whitespace", p.flagName)
	}

	// Early rejection of file-like paths (only for POSIX systems)
	if !isWindows() {
		// Strip any trailing slash to get the final path segment
		trimmed := strings.TrimRight(value, "/")
		if trimmed != "" {
			base := path.Base(trimmed)
			ext := path.Ext(base)
			// If there's a file extension and it's not "." or "..", reject as file-like
			if ext != "" && base != "." && base != ".." {
				platform := "POSIX/UNIX"
				msg := "(must end with '/' unless it's just '/')"
				if build_info.InCI() {
					msg = ""
				}
				return fmt.Errorf("the %s flag value '%s' is not a valid %s folder path %s", p.flagName, value, platform, msg)
			}
		}
	}

	// Use cross-platform validation
	if !validateFolderPath(value) {
		platform := "POSIX/UNIX"
		msg := "(must end with '/' unless it's just '/')"
		if isWindows() {
			platform = "Windows"
			msg = ""
		} else if build_info.InCI() {
			msg = ""
		}
		return fmt.Errorf("the %s flag value '%s' is not a valid %s folder path %s", p.flagName, value, platform, msg)
	}

	p.value = value
	return nil
}

// Type returns the flag type as a string
func (p folderPathFlag) Type() string {
	return "string"
}

// FlagName returns the flag's name for testing
func (p folderPathFlag) FlagName() string {
	return p.flagName
}

// emptyStringFlag represents a flag that cannot be empty or contain only whitespace
type emptyStringFlag struct {
	value    string
	flagName string
}

// NewEmptyStringFlag creates a new emptyStringFlag with the given flag name
func NewEmptyStringFlag(flagName string) emptyStringFlag {
	return emptyStringFlag{
		flagName: flagName,
	}
}

// String returns the flag's value as a string
func (t emptyStringFlag) String() string {
	return t.value
}

// Set validates and sets the flag's value, checking for empty/whitespace
func (t *emptyStringFlag) Set(value string) error {
	match, err := regexp.MatchString(`^\s+$`, value)
	if err != nil {
		return err
	}
	if match {
		return fmt.Errorf("the %s flag is empty", t.flagName)
	}
	t.value = value
	return nil
}

// Type returns the flag type as a string
func (t emptyStringFlag) Type() string {
	return "string"
}

// FlagName returns the flag's name for testing
func (t emptyStringFlag) FlagName() string {
	return t.flagName
}

// boolFlag represents a flag that must be either "true" or "false"
type boolFlag struct {
	value    string
	flagName string
}

// NewBoolFlag creates a new boolFlag with the given flag name
func NewBoolFlag(flagName string) boolFlag {
	return boolFlag{
		flagName: flagName,
	}
}

// String returns the flag's value as a string
func (c boolFlag) String() string {
	return c.value
}

// Set validates and sets the flag's value, ensuring it's a valid boolean
func (c *boolFlag) Set(value string) error {
	match, err := regexp.MatchString(`^\S+$`, value)
	if err != nil {
		return err
	}
	if match && !lo.Contains([]string{"true", "false"}, value) {
		return fmt.Errorf("%s flag must be one of %v", custom_errors.FlagName(c.flagName), []string{"true", "false"})
	}
	c.value = value
	return nil
}

// Type returns the flag type as a string
func (c boolFlag) Type() string {
	return "bool"
}

// Value returns the flag's value as a bool
func (c boolFlag) Value() bool {
	value, _ := strconv.ParseBool(c.value)
	return value
}

// FlagName returns the flag's name for testing
func (c boolFlag) FlagName() string {
	return c.flagName
}

// unionFlag represents a flag that must be one of a predefined set of values
type unionFlag struct {
	value         string
	allowedValues []string
	flagName      string
}

// NewUnionFlag creates a new unionFlag with the given allowed values and flag name
func NewUnionFlag(allowedValues []string, flagName string) unionFlag {
	return unionFlag{
		allowedValues: allowedValues,
		flagName:      flagName,
	}
}

// String returns the flag's value as a string
func (u unionFlag) String() string {
	return u.value
}

// Set validates and sets the flag's value, ensuring it's one of the allowed values
func (u *unionFlag) Set(value string) error {
	match, err := regexp.MatchString(`^\S+$`, value)
	if err != nil {
		return err
	}
	if match && !lo.Contains(u.allowedValues, value) {
		return fmt.Errorf("%s flag must be one of %v", custom_errors.FlagName(u.flagName), u.allowedValues)
	}
	u.value = value
	return nil
}

// Type returns the flag type as a string
func (u unionFlag) Type() string {
	return "string"
}

// FlagName returns the flag's name for testing
func (u unionFlag) FlagName() string {
	return u.flagName
}

// AllowedValues returns the allowed values for testing
func (u unionFlag) AllowedValues() []string {
	return u.allowedValues
}

// rangeFlag represents a flag that must be an integer within a specified range
type rangeFlag struct {
	value, min, max int
	flagName        string
}

// NewRangeFlag creates a new RangeFlag with the given flag name and range bounds
func NewRangeFlag(flagName string, min, max int) RangeFlag {
	if min > max {
		panic("min must be less than max")
	}
	if min < 0 || max < 0 {
		panic("min and max must be non-negative")
	}
	return &rangeFlag{
		min:      min,
		max:      max,
		flagName: flagName,
	}
}

// String returns the flag's value as a string
func (r rangeFlag) String() string {
	return fmt.Sprintf("%d", r.value)
}

// Value returns the flag's value as an int
func (r rangeFlag) Value() int {
	return r.value
}

// Set validates and sets the flag's value, ensuring it's within the allowed range
func (r *rangeFlag) Set(value string) error {
	match, err := regexp.MatchString(`^\d+$`, value)
	if err != nil {
		return err
	}
	if match {
		num, _ := strconv.Atoi(value)
		if num < r.min || num > r.max {
			return fmt.Errorf("%s flag must be between %d and %d", custom_errors.FlagName(r.flagName), r.min, r.max)
		}
		r.value = num
		return nil
	}
	return fmt.Errorf("%s flag must be an integer between %d and %d", custom_errors.FlagName(r.flagName), r.min, r.max)
}

// Type returns the flag type as a string
func (r rangeFlag) Type() string {
	return "string"
}

// FlagName returns the flag's name for testing
func (r rangeFlag) FlagName() string {
	return r.flagName
}

// Min returns the minimum value for testing
func (r rangeFlag) Min() int {
	return r.min
}

// Max returns the maximum value for testing
func (r rangeFlag) Max() int {
	return r.max
}
