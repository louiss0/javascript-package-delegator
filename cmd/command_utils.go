package cmd

import (
	"fmt"
	"strings"
)

// parseYarnMajor extracts the major version number from a yarn version string
func parseYarnMajor(version string) int {
	if version == "" {
		return 0
	}

	// Handle simple cases like "3" or "berry-3.1.0"
	if strings.HasPrefix(version, "berry-") {
		version = strings.TrimPrefix(version, "berry-")
	}

	// Extract first character and convert to int
	if len(version) > 0 && version[0] >= '1' && version[0] <= '9' {
		return int(version[0] - '0')
	}

	return 0 // unknown
}

// isURL checks if a string is a valid HTTP or HTTPS URL
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// buildExecCommand builds command line for running local dependencies
func buildExecCommand(pm, yarnVersion, bin string, args []string) (program string, argv []string, err error) {
	if bin == "" {
		return "", nil, fmt.Errorf("binary name is required for exec command")
	}

	switch pm {
	case "npm":
		argv = append([]string{"exec", bin, "--"}, args...)
		return "npm", argv, nil
	case "pnpm":
		argv = append([]string{"exec", bin}, args...)
		return "pnpm", argv, nil
	case "yarn":
		argv = append([]string{bin}, args...)
		return "yarn", argv, nil
	case "bun":
		argv = append([]string{"x", bin}, args...)
		return "bun", argv, nil
	case "deno":
		argv = append([]string{"run", bin}, args...)
		return "deno", argv, nil
	default:
		return "", nil, fmt.Errorf("unsupported package manager: %s", pm)
	}
}

// buildDLXCommand builds command line for running temporary packages
func buildDLXCommand(pm, yarnVersion, pkgOrURL string, args []string) (program string, argv []string, err error) {
	if pkgOrURL == "" {
		return "", nil, fmt.Errorf("package name or URL is required for dlx command")
	}

	switch pm {
	case "npm":
		argv = append([]string{"dlx", pkgOrURL}, args...)
		return "npm", argv, nil
	case "pnpm":
		argv = append([]string{"dlx", pkgOrURL}, args...)
		return "pnpm", argv, nil
	case "yarn":
		yarnMajor := parseYarnMajor(yarnVersion)
		if yarnMajor >= 2 {
			// Yarn v2+
			argv = append([]string{"dlx", pkgOrURL}, args...)
		} else {
			// Yarn v1 or unknown (default to v1)
			argv = append([]string{pkgOrURL}, args...)
		}
		return "yarn", argv, nil
	case "bun":
		argv = append([]string{pkgOrURL}, args...)
		return "bunx", argv, nil
	case "deno":
		if !isURL(pkgOrURL) {
			return "", nil, fmt.Errorf("deno dlx requires a URL")
		}
		argv = append([]string{"run", pkgOrURL}, args...)
		return "deno", argv, nil
	default:
		return "", nil, fmt.Errorf("unsupported package manager: %s", pm)
	}
}
