// Package custom_flags provides custom flag types for command-line argument parsing.
// It implements various flag types that can be used with the cobra CLI framework.
package custom_flags

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/louiss0/javascript-package-delegator/custom_errors"
	"github.com/samber/lo"
)

// filePathFlag represents a flag that must contain a valid POSIX/UNIX file path
type filePathFlag struct {
	value    string
	flagName string
}

// NewFilePathFlag creates a new FilePathFlag with the given flag name
func NewFilePathFlag(flagName string) filePathFlag {
	return filePathFlag{
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
		return fmt.Errorf("The %s flag cannot be empty or contain only whitespace", p.flagName)
	}

	// Regex for general POSIX/UNIX file paths (relative or absolute)
	// Allows:
	// - Optional leading slash (for absolute paths)
	// - Segments consisting of alphanumeric, underscore, hyphen, or dot, or '.'/'..'
	// - Segments separated by a single slash
	// - Ends with a filename (alphanumeric, underscore, hyphen, or dot - not '.' or '..')
	// - Does NOT allow consecutive slashes (//)
	// - Does NOT allow trailing slash
	posixUnixFilePathRegex := `^(?:/?(?:[a-zA-Z0-9._-]+|\.{1,2})(?:/(?:[a-zA-Z0-9._-]+|\.{1,2}))*)?/?([a-zA-Z0-9._-]+)$`

	match, err := regexp.MatchString(posixUnixFilePathRegex, value)
	if err != nil {
		// This error indicates a problem with the regex itself, not the input value.
		return fmt.Errorf("internal error: failed to compile file path regex: %w", err)
	}

	if !match {
		return fmt.Errorf("The %s flag value '%s' is not a valid POSIX/UNIX file path", p.flagName, value)
	}

	p.value = value
	return nil
}

// Type returns the flag type as a string
func (p filePathFlag) Type() string {
	return "string"
}

// folderPathFlag represents a flag that must contain a valid POSIX/UNIX path
type folderPathFlag struct {
	value    string
	flagName string
}

// NewFolderPathFlag creates a new PathFlag with the given flag name
func NewFolderPathFlag(flagName string) folderPathFlag {
	return folderPathFlag{
		flagName: flagName,
	}
}

// String returns the flag's value as a string
func (p folderPathFlag) String() string {
	return p.value
}

// Set validates and sets the flag's value, checking for valid path format
func (p *folderPathFlag) Set(value string) error {
	// First, check if the value is empty or just whitespace
	if len(value) == 0 || regexp.MustCompile(`^\s+$`).MatchString(value) {
		return fmt.Errorf("The %s flag cannot be empty or contain only whitespace", p.flagName)
	}

	// Regex for general POSIX/UNIX paths (relative or absolute)
	// Allows:
	// - Optional leading slash (for absolute paths)
	// - Segments consisting of alphanumeric, underscore, hyphen, or dot
	// - Segments can be '.' or '..'
	// - Segments separated by a single slash
	// - Does NOT allow consecutive slashes (//)
	// - Does NOT allow trailing slash unless it's just "/" (handled by the regex structure)
	posixUnixFolderPathRegex := `^(?:/?(?:[a-zA-Z0-9._-]+|\.{1,2})(?:/(?:[a-zA-Z0-9._-]+|\.{1,2}))*/|\/)$`

	match, err := regexp.MatchString(posixUnixFolderPathRegex, value)
	if err != nil {
		// This error indicates a problem with the regex itself, not the input value.
		return fmt.Errorf("internal error: failed to compile path regex: %w", err)
	}

	if !match {
		return fmt.Errorf("The %s flag value '%s' is not a valid POSIX/UNIX folder path (must end with '/' unless it's just '/')", p.flagName, value)
	}

	p.value = value
	return nil
}

// Type returns the flag type as a string
func (p folderPathFlag) Type() string {
	return "string"
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

	match, error := regexp.MatchString(`^\s+$`, value)

	if error != nil {
		return error
	}

	if match {
		return fmt.Errorf(
			"The %s is empty",
			t.flagName,
		)
	}

	t.value = value
	return nil
}

// Type returns the flag type as a string
func (t emptyStringFlag) Type() string {
	return "string"
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

	match, error := regexp.MatchString(`^\S+$`, value)

	if error != nil {
		return error
	}

	if match && !lo.Contains([]string{"true", "false"}, value) {
		return fmt.Errorf(
			"%sflag must be one of: %v",
			custom_errors.FlagName(c.flagName),
			[]string{"true", "false"},
		)
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
func (self unionFlag) String() string {
	return self.value
}

// Set validates and sets the flag's value, ensuring it's one of the allowed values
func (self *unionFlag) Set(value string) error {

	match, error := regexp.MatchString(`^\S+$`, value)

	if error != nil {
		return error
	}

	if match && !lo.Contains(self.allowedValues, value) {
		return fmt.Errorf(
			"%sflag must be one of: %v",
			custom_errors.FlagName(self.flagName),
			self.allowedValues,
		)

	}
	self.value = value
	return nil
}

// Type returns the flag type as a string
func (self unionFlag) Type() string {
	return "string"
}

// RangeFlag represents a flag that must be an integer within a specified range
type RangeFlag struct {
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

	if min > max {
		panic("min must be less than max")
	}

	if min < 0 || max < 0 {
		panic("min and max must be non-negative")
	}

	return RangeFlag{
		min:      min,
		max:      max,
		flagName: flagName,
	}
}

// String returns the flag's value as a string
func (self RangeFlag) String() string {
	return fmt.Sprintf("%d", self.value)
}

// Value returns the flag's value as an int
func (self RangeFlag) Value() int {
	return self.value
}

// Set validates and sets the flag's value, ensuring it's within the allowed range
func (self *RangeFlag) Set(value string) error {

	match, error := regexp.MatchString(`^\d+$`, value)

	if error != nil {
		return error
	}

	if match {
		num, _ := strconv.Atoi(value)
		if num < self.min || num > self.max {
			return fmt.Errorf(
				"%sflag must be between %d and %d",
				custom_errors.FlagName(self.flagName),
				self.min,
				self.max,
			)
		}
		self.value = num
		return nil
	}

	return fmt.Errorf(
		"%sflag must be an integer between %d and %d",
		custom_errors.FlagName(self.flagName),
		self.min,
		self.max,
	)
}

// Type returns the flag type as a string
func (self RangeFlag) Type() string {
	return "string"
}
