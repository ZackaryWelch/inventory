#!/bin/bash
# Automated Screenshot Comparison Script
# Uses headless Firefox to capture screenshots and ImageMagick to compare

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERIFICATION_DIR="$SCRIPT_DIR/../verification"
SCREENSHOTS_DIR="$VERIFICATION_DIR/screenshots"
COMPARISONS_DIR="$VERIFICATION_DIR/comparisons"
REPORTS_DIR="$VERIFICATION_DIR/reports"

# Create directories
mkdir -p "$SCREENSHOTS_DIR/react" "$SCREENSHOTS_DIR/go" "$COMPARISONS_DIR" "$REPORTS_DIR"

echo "=========================================="
echo "Automated Screenshot Comparison"
echo "=========================================="
echo ""

# Screen dimensions (matching typical mobile viewport)
WIDTH=414
HEIGHT=896
VIEWPORT="${WIDTH}x${HEIGHT}"

echo "Using viewport: $VIEWPORT"
echo ""

# URLs to test
REACT_BASE="http://localhost:3000"
GO_BASE="http://localhost:3002"

# Capture function
capture_screenshot() {
    local url="$1"
    local output_file="$2"
    local description="$3"

    echo "Capturing: $description"
    echo "  URL: $url"
    echo "  Output: $output_file"

    # Use Firefox headless mode with screenshot capability
    timeout 30 firefox --headless \
        --window-size=$WIDTH,$HEIGHT \
        --screenshot="$output_file" \
        "$url" 2>&1 | grep -v "Gtk-Message" || true

    if [ -f "$output_file" ]; then
        echo "  ✓ Screenshot captured successfully"
        # Crop to viewport size to ensure consistency
        convert "$output_file" -crop ${WIDTH}x${HEIGHT}+0+0 "$output_file"
    else
        echo "  ✗ Failed to capture screenshot"
        return 1
    fi

    sleep 2
}

# Compare function
compare_screenshots() {
    local react_img="$1"
    local go_img="$2"
    local diff_img="$3"
    local scene_name="$4"

    echo ""
    echo "Comparing: $scene_name"

    if [ ! -f "$react_img" ] || [ ! -f "$go_img" ]; then
        echo "  ✗ Missing screenshot files"
        return 1
    fi

    # Run comparison
    local diff_result=$(compare -metric AE -fuzz 5% \
        "$react_img" "$go_img" \
        "$diff_img" 2>&1 || true)

    # Calculate percentage difference
    local total_pixels=$((WIDTH * HEIGHT))
    local diff_pixels=$(echo "$diff_result" | grep -oE '[0-9]+' | head -1 || echo "0")
    local percent_diff=$(echo "scale=2; ($diff_pixels / $total_pixels) * 100" | bc)

    echo "  Pixel difference: $diff_pixels / $total_pixels ($percent_diff%)"

    if (( $(echo "$percent_diff < 5.0" | bc -l) )); then
        echo "  ✓ PASSED (< 5% difference)"
        return 0
    else
        echo "  ✗ FAILED (>= 5% difference)"
        return 1
    fi
}

# Test scenes
declare -A scenes=(
    ["login"]="$REACT_BASE/"
    ["dashboard"]="$REACT_BASE/dashboard"
    ["profile"]="$REACT_BASE/profile"
)

# Capture all screenshots
echo "=========================================="
echo "Step 1: Capturing Screenshots"
echo "=========================================="
echo ""

passed_tests=0
failed_tests=0
total_tests=${#scenes[@]}

for scene in "${!scenes[@]}"; do
    react_url="${scenes[$scene]}"
    go_url="$GO_BASE${react_url#$REACT_BASE}"

    react_output="$SCREENSHOTS_DIR/react/${scene}.png"
    go_output="$SCREENSHOTS_DIR/go/${scene}.png"

    # Capture React screenshot
    if capture_screenshot "$react_url" "$react_output" "React - $scene"; then
        :
    else
        ((failed_tests++))
        continue
    fi

    echo ""

    # Capture Go screenshot
    if capture_screenshot "$go_url" "$go_output" "Go/Cogent Core - $scene"; then
        :
    else
        ((failed_tests++))
        continue
    fi

    echo ""
done

# Compare screenshots
echo "=========================================="
echo "Step 2: Comparing Screenshots"
echo "=========================================="

for scene in "${!scenes[@]}"; do
    react_img="$SCREENSHOTS_DIR/react/${scene}.png"
    go_img="$SCREENSHOTS_DIR/go/${scene}.png"
    diff_img="$COMPARISONS_DIR/${scene}_diff.png"

    if compare_screenshots "$react_img" "$go_img" "$diff_img" "$scene"; then
        ((passed_tests++))
    else
        ((failed_tests++))
    fi
done

# Generate report
echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo "Total tests: $total_tests"
echo "Passed: $passed_tests"
echo "Failed: $failed_tests"
echo ""

if [ $failed_tests -eq 0 ]; then
    echo "✓ All visual tests PASSED!"
    exit 0
else
    echo "✗ Some visual tests FAILED"
    echo ""
    echo "Check the following locations for details:"
    echo "  Screenshots: $SCREENSHOTS_DIR"
    echo "  Comparisons: $COMPARISONS_DIR"
    exit 1
fi
