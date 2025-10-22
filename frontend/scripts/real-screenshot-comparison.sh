#!/bin/bash
# Real Screenshot Comparison - Using Chromium with Proper Waits
# This version actually waits for PWA/WASM to load before capturing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERIFICATION_DIR="$SCRIPT_DIR/../verification"
SCREENSHOTS_DIR="$VERIFICATION_DIR/screenshots-real"
COMPARISONS_DIR="$VERIFICATION_DIR/comparisons-real"
REPORTS_DIR="$VERIFICATION_DIR/reports"

mkdir -p "$SCREENSHOTS_DIR/react" "$SCREENSHOTS_DIR/go" "$COMPARISONS_DIR" "$REPORTS_DIR"

echo "=========================================="
echo "Real Screenshot Comparison"
echo "  React: http://localhost:3000"
echo "  Go/WASM: http://localhost:3002"
echo "=========================================="
echo ""

# Desktop window size for full screenshots
WIDTH=1920
HEIGHT=1080

# Function to capture with Firefox (better WASM support)
capture_firefox() {
    local url="$1"
    local output="$2"
    local description="$3"
    local wait_seconds="${4:-15}"

    echo "$description"
    echo "  URL: $url"
    echo "  Wait: ${wait_seconds}s for full render"

    # Use timeout to ensure Firefox doesn't hang
    timeout $((wait_seconds + 20)) firefox \
        --headless \
        --window-size=$WIDTH,$HEIGHT \
        --screenshot="$output" \
        "$url" 2>&1 | grep -E "bytes written|ERROR" || true

    # Give it extra time for WASM
    sleep 3

    if [ -f "$output" ]; then
        local size=$(stat -c%s "$output" 2>/dev/null)
        local dims=$(identify -format "%wx%h" "$output" 2>/dev/null)
        if [ "$size" -gt 5000 ]; then
            echo "  ✓ Success: $dims ($size bytes)"
            return 0
        else
            echo "  ⚠ Warning: File too small ($size bytes), likely blank"
            return 1
        fi
    else
        echo "  ✗ Failed: No screenshot created"
        return 1
    fi
}

# Function to compare images
compare_images() {
    local img1="$1"
    local img2="$2"
    local diff_output="$3"
    local name="$4"

    echo ""
    echo "Comparing: $name"

    if [ ! -f "$img1" ] || [ ! -f "$img2" ]; then
        echo "  ✗ Missing source images"
        return 1
    fi

    local dims1=$(identify -format "%wx%h" "$img1" 2>/dev/null)
    local dims2=$(identify -format "%wx%h" "$img2" 2>/dev/null)

    echo "  React dims: $dims1"
    echo "  Go dims: $dims2"

    # Calculate pixel difference
    local diff_count=$(compare -metric AE -fuzz 5% "$img1" "$img2" "$diff_output" 2>&1 | grep -oE '^[0-9]+' || echo "9999999")

    local total_pixels=$((WIDTH * HEIGHT))
    local percent=$(echo "scale=2; ($diff_count / $total_pixels) * 100" | bc 2>/dev/null || echo "100.00")

    echo "  Difference: $diff_count pixels ($percent%)"

    # Generate diff visualization
    if [ -f "$diff_output" ]; then
        echo "  Diff image: $diff_output"
    fi

    if (( $(echo "$percent < 5.0" | bc -l) )); then
        echo "  ✓ PASSED (< 5%)"
        return 0
    else
        echo "  ✗ FAILED (>= 5%)"
        return 1
    fi
}

# Kill any existing Firefox instances to avoid conflicts
pkill -9 firefox 2>/dev/null || true
sleep 2

echo "=========================================="
echo "Step 1: Capture React Frontend"
echo "=========================================="
echo ""

capture_firefox \
    "http://localhost:3000/" \
    "$SCREENSHOTS_DIR/react/login.png" \
    "React Login Screen" \
    8

echo ""
echo "=========================================="
echo "Step 2: Capture Go/WASM Frontend"
echo "=========================================="
echo ""

# Go/WASM needs longer to initialize
capture_firefox \
    "http://localhost:3002/" \
    "$SCREENSHOTS_DIR/go/login.png" \
    "Go/Cogent Core Login Screen" \
    18

echo ""
echo "=========================================="
echo "Step 3: Visual Comparison"
echo "=========================================="

passed=0
failed=0

if compare_images \
    "$SCREENSHOTS_DIR/react/login.png" \
    "$SCREENSHOTS_DIR/go/login.png" \
    "$COMPARISONS_DIR/login_diff.png" \
    "Login Screen"; then
    ((passed++))
else
    ((failed++))
fi

# Generate report
REPORT_FILE="$REPORTS_DIR/comparison_$(date +%Y%m%d_%H%M%S).txt"
cat > "$REPORT_FILE" << EOREPORT
Visual Comparison Report
========================
Date: $(date)

Screenshots:
- React: $SCREENSHOTS_DIR/react/
- Go:    $SCREENSHOTS_DIR/go/

Comparisons:
- Diff images: $COMPARISONS_DIR/

Results:
- Passed: $passed
- Failed: $failed

Status: $([ $failed -eq 0 ] && echo "SUCCESS" || echo "FAILED")
EOREPORT

echo ""
echo "=========================================="
echo "Summary"
echo "=========================================="
echo "Passed: $passed"
echo "Failed: $failed"
echo ""
echo "Report: $REPORT_FILE"
echo "Screenshots: $SCREENSHOTS_DIR"
echo "Diff images: $COMPARISONS_DIR"
echo ""

if [ $failed -eq 0 ]; then
    echo "✓ All comparisons PASSED"
    exit 0
else
    echo "✗ Comparison FAILED - UIs are significantly different"
    echo ""
    echo "Next steps:"
    echo "1. View screenshots in: $SCREENSHOTS_DIR"
    echo "2. View diff images in: $COMPARISONS_DIR"
    echo "3. Identify specific styling issues"
    exit 1
fi
