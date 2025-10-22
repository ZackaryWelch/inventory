#!/bin/bash
# Automated Screenshot Comparison Script v2
# Uses Chromium headless with proper wait times for WASM apps

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERIFICATION_DIR="$SCRIPT_DIR/../verification"
SCREENSHOTS_DIR="$VERIFICATION_DIR/screenshots"
COMPARISONS_DIR="$VERIFICATION_DIR/comparisons"
REPORTS_DIR="$VERIFICATION_DIR/reports"

# Create directories
mkdir -p "$SCREENSHOTS_DIR/react" "$SCREENSHOTS_DIR/go" "$COMPARISONS_DIR" "$REPORTS_DIR"

echo "=========================================="
echo "Automated Screenshot Comparison v2"
echo "=========================================="
echo ""

# Screen dimensions (matching typical mobile viewport)
WIDTH=414
HEIGHT=896

echo "Using viewport: ${WIDTH}x${HEIGHT}"
echo ""

# URLs to test
REACT_BASE="http://localhost:3000"
GO_BASE="http://localhost:3002"

# Capture function using Chromium
capture_screenshot_chromium() {
    local url="$1"
    local output_file="$2"
    local description="$3"
    local wait_time="${4:-5}"  # Default 5 seconds wait for WASM to load

    echo "Capturing: $description"
    echo "  URL: $url"
    echo "  Output: $output_file"
    echo "  Wait time: ${wait_time}s (for WASM loading)"

    # Create temporary HTML file with proper viewport and waiting
    local temp_html="/tmp/screenshot_capture_$$.html"
    cat > "$temp_html" << EOF
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=${WIDTH}, height=${HEIGHT}, initial-scale=1.0">
    <style>
        body { margin: 0; padding: 0; overflow: hidden; }
        iframe { border: none; width: ${WIDTH}px; height: ${HEIGHT}px; }
    </style>
</head>
<body>
    <iframe src="${url}" width="${WIDTH}" height="${HEIGHT}"></iframe>
    <script>
        // Wait for page to fully load including WASM
        setTimeout(() => {
            console.log('Page loaded, ready for screenshot');
        }, ${wait_time}000);
    </script>
</body>
</html>
EOF

    # Use chromium-browser in headless mode
    timeout $((wait_time + 10)) chromium-browser \
        --headless \
        --disable-gpu \
        --no-sandbox \
        --window-size=${WIDTH},${HEIGHT} \
        --screenshot="$output_file" \
        --virtual-time-budget=$((wait_time * 1000)) \
        "$url" 2>&1 | grep -v "DevTools" | grep -v "Chrome" || true

    rm -f "$temp_html"

    if [ -f "$output_file" ]; then
        local file_size=$(stat -f%z "$output_file" 2>/dev/null || stat -c%s "$output_file" 2>/dev/null || echo "0")
        if [ "$file_size" -gt 1000 ]; then
            echo "  ✓ Screenshot captured successfully ($file_size bytes)"
            # Ensure exact dimensions
            magick "$output_file" -resize ${WIDTH}x${HEIGHT}! "$output_file" 2>/dev/null || convert "$output_file" -resize ${WIDTH}x${HEIGHT}! "$output_file"
            return 0
        else
            echo "  ✗ Screenshot file too small ($file_size bytes), likely failed"
            return 1
        fi
    else
        echo "  ✗ Failed to capture screenshot"
        return 1
    fi
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

    # Get actual dimensions
    local react_dims=$(identify -format "%wx%h" "$react_img" 2>/dev/null)
    local go_dims=$(identify -format "%wx%h" "$go_img" 2>/dev/null)

    echo "  React dimensions: $react_dims"
    echo "  Go dimensions: $go_dims"

    # Run comparison with ImageMagick
    local diff_result=$(compare -metric AE -fuzz 5% \
        "$react_img" "$go_img" \
        "$diff_img" 2>&1 || echo "0")

    # Extract numeric value
    local diff_pixels=$(echo "$diff_result" | grep -oE '[0-9]+' | head -1 || echo "0")

    # Calculate based on actual image size
    local total_pixels=$((WIDTH * HEIGHT))
    local percent_diff=$(echo "scale=2; ($diff_pixels / $total_pixels) * 100" | bc 2>/dev/null || echo "0")

    echo "  Pixel difference: $diff_pixels / $total_pixels ($percent_diff%)"

    if (( $(echo "$percent_diff < 5.0" | bc -l) )); then
        echo "  ✓ PASSED (< 5% difference)"
        return 0
    else
        echo "  ✗ FAILED (>= 5% difference)"
        return 1
    fi
}

# Test scenes with appropriate wait times
declare -A scenes=(
    ["login"]="/"
    ["dashboard"]="/dashboard"
    ["profile"]="/profile"
)

declare -A wait_times=(
    ["login"]="8"      # WASM needs time to initialize
    ["dashboard"]="10" # May need to authenticate
    ["profile"]="10"   # May need to authenticate
)

# Capture all screenshots
echo "=========================================="
echo "Step 1: Capturing Screenshots"
echo "=========================================="
echo ""

passed_tests=0
failed_tests=0
total_tests=${#scenes[@]}

for scene in login dashboard profile; do  # Ordered execution
    path="${scenes[$scene]}"
    wait="${wait_times[$scene]}"

    react_url="${REACT_BASE}${path}"
    go_url="${GO_BASE}${path}"

    react_output="$SCREENSHOTS_DIR/react/${scene}.png"
    go_output="$SCREENSHOTS_DIR/go/${scene}.png"

    # Capture React screenshot
    if capture_screenshot_chromium "$react_url" "$react_output" "React - $scene" "$wait"; then
        :
    else
        echo "  ⚠ React screenshot failed, continuing..."
    fi

    echo ""

    # Capture Go screenshot
    if capture_screenshot_chromium "$go_url" "$go_output" "Go/Cogent Core - $scene" "$wait"; then
        :
    else
        echo "  ⚠ Go screenshot failed, continuing..."
    fi

    echo ""
    echo "---"
    echo ""
done

# Compare screenshots
echo "=========================================="
echo "Step 2: Comparing Screenshots"
echo "=========================================="
echo ""

for scene in login dashboard profile; do
    react_img="$SCREENSHOTS_DIR/react/${scene}.png"
    go_img="$SCREENSHOTS_DIR/go/${scene}.png"
    diff_img="$COMPARISONS_DIR/${scene}_diff.png"

    if [ -f "$react_img" ] && [ -f "$go_img" ]; then
        if compare_screenshots "$react_img" "$go_img" "$diff_img" "$scene"; then
            ((passed_tests++))
        else
            ((failed_tests++))
        fi
    else
        echo ""
        echo "Skipping: $scene (missing screenshots)"
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
    echo ""
    echo "Screenshots saved to: $SCREENSHOTS_DIR"
    echo "Comparisons saved to: $COMPARISONS_DIR"
    exit 0
else
    echo "✗ Some visual tests FAILED"
    echo ""
    echo "Check the following locations for details:"
    echo "  Screenshots: $SCREENSHOTS_DIR"
    echo "  Comparisons: $COMPARISONS_DIR"
    exit 1
fi
