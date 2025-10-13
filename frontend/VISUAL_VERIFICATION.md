# Visual Verification Plan

## Overview

This document outlines the process for verifying pixel-perfect visual parity between the React frontend (`nishiki-frontend/`) and the Go/Cogent Core WASM frontend (`frontend/`).

## Prerequisites

### React Frontend Requirements
- Node.js 18+ and npm
- Backend API running on `http://localhost:3001`
- Authentik OIDC provider configured

### Go Frontend Requirements
- Go 1.21+
- Cogent Core dependencies
- WebAssembly build tools
- Backend API running on `http://localhost:3001`
- Authentik OIDC provider configured

### Verification Tools
- Firefox (recommended for side-by-side comparison)
- Screenshot tool (scrot, gnome-screenshot, or Firefox DevTools)
- Image comparison tool (optional: ImageMagick for automated diff)

## Setup Process

### Automated Setup

Run the automated setup script:

```bash
cd /home/zwelch/projects/inventory/frontend
./scripts/setup-visual-verification.sh
```

This script will:
1. Check prerequisites
2. Install dependencies for both frontends
3. Build the Go WASM frontend
4. Verify configuration files
5. Display startup instructions

### Manual Setup

If you prefer manual setup, follow these steps:

#### 1. Setup React Frontend

```bash
cd /home/zwelch/projects/inventory/nishiki-frontend
npm install
npm run dev
# React frontend will run on http://localhost:3000
```

#### 2. Setup Go WASM Frontend

```bash
cd /home/zwelch/projects/inventory/frontend
go mod download
./bin/web
./bin/serve
# Go frontend will run on http://localhost:3000 (port from config.toml)
```

#### 3. Start Backend API

```bash
cd /home/zwelch/projects/inventory
go run main.go
# Backend API runs on http://localhost:3001
```

## Verification Process

### Phase 1: Side-by-Side Comparison

1. **Open Firefox with two windows side-by-side**
   ```bash
   firefox --new-window "http://localhost:3000" &
   firefox --new-window "http://localhost:3000" &
   ```

2. **Arrange windows**
   - Left: React frontend
   - Right: Go WASM frontend
   - Use Firefox's responsive design mode (Ctrl+Shift+M) to set exact dimensions

3. **Set consistent viewport**
   - Width: 375px (iPhone SE)
   - Width: 390px (iPhone 12/13/14)
   - Width: 414px (iPhone 14 Plus)

### Phase 2: Screen-by-Screen Verification

Follow the verification checklist in `VERIFICATION_CHECKLIST.md`

### Phase 3: Screenshot Comparison

Use the screenshot script:

```bash
./scripts/capture-screenshots.sh
```

This will:
1. Capture screenshots of key screens from both frontends
2. Save to `verification/screenshots/`
3. Generate a comparison report (optional)

### Phase 4: Automated Pixel Comparison

Use ImageMagick to compare screenshots:

```bash
./scripts/compare-screenshots.sh
```

This will:
1. Compare corresponding screenshots
2. Generate diff images highlighting differences
3. Calculate pixel difference percentage
4. Output a report

## Common Issues and Solutions

### Issue: Ports Already in Use

**React Frontend (port 3000)**
```bash
# Find process using port 3000
lsof -i :3000
# Kill the process
kill -9 <PID>
```

**Go Frontend (port from config.toml)**
```bash
# Check your configured port in frontend/config.toml
lsof -i :<port>
kill -9 <PID>
```

### Issue: Backend Not Running

```bash
# Verify backend is running
curl http://localhost:3001/health

# If not running, start it
cd /home/zwelch/projects/inventory
go run main.go
```

### Issue: Authentication Fails

1. Check Authentik is running and accessible
2. Verify `config.toml` settings match Authentik configuration
3. Check redirect URIs in Authentik match your frontend URLs
4. Clear browser localStorage and cookies

### Issue: WASM Build Fails

```bash
# Clean and rebuild
cd /home/zwelch/projects/inventory/frontend
rm -rf web/
go clean -cache
./bin/web
```

### Issue: Different Fonts Rendering

- Ensure system fonts match between screenshots
- Use Firefox font overrides if needed
- Check if Cogent Core is using system fonts correctly

## Verification Criteria

### Critical (Must Match Exactly)
- Layout structure (flex, positioning)
- Spacing (padding, margins, gaps)
- Colors (hex values)
- Border radius
- Font sizes
- Icon sizes

### Important (Should Match Closely)
- Font rendering (may vary slightly by platform)
- Line heights
- Letter spacing
- Shadows (if supported)

### Nice to Have (May Differ)
- Animations and transitions
- Hover states timing
- Focus indicators
- Scrollbar styling

## Success Criteria

Visual verification is considered successful when:

1. ✅ All critical elements match exactly (< 2% pixel difference)
2. ✅ Important elements match closely (< 5% pixel difference)
3. ✅ Layout structure is identical across all views
4. ✅ Color palette matches design tokens
5. ✅ Typography hierarchy is consistent
6. ✅ Spacing system is applied correctly
7. ✅ Components are visually indistinguishable
8. ✅ User flows work identically

## Documentation

After verification, document:

1. **Passed Screens**: List of screens with exact visual parity
2. **Minor Differences**: Document acceptable differences with justification
3. **Issues Found**: Any visual discrepancies requiring fixes
4. **Screenshots**: Archive comparison screenshots for reference

## Next Steps

After successful verification:

1. Mark Phase 1 as complete in `PORTING_PLAN.md`
2. Archive verification screenshots
3. Update `CLAUDE.md` with any architectural insights
4. Proceed to Phase 2: Groups Management

## Rollback Plan

If visual verification fails:

1. Document all discrepancies in `VISUAL_ISSUES.md`
2. Prioritize fixes based on severity
3. Fix styling in `ui/styles/` modules
4. Re-run verification after fixes
5. Iterate until success criteria met

## Verification Schedule

Recommended verification frequency:

- **After each Phase**: Full verification of new screens
- **After style changes**: Affected screens only
- **Before major releases**: Complete end-to-end verification
- **After Cogent Core updates**: Full verification (API changes may affect rendering)

## Contact and Support

If you encounter issues during verification:

1. Check existing GitHub issues in Cogent Core repo
2. Review CLAUDE.md for project-specific guidance
3. Consult PORTING_PLAN.md for phase-specific requirements
4. Document new issues in VISUAL_ISSUES.md
