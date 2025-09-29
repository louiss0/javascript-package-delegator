package cmd

import (
	// standard library
	"fmt"
	"os"

	// external
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/louiss0/javascript-package-delegator/custom_flags"
	integrations "github.com/louiss0/javascript-package-delegator/internal"
)

// NewIntegrateCmd creates the integrate command with subcommands for external integrations
func NewIntegrateCmd() *cobra.Command {

	outputFile := custom_flags.NewFilePathFlag("output")

	integrateCmd := &cobra.Command{
		Use:   "integrate",
		Short: "Generate integration files for external tools",
		Long: `Generate integration files for external tools like Warp terminal and Carapace.

Available integrations:
  warp      - Generate Warp terminal workflow files
  carapace  - Generate Carapace completion spec

Examples:
	jpd integrate warp                           # Install workflows to default directory
	jpd integrate warp --output-dir ./workflows  # Install workflows to custom directory
	jpd integrate carapace                       # Install spec to global Carapace directory
	jpd integrate carapace --stdout              # Print spec to stdout
	jpd integrate carapace --output ./jpd.yaml   # Write spec to custom file

Carapace spec installation locations:
  Linux/macOS: ~/.local/share/carapace/specs/javascript-package-delegator.yaml
  Windows:     %APPDATA%\carapace\specs\javascript-package-delegator.yaml
  Custom:      Set XDG_DATA_HOME to override on Unix systems
`,
		DisableFlagsInUseLine: true,
	}

	// Only add Carapace-specific flags to the root integrate command
	// The warp-specific output-dir flag is defined on the warp subcommand itself
	integrateCmd.Flags().VarP(outputFile, "output", "o", "Output file for Carapace spec")
	integrateCmd.Flags().Bool("stdout", false, "Print Carapace spec to stdout instead of installing")

	integrateCmd.AddCommand(NewIntegrateWarpCmd(), NewIntegrateCarapaceCmd())

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
create, update, uninstall, clean-install, agent) that can be used with Warp terminal.

By default, workflow files are installed to:
  ${XDG_DATA_HOME:-$HOME/.local/share}/warp-terminal/workflows

Examples:
	jpd integrate warp                           # Install to default directory
	jpd integrate warp --output-dir ./workflows  # Install to custom directory
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWarpIntegration(outputDirFlag.String(), cmd)
		},
	}

	// Add the --output-dir flag directly to the warp subcommand
	warpCmd.Flags().VarP(outputDirFlag, "output-dir", "o", "Output directory for workflow files")

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
			return runCarapaceIntegration(outputFileFlag.String(), cmd)
		},
	}

	carapaceCmd.Flags().VarP(outputFileFlag, "output", "o", "Output file for Carapace spec")
	carapaceCmd.Flags().Bool("stdout", false, "Print Carapace spec to stdout instead of installing")

	return carapaceCmd
}

func runWarpIntegration(output_dir string, cmd *cobra.Command) error {
	warpGenerator := integrations.NewWarpGenerator()
	goEnv := getGoEnvFromCommandContext(cmd)

	// Get output directory from flag (already validated by FolderPathFlag)
	var outDir string
	var err error

	if output_dir == "" {
		// No flag provided, use default directory
		outDir, err = integrations.DefaultWarpWorkflowsDir()
		if err != nil {
			return fmt.Errorf("failed to resolve default Warp workflows directory: %w", err)
		}
	} else {
		// Use the provided directory (already validated by FolderPathFlag)
		// Remove trailing slash for consistency
		outDir = output_dir
	}

	// Generate workflow files in the directory
	err = warpGenerator.GenerateJPDWorkflows(outDir)
	if err != nil {
		return fmt.Errorf("failed to generate Warp workflow files: %w", err)
	}

	goEnv.ExecuteIfModeIsProduction(func() {
		log.Info("Generated Warp workflow files", "directory", outDir)
	})
	return nil
}

// runCarapaceIntegration handles Carapace integration with three behavioral modes:
// 1. --output flag: Write to specified file (preserves existing behavior)
// 2. --stdout flag: Print to stdout (preserves old default behavior for tooling/pipes)
// 3. Default: Install to global Carapace specs directory (new default behavior)
func runCarapaceIntegration(outputFile string, cmd *cobra.Command) error {
	carapaceGenerator := integrations.NewCarapaceSpecGenerator()

	goEnv := getGoEnvFromCommandContext(cmd)

	// Generate the spec (needed for all modes)
	spec, err := carapaceGenerator.GenerateYAMLSpec(cmd.Root())
	if err != nil {
		return fmt.Errorf("failed to generate Carapace spec: %w", err)
	}

	// Check flag precedence: 1) --output, 2) --stdout, 3) default (global install)
	if outputFile != "" {
		// Mode 1: Write to custom file path
		err = writeToFile(outputFile, spec)
		if err != nil {
			return fmt.Errorf("failed to write Carapace spec to file: %w", err)
		}

		goEnv.ExecuteIfModeIsProduction(func() {

			log.Info("Generated Carapace spec file", "path", outputFile)
		})
		return nil
	}

	stdoutFlag, _ := cmd.Flags().GetBool("stdout")
	if stdoutFlag {
		// Mode 2: Print to stdout (old default behavior)
		cmd.Print(spec)
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

	goEnv.ExecuteIfModeIsProduction(func() {
		log.Info("Installed Carapace spec", "path", specPath)

	})
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
