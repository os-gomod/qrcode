#!/bin/bash
# verify_qr_codes.sh
# Verifies QR code files in ./output/ directory
# - Checks output directory exists
# - Counts generated files
# - Verifies file sizes are not zero
# - Checks file formats (PNG, SVG, TXT)
# - Validates PNG files are valid images
# - Validates SVG files contain valid SVG markup

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/output"
MIN_FILES=30

PASS=0
FAIL=0

pass() { ((PASS++)); echo "  [PASS] $1"; }
fail() { ((FAIL++)); echo "  [FAIL] $1"; }

echo "=== QR Code File Verification ==="
echo "Output directory: $OUTPUT_DIR"
echo ""

# 1. Check output directory exists
echo "--- Directory Check ---"
if [ -d "$OUTPUT_DIR" ]; then
    pass "Output directory exists"
else
    fail "Output directory does NOT exist"
    echo ""
    echo "=== Verification: FAILED ==="
    exit 1
fi

# 2. Count files
echo ""
echo "--- File Count ---"
FILE_COUNT=$(find "$OUTPUT_DIR" -maxdepth 1 -type f | wc -l)
echo "  Total files: $FILE_COUNT (minimum expected: $MIN_FILES)"

if [ "$FILE_COUNT" -ge "$MIN_FILES" ]; then
    pass "File count >= $MIN_FILES"
else
    fail "File count < $MIN_FILES (got $FILE_COUNT)"
fi

# 3. Check for zero-byte files
echo ""
echo "--- Zero-Byte Check ---"
ZERO_COUNT=0
while IFS= read -r -d '' f; do
    if [ ! -s "$f" ]; then
        ZERO_COUNT=$((ZERO_COUNT + 1))
        fail "Zero-byte file: $(basename "$f")"
    fi
done < <(find "$OUTPUT_DIR" -maxdepth 1 -type f -print0)

if [ "$ZERO_COUNT" -eq 0 ]; then
    pass "No zero-byte files"
fi

# 4. Validate PNG files
echo ""
echo "--- PNG Validation ---"
PNG_COUNT=0
while IFS= read -r -d '' f; do
    fname=$(basename "$f")
    PNG_COUNT=$((PNG_COUNT + 1))

    # Check PNG header magic bytes
    if file "$f" | grep -qi "PNG image data"; then
        pass "$fname is a valid PNG image"
    elif hexdump -C "$f" 2>/dev/null | head -1 | grep -q "89 50 4e 47"; then
        pass "$fname has valid PNG magic bytes"
    else
        # Fallback: check if file starts with PNG signature
        FIRST_BYTES=$(od -A n -t x1 -N 8 "$f" 2>/dev/null | tr -d ' ')
        if [ "$FIRST_BYTES" = "89504e470d0a1a0a" ]; then
            pass "$fname has valid PNG header"
        else
            fail "$fname: not a valid PNG"
        fi
    fi
done < <(find "$OUTPUT_DIR" -maxdepth 1 -name "*.png" -type f -print0)

echo "  Total PNG files: $PNG_COUNT"

# 5. Validate SVG files
echo ""
echo "--- SVG Validation ---"
SVG_COUNT=0
while IFS= read -r -d '' f; do
    fname=$(basename "$f")
    SVG_COUNT=$((SVG_COUNT + 1))

    if grep -qi '<?xml' "$f" && grep -qi '<svg' "$f"; then
        pass "$fname contains valid SVG markup"
    else
        fail "$fname missing SVG markup"
    fi
done < <(find "$OUTPUT_DIR" -maxdepth 1 -name "*.svg" -type f -print0)

echo "  Total SVG files: $SVG_COUNT"

# 6. Validate TXT files
echo ""
echo "--- TXT Validation ---"
TXT_COUNT=0
while IFS= read -r -d '' f; do
    fname=$(basename "$f")
    TXT_COUNT=$((TXT_COUNT + 1))

    if [ -s "$f" ]; then
        pass "$fname is non-empty text"
    else
        fail "$fname is empty"
    fi
done < <(find "$OUTPUT_DIR" -maxdepth 1 -name "*.txt" -type f -print0)

echo "  Total TXT files: $TXT_COUNT"

# Summary
echo ""
echo "====================================================="
echo "=== Verification Summary ==="
echo "  Passed: $PASS"
echo "  Failed: $FAIL"
echo "  Total files: $FILE_COUNT"
echo "  PNG files: $PNG_COUNT"
echo "  SVG files: $SVG_COUNT"
echo "  TXT files: $TXT_COUNT"
echo "====================================================="

if [ "$FAIL" -eq 0 ]; then
    echo "=== Result: ALL CHECKS PASSED ==="
    exit 0
else
    echo "=== Result: SOME CHECKS FAILED ==="
    exit 1
fi
