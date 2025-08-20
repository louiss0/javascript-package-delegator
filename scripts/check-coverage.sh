#!/bin/bash

# Script to check test coverage meets minimum threshold
# Used by CI to gate builds on coverage requirements

set -euo pipefail

THRESHOLD=${1:-80}
COVERAGE_FILE=${2:-coverage.out}

echo "Checking test coverage against ${THRESHOLD}% threshold..."

# Run tests with coverage, excluding infrastructure packages
go test $(go list ./... | grep -v -E '/(build_info|mock|testutil)$') -coverprofile="${COVERAGE_FILE}"

# Extract coverage percentage
coverage_line=$(go tool cover -func="${COVERAGE_FILE}" | grep '^total:')
coverage_percent=$(echo "${coverage_line}" | awk '{print $3}' | sed 's/%//')

echo "Current coverage: ${coverage_percent}%"
echo "Required threshold: ${THRESHOLD}%"

# Check if coverage meets threshold
if (( $(echo "${coverage_percent} < ${THRESHOLD}" | bc -l) )); then
    echo "❌ Coverage ${coverage_percent}% is below ${THRESHOLD}% threshold"
    echo
    echo "Coverage details:"
    go tool cover -func="${COVERAGE_FILE}" | tail -n 5
    exit 1
fi

echo "✅ Coverage ${coverage_percent}% meets ${THRESHOLD}% threshold"
