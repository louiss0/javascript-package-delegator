package cmd

import (
	"strings"
	"testing"
)

// Test strategy and mapping rules:
//
// EXEC (local dependencies): Run local deps using package manager's exec feature
// | Tool           | Command                                        |
// | -------------- | ---------------------------------------------- |
// | npm            | npm exec <bin> -- <args>                      |
// | pnpm           | pnpm exec <bin> <args>                        |
// | yarn (any)     | yarn <bin> <args>                             |
// | bun            | bun x <bin> <args>                            |
// | deno           | deno run <script> <args>                      |
//
// DLX (temporary packages): Run packages temporarily without install
// | Tool           | Command                                        |
// | -------------- | ---------------------------------------------- |
// | npm            | npm dlx <package> <args>                      |
// | pnpm           | pnpm dlx <package> <args>                     |
// | yarn v1        | yarn <package> <args> (no dlx)               |
// | yarn v2+       | yarn dlx <package> <args>                     |
// | bun            | bunx <package> <args>                         |
// | deno           | deno run <url> <args> (requires URL)          |

// Test fixtures and helpers
type pm string

const (
	npm  pm = "npm"
	pnpm pm = "pnpm"
	yarn pm = "yarn"
	bun  pm = "bun"
	deno pm = "deno"
)

type yarnInfo struct {
	version string
}

type wantCmd struct {
	program string
	args    []string
}

type harness struct {
	packageManager pm
	yarnVersion    string
}

func newHarness(packageManager pm, yarnVersion string) harness {
	return harness{
		packageManager: packageManager,
		yarnVersion:    yarnVersion,
	}
}

func assertCmd(t *testing.T, gotProg string, gotArgs []string, want wantCmd) {
	t.Helper()
	if gotProg != want.program {
		t.Errorf("program = %q, want %q", gotProg, want.program)
	}
	if len(gotArgs) != len(want.args) {
		t.Errorf("args length = %d, want %d\nGot: %v\nWant: %v", len(gotArgs), len(want.args), gotArgs, want.args)
		return
	}
	for i, gotArg := range gotArgs {
		if gotArg != want.args[i] {
			t.Errorf("args[%d] = %q, want %q", i, gotArg, want.args[i])
		}
	}
}

// Helper functions are now in command_utils.go

// buildExec wraps the shared function for testing
func buildExec(h harness, bin string, args []string) (program string, argv []string, err error) {
	return buildExecCommand(string(h.packageManager), h.yarnVersion, bin, args)
}

// buildDLX wraps the shared function for testing
func buildDLX(h harness, pkgOrURL string, args []string) (program string, argv []string, err error) {
	return buildDLXCommand(string(h.packageManager), h.yarnVersion, pkgOrURL, args)
}

// Using standard Go errors from shared functions

