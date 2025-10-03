#!/usr/bin/env bash
set -euo pipefail
# Usage: ./scripts/coverage.sh [threshold]
# Default threshold is 100 (fail unless 100% coverage). Pass e.g. 80 to require 80%.
threshold=${1:-100}
# optional comma-separated packages to exclude from coverage (e.g. cmd/gitimpl)
excludes=${2:-}

echo "Running go tests with coverage (threshold=${threshold}%)"

# build coverpkg list excluding any specified packages
if [ -n "$excludes" ]; then
	# build a grep -v pattern from comma-separated excludes
	IFS=',' read -ra EX <<< "$excludes"
	# list all packages and filter excludes
	pkgs=$(go list ./...)
	for e in "${EX[@]}"; do
		pkgs=$(echo "$pkgs" | grep -v "^$e$" || true)
	done
	coverpkg=$(echo "$pkgs" | paste -sd, -)
else
	coverpkg=./...
fi

# run tests across packages and capture coverage
go test ./... -covermode=count -coverpkg="$coverpkg" -coverprofile=coverage.out

echo "Coverage report (per-file):"
go tool cover -func=coverage.out | tee coverage.txt

total=$(awk '/^total:/ {gsub(/%/,"",$3); print $3}' coverage.txt)
echo "Total coverage: ${total}%"

# compare as floats via awk
awk -v total="$total" -v thr="$threshold" 'BEGIN {if (total+0 < thr+0) {print "Coverage " total "% is below threshold " thr "%"; exit 2} else {print "Coverage meets threshold"; exit 0}}'

echo "(coverage.out written)"
