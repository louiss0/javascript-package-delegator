# Go Coding Philosophy Violations Report

Generated on: 2025-08-16
Project: github.com/louiss0/javascript-package-delegator

This report lists instances where the current codebase deviates from your Go Coding Philosophy and related rules. Each section corresponds to a file and enumerates specific violations with line references, code snippets, and suggested corrections.

Notes
- Error strings should start with a lowercase letter and have no trailing punctuation.
- For interfaces and constructors, follow the Access Philosophy about pointer receivers and constructor return types.
- Import groups should be separated into: standard library, external, and internal packages (blank line between groups) and sorted by goimports.
- Package names should be snake_case per your rule. Multi-word packages without underscores are noted.

---

## main.go
- No specific violations found.

---

## go.mod
- No specific violations found (semantic rules apply to source files).

---

## env/pkg.go

1) Getter method naming (Go Function Naming – Getter Methods)
- Lines 16-18:
  code:
    func (e GoEnv) GetgoEnv() string {
        return e.goEnv
    }
  issue: Method uses a Get prefix and mixed-case name (GetgoEnv). Per rule, do not use Get for simple getters; prefer a property-style method name. Also casing should be PascalCase for exported methods.
  suggestion: Rename to Env() or GoEnv() if needed, or more explicit Mode(). Example:
    func (e GoEnv) Mode() string { return e.goEnv }

2) Import grouping (Go Import Grouping)
- Lines 3-5: Only an internal import exists; this is acceptable. No action required.

---

## build_info/pkg.go

1) Getter function naming vs. Go-specific getter method rule
- Lines 85-98:
  code:
    func GetVersion() string { return CLI_VERSION.String() }
    func GetGoMode() string { return GO_MODE.String() }
    func GetBuildDate() string { return BUILD_DATE.String() }
  issue: These are functions (not methods). Your general coding rules prefer function names to indicate returning a value (get...). Your Go-specific rule only bans Get for simple getters on methods. This is acceptable, but if you want consistency with Go idioms, consider Version(), GoMode(), BuildDate().
  suggestion: Optional refactor:
    func Version() string
    func GoMode() string
    func BuildDate() string

- No other clear violations.

---

## custom_errors/pkg.go

1) Exported error var naming (Go Error Handling – Exported Errors)
- Lines 11-14:
  code:
    var InvalidFlag = errors.New("Invalid Flag:")
    var InvalidArgument = errors.New("Invalid Argument:")
  issue: Exported errors should be named ErrXxx and error strings should start with lowercase and have no trailing punctuation.
  suggestion:
    var ErrInvalidFlag = errors.New("invalid flag")
    var ErrInvalidArgument = errors.New("invalid argument")

2) Error string formatting (Go Error Handling – Error String Formatting)
- Lines 11-14: Strings start uppercase and have a colon. See suggestion above.

3) Error wrapping and formatting
- Lines 26-27:
  code:
    return fmt.Errorf("%w %s a flag name must be alphanumeric from start to end %s", InvalidFlag, self, string(self))
  issue: Message starts with wrapped error token rendering first word from wrapped error; after renaming to ErrInvalidFlag with lowercase string, this improves. Consider clearer phrasing and colonless, lowercase message.
  suggestion:
    return fmt.Errorf("%w: %s must be alphanumeric: %s", ErrInvalidFlag, self, string(self))

4) API shape: function-typed vars
- Lines 34-42 and 45-49 define functions via var assignment. This is acceptable but can reduce clarity. Prefer named functions unless you need late binding or replacement.

---

## custom_flags/pkg.go

1) Duplicated validation blocks (Code Clarity and Consistency)
- Lines 269-283 are duplicated logic immediately after 267-271 and 273-275 checks in NewRangeFlag. This reduces clarity.
  code excerpt:
    if min > max { panic("min must be less than max") }
    if min < 0 || max < 0 { panic("min and max must be non-negative") }
    if min > max { panic("min must be less than max") }
    if min < 0 || max < 0 { panic("min and max must be non-negative") }
  suggestion: Remove duplicates and keep a single set of validations.

