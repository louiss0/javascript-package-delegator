/*
Copyright © 2025 Shelton Louis

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"os"
	"os/exec"

	// "log/slog"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

func NewAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Show the detected package manager agent",
		Long: `Show information about the detected package manager agent.
Equivalent to 'na' command - detects and disp"Agentlays the package manager being used.

This command shows which package manager would be used based on lock files in the current directory.

Examples:
  jpd    # Show detected package manager`,
		Aliases: []string{"a"},
		RunE: func(cmd *cobra.Command, args []string) error {

			pm := getPackageNameFromCommandContext(cmd)

			goMode := getGoModeFromCommandContext(cmd)

			if goMode != "development" {
				log.Infof("Detected package manager, now executing command: %s\n", pm)

			}
			// Show detailed information

			commmand := exec.Command(pm, args...)
			commmand.Stderr = os.Stderr
			commmand.Stdin = os.Stdin
			commmand.Stdout = os.Stdout

			return commmand.Run()
		},
	}

	return cmd
}
