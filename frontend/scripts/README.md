# Visual Verification Scripts

This directory contains automation scripts for visual verification of the Go/Cogent Core WASM frontend against the React frontend.

## Quick Start

### 1. Initial Setup

Run the setup script to check prerequisites and prepare both frontends:

```bash
cd /home/zwelch/projects/inventory/frontend
./scripts/setup-visual-verification.sh
```

This script will:
- ✅ Check all prerequisites (Go, Node.js, npm, Firefox)
- ✅ Verify project structure
- ✅ Check port availability
- ✅ Install React frontend dependencies
- ✅ Download Go dependencies
- ✅ Build the WASM frontend
- ✅ Create verification directories
- ✅ Generate quick-start helper scripts

### 2. Start All Services

Use the generated quick-start script to launch everything at once:

```bash
./scripts/start-verification.sh
```

This will:
- Start the backend API on http://localhost:3001
- Start the React frontend on http://localhost:3000
- Start the Go WASM frontend on http://localhost:[PORT]
- Open both frontends in Firefox for side-by-side comparison
- Save process IDs for easy cleanup

### 3. Manual Verification

Follow the comprehensive checklist:

```bash
# Open the checklist
cat VERIFICATION_CHECKLIST.md
# Or open in your editor
vim VERIFICATION_CHECKLIST.md
```

Work through each section, comparing React (left) vs Go WASM (right):
- Login screen
- Dashboard
- Groups list
- Collections list
- Profile screen
- Typography, colors, spacing, components

### 4. Capture Screenshots (Optional)

For automated pixel-by-pixel comparison:

```bash
./scripts/capture-screenshots.sh
```

This interactive script will:
- Prompt you to navigate to each key screen
- Capture screenshots from both frontends
- Save to `verification/screenshots/`

### 5. Generate Comparison Report (Optional)

After capturing screenshots, run the comparison:

```bash
./scripts/compare-screenshots.sh
```

This will:
- Compare React vs Go screenshots using ImageMagick
- Generate diff images highlighting pixel differences
- Calculate difference percentages
- Create a markdown report with statistics
- Save to `verification/reports/`

### 6. Stop All Services

Clean up when you're done:

```bash
./scripts/stop-verification.sh
```

This will kill all processes started by `start-verification.sh`.

## Script Reference

### setup-visual-verification.sh

**Purpose:** Initial setup and prerequisite checking

**What it does:**
- Checks for Go, Node.js, npm, Firefox, ImageMagick
- Verifies project directory structure
- Checks port availability (3000, 3001, config port)
- Installs React dependencies (`npm install`)
- Downloads Go dependencies (`go mod download`)
- Builds WASM frontend (`./bin/web`)
- Creates verification directories
- Displays configuration summary
- Generates helper scripts

**Usage:**
```bash
./scripts/setup-visual-verification.sh
```

**Output:**
- Colored status messages for each step
- Summary of configuration
- Next steps instructions
- Generated helper scripts in `scripts/`

**Requirements:**
- Go 1.21+
- Node.js 18+
- npm
- Disk space for dependencies and build artifacts

---

### start-verification.sh

**Purpose:** Start all services for verification

**What it does:**
- Starts backend API in background
- Starts React frontend in background
- Starts Go WASM frontend in background
- Opens both frontends in Firefox
- Saves PIDs to file for cleanup

**Usage:**
```bash
./scripts/start-verification.sh
```

**Output:**
- Process IDs for each service
- URLs for accessing frontends
- PID file at `logs/verification.pids`
- Log files in `logs/` directory

**Logs:**
- `logs/backend.log` - Backend API output
- `logs/react.log` - React frontend output
- `logs/go.log` - Go WASM frontend output

**Notes:**
- Services run in background
- Check logs if something doesn't work
- Use `stop-verification.sh` to clean up

---

### stop-verification.sh

**Purpose:** Stop all verification services

**What it does:**
- Reads PIDs from `logs/verification.pids`
- Kills backend, React, and Go processes
- Removes PID file

**Usage:**
```bash
./scripts/stop-verification.sh
```

**Manual cleanup (if script fails):**
```bash
# Find processes on ports
lsof -ti :3000 | xargs kill -9  # React
lsof -ti :3001 | xargs kill -9  # Backend
lsof -ti :8080 | xargs kill -9  # Go (replace 8080 with your port)
```