func TestExec_CommandMapping(t *testing.T) {
	tests := []struct {
		name string
		h    harness
		bin  string
		args []string
		want wantCmd
	}{
		{
			name: "npm exec with args",
			h:    newHarness(npm, ""),
			bin:  "ts-node",
			args: []string{"src/index.ts"},
			want: wantCmd{program: "npm", args: []string{"exec", "ts-node", "--", "src/index.ts"}},
		},
		{
			name: "npm exec no args",
			h:    newHarness(npm, ""),
			bin:  "eslint",
			args: []string{},
			want: wantCmd{program: "npm", args: []string{"exec", "eslint", "--"}},
		},
		{
			name: "pnpm exec with args",
			h:    newHarness(pnpm, ""),
			bin:  "eslint",
			args: []string{"--version"},
			want: wantCmd{program: "pnpm", args: []string{"exec", "eslint", "--version"}},
		},
		{
			name: "pnpm exec no args",
			h:    newHarness(pnpm, ""),
			bin:  "vite",
			args: []string{},
			want: wantCmd{program: "pnpm", args: []string{"exec", "vite"}},
		},
		{
			name: "yarn v1 exec",
			h:    newHarness(yarn, "1.22.19"),
			bin:  "vite",
			args: []string{"--help"},
			want: wantCmd{program: "yarn", args: []string{"vite", "--help"}},
		},
		{
			name: "yarn v3 exec",
			h:    newHarness(yarn, "3.6.1"),
			bin:  "vite",
			args: []string{"build"},
			want: wantCmd{program: "yarn", args: []string{"vite", "build"}},
		},
		{
			name: "bun exec",
			h:    newHarness(bun, ""),
			bin:  "tsx",
			args: []string{"file.ts"},
			want: wantCmd{program: "bun", args: []string{"x", "tsx", "file.ts"}},
		},
		{
			name: "deno run",
			h:    newHarness(deno, ""),
			bin:  "main.ts",
			args: []string{"--allow-all"},
			want: wantCmd{program: "deno", args: []string{"run", "main.ts", "--allow-all"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prog, argv, err := buildExec(tt.h, tt.bin, tt.args)
			if err != nil {
				t.Fatalf("buildExec() error = %v", err)
			}

			assertCmd(t, prog, argv, tt.want)
		})
	}
}

func TestDLX_CommandMapping(t *testing.T) {
	tests := []struct {
		name     string
		h        harness
		pkgOrURL string
		args     []string
		want     wantCmd
	}{
		{
			name:     "npm dlx with args",
			h:        newHarness(npm, ""),
			pkgOrURL: "create-vite@latest",
			args:     []string{"my-app"},
			want:     wantCmd{program: "npm", args: []string{"dlx", "create-vite@latest", "my-app"}},
		},
		{
			name:     "npm dlx no args",
			h:        newHarness(npm, ""),
			pkgOrURL: "@angular/cli",
			args:     []string{},
			want:     wantCmd{program: "npm", args: []string{"dlx", "@angular/cli"}},
		},
		{
			name:     "pnpm dlx with args",
			h:        newHarness(pnpm, ""),
			pkgOrURL: "create-next-app",
			args:     []string{"my-app"},
			want:     wantCmd{program: "pnpm", args: []string{"dlx", "create-next-app", "my-app"}},
		},
		{
			name:     "yarn v1 dlx (no dlx subcommand)",
			h:        newHarness(yarn, "1.22.19"),
			pkgOrURL: "create-vite",
			args:     []string{"my-app"},
			want:     wantCmd{program: "yarn", args: []string{"create-vite", "my-app"}},
		},
		{
			name:     "yarn v3 dlx",
			h:        newHarness(yarn, "3.6.1"),
			pkgOrURL: "create-vite",
			args:     []string{"my-app"},
			want:     wantCmd{program: "yarn", args: []string{"dlx", "create-vite", "my-app"}},
		},
		{
			name:     "yarn berry dlx",
			h:        newHarness(yarn, "berry-3.1.0"),
			pkgOrURL: "create-vue",
			args:     []string{},
			want:     wantCmd{program: "yarn", args: []string{"dlx", "create-vue"}},
		},
		{
			name:     "bun dlx",
			h:        newHarness(bun, ""),
			pkgOrURL: "create-vite",
			args:     []string{"my-app"},
			want:     wantCmd{program: "bunx", args: []string{"create-vite", "my-app"}},
		},
		{
			name:     "deno dlx with URL",
			h:        newHarness(deno, ""),
			pkgOrURL: "https://deno.land/x/xyz/mod.ts",
			args:     []string{"--help"},
			want:     wantCmd{program: "deno", args: []string{"run", "https://deno.land/x/xyz/mod.ts", "--help"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prog, argv, err := buildDLX(tt.h, tt.pkgOrURL, tt.args)
			if err != nil {
				t.Fatalf("buildDLX() error = %v", err)
			}

			assertCmd(t, prog, argv, tt.want)
		})
	}
}

func TestYarnVersionDetection(t *testing.T) {
	tests := []struct {
		version string
		want    int
	}{
		{"1.22.19", 1},
		{"2.4.3", 2},
		{"3.6.1", 3},
		{"berry-3.1.0", 3},
		{"3", 3},
		{"", 0},
		{"invalid", 0},
		{"0.9.0", 0}, // Pre-1.0 versions
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := parseYarnMajor(tt.version)
			if got != tt.want {
				t.Errorf("parseYarnMajor(%q) = %d, want %d", tt.version, got, tt.want)
			}
		})
	}
}

