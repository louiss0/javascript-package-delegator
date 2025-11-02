package integrations

import (
	// standard library
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// external
	"gopkg.in/yaml.v3"
)

// WarpGenerator provides methods for generating Warp workflow YAML files.
type WarpGenerator interface {
	// GenerateJPDWorkflows writes individual .yaml workflow files to the specified directory.
	GenerateJPDWorkflows(outDir string) error

	// RenderJPDWorkflowsMultiDoc returns all workflows as a YAML multi-document string.
	RenderJPDWorkflowsMultiDoc() (string, error)
}

// warpGenerator is the concrete implementation of WarpGenerator.
type warpGenerator struct{}

// NewWarpGenerator creates a new Warp workflow generator instance.
func NewWarpGenerator() WarpGenerator {
	return &warpGenerator{}
}

// WarpWorkflow represents the structure of a Warp workflow YAML.
type WarpWorkflow struct {
	Name        string                 `yaml:"name"`
	Command     string                 `yaml:"command"`
	Tags        []string               `yaml:"tags"`
	Description string                 `yaml:"description"`
	Arguments   []WarpWorkflowArgument `yaml:"arguments,omitempty"`
	Author      string                 `yaml:"author"`
	AuthorURL   string                 `yaml:"author_url,omitempty"`
	Shells      []string               `yaml:"shells"`
}

// WarpWorkflowArgument represents an argument in a Warp workflow.
type WarpWorkflowArgument struct {
	Name         string      `yaml:"name"`
	Description  string      `yaml:"description"`
	DefaultValue interface{} `yaml:"default_value,omitempty"`
}

// workflowDefinition holds metadata for generating a workflow.
type workflowDefinition struct {
	Name        string
	Command     string
	Description string
	Arguments   []WarpWorkflowArgument
}

// getWorkflowDefinitions returns all JPD workflow definitions.
func (g *warpGenerator) getWorkflowDefinitions() []workflowDefinition {
	return []workflowDefinition{
		{
			Name:        "JPD Install",
			Command:     "jpd install {{package}}",
			Description: "Install packages using JPD (JavaScript Package Delegator)",
			Arguments: []WarpWorkflowArgument{
				{
					Name:         "package",
					Description:  "Package name to install",
					DefaultValue: nil, // Use nil for ~ in YAML
				},
			},
		},
		{
			Name:        "JPD Run",
			Command:     "jpd run {{script}}",
			Description: "Run package.json scripts using JPD",
			Arguments: []WarpWorkflowArgument{
				{
					Name:         "script",
					Description:  "Script name to run",
					DefaultValue: nil,
				},
			},
		},
		{
			Name:        "JPD Exec",
			Command:     "jpd exec {{package}} {{args}}",
			Description: "Execute packages using JPD",
			Arguments: []WarpWorkflowArgument{
				{
					Name:         "package",
					Description:  "Package name to execute",
					DefaultValue: nil,
				},
				{
					Name:         "args",
					Description:  "Additional arguments",
					DefaultValue: nil,
				},
			},
		},
		{
			Name:        "JPD DLX",
			Command:     "jpd dlx {{package}} {{args}}",
			Description: "Execute packages with package runner using JPD",
			Arguments: []WarpWorkflowArgument{
				{
					Name:         "package",
					Description:  "Package name to execute",
					DefaultValue: nil,
				},
				{
					Name:         "args",
					Description:  "Additional arguments",
					DefaultValue: nil,
				},
			},
		},
		{
			Name:        "JPD Update",
			Command:     "jpd update {{package}}",
			Description: "Update packages using JPD",
			Arguments: []WarpWorkflowArgument{
				{
					Name:         "package",
					Description:  "Package name to update",
					DefaultValue: nil,
				},
			},
		},
		{
			Name:        "JPD Uninstall",
			Command:     "jpd uninstall {{package}}",
			Description: "Uninstall packages using JPD",
			Arguments: []WarpWorkflowArgument{
				{
					Name:         "package",
					Description:  "Package name to uninstall",
					DefaultValue: nil,
				},
			},
		},
		{
			Name:        "JPD Clean Install",
			Command:     "jpd clean-install",
			Description: "Clean install with frozen lockfile using JPD",
			Arguments:   nil, // No arguments for clean-install
		},
		{
			Name:        "JPD Create",
			Command:     "jpd create {{name}} {{args}}",
			Description: "Scaffold new projects using JPD (supports package names and URLs for deno)",
			Arguments: []WarpWorkflowArgument{
				{
					Name:         "name",
					Description:  "Package name (e.g., react-app) or URL for deno",
					DefaultValue: nil,
				},
				{
					Name:         "args",
					Description:  "Project name and additional arguments",
					DefaultValue: nil,
				},
			},
		},
		{
			Name:        "JPD Agent",
			Command:     "jpd agent",
			Description: "Show detected package manager using JPD",
			Arguments:   nil, // No arguments for agent
		},
	}
}

