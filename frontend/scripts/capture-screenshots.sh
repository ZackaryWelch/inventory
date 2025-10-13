#!/bin/bash
# Screenshot Capture Script
# Captures screenshots of key screens from both React and Go WASM frontends

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project paths
GO_FRONTEND="/home/zwelch/projects/inventory/frontend"
OUTPUT_DIR="$GO_FRONTEND/verification/screenshots"
REACT_DIR="$OUTPUT_DIR/react"
GO_DIR="$OUTPUT_DIR/go"

# Get Go frontend port from config
GO_PORT=$(grep '^port = ' "$GO_FRONTEND/config.toml" | sed 's/port = "\(.*\)"/\1/' || echo "8080")

# URLs
REACT_URL="http://localhost:3000"
GO_URL="http://localhost:$GO_PORT"

echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}    Screenshot Capture Script${NC}"
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

# Check if screenshot tool is available
if command -v gnome-screenshot >/dev/null 2>&1; then
    SCREENSHOT_TOOL="gnome-screenshot"
    print_success "Screenshot tool found: gnome-screenshot"
elif command -v scrot >/dev/null 2>&1; then
    SCREENSHOT_TOOL="scrot"
    print_success "Screenshot tool found: scrot"
elif command -v import >/dev/null 2>&1; then
    SCREENSHOT_TOOL="import"
    print_success "Screenshot tool found: ImageMagick import"
else
    print_error "No screenshot tool found. Please install gnome-screenshot, scrot, or ImageMagick."
    print_status "Alternatively, you can use Firefox DevTools (Shift+F2, then 'screenshot --fullpage')"
    exit 1
fi

# Create output directories
mkdir -p "$REACT_DIR"
mkdir -p "$GO_DIR"
print_success "Output directories ready"
echo ""

# Screen list for verification
declare -a SCREENS=(
    "login:Login Screen"
    "callback:Callback/Loading Screen"
    "dashboard:Dashboard Screen"
    "groups:Groups List Screen"
    "collections:Collections List Screen"
    "profile:Profile Screen"
)

print_status "This script will guide you through capturing screenshots"
print_status "React frontend: $REACT_URL"
print_status "Go WASM frontend: $GO_URL"
echo ""

print_warning "IMPORTANT: You must manually navigate to each screen in both browsers"
print_warning "This script will prompt you for each screenshot"
echo ""

# Function to capture screenshot
capture_screenshot() {
    local frontend=$1  # "react" or "go"
    local screen_id=$2
    local screen_name=$3
    local output_file=$4

    echo -e "${YELLOW}=== Capture: $screen_name ($frontend) ===${NC}"
    echo "1. Navigate to: $screen_name"
    echo "2. Ensure the screen is fully loaded"
    echo "3. Press Enter when ready to capture"
    read -r

    if [ "$SCREENSHOT_TOOL" = "gnome-screenshot" ]; then
        echo "Click on the browser window to capture in 3 seconds..."
        sleep 3
        gnome-screenshot -w -f "$output_file"
    elif [ "$SCREENSHOT_TOOL" = "scrot" ]; then
        echo "Click on the browser window to capture..."
        scrot -s "$output_file"
    elif [ "$SCREENSHOT_TOOL" = "import" ]; then
        echo "Click on the browser window to capture..."
        import "$output_file"
    fi

    if [ -f "$output_file" ]; then
        print_success "Screenshot saved: $output_file"
    else
        print_error "Screenshot failed: $output_file"
    fi
    echo ""
}

# Capture screenshots for each screen
for screen_info in "${SCREENS[@]}"; do
    IFS=':' read -r screen_id screen_name <<< "$screen_info"

    # Capture React screenshot
    react_file="$REACT_DIR/${screen_id}.png"
    capture_screenshot "React" "$screen_id" "$screen_name" "$react_file"

    # Capture Go screenshot
    go_file="$GO_DIR/${screen_id}.png"
    capture_screenshot "Go WASM" "$screen_id" "$screen_name" "$go_file"
done

# Summary
echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}    Screenshot Capture Complete${NC}"
echo -e "${GREEN}==================================================${NC}"
echo ""

print_success "Screenshots saved to:"
echo "  React: $REACT_DIR"
echo "  Go WASM: $GO_DIR"
echo ""

# List captured screenshots
print_status "React screenshots:"
ls -lh "$REACT_DIR"/*.png 2>/dev/null || echo "  No screenshots found"
echo ""

print_status "Go WASM screenshots:"
ls -lh "$GO_DIR"/*.png 2>/dev/null || echo "  No screenshots found"
echo ""

print_status "Next steps:"
echo "  1. Review screenshots visually"
echo "  2. Run comparison script: ./scripts/compare-screenshots.sh"
echo "  3. Fill out VERIFICATION_CHECKLIST.md"
echo ""
