package cmd

import (
	// standard library
	"fmt"
	"os"

	// external
	"github.com/spf13/cobra"

	// internal
	"github.com/louiss0/javascript-package-delegator/custom_errors"
	"github.com/louiss0/javascript-package-delegator/custom_flags"
	"github.com/louiss0/javascript-package-delegator/internal/integrations"
)

// NewIntegrateCmd creates the integrate command with subcommands for external integrations
func NewIntegrateCmd() *cobra.Command {
	outputDirFlag := custom_flags.NewFolderPathFlag("output-dir")
	outputFileFlag := custom_flags.NewFilePathFlag("output")

	integrateCmd := &cobra.Command{
		Use:   "integrate <target>",
		Short: "Generate integration files for external tools",
		Long: `Generate integration files for external tools like Warp terminal and Carapace.

Available integrations:
  warp      - Generate Warp terminal workflow files
  carapace  - Generate Carapace completion spec

Examples:
	jpd integrate warp --output-dir ./workflows/  # Generate Warp workflow files
	jpd integrate warp                           # Print Warp workflows as multi-doc YAML
	jpd integrate carapace                       # Install spec to global Carapace directory
	jpd integrate carapace --stdout              # Print spec to stdout
	jpd integrate carapace --output ./jpd.yaml   # Write spec to custom file

Carapace spec installation locations:
  Linux/macOS: ~/.local/share/carapace/specs/javascript-package-delegator.yaml
  Windows:     %APPDATA%\carapace\specs\javascript-package-delegator.yaml
  Custom:      Set XDG_DATA_HOME to override on Unix systems
`,
		DisableFlagsInUseLine: true,
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			switch target {
			case "warp":
				return runWarpIntegration(cmd)
			case "carapace":
				return runCarapaceIntegration(cmd)
			default:
				return custom_errors.CreateInvalidArgumentErrorWithMessage(
					fmt.Sprintf("unsupported integration target: '%s'. Supported targets are: warp, carapace", target))
			}
		},
	}

	// Add flags for both warp and carapace integrations
	integrateCmd.Flags().Var(&outputDirFlag, "output-dir", "Output directory for Warp workflow files")
	integrateCmd.Flags().VarP(&outputFileFlag, "output", "o", "Output file for Carapace spec")
	integrateCmd.Flags().Bool("stdout", false, "Print Carapace spec to stdout instead of installing")

	return integrateCmd
}

// NewIntegrateWarpCmd creates the warp integration subcommand
func NewIntegrateWarpCmd() *cobra.Command {
	outputDirFlag := custom_flags.NewFolderPathFlag("output-dir")

	warpCmd := &cobra.Command{
		Use:   "warp",
		Short: "Generate Warp terminal workflow files",
		Long: `Generate Warp terminal workflow files for JPD commands.

This generates individual .yaml workflow files for each JPD command (install, run, exec, dlx, 
update, uninstall, clean-install, agent) that can be used with Warp terminal.

Examples:
	jpd integrate warp --output-dir ./workflows/  # Generate individual workflow files
	jpd integrate warp                           # Print workflows as multi-doc YAML
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWarpIntegration(cmd)
		},
	}

	warpCmd.Flags().VarP(&outputDirFlag, "output-dir", "o", "Output directory for workflow files")

	return warpCmd
}

// NewIntegrateCarapaceCmd creates the carapace integration subcommand
func NewIntegrateCarapaceCmd() *cobra.Command {
	outputFileFlag := custom_flags.NewFilePathFlag("output")

	carapaceCmd := &cobra.Command{
		Use:   "carapace",
		Short: "Generate Carapace completion spec file",
		Long: `Generate a Carapace YAML specification file for JPD.

This generates a YAML spec file that can be used with carapace-bin for intelligent
completions across multiple shells. By default, the spec is installed to the global
Carapace specs directory as "javascript-package-delegator.yaml".

Examples:
	jpd integrate carapace                       # Install spec to global Carapace directory
	jpd integrate carapace --stdout              # Print spec to stdout
	jpd integrate carapace --output ./jpd.yaml   # Write spec to custom file

Global installation locations:
  Linux/macOS: ~/.local/share/carapace/specs/javascript-package-delegator.yaml
  Windows:     %APPDATA%\carapace\specs\javascript-package-delegator.yaml
  Custom:      Set XDG_DATA_HOME to override on Unix systems
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCarapaceIntegration(cmd)
		},
	}

	carapaceCmd.Flags().VarP(&outputFileFlag, "output", "o", "Output file for Carapace spec")
	carapaceCmd.Flags().Bool("stdout", false, "Print Carapace spec to stdout instead of installing")

	return carapaceCmd
}

