#!/bin/bash

# Script to clean up test artifacts and accidentally generated binaries
# This should be run after tests to ensure no binary artifacts remain

set -euo pipefail

PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." &> /dev/null && pwd)
cd "$PROJECT_ROOT"

echo "Cleaning up test artifacts and binaries..."

# Remove the main binary if it exists (generated during tests)
if [[ -f "javascript-package-delegator" ]]; then
    echo "Removing javascript-package-delegator binary..."
    rm -f "javascript-package-delegator"
fi

# Remove any test binaries
if [[ -f "jpd" ]]; then
    echo "Removing jpd binary..."
    rm -f "jpd"
fi

# Remove coverage files if they exist
if [[ -f "coverage.out" ]]; then
    echo "Removing coverage.out..."
    rm -f "coverage.out"
fi

if [[ -f "coverage.html" ]]; then
    echo "Removing coverage.html..."
    rm -f "coverage.html"
fi

# Remove any .test files (test binaries)
find . -name "*.test" -type f -delete 2>/dev/null || true

echo "✅ Cleanup complete"
