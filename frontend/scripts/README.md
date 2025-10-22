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

### 4. Capture Screenshots with Selenium (Recommended)

For automated screenshot capture of the PWA:

```bash
# Install Python dependencies first
pip install selenium

# Capture a single screenshot
python3 scripts/selenium-screenshot.py <url> <output_path> [wait_seconds] [description]

# Example:
python3 scripts/selenium-screenshot.py http://localhost:3000 screenshots/react-login.png 20 "Login screen"
```

This Selenium-based script will:
- Use headless Chrome/Chromium to capture screenshots
- Wait for app to fully render (configurable wait time)
- Save high-quality screenshots
- Work properly with WebAssembly PWAs

**Note:** Manual screenshot tools (gnome-screenshot, scrot) don't work reliably with PWAs due to WebAssembly rendering timing issues. Always use Selenium for consistent results.

### 6. Stop All Services

Clean up when you're done:

```bash
./scripts/stop-verification.sh
```

This will kill all processes started by `start-verification.sh`.

## Script Reference

### selenium-screenshot.py

**Purpose:** Automated screenshot capture using Selenium WebDriver

**What it does:**
- Uses headless Chrome to capture screenshots
- Waits for page load completion
- Configurable wait time for WASM/PWA rendering
- Generates high-quality PNG screenshots

**Usage:**
```bash
python3 scripts/selenium-screenshot.py <url> <output_path> [wait_seconds] [description]
```

**Parameters:**
- `url` - URL to capture (required)
- `output_path` - Output file path (required)
- `wait_seconds` - Time to wait for app rendering (optional, default: 20)
- `description` - Description for logging (optional)

**Example:**
```bash
# Capture login screen with 20 second wait
python3 scripts/selenium-screenshot.py \
  http://localhost:3000 \
  verification/screenshots/react-login.png \
  20 \
  "React login screen"

# Capture dashboard with custom wait time
python3 scripts/selenium-screenshot.py \
  http://localhost:3002/dashboard \
  verification/screenshots/go-dashboard.png \
  30 \
  "Go WASM dashboard"
```

**Requirements:**
- Python 3.7+
- Selenium (`pip install selenium`)
- Chrome/Chromium browser
- ChromeDriver (usually auto-managed by Selenium)

**Why Selenium?**
- PWAs with WebAssembly need time to initialize
- Headless browser ensures consistent rendering
- Proper waiting for `document.readyState`
- No manual interaction required

---

### setup-visual-verification.sh

**Purpose:** Initial setup and prerequisite checking

**What it does:**
- Checks for Go, Node.js, npm, Firefox
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
- Python 3.7+ (for Selenium screenshots)
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

## Directory Structure

After running all scripts, you'll have:

```
frontend/
├── scripts/
│   ├── README.md                          (this file)
│   ├── selenium-screenshot.py             (Selenium screenshot tool)
│   ├── setup-visual-verification.sh       (initial setup)
│   ├── start-verification.sh              (auto-generated)
│   └── stop-verification.sh               (auto-generated)
├── verification/
│   ├── screenshots/
│   │   ├── react/                         (React screenshots)
│   │   ├── go/                            (Go WASM screenshots)
│   │   └── diff/                          (diff images - manual comparison)
│   └── reports/
│       └── verification-notes.md          (manual verification notes)
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

### Selenium WebDriver Issues

**Error:** `selenium module not found` or `WebDriver not found`

**Solution:**
```bash
# Install Selenium
pip install selenium

# Or with pip3
pip3 install selenium

# Install Chrome/Chromium browser
# Fedora
sudo dnf install chromium

# Ubuntu
sudo apt install chromium-browser
```

**ChromeDriver Issues:**
Selenium 4+ manages ChromeDriver automatically. If you encounter issues:
```bash
# Check Chrome version
chromium --version

# Selenium will download matching ChromeDriver automatically
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

### Screenshot Timing Issues

**Problem:** Screenshots are blank or show loading state

**Solution:**
Increase wait time in Selenium script:
```bash
# Default 20 seconds
python3 scripts/selenium-screenshot.py http://localhost:3000 output.png 20

# Increase to 30 seconds for slower WASM loads
python3 scripts/selenium-screenshot.py http://localhost:3002 output.png 30
```

**For PWAs:** WebAssembly apps need extra time to initialize. Always use at least 20-30 seconds wait time.

