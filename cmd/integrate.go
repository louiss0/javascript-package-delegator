package cmd

import (
	// standard library
	"fmt"

	// external
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/louiss0/javascript-package-delegator/custom_flags"
	integrations "github.com/louiss0/javascript-package-delegator/internal"
)

// NewIntegrateCmd creates the integrate command with subcommands for external integrations
func NewIntegrateCmd() *cobra.Command {

	integrateCmd := &cobra.Command{
		Use:   "integrate",
		Short: "Generate integration files for external tools",
		Long: `Generate integration files for external tools like Warp terminal.

Available integrations:
  warp  - Generate Warp terminal workflow files

Examples:
	jpd integrate warp                           # Install workflows to default directory
	jpd integrate warp --output-dir ./workflows  # Install workflows to custom directory
`,
		DisableFlagsInUseLine: true,
	}
	integrateCmd.AddCommand(NewIntegrateWarpCmd())

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
