#!/bin/bash
# Screenshot Comparison Script
# Compares React and Go WASM screenshots using ImageMagick

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project paths
GO_FRONTEND="/home/zwelch/projects/inventory/frontend"
SCREENSHOTS_DIR="$GO_FRONTEND/verification/screenshots"
REACT_DIR="$SCREENSHOTS_DIR/react"
GO_DIR="$SCREENSHOTS_DIR/go"
DIFF_DIR="$SCREENSHOTS_DIR/diff"
REPORTS_DIR="$GO_FRONTEND/verification/reports"

echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}    Screenshot Comparison Script${NC}"
echo -e "${BLUE}==================================================${NC}"
echo ""

# Print functions
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if ImageMagick is installed
if ! command -v compare >/dev/null 2>&1; then
    print_error "ImageMagick 'compare' command not found"
    print_status "Install ImageMagick: sudo dnf install ImageMagick (Fedora)"
    print_status "                    sudo apt install imagemagick (Ubuntu/Debian)"
    exit 1
fi

if ! command -v convert >/dev/null 2>&1; then
    print_error "ImageMagick 'convert' command not found"
    exit 1
fi

print_success "ImageMagick is installed"
echo ""

# Create output directories
mkdir -p "$DIFF_DIR"
mkdir -p "$REPORTS_DIR"

# Check if screenshots exist
if [ ! -d "$REACT_DIR" ] || [ -z "$(ls -A "$REACT_DIR" 2>/dev/null)" ]; then
    print_error "No React screenshots found in $REACT_DIR"
    print_status "Run ./scripts/capture-screenshots.sh first"
    exit 1
fi

if [ ! -d "$GO_DIR" ] || [ -z "$(ls -A "$GO_DIR" 2>/dev/null)" ]; then
    print_error "No Go WASM screenshots found in $GO_DIR"
    print_status "Run ./scripts/capture-screenshots.sh first"
    exit 1
fi

print_success "Screenshot directories found"
echo ""

# Initialize report file
REPORT_FILE="$REPORTS_DIR/comparison-$(date +%Y%m%d-%H%M%S).md"
cat > "$REPORT_FILE" << EOF
# Visual Comparison Report

**Generated:** $(date)

## Summary

| Screen | Difference % | Status | Notes |
|--------|-------------|--------|-------|
EOF

print_status "Generating comparison report: $REPORT_FILE"
echo ""

# Comparison threshold (percentage)
THRESHOLD_PERFECT=1.0
THRESHOLD_CLOSE=5.0

# Counters
TOTAL=0
PERFECT=0
CLOSE=0
FAIL=0