---

## Advanced Usage

### Custom Viewport Sizes with Selenium

The Selenium script uses a default 1920x1080 viewport. To customize:

**Edit `selenium-screenshot.py`:**
```python
# Change this line (around line 29)
chrome_options.add_argument('--window-size=1920,1080')

# To:
chrome_options.add_argument('--window-size=375,667')  # iPhone SE
# Or:
chrome_options.add_argument('--window-size=414,896')  # iPhone 11 Pro Max
```

**Mobile Testing:**
```bash
# Capture with mobile viewport
python3 scripts/selenium-screenshot.py http://localhost:3000 mobile-login.png 20
```

### Batch Screenshot Capture

Create a simple batch script:

```bash
#!/bin/bash
# batch-screenshots.sh

REACT_BASE="http://localhost:3000"
GO_BASE="http://localhost:3002"
WAIT_TIME=25

declare -a SCREENS=(
    "/:login"
    "/dashboard:dashboard"
    "/groups:groups"
    "/collections:collections"
    "/profile:profile"
)

for screen in "${SCREENS[@]}"; do
    IFS=':' read -r path name <<< "$screen"

    # Capture React
    python3 scripts/selenium-screenshot.py \
        "${REACT_BASE}${path}" \
        "verification/screenshots/react/${name}.png" \
        $WAIT_TIME \
        "React ${name}"

    # Capture Go
    python3 scripts/selenium-screenshot.py \
        "${GO_BASE}${path}" \
        "verification/screenshots/go/${name}.png" \
        $WAIT_TIME \
        "Go ${name}"
done
```

### CI/CD Integration

For automated testing in CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
name: Visual Verification
on: [push, pull_request]

jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.10'

      - name: Install Selenium
        run: pip install selenium

      - name: Start services
        run: ./scripts/start-verification.sh

      - name: Wait for services
        run: sleep 30

      - name: Capture screenshots
        run: |
          python3 scripts/selenium-screenshot.py \
            http://localhost:3000 \
            screenshots/react-login.png \
            30 \
            "React login"

      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: screenshots
          path: verification/screenshots/
```

## Best Practices

1. **Selenium Screenshot Capture**
   - Always use Selenium for PWA/WASM apps (never manual tools)
   - Use consistent wait times (20-30 seconds minimum)
   - Set explicit viewport sizes for consistency
   - Run in headless mode for reproducibility

2. **Consistent Environment**
   - Use same Chrome/Chromium version
   - Same viewport size across all captures
   - Disable browser extensions in test environment
   - Use consistent test data and auth state

3. **Timing Considerations**
   - WASM apps need longer initialization times
   - Increase wait time if screenshots show loading states
   - Wait for document.readyState === 'complete'
   - Add extra buffer for heavy JavaScript apps

4. **Documentation**
   - Fill out VERIFICATION_CHECKLIST.md completely
   - Document any acceptable differences
   - Save all screenshots with descriptive names
   - Keep notes on visual discrepancies

5. **Iteration**
   - Compare screenshots manually or with image diff tools
   - Fix critical visual bugs first
   - Re-run screenshots after styling changes
   - Track progress over time with dated screenshot sets

## Integration with Phase 1

This verification process is **Task 10 of 10** in Phase 1 (Foundation) of the porting plan.

**Success Criteria:**
- ✅ All critical screens captured with Selenium
- ✅ Visual comparison shows close match
- ✅ Component library is visually consistent
- ✅ Design tokens are applied correctly
- ✅ VERIFICATION_CHECKLIST.md is complete

**After Successful Verification:**
1. Mark Phase 1 as complete in PORTING_PLAN.md
2. Archive screenshots with timestamps
3. Update CLAUDE.md with insights
4. Proceed to Phase 2: Groups Management

## Screenshot Comparison Tools

While this repository includes Selenium for screenshot capture, you can use external tools for comparison:

**Command-line tools:**
```bash
# ImageMagick compare
compare react-login.png go-login.png diff.png

# Perceptual diff
perceptualdiff react-login.png go-login.png -output diff.png
```

**Online tools:**
- [Pixelmatch](https://github.com/mapbox/pixelmatch) - JavaScript pixel-level comparison
- [resemblejs](https://github.com/rsmbl/Resemble.js) - Image analysis and comparison
- Browser extensions for side-by-side comparison

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