2) Error string formatting and messages (Go Error Handling)
- Multiple Set methods return messages starting with uppercase or include awkward concatenations:
  e.g., Lines 194-197:
    "%sflag must be one of: %v", custom_errors.FlagName(c.flagName), []string{"true", "false"}
  issue: Start lowercase and ensure human-friendly phrasing.
  suggestion:
    return fmt.Errorf("%s flag must be one of %v", custom_errors.FlagName(c.flagName), []string{"true", "false"})

3) Variable naming clarity
- Several receivers named self or t; keep consistency. Prefer concise, conventional names (p, f, r) and avoid self unless method-chaining semantics.

4) Package interface/constructor abstraction (Access Philosophy)
- New... constructors return concrete types for private structs (filePathFlag, folderPathFlag, emptyStringFlag, boolFlag, unionFlag, RangeFlag). This is okay when no abstraction is needed. No change required.

---

## detect/pkg.go

1) Spelling and naming clarity (Code Clarity and Consistency)
- Lines 129-150: DetectJSPacakgeManagerBasedOnLockFile – "Pacakge" misspelled; similarly used elsewhere (root.go). Also VOLTA_RUN_COMMNAD misspelled (lines 195-196).
  suggestion: Rename to DetectJSPackageManagerBasedOnLockFile and VOLTA_RUN_COMMAND. Ensure all call sites updated.

2) Error string formatting (Go Error Handling)
- Line 99:
  code: return "", fmt.Errorf("No lock file found")
  issue: Should start lowercase and no trailing punctuation.
  suggestion: return "", fmt.Errorf("no lock file found")

3) Interface placement (Go Interface Design – Definition Location)
- PathLookup and FileSystem are defined where implemented/used (detect), which is fine. No issue.

4) SupportedJSPackageManagers order comment
- The comment indicates a reliance on order. Consider documenting Why more clearly per Comments rule.

---

## services/npm-registry.go

1) Interface definition location (Go Interface Design – Definition Location)
- Lines 21-26: NpmRegistryService interface is defined in the services package (implementation package). Your rule prefers interfaces defined where they are consumed. Here, consumers are in cmd/install.go.
  suggestion: Define a minimal interface at the consumption site (cmd) and have services expose a concrete type. Alternatively, keep the interface but ensure primary usage interfaces live near the consumers.

2) Import grouping
- Lines 3-11: stdlib then blank line then external is correct. No violation.

3) Error string formatting
- Already uses lowercase starts and wraps with %w where appropriate. Good.

---

## cmd/root.go

1) Import grouping (Go Import Grouping)
- Lines 24-43: Standard library and external and internal packages are mixed in one block. Internal packages (github.com/louiss0/javascript-package-delegator/...) should be in a separate group after third-party imports.
  suggestion: Group as
    // stdlib
    context, errors, fmt, os, os/exec, regexp, strings

    // external
    github.com/charmbracelet/huh, github.com/charmbracelet/log, github.com/joho/godotenv, github.com/samber/lo, github.com/spf13/cobra

    // internal
    github.com/louiss0/javascript-package-delegator/build_info
    github.com/louiss0/javascript-package-delegator/custom_flags
    github.com/louiss0/javascript-package-delegator/detect
    github.com/louiss0/javascript-package-delegator/env
    github.com/louiss0/javascript-package-delegator/services

2) Access Philosophy – Constructor return type
- Lines 80-84:
  code:
    func newCommandRunner(execCommandFunc _ExecCommandFunc) *commandRunner { ... }
  issue: Private struct has methods and an interface (CommandRunner) exists. Constructor should return the public interface, not the concrete type.
  suggestion:
    func newCommandRunner(execCommandFunc _ExecCommandFunc) CommandRunner { return &commandRunner{...} }

3) Method receiver selection consistency (Access Philosophy – Method Receiver)
- Lines 86-125: commandRunner has pointer receivers for Command and SetTargetDir, but value receiver for Run.
  issue: Mixing receiver types can cause subtle bugs; the struct holds state (cmd, targetDir). Per rule, use pointer receivers for methods that read or might mutate state; and if any method uses a pointer receiver, prefer pointer receivers consistently.
  suggestion:
    func (e *commandRunner) Run() error { ... }

