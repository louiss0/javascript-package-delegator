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

// Package cmd provides command-line interface implementations for the JavaScript package delegator.
package cmd

import (
	// standard library
	"fmt"
	"strings"

	// external

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	// internal
	"github.com/louiss0/javascript-package-delegator/custom_errors"
	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/services"
)

// BuildCreateCommand builds command line for running package create commands
func BuildCreateCommand(pm, yarnVersion, name string, args []string) (program string, argv []string, err error) {
	// Special handling for deno - requires URL as name
	if pm == "deno" {
		if name == "" {
			return "", nil, fmt.Errorf("deno create requires a URL as the first argument")
		}
		if !isURL(name) {
			return "", nil, fmt.Errorf("deno create requires a valid URL, got: %s", name)
		}
		return "deno", append([]string{"run", name}, args...), nil
	}

	// For all other package managers, require name and reject URLs
	if name == "" {
		return "", nil, fmt.Errorf("package name is required for create command")
	}
	if isURL(name) {
		return "", nil, fmt.Errorf("URLs are not supported for %s, use deno instead", pm)
	}

	// Determine the actual binary to execute (create-<name> for unscoped, leave scoped and already-prefixed names as-is)
	bin := name
	// identify scoped package like @scope/name (optionally with @version)
	isScoped := strings.HasPrefix(name, "@") && strings.Contains(name, "/")
	// strip version/dist tag to check base name prefixing rules
	base := name
	if at := strings.Index(base, "@"); at > 0 { // ignore leading '@' for scoped names
		base = base[:at]
	}
	if !isScoped {
		// only prefix for unscoped names when they don't already start with create-
		if !strings.HasPrefix(base, "create-") {
			bin = "create-" + name
		}
	}

	// Build command based on package manager
	switch pm {
	case "npm":
		// npm exec create-<name> -- <args>
		argv = append([]string{"exec", bin, "--"}, args...)
		return "npm", argv, nil
	case "pnpm":
		// Prefer pnpm dlx for create runners: pnpm dlx <bin> <args>
		argv = append([]string{"dlx", bin}, args...)
		return "pnpm", argv, nil
	case "yarn":
		yarnMajor := ParseYarnMajor(yarnVersion)
		if yarnMajor <= 1 {
			// Yarn v1 -> use npx: npx <bin> <args>
			argv = append([]string{bin}, args...)
			return "npx", argv, nil
		} else {
			// Yarn v2+ -> use dlx: yarn dlx <bin> <args>
			argv = append([]string{"dlx", bin}, args...)
			return "yarn", argv, nil
		}
	case "bun":
		// bunx <bin> <args>
		argv = append([]string{bin}, args...)
		return "bunx", argv, nil
	default:
		return "", nil, fmt.Errorf("unsupported package manager: %s", pm)
	}
}

// CreateAppSelector provides an interface for selecting a create app package.
// It follows Go Writing Philosophy: defined at point of use, with clean methods.
type CreateAppSelector interface {
	Run() error
	Value() string
}

// CreateAppSearcher is a minimal interface to search for create-* packages.
// Defined at point of consumption for testability.
type CreateAppSearcher interface {
	SearchCreateApps(query string, size int) ([]services.PackageInfo, error)
}

// createAppSelector is a private struct implementing CreateAppSelector.
// All fields are unexported to comply with Go Writing Philosophy.
type createAppSelector struct {
	sel   *huh.Select[string]
	value string
}

// NewCreateAppSelector creates a new CreateAppSelector with pre-fetched packages and title.
// Constructor returns the interface (struct stays private) following Go Writing Philosophy.
func NewCreateAppSelector(packageInfo []services.PackageInfo) CreateAppSelector {

	// Map []services.PackageInfo to []huh.Option[string] using lo.Map
	opts := lo.Map(packageInfo, func(p services.PackageInfo, _ int) huh.Option[string] {
		// Use Name for both label and value; add description for better UX
		return huh.NewOption(p.Name, fmt.Sprintf("%s %s", p.Name, p.Description))
	})

	// Build the huh.Select with Title and Options
	sel := huh.NewSelect[string]().
		Title("Select app package").
		Options(opts...)

	// Return createAppSelector as CreateAppSelector interface
	return &createAppSelector{sel: sel}
}

// Run executes the interactive UI and stores the selected value.
// Uses pointer receiver since Run mutates internal state via s.value.
func (s *createAppSelector) Run() error {
	// Bind the pointer - huh.Select.Value takes *string
	s.sel.Value(&s.value)
	return s.sel.Run()
}

// Value returns the selected package name.
// No Get prefix - follows Go naming conventions.
func (s createAppSelector) Value() string {
	return s.value
}

