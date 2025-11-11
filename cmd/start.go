package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/louiss0/javascript-package-delegator/env"
	"github.com/louiss0/javascript-package-delegator/internal/deps"
)

// NewStartCmd creates a command dedicated to bootstrapping start scripts.
// It layers Deno cache warming on top of the standard run flow while delegating
// the actual script execution to the run command implementation.
func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [args...]",
		Short: "Run the start task with package manager specific preflight",
		Long: `Start projects with optional preflight. For Deno projects this warms the cache
based on deno.json imports before delegating to 'jpd run start'. Node-based projects are
forwarded directly to the run command with the provided flags.`,
		Aliases: []string{"s"},
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, _ := cmd.Flags().GetString(AGENT_FLAG)
			cmdRunner := getCommandRunnerFromCommandContext(cmd)
			goEnv := getGoEnvFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)

			baseDir, err := resolveBaseDir(cmd)
			if err != nil {
				return err
			}

			reloadCache, err := cmd.Flags().GetBool("reload-cache")
			if err != nil {
				return fmt.Errorf("failed to parse --reload-cache flag: %w", err)
			}

			if pm == "deno" {
				if err := warmDenoCache(cmdRunner, goEnv, de, baseDir, reloadCache); err != nil {
					return err
				}
			}

			// Delegate to the run command by re-invoking the root command with adjusted args.
			forwardArgs := buildRunArgsFromStart(cmd, args)
			root := cmd.Root()
			root.SetArgs(forwardArgs)
			return root.ExecuteContext(cmd.Context())
		},
	}

	cmd.Flags().Bool("auto-install", false, "Forward to run command to control auto-install behaviour")
	cmd.Flags().Bool("no-volta", false, "Disable Volta integration during auto-install (forwarded to run)")
	cmd.Flags().Bool("if-present", false, "Run script only if it exists (forwarded to run command)")
	cmd.Flags().Bool("reload-cache", false, "Force Deno cache reload before executing the start task")

	return cmd
}

func resolveBaseDir(cmd *cobra.Command) (string, error) {
	cwdFlagValue, err := cmd.Flags().GetString(_CWD_FLAG)
	if err != nil {
		return "", fmt.Errorf("failed to parse --%s flag: %w", _CWD_FLAG, err)
	}

	if cwdFlagValue != "" {
		return cwdFlagValue, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return cwd, nil
}

func warmDenoCache(cmdRunner CommandRunner, goEnv env.GoEnv, de DebugExecutor, baseDir string, reloadCache bool) error {
	currentHash, err := deps.ComputeDenoImportsHash(baseDir)
	if err != nil {
		if errors.Is(err, deps.ErrDenoConfigNotFound) {
			return fmt.Errorf("no Deno configuration found in %s: %w", baseDir, err)
		}
		return fmt.Errorf("failed to compute Deno imports hash: %w", err)
	}

	storedHash, err := deps.ReadStoredDenoDepsHash(baseDir)
	if err != nil {
		return fmt.Errorf("failed to read stored Deno hash: %w", err)
	}

	shouldCache := reloadCache || storedHash == "" || storedHash != currentHash
	forceReload := reloadCache || (storedHash != "" && storedHash != currentHash)

	if de != nil {
		shorten := func(hash string) string {
			if len(hash) >= 8 {
				return hash[:8]
			}
			return hash
		}

		de.LogDebugMessageIfDebugIsTrue(
			"Deno cache decision",
			"stored", shorten(storedHash),
			"current", shorten(currentHash),
			"should_cache", shouldCache,
			"force_reload", forceReload,
		)
	}

	if !shouldCache {
		return nil
	}

	cacheArgs := []string{"cache"}
	if forceReload {
		cacheArgs = append(cacheArgs, "--reload")
	}
	cacheArgs = append(cacheArgs, "deno.json")

	if de != nil {
		de.LogJSCommandIfDebugIsTrue("deno", cacheArgs...)
	}

	cmdRunner.Command("deno", cacheArgs...)
	if err := cmdRunner.Run(); err != nil {
		return fmt.Errorf("failed to warm Deno cache: %w", err)
	}

	if err := deps.WriteStoredDenoDepsHash(baseDir, currentHash); err != nil {
		if goEnv.IsDevelopmentMode() && de != nil {
			de.LogDebugMessageIfDebugIsTrue("Failed to persist Deno hash", "error", err)
		}
	} else {
		if goEnv.IsDevelopmentMode() && de != nil {
			de.LogDebugMessageIfDebugIsTrue("Updated Deno imports hash", "hash", currentHash)
		}
	}

	goEnv.ExecuteIfModeIsProduction(func() {
		log.Info("Warming Deno cache", "dir", baseDir, "reload", forceReload)
	})

	return nil
}

func buildRunArgsFromStart(cmd *cobra.Command, scriptArgs []string) []string {
	args := []string{}

	args = append(args, collectChangedPersistentFlags(cmd.Root().PersistentFlags())...)
	args = append(args, "run")
	args = append(args, collectForwardedStartFlags(cmd.Flags())...)
	args = append(args, "start")
	args = append(args, scriptArgs...)

	return args
}

func collectChangedPersistentFlags(flags *pflag.FlagSet) []string {
	results := []string{}
	if flags == nil {
		return results
	}

	flags.Visit(func(f *pflag.Flag) {
		results = append(results, formatFlagForArgs(f)...)
	})

	return results
}

func collectForwardedStartFlags(flags *pflag.FlagSet) []string {
	results := []string{}
	if flags == nil {
		return results
	}

	for _, name := range []string{"auto-install", "no-volta", "if-present"} {
		flag := flags.Lookup(name)
		if flag == nil || !flag.Changed {
			continue
		}
		results = append(results, formatFlagForArgs(flag)...)
	}

	return results
}

func formatFlagForArgs(flag *pflag.Flag) []string {
	if flag == nil {
		return nil
	}

	name := "--" + flag.Name
	value := flag.Value.String()

	switch flag.Value.Type() {
	case "bool":
		if value == "true" {
			return []string{name}
		}
		return []string{fmt.Sprintf("%s=%s", name, value)}
	default:
		if flag.NoOptDefVal != "" && value == flag.NoOptDefVal {
			return []string{name}
		}
		return []string{name, value}
	}
}