4) Getter method naming (Go Function Naming – Getter Methods)
- Lines 558-576: Helper functions getDebugExecutorFromCommandContext, getCommandRunnerFromCommandContext, getYarnVersionRunnerCommandContext, getGoEnvFromCommandContext are simple accessors.
  note: These are functions (not methods) so the "no Get" rule does not strictly apply. If you want to align with your general coding rule, keeping get- prefix is acceptable. If aligning to Go idioms, consider dropping get.

5) Variable naming – avoid using "error" as an identifier (Go Variable Naming)
- Multiple occurrences, e.g., Lines 328-334, 282-287, 399-407, 462-471:
  code:
    agent, error := persistentFlags.GetString(AGENT_FLAG)
  issue: Using "error" as a variable name conflicts with the built-in type and reduces clarity.
  suggestion: Rename to err.

6) Comments – explain Why, not What (Comments)
- Multiple large comment blocks (e.g., Lines 57-70, 127-141, 249-274) explain what code does in detail rather than the design intent/why.
  suggestion: Trim or rewrite comments to focus on rationale, invariants, and non-obvious decisions.

7) Misspellings (Code Clarity and Consistency)
- Lines 300-303, 520-525, 529-538: "DetectJSPacakgeManagerBasedOnLockFile" misspelling propagates; similarly VOLTA_RUN_COMMNAD appears in other files.
  suggestion: Fix spelling everywhere.

8) Error string formatting
- Lines 354-356 in an error: ensure lowercase first letter and no trailing punctuation.

---

## cmd/install.go

1) Variable naming – avoid using "error" as identifier (Go Variable Naming)
- Lines 140-157, 282-287, 297-311:
  code:
    packageInfo, error := npmRegistryService.SearchPackages(searchFlag.String())
    if error != nil { return error }
  issue: Use err instead of error.
  suggestion:
    packageInfo, err := ...
    if err != nil { return err }

2) Error string formatting (Go Error Handling)
- Line 149:
  code: return fmt.Errorf("Your query has failed %s", searchFlag.String())
  issue: Start lowercase; consider clearer phrasing.
  suggestion: return fmt.Errorf("search failed for %q", searchFlag.String())

- Line 256-263 (deno production flag): error message text should start lowercase, and be concise.
  suggestion: return custom_errors.CreateInvalidFlagErrorWithMessage(custom_errors.FlagName("production"), "deno does not support --production")

3) Import grouping
- Lines 24-32: stdlib fmt separated, then third-party and internal mixed without a separate internal group.
  suggestion: separate internal (github.com/louiss0/javascript-package-delegator/...) from external imports with blank line.

4) General clarity
- The switch blocks have repeated patterns for building cmdArgs; consider extracting helper functions per package manager to reduce duplication (Code Clarity and Consistency).

---

## cmd/clean-install.go

1) Variable naming – avoid "error"
- Lines 90-95:
  code: noVolta, error := cmd.Flags().GetBool(_NO_VOLTA_FLAG)
  suggestion: noVolta, err := ...; if err != nil { return err }

2) Error string formatting
- Line 84: return fmt.Errorf("%s doesn't support this command", "deno") → should be lowercase and possibly without contraction.
  suggestion: return fmt.Errorf("deno does not support this command")

3) Import grouping
- Internal vs external packages should be separated.

---

## cmd/exec.go

1) Error string formatting
- Line 108: return fmt.Errorf("Deno doesn't have a dlx or x like the others") → start lowercase; avoid contraction.
  suggestion: return fmt.Errorf("deno does not have a dlx/x equivalent")

2) Import grouping
- Separate internal imports from external.

3) General duplication with cmd/dlx.go
- exec and dlx share large amounts of logic. Consider consolidation to reduce duplication (Code Clarity and Consistency).

---

## cmd/dlx.go

1) Error string formatting
- Line 107: return fmt.Errorf("Deno doesn't have a dlx or x like the others") → lowercase and no contraction.

2) Import grouping
- Separate internal imports from external.