// NewCreateCmd creates a new Cobra command for the "create" functionality
func NewCreateCmd(searcher CreateAppSearcher, newCreateAppSelector func([]services.PackageInfo) CreateAppSelector) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name|url] [args...]",
		Short: "Scaffold a new project using create runners",
		Long: `Scaffold a new project using the appropriate package manager's create command.
This command delegates to the package manager's create functionality to bootstrap new projects.

Package Manager Behavior:
- npm: Runs 'npm exec create-<name> -- <args>'
- pnpm: Runs 'pnpm exec create-<name> <args>'
- yarn v1: Runs 'npx create-<name> <args>'
- yarn v2+: Runs 'yarn dlx create-<name> <args>'
- bun: Runs 'bunx create-<name> <args>'
- deno: Runs 'deno run <url> <args>' (expects URL as first argument)

JPD flags (for this command):
  --search, -s    Search npm for popular "create-*" packages and select interactively
  --size <n>      Number of results to show when using --search (default: 25)

Passing flags to scaffolding tools:
- npm: JPD automatically inserts the -- separator before the app name so flags go to the scaffolder.
       Do not add another -- yourself; if you do, JPD will normalize it.
- pnpm / yarn v2+ / bun: pass flags directly after your app name.
- yarn v1: uses npx under the hood; pass flags directly after your app name.
- deno: pass arguments directly after the URL.

Examples:
  jpd create react-app my-app
  jpd create vite@latest my-app -- --template react-swc
  jpd create next-app myapp --typescript --tailwind
  jpd -a deno create https://deno.land/x/fresh/init.ts my-fresh-app`,
		Aliases: []string{"c"},
		// Allow passing through unknown flags (e.g., flags intended for the underlying create tools)
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		DisableFlagParsing: true,
		Args:               cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Manually parse flags since we disabled flag parsing
			search := false
			size := 0
			createAppQuery := ""
			packageArgs := []string{}

			// Parse arguments manually to separate flags from create arguments
			for i := 0; i < len(args); i++ {
				arg := args[i]
				switch {
				case arg == "--search" || arg == "-s":
					search = true
				case arg == "--size":
					if i+1 < len(args) {
						i++
						if s, err := fmt.Sscanf(args[i], "%d", &size); err != nil || s != 1 {
							return fmt.Errorf("invalid size value: %s", args[i])
						}
					} else {
						return fmt.Errorf("--size requires a value")
					}
				case arg == "-h" || arg == "--help":
					return cmd.Help()
				// Skip global flags - they're handled by the root command
				case arg == "-a" || arg == "--agent":
					if arg == "-a" || arg == "--agent" {
						i++ // skip the value
					}
				case arg == "-C" || arg == "--cwd":
					if arg == "-C" || arg == "--cwd" {
						i++ // skip the value
					}
				case arg == "-d" || arg == "--debug":
					// skip debug flag
				case strings.HasPrefix(arg, "-a=") || strings.HasPrefix(arg, "--agent="):
					// skip combined flag=value
				case strings.HasPrefix(arg, "-C=") || strings.HasPrefix(arg, "--cwd="):
					// skip combined flag=value
				default:
					// This is either the package name or arguments to pass through
					if createAppQuery == "" {
						createAppQuery = arg
					} else {
						packageArgs = append(packageArgs, arg)
					}
				}
			}

			// Validate arguments
			if !search && createAppQuery == "" {
				return fmt.Errorf("requires at least 1 arg(s), only received 0")
			}

			pm, _ := cmd.Flags().GetString(AGENT_FLAG)
			goEnv := getGoEnvFromCommandContext(cmd)
			cmdRunner := getCommandRunnerFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Info("Using package manager", "pm", pm)
			})

			// Variables are already parsed above

			if search {

				goEnv.ExecuteIfModeIsProduction(func() {
					log.Info(
						"Searching for create app packages based on your search",
						"search",
						createAppQuery,
					)
				})
				packageInfo, err := searcher.SearchCreateApps(createAppQuery, size)

				if err != nil {
					return err
				}

				if len(packageInfo) == 0 {
					return custom_errors.CreateInvalidArgumentErrorWithMessage(
						fmt.Sprintf("no packages found matching: %s", createAppQuery),
					)
				}

				// Use new CreateAppSelector implementation
				selector := newCreateAppSelector(packageInfo)

				if err := selector.Run(); err != nil {
					return err
				}

				chosen := selector.Value()
				createAppQuery = strings.Split(chosen, " ")[0]

			}

			if createAppQuery == "" {

				goEnv.ExecuteIfModeIsProduction(func() {
					log.Info(
						"You have not provided a package to create from, searching for popular create app packages instead",
					)
				})

				packageInfo, err := searcher.SearchCreateApps(createAppQuery, size)

				if err != nil {
					return err
				}

				if len(packageInfo) == 0 {
					return custom_errors.CreateInvalidArgumentErrorWithMessage(
						fmt.Sprintf("no packages found matching: %s", createAppQuery),
					)
				}

				// Use new CreateAppSelector implementation
				selector := newCreateAppSelector(packageInfo)

				if err := selector.Run(); err != nil {
					return err
				}

				chosen := selector.Value()
				createAppQuery = strings.Split(chosen, " ")[0]
			}

			// Get yarn version if needed
			yarnVersion := ""
			if pm == "yarn" {
				if version, err := detect.DetectYarnVersion(
					getYarnVersionRunnerCommandContext(cmd),
				); err == nil {
					yarnVersion = version
				}
			}

			// Normalize extra "--" for npm (users sometimes add it themselves).
			// npm mapping already injects one "--" internally, so remove any user-provided separators.
			if pm == "npm" && len(packageArgs) > 0 {
				filtered := make([]string, 0, len(packageArgs))
				for _, a := range packageArgs {
					if a == "--" {
						continue // drop
					}
					filtered = append(filtered, a)
				}
				packageArgs = filtered
			}

			// Build command for creating projects
			execCommand, cmdArgs, err := BuildCreateCommand(pm, yarnVersion, createAppQuery, packageArgs)
			if err != nil {
				return err
			}

			// Execute the command
			de.LogJSCommandIfDebugIsTrue(execCommand, cmdArgs...)
			cmdRunner.Command(execCommand, cmdArgs...)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Info("Running command", "cmd", execCommand, "args", strings.Join(cmdArgs, " "))
			})

			return cmdRunner.Run()
		},
	}

	// Add help-visible flags (parsing remains manual to allow passthrough)
	cmd.Flags().BoolP("search", "s", false, "Search npm for create packages (interactive)")
	cmd.Flags().Int("size", 25, "Number of results to show with --search")

	return cmd
}
