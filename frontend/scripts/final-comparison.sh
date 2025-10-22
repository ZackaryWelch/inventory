#!/bin/bash
# Final Visual Comparison using Selenium with Real Waits

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERIFICATION_DIR="$SCRIPT_DIR/../verification"
SCREENSHOTS_DIR="$VERIFICATION_DIR/screenshots-real"
COMPARISONS_DIR="$VERIFICATION_DIR/comparisons-real"
REPORTS_DIR="$VERIFICATION_DIR/reports"

mkdir -p "$SCREENSHOTS_DIR/react" "$SCREENSHOTS_DIR/go" "$COMPARISONS_DIR" "$REPORTS_DIR"

echo "=========================================="
echo "Visual Comparison - Selenium (Real Waits)"
echo "=========================================="
echo ""

REACT_URL="http://localhost:3000/"
GO_URL="http://localhost:3002/"
REACT_OUT="$SCREENSHOTS_DIR/react/login.png"
GO_OUT="$SCREENSHOTS_DIR/go/login.png"
DIFF_OUT="$COMPARISONS_DIR/login_diff.png"

echo "Step 1: Capture React Frontend"
echo "---"
python3 "$SCRIPT_DIR/selenium-screenshot.py" \
    "$REACT_URL" \
    "$REACT_OUT" \
    10 \
    "React Login (localhost:3000)"

echo ""
echo "Step 2: Capture Go/WASM Frontend"
echo "---"
python3 "$SCRIPT_DIR/selenium-screenshot.py" \
    "$GO_URL" \
    "$GO_OUT" \
    20 \
    "Go/Cogent Core Login (localhost:3002)"

echo ""
echo "=========================================="
echo "Step 3: Visual Comparison"
echo "=========================================="
echo ""

if [ ! -f "$REACT_OUT" ] || [ ! -f "$GO_OUT" ]; then
    echo "✗ Missing screenshots"
    exit 1
fi

# Get dimensions
react_dims=$(identify -format "%wx%h" "$REACT_OUT")
go_dims=$(identify -format "%wx%h" "$GO_OUT")
echo "React dimensions: $react_dims"
echo "Go dimensions: $go_dims"

# Compare with ImageMagick
echo ""
echo "Comparing pixels..."
diff_count=$(compare -metric AE -fuzz 5% "$REACT_OUT" "$GO_OUT" "$DIFF_OUT" 2>&1 || echo "999999999")
diff_pixels=$(echo "$diff_count" | grep -oE '^[0-9]+' || echo "999999999")

total_pixels=$((1920 * 1080))
percent=$(echo "scale=2; ($diff_pixels / $total_pixels) * 100" | bc)

echo ""
echo "Different pixels: $diff_pixels / $total_pixels"
echo "Difference: $percent%"
echo ""

if (( $(echo "$percent < 5.0" | bc -l) )); then
    status="✓ PASSED"
    exit_code=0
else
    status="✗ FAILED"
    exit_code=1
fi

echo "Result: $status"
echo ""

# Generate report
REPORT_FILE="$REPORTS_DIR/selenium_comparison_$(date +%Y%m%d_%H%M%S).txt"
cat > "$REPORT_FILE" << EOREPORT
Visual Comparison Report (Selenium)
====================================
Date: $(date)
Tool: Selenium WebDriver with Python 3

URLs Compared:
- React:  $REACT_URL
- Go/WASM: $GO_URL

Wait Times:
- React: 10 seconds
- Go/WASM: 20 seconds

Screenshots:
- React: $REACT_OUT ($react_dims)
- Go:    $GO_OUT ($go_dims)

Comparison:
- Different pixels: $diff_pixels / $total_pixels
- Percentage: $percent%
- Diff image: $DIFF_OUT

Result: $status
EOREPORT

echo "Report saved: $REPORT_FILE"
echo "Screenshots: $SCREENSHOTS_DIR"
echo "Diff image: $DIFF_OUT"
echo ""

if [ $exit_code -eq 0 ]; then
    echo "✓ Visual comparison PASSED (< 5% difference)"
else
    echo "✗ Visual comparison FAILED (>= 5% difference)"
    echo ""
    echo "Next steps:"
    echo "1. Open $REACT_OUT and $GO_OUT side-by-side"
    echo "2. Review $DIFF_OUT to see highlighted differences"
    echo "3. Identify specific styling issues in Go/Cogent Core frontend"
fi

exit $exit_code
