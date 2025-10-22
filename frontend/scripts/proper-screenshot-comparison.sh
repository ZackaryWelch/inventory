#!/bin/bash
# Proper Screenshot Comparison Script
# Waits for actual app rendering before capturing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERIFICATION_DIR="$SCRIPT_DIR/../verification"
SCREENSHOTS_DIR="$VERIFICATION_DIR/screenshots"
COMPARISONS_DIR="$VERIFICATION_DIR/comparisons"

mkdir -p "$SCREENSHOTS_DIR/react" "$SCREENSHOTS_DIR/go" "$COMPARISONS_DIR"

echo "=========================================="
echo "Proper Screenshot Comparison"
echo "=========================================="
echo ""

# Full desktop window size
WIDTH=1920
HEIGHT=1080

echo "Window size: ${WIDTH}x${HEIGHT}"
echo ""

REACT_URL="http://localhost:3000"
GO_URL="http://localhost:3002"

# Function to capture screenshot with proper waiting
capture_with_wait() {
    local url="$1"
    local output="$2"
    local description="$3"
    local wait_time="${4:-15}"  # Default 15 seconds for WASM

    echo "Capturing: $description"
    echo "  URL: $url"
    echo "  Wait time: ${wait_time}s"
    echo "  Output: $output"

    # Create a temporary HTML file that waits for content
    local temp_script="/tmp/capture_${RANDOM}.js"
    cat > "$temp_script" << 'EOF'
const puppeteer = require('puppeteer');

(async () => {
    const browser = await puppeteer.launch({
        headless: true,
        args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-dev-shm-usage']
    });

    const page = await browser.newPage();
    await page.setViewport({ width: parseInt(process.env.WIDTH), height: parseInt(process.env.HEIGHT) });

    console.log(`Navigating to ${process.env.URL}...`);
    await page.goto(process.env.URL, {
        waitUntil: 'networkidle2',
        timeout: 30000
    });

    // Wait additional time for WASM to initialize and render
    console.log(`Waiting ${process.env.WAIT_TIME}s for app to fully render...`);
    await page.waitForTimeout(parseInt(process.env.WAIT_TIME) * 1000);

    // Take screenshot
    await page.screenshot({
        path: process.env.OUTPUT,
        fullPage: false
    });

    console.log('Screenshot captured successfully');
    await browser.close();
})();
EOF

    # Try using puppeteer if available
    if command -v node &> /dev/null && npm list -g puppeteer &> /dev/null; then
        WIDTH=$WIDTH HEIGHT=$HEIGHT URL="$url" OUTPUT="$output" WAIT_TIME="$wait_time" node "$temp_script" 2>&1 | grep -v "Warning" || true
    else
        # Fallback to chromium with timeout
        echo "  Using chromium fallback (puppeteer not available)"
        timeout $((wait_time + 10)) chromium-browser \
            --headless \
            --disable-gpu \
            --no-sandbox \
            --window-size=${WIDTH},${HEIGHT} \
            --virtual-time-budget=$((wait_time * 1000)) \
            --screenshot="$output" \
            "$url" 2>&1 | grep -v "DevTools" | grep -v "Chrome" || true

        # Additional real-time wait for WASM
        sleep 2
    fi

    rm -f "$temp_script"

    if [ -f "$output" ]; then
        local file_size=$(stat -c%s "$output" 2>/dev/null || echo "0")
        echo "  ✓ Captured ($file_size bytes)"
        return 0
    else
        echo "  ✗ Failed to capture"
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
        echo "  ✗ Missing screenshots"
        return 1
    fi

    # Get dimensions
    local react_dims=$(identify -format "%wx%h" "$react_img" 2>/dev/null)
    local go_dims=$(identify -format "%wx%h" "$go_img" 2>/dev/null)

    echo "  React: $react_dims"
    echo "  Go: $go_dims"

    # Run comparison
    local diff_result=$(compare -metric AE -fuzz 5% \
        "$react_img" "$go_img" \
        "$diff_img" 2>&1 | grep -oE '[0-9]+' | head -1 || echo "999999999")

    local total_pixels=$((WIDTH * HEIGHT))
    local percent_diff=$(echo "scale=2; ($diff_result / $total_pixels) * 100" | bc 2>/dev/null || echo "100")

    echo "  Different pixels: $diff_result / $total_pixels ($percent_diff%)"

    if (( $(echo "$percent_diff < 5.0" | bc -l) )); then
        echo "  ✓ PASSED (< 5% difference)"
        return 0
    else
        echo "  ✗ FAILED (>= 5% difference)"
        return 1
    fi
}

echo "=========================================="
echo "Step 1: Capturing Screenshots"
echo "=========================================="
echo ""

# Capture React (quick load, it's Next.js)
if capture_with_wait "$REACT_URL" "$SCREENSHOTS_DIR/react/login.png" "React Login" 8; then
    :
fi

echo ""
echo "---"
echo ""

# Capture Go (needs longer for WASM)
if capture_with_wait "$GO_URL" "$SCREENSHOTS_DIR/go/login.png" "Go/WASM Login" 15; then
    :
fi

echo ""
echo "=========================================="
echo "Step 2: Comparing Screenshots"
echo "=========================================="

passed=0
failed=0

if compare_screenshots \
    "$SCREENSHOTS_DIR/react/login.png" \
    "$SCREENSHOTS_DIR/go/login.png" \
    "$COMPARISONS_DIR/login_diff.png" \
    "Login Screen"; then
    ((passed++))
else
    ((failed++))
fi

echo ""
echo "=========================================="
echo "Results"
echo "=========================================="
echo "Passed: $passed"
echo "Failed: $failed"
echo ""
echo "Screenshots: $SCREENSHOTS_DIR"
echo "Comparisons: $COMPARISONS_DIR"
echo ""

if [ $failed -eq 0 ]; then
    echo "✓ Visual comparison PASSED"
    exit 0
else
    echo "✗ Visual comparison FAILED"
    echo ""
    echo "Review difference images in: $COMPARISONS_DIR"
    exit 1
fi
