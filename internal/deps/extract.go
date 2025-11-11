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
func ExtractProdAndDevDependenciesFromPackageJSON(baseDir string) ([]string, error) {
	type PackageJSONDependencies struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if baseDir == "" {
		var err error
		baseDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	packageJSONPath := filepath.Join(baseDir, "package.json")
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
func ExtractImportsFromDenoJSON(baseDir string) ([]string, error) {
	type DenoJSONDependencies struct {
		Imports map[string]string `json:"imports"`
	}

	if baseDir == "" {
		var err error
		baseDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	// Try deno.json first, then deno.jsonc
	var denoFilePath string
	denoJSONPath := filepath.Join(baseDir, "deno.json")
	denoJSONCPath := filepath.Join(baseDir, "deno.jsonc")

	if _, err := os.Stat(denoJSONPath); err == nil {
		denoFilePath = denoJSONPath
	} else if _, err := os.Stat(denoJSONCPath); err == nil {
		denoFilePath = denoJSONCPath
	} else {
		return nil, fmt.Errorf("%w: %s", ErrDenoConfigNotFound, baseDir)
	}

	data, err := os.ReadFile(denoFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", denoFilePath, err)
	}

	// If this is a deno.jsonc file, normalize it to JSON
	if filepath.Ext(denoFilePath) == ".jsonc" {
		data = NormalizeJSONCToJSON(data)
	}

	var pkg DenoJSONDependencies

	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", denoFilePath, err)
	}

	importValues := lo.Values(pkg.Imports)

	return importValues, nil
}