func TestArgumentValidation(t *testing.T) {
	t.Run("exec missing binary", func(t *testing.T) {
		h := newHarness(npm, "")
		_, _, err := buildExec(h, "", []string{})
		if err == nil {
			t.Error("buildExec() with empty binary should return error")
		}
		if !strings.Contains(err.Error(), "binary name is required") {
			t.Errorf("buildExec() error = %v, should mention missing binary", err)
		}
	})

	t.Run("dlx missing package", func(t *testing.T) {
		h := newHarness(npm, "")
		_, _, err := buildDLX(h, "", []string{})
		if err == nil {
			t.Error("buildDLX() with empty package should return error")
		}
		if !strings.Contains(err.Error(), "package name or URL is required") {
			t.Errorf("buildDLX() error = %v, should mention missing package", err)
		}
	})

	t.Run("deno dlx requires URL", func(t *testing.T) {
		h := newHarness(deno, "")
		_, _, err := buildDLX(h, "not-a-url", []string{})
		if err == nil {
			t.Error("buildDLX() with deno and non-URL should return error")
		}
		if !strings.Contains(err.Error(), "deno dlx requires a URL") {
			t.Errorf("buildDLX() error = %v, should mention URL requirement", err)
		}
	})

	t.Run("unsupported package manager exec", func(t *testing.T) {
		h := newHarness("unknown", "")
		_, _, err := buildExec(h, "test", []string{})
		if err == nil {
			t.Error("buildExec() with unknown PM should return error")
		}
		if !strings.Contains(err.Error(), "unsupported package manager") {
			t.Errorf("buildExec() error = %v, should mention unsupported PM", err)
		}
	})

	t.Run("unsupported package manager dlx", func(t *testing.T) {
		h := newHarness("unknown", "")
		_, _, err := buildDLX(h, "test", []string{})
		if err == nil {
			t.Error("buildDLX() with unknown PM should return error")
		}
		if !strings.Contains(err.Error(), "unsupported package manager") {
			t.Errorf("buildDLX() error = %v, should mention unsupported PM", err)
		}
	})
}

func TestWindowsArgumentPassthrough(t *testing.T) {
	// Test that arguments with spaces and special characters pass through unchanged
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "args with spaces",
			args: []string{"--flag=a b", "c=d"},
		},
		{
			name: "args with quotes",
			args: []string{"--message=\"hello world\"", "--path=C:\\Program Files\\test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(npm, "")
			_, argv, err := buildExec(h, "test-bin", tt.args)
			if err != nil {
				t.Fatalf("buildExec() error = %v", err)
			}

			// Args should be passed through unchanged after the exec portion
			expectedStart := []string{"exec", "test-bin", "--"}
			if len(argv) < len(expectedStart)+len(tt.args) {
				t.Fatalf("argv length = %d, too short", len(argv))
			}

			for i, expected := range expectedStart {
				if argv[i] != expected {
					t.Errorf("argv[%d] = %q, want %q", i, argv[i], expected)
				}
			}

			// Check that the remaining args match exactly
			actualArgs := argv[len(expectedStart):]
			for i, expected := range tt.args {
				if actualArgs[i] != expected {
					t.Errorf("args[%d] = %q, want %q", i, actualArgs[i], expected)
				}
			}
		})
	}
}
