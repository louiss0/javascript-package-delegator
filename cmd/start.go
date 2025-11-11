package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/env"
	"github.com/louiss0/javascript-package-delegator/internal/deps"
)

// NewStartCmd wires up the dedicated `start` command that handles dev/start scripts
// plus dependency installation preflight for local development.
func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [args...]",
		Short: "Run your dev/start script with automatic dependency checks",
		Long: `Start runs the dev/start script (or a user-specified script) after ensuring
dependencies are installed for the detected package manager. It replaces the implicit
auto-install behavior that previously lived in the run command.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, _ := cmd.Flags().GetString(AGENT_FLAG)
			if pm == "" {
				return fmt.Errorf("no package manager detected for start command")
			}

			scriptFlag, err := cmd.Flags().GetString("script")
			if err != nil {
				return fmt.Errorf("failed to parse --script flag: %w", err)
			}
			noVoltaFlag, err := cmd.Flags().GetBool("no-volta")
			if err != nil {
				return fmt.Errorf("failed to parse --no-volta flag: %w", err)
			}

			targetDir, err := cmd.Flags().GetString(_CWD_FLAG)
			if err != nil {
				return fmt.Errorf("failed to parse --%s flag: %w", _CWD_FLAG, err)
			}
			if targetDir == "" {
				targetDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to determine working directory: %w", err)
				}
			}

			scriptName, err := resolveStartScript(pm, targetDir, scriptFlag)
			if err != nil {
				return err
			}

			cmdRunner := getCommandRunnerFromCommandContext(cmd)
			goEnv := getGoEnvFromCommandContext(cmd)
			de := getDebugExecutorFromCommandContext(cmd)

			if err := autoInstallDependenciesIfNeeded(pm, scriptName, targetDir, cmdRunner, goEnv, de, noVoltaFlag); err != nil {
				return err
			}

			type commandArgsResult struct {
				args []string
				err  error
			}

			withScriptArgs := func(base []string, includeDoubleDash bool) commandArgsResult {
				full := append([]string{}, base...)
				if len(args) > 0 {
					if includeDoubleDash {
						full = append(full, "--")
					}
					full = append(full, args...)
				}

				return commandArgsResult{args: full}
			}

			cmdResult := lo.Switch[string, commandArgsResult](pm).
				Case("npm", withScriptArgs([]string{"run", scriptName}, true)).
				Case("pnpm", withScriptArgs([]string{"run", scriptName}, true)).
				Case("yarn", withScriptArgs([]string{"run", scriptName}, false)).
				Case("bun", withScriptArgs([]string{"run", scriptName}, false)).
				CaseF("deno", func() commandArgsResult {
					if lo.Contains(args, "--eval") {
						return commandArgsResult{err: fmt.Errorf("don't pass --eval here use the exec command instead")}
					}

					base := []string{"task", scriptName}
					return commandArgsResult{args: append(base, args...)}
				}).
				Default(commandArgsResult{err: fmt.Errorf("start command does not support package manager %q", pm)})

			if cmdResult.err != nil {
				return cmdResult.err
			}

			cmdArgs := cmdResult.args

			cmdRunner.Command(pm, cmdArgs...)
			de.LogJSCommandIfDebugIsTrue(pm, cmdArgs...)

			goEnv.ExecuteIfModeIsProduction(func() {
				log.Info("Running command", "pm", pm, "args", strings.Join(cmdArgs, " "))
			})

			return cmdRunner.Run()
		},
	}

	cmd.Flags().String("script", "", "Script name to run (overrides the automatic dev/start detection)")
	cmd.Flags().Bool("no-volta", false, "Disable Volta integration during auto-install")

	return cmd
}

func resolveStartScript(pm string, targetDir string, override string) (string, error) {
	switch pm {
	case "deno":
		pkg, err := readDenoJSONFrom(targetDir)
		if err != nil {
			return "", err
		}
		return selectScriptCandidate(pkg.Tasks, override, "deno.json tasks")
	default:
		pkg, err := readPackageJSONAndUnmarshalScriptsFrom(targetDir)
		if err != nil {
			return "", err
		}
		return selectScriptCandidate(pkg.Scripts, override, "package.json scripts")
	}
}

func selectScriptCandidate(scripts map[string]string, override string, label string) (string, error) {
	if len(scripts) == 0 {
		return "", fmt.Errorf("no scripts found in %s", label)
	}

	if override != "" {
		if _, ok := scripts[override]; !ok {
			return "", fmt.Errorf("script %q not found in %s", override, label)
		}
		return override, nil
	}

	keys := lo.Keys(scripts)
	sort.Strings(keys)

	for _, candidate := range []string{"dev", "start"} {
		if _, ok := scripts[candidate]; ok {
			return candidate, nil
		}
	}

	if script := findScriptByKeyword(keys, "dev"); script != "" {
		return script, nil
	}
	if script := findScriptByKeyword(keys, "start"); script != "" {
		return script, nil
	}

	return "", fmt.Errorf("no script containing 'dev' or 'start' found in %s (use --script)", label)
}

func findScriptByKeyword(keys []string, keyword string) string {
	match, found := lo.Find(keys, func(name string) bool {
		return strings.Contains(strings.ToLower(name), keyword)
	})
	if !found {
		return ""
	}

	return match
}

// autoInstallDependenciesIfNeeded performs the dependency install preflight that backs the
// start command. The helper stays reusable so future commands can opt-in to the same behavior.
func autoInstallDependenciesIfNeeded(
	pm string,
	scriptName string,
	baseDir string,
	cmdRunner CommandRunner,
	goEnv env.GoEnv,
	de DebugExecutor,
	noVoltaFlag bool,
) error {
	shouldInstall := false
	var installReason strings.Builder

	if goEnv.IsDevelopmentMode() {
		de.LogDebugMessageIfDebugIsTrue("Auto-install check", "script", scriptName, "pm", pm, "enabled", true)
	}

	if pm != "deno" {
		isYarnPnp := pm == "yarn" && IsYarnPnpProject(baseDir)

		if goEnv.IsDevelopmentMode() {
			de.LogDebugMessageIfDebugIsTrue("Node PM check", "yarn_pnp", isYarnPnp)
		}

		if !isYarnPnp {
			nmPath := filepath.Join(baseDir, "node_modules")
			info, err := os.Stat(nmPath)
			missingNodeModules := err != nil || !info.IsDir()

			if missingNodeModules {
				shouldInstall = true
				installReason.WriteString("missing node_modules; ")
			}

			if goEnv.IsDevelopmentMode() {
				de.LogDebugMessageIfDebugIsTrue("Node modules check", "missing", missingNodeModules)
			}
		}

		depsWithVersions, err := deps.ExtractProdAndDevDependenciesFromPackageJSON()
		if err == nil && len(depsWithVersions) > 0 {
			names := ParsePackageNames(depsWithVersions)

			if !isYarnPnp {
				missing := MissingNodePackages(baseDir, names)
				if len(missing) > 0 {
					shouldInstall = true
					installReason.WriteString(fmt.Sprintf("%d missing packages; ", len(missing)))

					if goEnv.IsDevelopmentMode() {
						firstFew := lo.Slice(missing, 0, lo.Min(len(missing), 3))
						de.LogDebugMessageIfDebugIsTrue("Missing packages", "count", len(missing), "examples", firstFew)
					}
				}
			}
		}

		currentHash, err := deps.ComputeNodeDepsHash(baseDir)
		if err == nil {
			storedHash, err := deps.ReadStoredDepsHash(baseDir)
			if err == nil {
				hashMismatch := storedHash == "" || currentHash != storedHash
				if hashMismatch {
					shouldInstall = true
					installReason.WriteString(lo.Ternary(storedHash == "", "no stored hash; ", "dependencies changed; "))
				}

				if goEnv.IsDevelopmentMode() {
					currentShort := lo.TernaryF(len(currentHash) >= 8, func() string { return currentHash[:8] }, func() string { return "" })
					storedShort := lo.TernaryF(len(storedHash) >= 8, func() string { return storedHash[:8] }, func() string { return "" })
					de.LogDebugMessageIfDebugIsTrue(
						"Hash comparison",
						"current", currentShort,
						"stored", storedShort,
						"mismatch", hashMismatch,
					)
				}
			}
		}
	} else {
		importValues, err := deps.ExtractImportsFromDenoJSON(baseDir)
		if err == nil && len(importValues) > 0 {
			const maxImportChecks = 5
			checkLimit := lo.Min(len(importValues), maxImportChecks)
			checksToRun := lo.Slice(importValues, 0, checkLimit)

			missingImports := lo.Reduce(checksToRun, func(acc int, importURL string, _ int) int {
				de.LogJSCommandIfDebugIsTrue("deno", "info", "--json", importURL)
				infoCmd := cmdRunner
				infoCmd.Command("deno", "info", "--json", importURL)
				if err := infoCmd.Run(); err != nil {
					return acc + 1
				}

				return acc
			}, 0)

			if missingImports > 0 {
				shouldInstall = true
				installReason.WriteString(fmt.Sprintf("%d unresolvable imports; ", missingImports))

				if goEnv.IsDevelopmentMode() {
					de.LogDebugMessageIfDebugIsTrue("Import check", "checked", len(checksToRun), "missing", missingImports)
				}
			}
		}

		currentHash, err := deps.ComputeDenoImportsHash(baseDir)
		if err == nil {
			storedHash, err := deps.ReadStoredDenoDepsHash(baseDir)
			if err == nil {
				hashMismatch := storedHash == "" || currentHash != storedHash
				if hashMismatch {
					shouldInstall = true
					installReason.WriteString(lo.Ternary(storedHash == "", "no stored hash; ", "imports changed; "))
				}

				if goEnv.IsDevelopmentMode() {
					currentShort := lo.TernaryF(len(currentHash) >= 8, func() string { return currentHash[:8] }, func() string { return "" })
					storedShort := lo.TernaryF(len(storedHash) >= 8, func() string { return storedHash[:8] }, func() string { return "" })
					de.LogDebugMessageIfDebugIsTrue(
						"Deno hash comparison",
						"current", currentShort,
						"stored", storedShort,
						"mismatch", hashMismatch,
					)
				}
			}
		}
	}

	if !shouldInstall {
		return nil
	}

	reason := strings.TrimSuffix(installReason.String(), "; ")
	goEnv.ExecuteIfModeIsProduction(func() {
		log.Info("Auto-installing dependencies", "reason", reason, "pm", pm, "dir", baseDir)
	})

	if pm != "deno" {
		useVolta := detect.DetectVolta(detect.RealPathLookup{}) &&
			lo.Contains([]string{"npm", "pnpm", "yarn"}, pm) &&
			!noVoltaFlag

		var name string
		var args []string
		if useVolta {
			name = detect.VOLTA_RUN_COMMAND[0]
			args = append(detect.VOLTA_RUN_COMMAND[1:], pm, "install")
		} else {
			name = pm
			args = []string{"install"}
		}

		de.LogJSCommandIfDebugIsTrue(name, args...)
		cmdRunner.Command(name, args...)
		if err := cmdRunner.Run(); err != nil {
			return fmt.Errorf("failed to install dependencies: %w", err)
		}

		if newHash, err := deps.ComputeNodeDepsHash(baseDir); err == nil {
			if err := deps.WriteStoredDepsHash(baseDir, newHash); err == nil {
				if goEnv.IsDevelopmentMode() {
					hashShort := lo.TernaryF(len(newHash) >= 8, func() string { return newHash[:8] }, func() string { return "" })
					de.LogDebugMessageIfDebugIsTrue("Updated dependency hash", "hash", hashShort)
				}
			}
		}
	} else {
		denoConfigPath, err := deps.DenoConfigPath(baseDir)
		if err != nil {
			return fmt.Errorf("failed to locate Deno config for caching: %w", err)
		}

		de.LogJSCommandIfDebugIsTrue("deno", "cache", denoConfigPath)
		cmdRunner.Command("deno", "cache", denoConfigPath)
		if err := cmdRunner.Run(); err != nil {
			if goEnv.IsDevelopmentMode() {
				de.LogDebugMessageIfDebugIsTrue("Deno cache failed, continuing", "error", err)
			}
		} else {
			if newHash, err := deps.ComputeDenoImportsHash(baseDir); err == nil {
				if err := deps.WriteStoredDenoDepsHash(baseDir, newHash); err == nil {
					if goEnv.IsDevelopmentMode() {
						hashShort := lo.TernaryF(len(newHash) >= 8, func() string { return newHash[:8] }, func() string { return "" })
						de.LogDebugMessageIfDebugIsTrue("Updated Deno imports hash", "hash", hashShort)
					}
				}
			}
		}
	}

	return nil
}