// createWorkflow creates a WarpWorkflow from a workflow definition.
func (g *warpGenerator) createWorkflow(def workflowDefinition) WarpWorkflow {
	return WarpWorkflow{
		Name:        def.Name,
		Command:     def.Command,
		Tags:        []string{"jpd", "javascript", "package-manager"},
		Description: def.Description,
		Arguments:   def.Arguments,
		Author:      "the-code-fixer-23",
		AuthorURL:   "",         // Empty as requested
		Shells:      []string{}, // Empty array means all shells
	}
}

// GenerateJPDWorkflows writes individual .yaml workflow files to the specified directory.
func (g *warpGenerator) GenerateJPDWorkflows(outDir string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outDir, err)
	}

	definitions := g.getWorkflowDefinitions()
	filenameMap := map[string]string{
		"JPD Install":       "jpd-install.yaml",
		"JPD Run":           "jpd-run.yaml",
		"JPD Exec":          "jpd-exec.yaml",
		"JPD DLX":           "jpd-dlx.yaml",
		"JPD Create":        "jpd-create.yaml",
		"JPD Update":        "jpd-update.yaml",
		"JPD Uninstall":     "jpd-uninstall.yaml",
		"JPD Clean Install": "jpd-clean-install.yaml",
		"JPD Agent":         "jpd-agent.yaml",
	}

	for _, def := range definitions {
		workflow := g.createWorkflow(def)

		// Marshal to YAML
		yamlBytes, err := yaml.Marshal(&workflow)
		if err != nil {
			return fmt.Errorf("failed to marshal workflow %s to YAML: %w", def.Name, err)
		}

		// Write to file
		filename := filenameMap[def.Name]
		filepath := filepath.Join(outDir, filename)
		if err := os.WriteFile(filepath, yamlBytes, 0644); err != nil {
			return fmt.Errorf("failed to write workflow file %s: %w", filepath, err)
		}
	}

	return nil
}

// RenderJPDWorkflowsMultiDoc returns all workflows as a YAML multi-document string.
func (g *warpGenerator) RenderJPDWorkflowsMultiDoc() (string, error) {
	definitions := g.getWorkflowDefinitions()
	var parts []string

	for i, def := range definitions {
		workflow := g.createWorkflow(def)

		// Marshal to YAML
		yamlBytes, err := yaml.Marshal(&workflow)
		if err != nil {
			return "", fmt.Errorf("failed to marshal workflow %s to YAML: %w", def.Name, err)
		}

		// Add document separator for multi-doc format
		if i == 0 {
			parts = append(parts, "---")
		}
		parts = append(parts, string(yamlBytes))
		if i < len(definitions)-1 {
			parts = append(parts, "---")
		}
	}

	return strings.Join(parts, "\n"), nil
}

// DefaultWarpWorkflowsDir returns the default Warp workflows directory:
// ${XDG_DATA_HOME:-$HOME/.local/share}/warp-terminal/workflows
func DefaultWarpWorkflowsDir() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to resolve user home: %w", err)
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "warp-terminal", "workflows"), nil
}
