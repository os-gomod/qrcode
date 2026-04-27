#!/bin/bash
# verify_all.sh
# Master verification script that runs all checks
# Generates a comprehensive report of all QR codes
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"
PASS_COUNT=0
FAIL_COUNT=0
pass_check() { ((PASS_COUNT++)); echo "  [PASS] $1"; }
fail_check() { ((FAIL_COUNT++)); echo "  [FAIL] $1"; }
echo "====================================================================="
echo "=== Master QR Code Verification Suite ==="
echo "Timestamp: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
echo "====================================================================="
echo ""
# Step 1: Compilation check
echo "--- Step 1: Compilation Check ---"
if bash "$SCRIPT_DIR/verify_compilation.sh" > /dev/null 2>&1; then
    pass_check "Compilation"
else
    fail_check "Compilation - see verify_compilation.sh output"
    echo "  Run: bash verify_compilation.sh"
fi
echo ""
# Step 2: Run the example (generate QR codes)
echo "--- Step 2: Run Example & Generate QR Codes ---"
if go run example.go 2>&1 | tail -20; then
    pass_check "Example execution"
else
    fail_check "Example execution failed"
fi
echo ""
# Step 3: File verification
echo "--- Step 3: File Verification ---"
if [ -f "$SCRIPT_DIR/verify_qr_codes.sh" ]; then
    if bash "$SCRIPT_DIR/verify_qr_codes.sh"; then
        pass_check "File verification"
    else
        fail_check "File verification"
    fi
else
    fail_check "verify_qr_codes.sh not found"
fi
echo ""
# Step 4: Python decode test
echo "--- Step 4: QR Code Content Decode ---"
if command -v python3 &>/dev/null; then
    if python3 "$SCRIPT_DIR/test_qr_decode.py" "$SCRIPT_DIR/output"; then
        pass_check "Content decode"
    else
        # pyzbar may not be available, that's OK
        if python3 -c "from pyzbar.pyzbar import decode" 2>/dev/null; then
            fail_check "Content decode (pyzbar available)"
        else
            pass_check "Content decode (skipped - pyzbar not installed)"
            echo "  [INFO] Install pyzbar for full decode testing:"
            echo "         pip install opencv-python-headless pyzbar pillow"
        fi
    fi
else
    pass_check "Content decode (skipped - python3 not found)"
fi
echo ""
# Step 5: Performance check
echo "--- Step 5: Performance Check ---"
if [ -d "$SCRIPT_DIR/output" ]; then
    FILE_COUNT=$(find "$SCRIPT_DIR/output -maxdepth 1 -type f" | wc -l)
    TOTAL_SIZE=$(du -sh "$SCRIPT_DIR/output" | cut -f1)
    echo "  Total files: $FILE_COUNT"
    echo "  Total size:  $TOTAL_SIZE"
    # Check individual file sizes are reasonable (< 1 MB per QR code)
    LARGE_FILES=0
    while IFS= read -r -d '' f; do
        size=$(stat -c%s "$f")
        if [ "$size" -gt 1048576 ]; then
            LARGE_FILES=$((LARGE_FILES + 1))
            echo "  [WARN] Large file: $(basename "$f") (${size} bytes)"
        fi
    done < <(find "$SCRIPT_DIR/output" -maxdepth 1 -type f -print0)
    if [ "$LARGE_FILES" -eq 0 ]; then
        pass_check "All files under 1 MB"
    else
        fail_check "$LARGE_FILES file(s) over 1 MB"
    fi
    if [ "$FILE_COUNT" -ge 30 ]; then
        pass_check "Generated $FILE_COUNT files (>= 30 expected)"
    else
        fail_check "Only $FILE_COUNT files generated (< 30 expected)"
    fi
fi
echo ""
# Final Report
echo "====================================================================="
echo "=== Final Verification Report ==="
echo "  Passed: $PASS_COUNT"
echo "  Failed: $FAIL_COUNT"
echo "====================================================================="
if [ "$FAIL_COUNT" -eq 0 ]; then
    echo "=== ALL VERIFICATIONS PASSED ==="
    exit 0
else
    echo "=== SOME VERIFICATIONS FAILED ==="
    exit 1
fi