# Compare each screenshot
for react_file in "$REACT_DIR"/*.png; do
    if [ ! -f "$react_file" ]; then
        continue
    fi

    filename=$(basename "$react_file")
    screen_name="${filename%.png}"
    go_file="$GO_DIR/$filename"
    diff_file="$DIFF_DIR/${screen_name}-diff.png"
    metrics_file="$DIFF_DIR/${screen_name}-metrics.txt"

    TOTAL=$((TOTAL + 1))

    echo -e "${BLUE}Comparing: $screen_name${NC}"

    if [ ! -f "$go_file" ]; then
        print_warning "Go WASM screenshot not found: $filename"
        echo "| $screen_name | N/A | ⏭️ SKIP | Go screenshot missing |" >> "$REPORT_FILE"
        continue
    fi

    # Get image dimensions
    react_dims=$(identify -format "%wx%h" "$react_file")
    go_dims=$(identify -format "%wx%h" "$go_file")

    print_status "React: $react_dims"
    print_status "Go WASM: $go_dims"

    # Check if dimensions match
    if [ "$react_dims" != "$go_dims" ]; then
        print_warning "Dimension mismatch! Resizing Go screenshot to match React..."

        # Resize Go screenshot to match React dimensions
        go_file_resized="$DIFF_DIR/${screen_name}-go-resized.png"
        convert "$go_file" -resize "$react_dims!" "$go_file_resized"
        go_file="$go_file_resized"

        print_status "Resized to: $react_dims"
    fi

    # Compare images and calculate difference
    print_status "Running comparison..."

    # Use compare with metric output
    compare -metric RMSE "$react_file" "$go_file" "$diff_file" 2> "$metrics_file" || true

    # Extract difference percentage from metrics
    if [ -f "$metrics_file" ]; then
        # Parse RMSE output: "1234.56 (0.0123)" where 0.0123 is the normalized difference
        diff_raw=$(cat "$metrics_file" | grep -oP '\(.*\)' | tr -d '()' || echo "0")

        # Convert to percentage (multiply by 100)
        diff_percent=$(echo "$diff_raw * 100" | bc -l)

        # Round to 2 decimal places
        diff_percent=$(printf "%.2f" "$diff_percent")

        print_status "Difference: ${diff_percent}%"

        # Categorize result
        if (( $(echo "$diff_percent < $THRESHOLD_PERFECT" | bc -l) )); then
            status="✅ PERFECT"
            PERFECT=$((PERFECT + 1))
            print_success "Perfect match!"
        elif (( $(echo "$diff_percent < $THRESHOLD_CLOSE" | bc -l) )); then
            status="⚠️ CLOSE"
            CLOSE=$((CLOSE + 1))
            print_warning "Close match with minor differences"
        else
            status="❌ FAIL"
            FAIL=$((FAIL + 1))
            print_error "Significant visual differences detected"
        fi

        # Add to report
        echo "| $screen_name | ${diff_percent}% | $status | See diff image |" >> "$REPORT_FILE"

        # Create annotated diff image with percentage overlay
        convert "$diff_file" \
            -gravity North \
            -pointsize 20 \
            -fill red \
            -annotate +0+10 "Difference: ${diff_percent}%" \
            "$diff_file"

        print_success "Diff image saved: $diff_file"
    else
        print_error "Failed to calculate difference"
        echo "| $screen_name | ERROR | ❌ FAIL | Comparison failed |" >> "$REPORT_FILE"
    fi

    echo ""
done

# Finalize report
cat >> "$REPORT_FILE" << EOF

## Statistics

- **Total Screens:** $TOTAL
- **Perfect Match (✅):** $PERFECT (< ${THRESHOLD_PERFECT}% difference)
- **Close Match (⚠️):** $CLOSE (${THRESHOLD_PERFECT}%-${THRESHOLD_CLOSE}% difference)
- **Failed (❌):** $FAIL (> ${THRESHOLD_CLOSE}% difference)

## Overall Visual Parity

EOF

if [ $TOTAL -eq 0 ]; then
    echo "**No screens compared**" >> "$REPORT_FILE"
    overall_percent="N/A"
else
    success_count=$((PERFECT + CLOSE))
    overall_percent=$(echo "scale=2; $success_count * 100 / $TOTAL" | bc)
    echo "**${overall_percent}%** of screens passed verification" >> "$REPORT_FILE"
fi

cat >> "$REPORT_FILE" << EOF

## Diff Images

All difference images are saved in:
\`verification/screenshots/diff/\`

Red areas in diff images indicate pixel differences.
White/black areas indicate matching pixels.

## Recommendations

EOF

if [ $FAIL -gt 0 ]; then
    cat >> "$REPORT_FILE" << EOF
### Critical Issues

Review the following screens with significant differences:

EOF
    for react_file in "$REACT_DIR"/*.png; do
        if [ ! -f "$react_file" ]; then
            continue
        fi
        filename=$(basename "$react_file")
        screen_name="${filename%.png}"
        metrics_file="$DIFF_DIR/${screen_name}-metrics.txt"

        if [ -f "$metrics_file" ]; then
            diff_raw=$(cat "$metrics_file" | grep -oP '\(.*\)' | tr -d '()' || echo "0")
            diff_percent=$(echo "$diff_raw * 100" | bc -l)
            diff_percent=$(printf "%.2f" "$diff_percent")

            if (( $(echo "$diff_percent >= $THRESHOLD_CLOSE" | bc -l) )); then
                echo "- **${screen_name}**: ${diff_percent}% difference" >> "$REPORT_FILE"
            fi
        fi
    done

    cat >> "$REPORT_FILE" << EOF

### Action Items

1. Review diff images for failed screens
2. Identify root cause of differences (layout, colors, spacing, etc.)
3. Update styles in \`ui/styles/\` modules
4. Re-run verification after fixes

EOF
fi

if [ $CLOSE -gt 0 ]; then
    cat >> "$REPORT_FILE" << EOF
### Minor Differences

The following screens have minor differences that should be reviewed:

EOF
    for react_file in "$REACT_DIR"/*.png; do
        if [ ! -f "$react_file" ]; then
            continue
        fi
        filename=$(basename "$react_file")
        screen_name="${filename%.png}"
        metrics_file="$DIFF_DIR/${screen_name}-metrics.txt"

        if [ -f "$metrics_file" ]; then
            diff_raw=$(cat "$metrics_file" | grep -oP '\(.*\)' | tr -d '()' || echo "0")
            diff_percent=$(echo "$diff_raw * 100" | bc -l)
            diff_percent=$(printf "%.2f" "$diff_percent")

            if (( $(echo "$diff_percent >= $THRESHOLD_PERFECT && $diff_percent < $THRESHOLD_CLOSE" | bc -l) )); then
                echo "- **${screen_name}**: ${diff_percent}% difference" >> "$REPORT_FILE"
            fi
        fi
    done

    cat >> "$REPORT_FILE" << EOF

These differences may be acceptable if they are due to:
- Font rendering variations
- Anti-aliasing differences
- Minor spacing variations (< 2px)

EOF
fi

cat >> "$REPORT_FILE" << EOF

## Sign-off

**Verified By:** ________________
**Date:** $(date +%Y-%m-%d)
**Status:** [ ] PASS [ ] NEEDS FIXES [ ] FAIL

EOF

# Print summary
echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}    Comparison Complete${NC}"
echo -e "${GREEN}==================================================${NC}"
echo ""

print_success "Comparison report generated: $REPORT_FILE"
echo ""

print_status "Summary:"
echo "  Total screens: $TOTAL"
echo "  Perfect match: $PERFECT (< ${THRESHOLD_PERFECT}%)"
echo "  Close match: $CLOSE (${THRESHOLD_PERFECT}%-${THRESHOLD_CLOSE}%)"
echo "  Failed: $FAIL (> ${THRESHOLD_CLOSE}%)"
echo ""

if [ $TOTAL -gt 0 ]; then
    echo "  Overall: ${overall_percent}% passed"
    echo ""
fi

print_status "Diff images saved to: $DIFF_DIR"
echo ""

print_status "Next steps:"
echo "  1. Open report: $REPORT_FILE"
echo "  2. Review diff images in: $DIFF_DIR"
echo "  3. Complete VERIFICATION_CHECKLIST.md"
if [ $FAIL -gt 0 ]; then
    echo "  4. Fix styling issues in ui/styles/"
    echo "  5. Re-run verification"
fi
echo ""

# Open report in default text editor (optional)
if command -v xdg-open >/dev/null 2>&1; then
    read -p "Open report now? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        xdg-open "$REPORT_FILE" &
    fi
fi
