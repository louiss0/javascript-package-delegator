// Package env provides environment configuration utilities.
package env

// PackageJSON represents a package.json file structure with name and version
type PackageJSON struct {
	Name    string
	Version string
}

// Dependency represents a package dependency with name and version
type Dependency struct {
	Name    string
	Version string
}