3) General duplication with cmd/exec.go
- Consider factoring common logic (package runner selection).

---

## cmd/run.go

1) Error string formatting
- Line 102-104: return fmt.Errorf("No tasks found in deno.json") → lowercase.
- Line 135-136: return fmt.Errorf("No scripts found in package.json") → lowercase.
- Line 231: return fmt.Errorf("Don't pass %s here use the exec command instead", "--eval") → lowercase and clearer phrasing.
  suggestion: return fmt.Errorf("--eval is not supported here; use the exec command instead")

2) Import grouping
- Separate internal imports if present (this file uses only stdlib + external; OK).

3) Comments – explain Why
- Consider trimming verbose comments that duplicate obvious behavior.

---

## cmd/update.go

1) Error string formatting
- Lines 69-70: return fmt.Errorf("npm does not support interactive updates") – already lowercase; OK.
- Review all messages for lowercase/no punctuation.

2) Import grouping
- Separate internal imports if present (none here).

---

## cmd/uninstall.go

1) Error string formatting
- Lines 212-213: return fmt.Errorf("No packages found for interactive uninstall.") → lowercase, no trailing period.
  suggestion: return fmt.Errorf("no packages found for interactive uninstall")

2) Import grouping
- Separate internal imports from external (detect is internal; others are external). Add blank line groups.

---

## cmd/agent.go

1) Error string formatting
- Lines 48-49: "Please ensure you have a lock file ..." – starts lowercase? The message begins with "no package manager detected." which is good. Ensure all errors start lowercase.

2) Import grouping
- Separate external from stdlib; there is no internal here.

---

## cmd/completion.go

1) Import grouping
- Lines 3-14: stdlib, internal (custom_errors, custom_flags), and external (cobra) are mixed. Separate into three groups.

2) Error string formatting
- Validate all messages start lowercase; they appear good.

3) File path validation behavior
- The command always creates the output file (lines 121-128) even when writing to stdout might be expected. This is by design per your help text, but ensure this is the intended UX.

---

## mock/pkg.go

1) Import grouping
- Lines 3-16: stdlib, internal, and external mixed. Separate into three groups.

2) Comments – Why vs What
- Several comments describe what the mock does in obvious terms. Consider focusing on why mocks are lenient/strict to align with your Comments rule.

---

## testutil/pkg.go

1) Package naming (Go Package Naming and Structure)
- Package name is testutil (single word). Your rule prefers snake_case for multi-word packages. Consider test_util.
  note: This conflicts with common Go idioms (lowercase, no underscores). If your rule is authoritative, rename; otherwise, consider an exception.

2) Import grouping
- Internal and external are mixed. Separate groups.

---

## testutil/cleancheck.go

1) Package naming (Go Package Naming and Structure)
- Same as testutil above.

2) Import grouping
- Lines 5-9: stdlib only; OK.

3) Comments – Why vs What
- The comments describe what functions do. Consider adding why this check is useful in tests (e.g., enforcing clean working tree).

---

# Summary

Files analyzed: 17
Files with violations: 13

Most common violations
- Import grouping not separated into stdlib / external / internal.
- Error string formatting (must start lowercase, no trailing punctuation).
- Variable naming: using "error" as an identifier.
- Access philosophy consistency: constructor return type should be interface; consistent pointer receivers for stateful structs.
- Spelling issues reducing clarity (DetectJSPacakgeManager..., VOLTA_RUN_COMMNAD).
- Package naming for multi-word testutil (per your snake_case rule).
- Getter method naming on methods (GetgoEnv, and general use of Get on methods).

Top files by violations
- cmd/root.go
- cmd/install.go
- custom_flags/pkg.go
- detect/pkg.go
- mock/pkg.go

Recommended next steps
- Apply goimports with proper group configuration; enforce via CI.
- Standardize error message style; consider a static check in CI.
- Rename misnamed identifiers (spelling) and avoid "error" variable name.
- Adjust commandRunner receivers; return interface from constructors of private structs with methods.
- Decide on package naming policy trade-off vs. Go ecosystem norms (snake_case vs. standard Go lowercase names).

