// This package defines all the build info that this repo needs
// All variables must be capitalised and must use the `BuildInfo` type
// All validation of individual types must be handled by the init function
// javascript-package-delegator/build_info/root.go
package build_info

import (
	"fmt"
	"regexp"
	"time" // Only for GetFormattedBuildDate in this version

	"github.com/samber/lo"
)

// BuildInfo is a type alias for string, ensuring consistency.
type BuildInfo string

func (value BuildInfo) String() string {
	return string(value)
}

// Raw variables to be populated by LDFLAGS.
// They must be of type string and package-level.
// Initialize with default/placeholder values that indicate they weren't set.
var (
	rawCLI_VERSION = "dev"         // Default for local development
	rawGO_MODE     = "development" // Default for local development
	rawBUILD_DATE  = "unknown"     // Default, will be overwritten by ldflags for releases
)

// Public variables that expose the validated and potentially parsed values.
var (
	CLI_VERSION BuildInfo
	GO_MODE     BuildInfo
	BUILD_DATE  BuildInfo // This will now hold the ldflags-injected date
)

// init function runs automatically when the package is initialized (before main).
func init() {
	// Assign raw values to public BuildInfo typed variables
	CLI_VERSION = BuildInfo(rawCLI_VERSION)
	GO_MODE = BuildInfo(rawGO_MODE)
	BUILD_DATE = BuildInfo(rawBUILD_DATE)

	// --- GO_MODE Validation ---
	// Check GO_MODE against allowed modes
	allowedModes := []string{"development", "production", "debug"}
	if !lo.Contains(allowedModes, GO_MODE.String()) { // CORRECTED: Checking GO_MODE
		panic(fmt.Sprintf("build_info: invalid GO_MODE: '%s'. Must be one of: %v", GO_MODE.String(), allowedModes))
	}

	// --- CLI_VERSION Validation ---
	// Check CLI_VERSION against semver regex (unless it's "dev")
	if CLI_VERSION.String() == "dev" {
		// "dev" is explicitly allowed for local development builds.
	} else {
		// Full semver regex including optional 'v' prefix
		semverRegex := `^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
		match, err := regexp.MatchString(semverRegex, CLI_VERSION.String()) // CORRECTED: Checking CLI_VERSION
		if err != nil {
			panic(fmt.Errorf("build_info: internal regex error for CLI_VERSION validation: %w", err))
		}
		if !match {
			panic(fmt.Sprintf("build_info: invalid CLI_VERSION format: '%s'. Must be a valid semver string (e.g., v1.2.3 or 1.2.3-beta.1)", CLI_VERSION.String()))
		}
	}

	// --- BUILD_DATE Validation ---
	// Ensure BUILD_DATE is in a valid format (e.g., YYYY-MM-DD) or "unknown" for dev builds.
	// If it's not "unknown", try parsing it.
	if BUILD_DATE.String() == "unknown" {
		// Allowed for dev builds, but for release builds it should be set by ldflags
		if GO_MODE.String() == "production" { // Optional: enforce date for production
			panic("build_info: BUILD_DATE is 'unknown' in production mode. It must be set via ldflags.")
		}
	} else {
		// Attempt to parse the date to ensure its format is correct
		_, err := time.Parse(time.RFC3339, BUILD_DATE.String())
		if err != nil {
			panic(fmt.Sprintf("build_info: invalid BUILD_DATE format: '%s'. Must be YYYY-MM-DD or 'unknown': %v", BUILD_DATE.String(), err))
		}
	}
}

// GetVersion returns the application's CLI version.
func GetVersion() string {
	return CLI_VERSION.String()
}

// GetGoMode returns the application's Go environment mode.
func GetGoMode() string {
	return GO_MODE.String()
}

// GetBuildDate returns the build date.
func GetBuildDate() string {
	return BUILD_DATE.String()
}