func runWarpIntegration(cmd *cobra.Command) error {
	warpGenerator := integrations.NewWarpGenerator()

	outputDirFlag, err := cmd.Flags().GetString("output-dir")
	if err != nil {
		// Flag not set, output to stdout
		multiDoc, err := warpGenerator.RenderJPDWorkflowsMultiDoc()
		if err != nil {
			return fmt.Errorf("failed to generate Warp workflows: %w", err)
		}
		fmt.Print(multiDoc)
		return nil
	}

	if outputDirFlag == "" {
		// No output directory specified, output to stdout
		multiDoc, err := warpGenerator.RenderJPDWorkflowsMultiDoc()
		if err != nil {
			return fmt.Errorf("failed to generate Warp workflows: %w", err)
		}
		fmt.Print(multiDoc)
		return nil
	}

	// Generate individual workflow files in directory
	err = warpGenerator.GenerateJPDWorkflows(outputDirFlag)
	if err != nil {
		return fmt.Errorf("failed to generate Warp workflow files: %w", err)
	}

	fmt.Printf("Generated Warp workflow files in: %s\n", outputDirFlag)
	return nil
}

// runCarapaceIntegration handles Carapace integration with three behavioral modes:
// 1. --output flag: Write to specified file (preserves existing behavior)
// 2. --stdout flag: Print to stdout (preserves old default behavior for tooling/pipes)
// 3. Default: Install to global Carapace specs directory (new default behavior)
func runCarapaceIntegration(cmd *cobra.Command) error {
	carapaceGenerator := integrations.NewCarapaceSpecGenerator()

	// Generate the spec (needed for all modes)
	spec, err := carapaceGenerator.GenerateYAMLSpec(cmd.Root())
	if err != nil {
		return fmt.Errorf("failed to generate Carapace spec: %w", err)
	}

	// Check flag precedence: 1) --output, 2) --stdout, 3) default (global install)
	outputFileFlag, _ := cmd.Flags().GetString("output")
	if outputFileFlag != "" {
		// Mode 1: Write to custom file path
		err = writeToFile(outputFileFlag, spec)
		if err != nil {
			return fmt.Errorf("failed to write Carapace spec to file: %w", err)
		}
		fmt.Printf("Generated Carapace spec file: %s\n", outputFileFlag)
		return nil
	}

	stdoutFlag, _ := cmd.Flags().GetBool("stdout")
	if stdoutFlag {
		// Mode 2: Print to stdout (old default behavior)
		fmt.Print(spec)
		return nil
	}

	// Mode 3: Default - Install to global Carapace specs directory
	specPath, err := integrations.DefaultCarapaceSpecPath()
	if err != nil {
		return fmt.Errorf("failed to resolve Carapace spec path: %w", err)
	}

	// Ensure the parent directory exists
	specDir, err := integrations.CarapaceSpecsDir()
	if err != nil {
		return fmt.Errorf("failed to resolve Carapace specs directory: %w", err)
	}

	err = integrations.EnsureDir(specDir)
	if err != nil {
		return fmt.Errorf("failed to create Carapace specs directory: %w", err)
	}

	// Write the spec file to the global location
	err = writeToFile(specPath, spec)
	if err != nil {
		return fmt.Errorf("failed to write Carapace spec to global location: %w", err)
	}

	fmt.Printf("Installed Carapace spec: %s\n", specPath)
	return nil
}

// Helper function to write content to file
func writeToFile(filename, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			if err == nil {
				err = cerr
			} else {
				err = fmt.Errorf("%w; %w", err, cerr)
			}
		}
	}()

	_, err = file.WriteString(content)
	return err
}
