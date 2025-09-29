package cmd

import (
	"testing"
)

func TestBuildCreateCommand_PNPM_UnscopedAddsCreatePrefix(t *testing.T) {
	prog, argv, err := BuildCreateCommand("pnpm", "", "vite@latest", []string{"myapp", "--", "--template", "react"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prog != "pnpm" {
		t.Fatalf("expected pnpm, got %s", prog)
	}
	if len(argv) < 2 || argv[0] != "dlx" || argv[1] != "create-vite@latest" {
		t.Fatalf("unexpected argv: %#v", argv)
	}
}

func TestBuildCreateCommand_PNPM_ScopedNoPrefix(t *testing.T) {
	prog, argv, err := BuildCreateCommand("pnpm", "", "@sveltejs/create-svelte@latest", []string{"myapp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prog != "pnpm" {
		t.Fatalf("expected pnpm, got %s", prog)
	}
	if len(argv) < 2 || argv[0] != "dlx" || argv[1] != "@sveltejs/create-svelte@latest" {
		t.Fatalf("unexpected argv: %#v", argv)
	}
}

func TestBuildCreateCommand_YarnV1_UsesNpx(t *testing.T) {
	prog, argv, err := BuildCreateCommand("yarn", "1.22.19", "vite@latest", []string{"myapp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prog != "npx" {
		t.Fatalf("expected npx for yarn v1, got %s", prog)
	}
	if len(argv) == 0 || argv[0] != "create-vite@latest" {
		t.Fatalf("unexpected argv: %#v", argv)
	}
}

func TestBuildCreateCommand_YarnV2_UsesYarnDlx(t *testing.T) {
	prog, argv, err := BuildCreateCommand("yarn", "3.1.0", "vite", []string{"myapp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prog != "yarn" {
		t.Fatalf("expected yarn, got %s", prog)
	}
	if len(argv) < 2 || argv[0] != "dlx" || argv[1] != "create-vite" {
		t.Fatalf("unexpected argv: %#v", argv)
	}
}

func TestBuildCreateCommand_NPM_InsertsSeparator(t *testing.T) {
	prog, argv, err := BuildCreateCommand("npm", "", "vite", []string{"myapp", "--", "--template", "react-swc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prog != "npm" {
		t.Fatalf("expected npm, got %s", prog)
	}
	if len(argv) < 3 || argv[0] != "exec" || argv[1] != "create-vite" || argv[2] != "--" {
		t.Fatalf("expected exec create-vite -- ..., got %#v", argv)
	}
}

func TestBuildCreateCommand_Deno_URLValidation(t *testing.T) {
	// valid URL
	if _, _, err := BuildCreateCommand("deno", "", "https://deno.land/x/fresh/init.ts", []string{"myapp"}); err != nil {
		t.Fatalf("unexpected error for valid deno URL: %v", err)
	}
	// invalid name
	if _, _, err := BuildCreateCommand("deno", "", "fresh", []string{"myapp"}); err == nil {
		t.Fatalf("expected error for non-URL deno create input")
	}
}
