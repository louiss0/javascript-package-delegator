# Nix Environment Preservation for GoReleaser

This document explains how the GitHub Actions workflow ensures that Nix is available when GoReleaser runs for package generation.

## Overview

The release workflow uses both Nix and GoReleaser. GoReleaser needs access to Nix commands to generate Nix packages as configured in `goreleaser.yaml`. The challenge is ensuring that the Nix environment variables and PATH modifications persist between GitHub Actions steps.

## Implementation Details

### 1. Nix Installation
The workflow uses `cachix/install-nix-action@v27` to install Nix with:
- A specific Nix version (2.18.1) for consistency
- Experimental features enabled (nix-command and flakes)
- GitHub token for authenticated access

### 2. Environment Preservation
After Nix installation, the workflow:
- Verifies Nix is installed correctly
- Captures critical environment variables:
  - `PATH` - Contains Nix binary locations
  - `NIX_PATH` - Nix's search path for packages
  - `NIX_PROFILES` - User profile paths (if set)
  - `NIX_SSL_CERT_FILE` - SSL certificates for HTTPS (if set)
- Exports these variables to `$GITHUB_ENV` for use in subsequent steps

### 3. Verification Step
A dedicated step tests that:
- The `nix` command is available in PATH
- Nix version can be displayed
- Additional Nix tools like `nix-prefetch-url` are accessible

### 4. GoReleaser Execution
GoReleaser runs with:
- Access to all preserved Nix environment variables
- Ability to execute Nix commands for package generation
- Configuration from `goreleaser.yaml` including the `nix:` section

## Testing

To test this setup locally:

```bash
# Install Nix if not already installed
curl -L https://nixos.org/nix/install | sh

# Source Nix environment
source ~/.nix-profile/etc/profile.d/nix.sh

# Verify Nix is available
nix --version

# Run GoReleaser in snapshot mode (doesn't publish)
goreleaser release --snapshot --clean
```

## Troubleshooting

If Nix commands are not available to GoReleaser:

1. Check the workflow logs for the "Verify Nix installation" step
2. Ensure all environment variables are properly exported
3. Verify the Nix installation URL is accessible
4. Check that experimental features are enabled if using flakes

## Related Files

- `.github/workflows/release.yml` - The GitHub Actions workflow
- `goreleaser.yaml` - GoReleaser configuration with Nix settings
- `default.nix` - Nix package definition