---

### capture-screenshots.sh

**Purpose:** Interactive screenshot capture

**What it does:**
- Prompts for each screen (login, dashboard, groups, etc.)
- Captures window screenshots using gnome-screenshot/scrot/ImageMagick
- Saves React screenshots to `verification/screenshots/react/`
- Saves Go screenshots to `verification/screenshots/go/`

**Usage:**
```bash
./scripts/capture-screenshots.sh
```

**Workflow:**
1. Script lists screens to capture
2. For each screen:
   - Navigate to the screen in React browser
   - Press Enter
   - Click on browser window to capture
   - Navigate to same screen in Go browser
   - Press Enter
   - Click on browser window to capture
3. Script saves all screenshots

**Screenshot tools (in priority order):**
- gnome-screenshot (GNOME)
- scrot (lightweight)
- ImageMagick import (universal)

**Install screenshot tool:**
```bash
# Fedora
sudo dnf install gnome-screenshot

# Ubuntu/Debian
sudo apt install scrot

# ImageMagick (all distros)
sudo dnf install ImageMagick  # Fedora
sudo apt install imagemagick  # Ubuntu
```

**Output files:**
- `verification/screenshots/react/login.png`
- `verification/screenshots/react/dashboard.png`
- `verification/screenshots/go/login.png`
- `verification/screenshots/go/dashboard.png`
- etc.

---

### compare-screenshots.sh

**Purpose:** Automated screenshot comparison

**What it does:**
- Compares React and Go screenshots pixel-by-pixel
- Generates diff images with visual highlighting
- Calculates difference percentages (RMSE metric)
- Categorizes results (Perfect < 1%, Close < 5%, Fail > 5%)
- Creates markdown report with statistics

**Usage:**
```bash
./scripts/compare-screenshots.sh
```

**Requirements:**
- ImageMagick (compare and convert commands)
- bc (calculator for percentage math)

**Install ImageMagick:**
```bash
# Fedora
sudo dnf install ImageMagick

# Ubuntu/Debian
sudo apt install imagemagick

# Arch
sudo pacman -S imagemagick
```

**Output:**
- Diff images in `verification/screenshots/diff/`
- Comparison report in `verification/reports/`
- Statistics summary in terminal

**Diff image format:**
- Red areas = pixel differences
- White/black areas = matching pixels
- Percentage overlay shows total difference

**Report includes:**
- Table of all screens with diff percentages
- Pass/fail status for each screen
- Overall statistics
- Recommendations for fixes
- Sign-off section

**Thresholds:**
- **Perfect:** < 1.0% difference
- **Close:** 1.0% - 5.0% difference
- **Fail:** > 5.0% difference

---

## Directory Structure

After running all scripts, you'll have:

```
frontend/
├── scripts/
│   ├── README.md                          (this file)
│   ├── setup-visual-verification.sh       (initial setup)
│   ├── start-verification.sh              (auto-generated)
│   ├── stop-verification.sh               (auto-generated)
│   ├── capture-screenshots.sh             (screenshot capture)
│   └── compare-screenshots.sh             (comparison)
├── verification/
│   ├── screenshots/
│   │   ├── react/                         (React screenshots)
│   │   ├── go/                            (Go WASM screenshots)
│   │   └── diff/                          (diff images)
│   └── reports/
│       └── comparison-YYYYMMDD-HHMMSS.md  (comparison reports)
├── logs/
│   ├── backend.log                        (backend output)
│   ├── react.log                          (React output)
│   ├── go.log                             (Go WASM output)
│   └── verification.pids                  (process IDs)
├── VISUAL_VERIFICATION.md                 (main documentation)
└── VERIFICATION_CHECKLIST.md              (manual checklist)
```

## Troubleshooting

### Ports Already in Use

**Error:** `Port 3000 already in use`

**Solution:**
```bash
lsof -ti :3000 | xargs kill -9
```

Repeat for ports 3001 and your Go frontend port.

---

### Backend Not Running

**Error:** `Failed to connect to backend`

**Solution:**
```bash
# Check if backend is running
curl http://localhost:3001/health

# If not, start it manually
cd /home/zwelch/projects/inventory
go run main.go
```

---

### Screenshot Tool Not Found

**Error:** `No screenshot tool found`

**Solution:**
Install a screenshot tool:
```bash
# Fedora
sudo dnf install gnome-screenshot

# Ubuntu
sudo apt install scrot
```

