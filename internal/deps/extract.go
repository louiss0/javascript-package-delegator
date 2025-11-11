// Package deps provides functionality for dependency management and detection
// across different JavaScript package managers and runtime environments.
package deps

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/samber/lo"
)

// ExtractProdAndDevDependenciesFromPackageJSON reads package.json and extracts
// both production and development dependencies with their versions.
// Returns a slice of strings in format "name@version".
func ExtractProdAndDevDependenciesFromPackageJSON() ([]string, error) {
	type PackageJSONDependencies struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	packageJSONPath := filepath.Join(cwd, "package.json")
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg PackageJSONDependencies

	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}

	prodAndDevDependenciesMerged := lo.Map(
		lo.Entries(lo.Assign(pkg.Dependencies, pkg.DevDependencies)),
		func(item lo.Entry[string, string], index int) string {
			return fmt.Sprintf("%s@%s", item.Key, item.Value)
		},
	)

	return prodAndDevDependenciesMerged, nil
}

// ExtractImportsFromDenoJSON reads deno.json (or deno.jsonc if deno.json doesn't exist)
// and extracts import values from the "imports" field.
// Returns a slice of import URLs/paths.
func ExtractImportsFromDenoJSON(cwd string) ([]string, error) {
	type DenoJSONDependencies struct {
		Imports map[string]string `json:"imports"`
	}

	denoFilePath, err := DenoConfigPath(cwd)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(denoFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", denoFilePath, err)
	}

	// If this is a deno.jsonc file, normalize it to JSON
	data = lo.TernaryF(
		filepath.Ext(denoFilePath) == ".jsonc",
		func() []byte { return NormalizeJSONCToJSON(data) },
		func() []byte { return data },
	)

	var pkg DenoJSONDependencies

	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", denoFilePath, err)
	}

	importValues := lo.Values(pkg.Imports)

	return importValues, nil
}

// DenoConfigPath returns the path to deno.json (preferring it over deno.jsonc) for the
// provided working directory.
func DenoConfigPath(cwd string) (string, error) {
	denoJSONPath := filepath.Join(cwd, "deno.json")
	denoJSONCPath := filepath.Join(cwd, "deno.jsonc")

	fileExists := func(path string) bool {
		info, err := os.Stat(path)
		return err == nil && !info.IsDir()
	}

	selectedPath := lo.Switch[bool, string](true).
		Case(fileExists(denoJSONPath), denoJSONPath).
		Case(fileExists(denoJSONCPath), denoJSONCPath).
		Default("")

	if selectedPath == "" {
		return "", fmt.Errorf("failed to find deno.json or deno.jsonc in %s", cwd)
	}

	return selectedPath, nil
}
