package integrations

import (
	// standard library
	_ "embed" // Required for the embed directive
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	// external
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	// CarapaceSpecFileName is the filename for the Carapace spec file
	CarapaceSpecFileName = "javascript-package-delegator.yaml"
)

//go:embed assets/jpd-extern.nu
var nushellCompletionScript string

// NushellCompletionScript returns the embedded Nushell completion script content.
func NushellCompletionScript() string {
	return nushellCompletionScript
}

// CarapaceSpecsDir returns the OS-appropriate directory path for Carapace specs.
// Uses XDG_DATA_HOME on Unix systems if set, otherwise falls back to standard locations:
// - Linux/macOS: ~/.local/share/carapace/specs
// - Windows: %APPDATA%\carapace\specs or %USERPROFILE%\AppData\Roaming\carapace\specs
func CarapaceSpecsDir() (string, error) {
	return resolveCarapaceSpecsDirFor(runtime.GOOS, os.Getenv, os.UserHomeDir)
}

// resolveCarapaceSpecsDirFor is a testable version of CarapaceSpecsDir with injected dependencies.
func resolveCarapaceSpecsDirFor(goos string, getenv func(string) string, home func() (string, error)) (string, error) {
	// Check for XDG_DATA_HOME first (Unix systems)
	if xdgDataHome := getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "carapace", "specs"), nil
	}

	// Get user home directory
	homeDir, err := home()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Platform-specific paths
	switch goos {
	case "windows":
		// Prefer APPDATA if available
		if appData := getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "carapace", "specs"), nil
		}
		// Fallback to standard Windows location
		return filepath.Join(homeDir, "AppData", "Roaming", "carapace", "specs"), nil
	default:
		// Unix-like systems (Linux, macOS, etc.)
		return filepath.Join(homeDir, ".local", "share", "carapace", "specs"), nil
	}
}

// DefaultCarapaceSpecPath returns the full path where the Carapace spec should be installed.
func DefaultCarapaceSpecPath() (string, error) {
	specsDir, err := CarapaceSpecsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(specsDir, CarapaceSpecFileName), nil
}

// EnsureDir creates the directory and all parent directories if they don't exist.
func EnsureDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

// CarapaceSpecGenerator provides methods for generating Carapace YAML specifications.
type CarapaceSpecGenerator interface {
	// GenerateYAMLSpec generates a Carapace YAML spec from the given cobra command.
	GenerateYAMLSpec(cmd *cobra.Command) (string, error)
}

// carapaceSpecGenerator is the concrete implementation of CarapaceSpecGenerator.
type carapaceSpecGenerator struct{}

// NewCarapaceSpecGenerator creates a new Carapace spec generator instance.
func NewCarapaceSpecGenerator() *carapaceSpecGenerator {
	return &carapaceSpecGenerator{}
}

// CarapaceSpec represents the structure of a Carapace YAML specification.
type CarapaceSpec struct {
	Name            string                 `yaml:"name"`
	Description     string                 `yaml:"description"`
	PersistentFlags map[string]FlagSpec    `yaml:"persistentFlags,omitempty"`
	Commands        map[string]CommandSpec `yaml:"commands,omitempty"`
}

// FlagSpec represents a flag specification in Carapace YAML.
type FlagSpec struct {
	Shorthand   string   `yaml:"shorthand,omitempty"`
	Description string   `yaml:"description"`
	Completion  string   `yaml:"completion,omitempty"`
	Enum        []string `yaml:"enum,omitempty"`
}

// CommandSpec represents a command specification in Carapace YAML.
type CommandSpec struct {
	Description string              `yaml:"description"`
	Flags       map[string]FlagSpec `yaml:"flags,omitempty"`
	Completion  string              `yaml:"completion,omitempty"`
}

// GenerateYAMLSpec generates a Carapace YAML spec from the given cobra command.
func (g *carapaceSpecGenerator) GenerateYAMLSpec(cmd *cobra.Command) (string, error) {
	spec := CarapaceSpec{
		Name:        "jpd",
		Description: "JavaScript Package Delegator - A universal package manager interface",
		PersistentFlags: map[string]FlagSpec{
			"agent": {
				Shorthand:   "a",
				Description: "Select the JS package manager you want to use",
				Enum:        []string{"npm", "yarn", "pnpm", "bun", "deno"},
			},
			"debug": {
				Shorthand:   "d",
				Description: "Make commands run in debug mode",
			},
			"cwd": {
				Shorthand:   "C",
				Description: "Set the working directory for commands",
				Completion:  "$carapace.directories",
			},
		},
		Commands: map[string]CommandSpec{
			"install": {
				Description: "Install packages (equivalent to 'ni')",
				Completion:  "$carapace.packages.npm",
			},
			"run": {
				Description: "Run package.json scripts (equivalent to 'nr')",
				Completion:  "$carapace.scripts.npm",
			},
			"exec": {
				Description: "Execute packages (equivalent to 'nlx')",
				Completion:  "$carapace.packages.npm",
			},
			"dlx": {
				Description: "Execute packages with package runner (dedicated package-runner command)",
				Completion:  "$carapace.packages.npm",
			},
			"update": {
				Description: "Update packages (equivalent to 'nup')",
				Completion:  "$carapace.packages.npm",
			},
			"uninstall": {
				Description: "Uninstall packages (equivalent to 'nun')",
				Completion:  "$carapace.packages.npm",
			},
			"clean-install": {
				Description: "Clean install with frozen lockfile (equivalent to 'nci')",
			},
			"agent": {
				Description: "Show detected package manager (equivalent to 'na')",
			},
			"completion": {
				Description: "Generate shell completion scripts",
				Flags: map[string]FlagSpec{
					"filename": {
						Shorthand:   "f",
						Description: "Output completion script to file",
						Completion:  "$carapace.files",
					},
					"with-shorthand": {
						Shorthand:   "s",
						Description: "Include shorthand alias functions",
					},
				},
				Completion: "$carapace.shells",
			},
		},
	}

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(&spec)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Carapace spec to YAML: %w", err)
	}

	// Add header comment
	header := "# Carapace completion spec for jpd\n# Generated by jpd integrate carapace\n---\n"
	return header + string(yamlBytes), nil
}
