module jpd_completions_module {
    # Define a custom completer for the --agent flag
    # This provides the list of supported package managers.
    export def "complete_jpd_agent_types" [] {
        [
            "npm",
            "yarn",
            "pnpm",
            "bun",
            "deno"
        ]
    }

    # Define a completer for the top-level subcommands of 'jpd'
    export def "complete_jpd_subcommands" [] {
        [
            "agent",
            "clean-install",
            "exec",
            "install",
            "run",
            "uninstall",
            "update",
            "completion"
        ]
    }

    # Define a completer for the 'jpd completion' subcommand's shell types
    export def "complete_jpd_completion_shells" [] {
        [
            "nushell",
            "bash",
            "zsh",
            "fish",
            "powershell"
        ]
    }

    # Define the 'jpd' extern command and its global flags.
    # Subcommands are handled by separate extern definitions (e.g., 'jpd install').
    export extern "jpd" [
        # Global flags
        --debug(-d)                  # Make commands run in debug mode
        --agent(-a): string@complete_jpd_agent_types # Select the JS package manager you want to use
        --cwd(-C): path              # Run command in a specific directory (must end with '/')

        # First positional argument is expected to be a subcommand
        subcommand: string@complete_jpd_subcommands # The subcommand to run (e.g., "install", "run")
        ...args: string              # Remaining arguments and flags for the subcommand

        # Basic help flags
        --help(-h)                   # Show help for command
        --version(-v)                # Show version for command
    ]

    # Define separate externs for each subcommand for specific flag completions
    export extern "jpd agent" [
        # This subcommand primarily uses global flags from 'jpd'
    ] # Show the detected package manager agent

    export extern "jpd clean-install" [
        --no-volta                   # Disable Volta integration for this command
    ] # Clean install packages using the detected package manager

    export extern "jpd exec" [
        ...args: string              # Package to execute and its arguments
    ] # Execute packages using the detected package manager

    export extern "jpd install" [
        ...packages: string          # Packages to install
        --dev(-D)                    # Install as dev dependency
        --global(-g)                 # Install globally
        --production(-P)             # Install production dependencies only
        --frozen                     # Install with frozen lockfile
        --search(-s): string         # Interactive package search selection
        --no-volta                   # Disable Volta integration for this command
    ] # Install packages using the detected package manager

    export extern "jpd run" [
        script?: string              # Script to execute
        ...args: string              # Arguments for the script
        --if-present                 # Run script only if it exists
    ] # Run scripts using the detected package manager

    export extern "jpd uninstall" [
        ...packages: string          # Packages to uninstall
        --global(-g)                 # Uninstall global packages
        --interactive(-i)            # Uninstall packages interactively
    ] # Uninstall packages using the detected package manager

    export extern "jpd update" [
        ...packages: string          # Packages to update
        --interactive(-i)            # Interactive update (where supported)
        --global(-g)                 # Update global packages
        --latest                     # Update to latest version (ignoring version ranges)
    ] # Update packages using the detected package manager

    # Special "completion" subcommand for generating completion scripts
    export extern "jpd completion" [
        shell_type: string@complete_jpd_completion_shells # Type of shell to generate completion script for
        output_file?: path           # Optional output file path
    ] # Generate shell completion scripts
}

# Source this file in your Nushell config (e.g., ~/.config/nushell/config.nu or env.nu):
# source /path/to/jpd_completions.nu
# Then use the module to bring the externs into scope:
use "jpd_completions_module" *