Or use Firefox DevTools:
1. Open DevTools (F12)
2. Press Shift+F2
3. Type: `screenshot --fullpage filename.png`

---

### ImageMagick Not Found

**Error:** `ImageMagick 'compare' command not found`

**Solution:**
```bash
# Fedora
sudo dnf install ImageMagick

# Ubuntu
sudo apt install imagemagick
```

---

### Permission Denied on Scripts

**Error:** `Permission denied: ./scripts/setup-visual-verification.sh`

**Solution:**
```bash
chmod +x ./scripts/*.sh
```

---

### WASM Build Fails

**Error:** `./bin/web: No such file or directory`

**Solution:**
Build the web tool first:
```bash
cd /home/zwelch/projects/inventory/frontend
go build -o bin/web cmd/web/main.go
go build -o bin/serve cmd/serve/main.go
```

Then run `./bin/web` to build the WASM frontend.

---

### Different Image Sizes

**Warning:** `Dimension mismatch! Resizing...`

**Explanation:** The React and Go screenshots have different dimensions. The comparison script automatically resizes the Go screenshot to match React for fair comparison.

**Action:** Review your viewport settings. Both browsers should use the same responsive design mode dimensions (e.g., 375x667 for iPhone SE).

---

## Advanced Usage

### Custom Viewport Sizes

To test different device sizes:

1. In Firefox DevTools (Ctrl+Shift+M)
2. Set custom dimensions (e.g., 375x667, 390x844, 414x896)
3. Ensure both browsers use the same size
4. Capture screenshots

### Automated Testing

For CI/CD integration, you can automate the entire process:

```bash
#!/bin/bash
# CI/CD verification script

cd /home/zwelch/projects/inventory/frontend

# Setup
./scripts/setup-visual-verification.sh

# Start services
./scripts/start-verification.sh

# Wait for services to be ready
sleep 10

# Note: Screenshot capture requires manual interaction
# For CI/CD, use headless browser automation (Selenium/Playwright)

# Stop services
./scripts/stop-verification.sh

# Check results
if [ -f verification/reports/comparison-*.md ]; then
    # Parse report and exit with appropriate code
    echo "Verification complete - check report"
fi
```

### Filtering Screens

To compare only specific screens, modify `compare-screenshots.sh`:

```bash
# Only compare login and dashboard
for screen in login dashboard; do
    # comparison logic
done
```

## Best Practices

1. **Consistent Environment**
   - Use the same browser for both frontends
   - Same viewport size and zoom level
   - Same system fonts and display settings

2. **Clean State**
   - Clear browser cache before starting
   - Use consistent test data
   - Ensure backend is in known state

3. **Timing**
   - Wait for all content to load before screenshots
   - Let animations complete
   - Ensure loading states finish

4. **Documentation**
   - Fill out VERIFICATION_CHECKLIST.md completely
   - Document any acceptable differences
   - Save all reports and screenshots

5. **Iteration**
   - Fix critical issues first (> 5% difference)
   - Review minor issues (1-5% difference)
   - Re-run verification after fixes
   - Track progress over time

## Integration with Phase 1

This verification process is **Task 10 of 10** in Phase 1 (Foundation) of the porting plan.

**Success Criteria:**
- ✅ All critical screens match exactly (< 2% difference)
- ✅ Important screens match closely (< 5% difference)
- ✅ Component library is visually consistent
- ✅ Design tokens are applied correctly
- ✅ VERIFICATION_CHECKLIST.md is complete

**After Successful Verification:**
1. Mark Phase 1 as complete in PORTING_PLAN.md
2. Archive screenshots and reports
3. Update CLAUDE.md with insights
4. Proceed to Phase 2: Groups Management

## Support

**Documentation:**
- `VISUAL_VERIFICATION.md` - Main verification documentation
- `VERIFICATION_CHECKLIST.md` - Detailed screen-by-screen checklist
- `PORTING_PLAN.md` - Overall porting plan and phases

**Troubleshooting:**
- Check logs in `logs/` directory
- Review GitHub issues for Cogent Core
- Consult CLAUDE.md for project-specific guidance

**Questions:**
- Document issues in verification/reports/
- Note discrepancies in VERIFICATION_CHECKLIST.md
- Create GitHub issues for blocking problems
