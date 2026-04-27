#!/bin/bash
# verify_compilation.sh
# Checks if the example compiles successfully
# Exit code 0 = success, 1 = failure
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"
echo "=== Compilation Verification ==="
echo "Working directory: $(pwd)"
echo "Go version: $(go version)"
echo ""
echo "Running 'go vet'..."
if go vet ./...; then
    echo "  go vet: PASSED"
else
    echo "  go vet: FAILED"
    exit 1
fi
echo ""
echo "Running 'go build'..."
if go build -o example_bin ./...; then
    echo "  go build: PASSED"
    rm -f example_bin
else
    echo "  go build: FAILED"
    exit 1
fi
echo ""
echo "=== Compilation Verification: PASSED ==="
exit 0
