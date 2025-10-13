#!/bin/bash
# Visual Verification Setup Script
# Automates the setup process for comparing React and Go WASM frontends

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project paths
PROJECT_ROOT="/home/zwelch/projects/inventory"
REACT_FRONTEND="$PROJECT_ROOT/nishiki-frontend"
GO_FRONTEND="$PROJECT_ROOT/frontend"
BACKEND="$PROJECT_ROOT"

echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}    Visual Verification Setup Script${NC}"
echo -e "${BLUE}==================================================${NC}"
echo ""

# Function to print status messages
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

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if port is in use
port_in_use() {
    lsof -i :"$1" >/dev/null 2>&1
}

# =================================================
# Step 1: Check Prerequisites
# =================================================

print_status "Checking prerequisites..."
echo ""

# Check Go
if command_exists go; then
    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go installed: $GO_VERSION"
else
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Check Node.js
if command_exists node; then
    NODE_VERSION=$(node --version)
    print_success "Node.js installed: $NODE_VERSION"
else
    print_error "Node.js is not installed or not in PATH"
    exit 1
fi

# Check npm
if command_exists npm; then
    NPM_VERSION=$(npm --version)
    print_success "npm installed: $NPM_VERSION"
else
    print_error "npm is not installed or not in PATH"
    exit 1
fi

# Check Firefox (optional but recommended)
if command_exists firefox; then
    print_success "Firefox installed (recommended for verification)"
else
    print_warning "Firefox not found. Any modern browser will work, but Firefox is recommended."
fi

# Check lsof for port checking
if command_exists lsof; then
    print_success "lsof available for port checking"
else
    print_warning "lsof not found. Port conflict detection will be limited."
fi

echo ""

# =================================================
# Step 2: Check Project Structure
# =================================================

print_status "Checking project structure..."
echo ""

# Check React frontend directory
if [ -d "$REACT_FRONTEND" ]; then
    print_success "React frontend directory found: $REACT_FRONTEND"
else
    print_error "React frontend directory not found: $REACT_FRONTEND"
    exit 1
fi

# Check Go frontend directory
if [ -d "$GO_FRONTEND" ]; then
    print_success "Go frontend directory found: $GO_FRONTEND"
else
    print_error "Go frontend directory not found: $GO_FRONTEND"
    exit 1
fi

# Check Backend directory
if [ -d "$BACKEND" ]; then
    print_success "Backend directory found: $BACKEND"
else
    print_error "Backend directory not found: $BACKEND"
    exit 1
fi

# Check Go frontend config
if [ -f "$GO_FRONTEND/config.toml" ]; then
    print_success "Go frontend config.toml found"
    GO_PORT=$(grep '^port = ' "$GO_FRONTEND/config.toml" | sed 's/port = "\(.*\)"/\1/')
    if [ -n "$GO_PORT" ]; then
        print_status "Go frontend configured for port: $GO_PORT"
    else
        print_warning "Port not found in config.toml, will default to 8080"
        GO_PORT="8080"
    fi
else
    print_error "Go frontend config.toml not found"
    exit 1
fi

echo ""

# =================================================
# Step 3: Check Port Availability
# =================================================

print_status "Checking port availability..."
echo ""

# Check React frontend port (3000)
if command_exists lsof && port_in_use 3000; then
    print_warning "Port 3000 is already in use"
    print_status "React frontend may not start. Kill the process with: lsof -ti :3000 | xargs kill -9"
else
    print_success "Port 3000 is available (React frontend)"
fi

# Check Go frontend port
if command_exists lsof && port_in_use "$GO_PORT"; then
    print_warning "Port $GO_PORT is already in use"
    print_status "Go frontend may not start. Kill the process with: lsof -ti :$GO_PORT | xargs kill -9"
else
    print_success "Port $GO_PORT is available (Go frontend)"
fi

# Check backend port (3001)
if command_exists lsof && port_in_use 3001; then
    print_warning "Port 3001 is already in use"
    print_status "Backend API may not start. Kill the process with: lsof -ti :3001 | xargs kill -9"
else
    print_success "Port 3001 is available (Backend API)"
fi

echo ""

# =================================================
# Step 4: Setup React Frontend
# =================================================

print_status "Setting up React frontend..."
echo ""

cd "$REACT_FRONTEND"

# Check if node_modules exists
if [ ! -d "node_modules" ]; then
    print_status "Installing React frontend dependencies..."
    npm install
    print_success "React frontend dependencies installed"
else
    print_success "React frontend dependencies already installed"
fi

# Check package.json
if [ -f "package.json" ]; then
    print_success "React frontend package.json found"
else
    print_error "React frontend package.json not found"
    exit 1
fi

echo ""

# =================================================
# Step 5: Setup Go Frontend
# =================================================

print_status "Setting up Go frontend..."
echo ""

cd "$GO_FRONTEND"

# Download Go dependencies
print_status "Downloading Go dependencies..."
go mod download
print_success "Go dependencies downloaded"

# Check if build tools exist
if [ -f "bin/web" ]; then
    print_success "Go build tool (bin/web) found"
else
    print_warning "Go build tool (bin/web) not found"
    print_status "You may need to build it first: go build -o bin/web cmd/web/main.go"
fi

if [ -f "bin/serve" ]; then
    print_success "Go serve tool (bin/serve) found"
else
    print_warning "Go serve tool (bin/serve) not found"
    print_status "You may need to build it first: go build -o bin/serve cmd/serve/main.go"
fi

# Build WASM frontend
print_status "Building Go WASM frontend..."
if [ -f "bin/web" ]; then
    ./bin/web
    print_success "Go WASM frontend built successfully"
else
    print_warning "Skipping WASM build - bin/web not found"
fi

# Check web output directory
if [ -d "web" ]; then
    print_success "WASM output directory (web/) exists"

    # Check for key files
    if [ -f "web/app.wasm" ]; then
        print_success "app.wasm built successfully"
    else
        print_warning "app.wasm not found in web/ directory"
    fi

    if [ -f "web/index.html" ]; then
        print_success "index.html found"
    else
        print_warning "index.html not found in web/ directory"
    fi
else
    print_warning "WASM output directory (web/) not found"
fi

echo ""

# =================================================
# Step 6: Verify Configuration Files
# =================================================

print_status "Verifying configuration files..."
echo ""

# Check React .env or config
if [ -f "$REACT_FRONTEND/.env" ] || [ -f "$REACT_FRONTEND/.env.local" ]; then
    print_success "React environment config found"
else
    print_warning "React .env file not found. Make sure API URLs are configured."
fi

# Check Go config.toml
if [ -f "$GO_FRONTEND/config.toml" ]; then
    print_success "Go config.toml found"

    # Display key config values
    print_status "Go Frontend Configuration:"
    echo "  Port: $GO_PORT"
    BACKEND_URL=$(grep '^backend_url = ' "$GO_FRONTEND/config.toml" | sed 's/backend_url = "\(.*\)"/\1/' || echo "not found")
    echo "  Backend URL: $BACKEND_URL"
    AUTH_URL=$(grep '^auth_url = ' "$GO_FRONTEND/config.toml" | sed 's/auth_url = "\(.*\)"/\1/' || echo "not found")
    echo "  Auth URL: $AUTH_URL"
else
    print_error "Go config.toml not found"
    exit 1
fi

# Check Backend config
if [ -f "$BACKEND/app.toml" ]; then
    print_success "Backend app.toml found"
else
    print_warning "Backend app.toml not found. Backend may use default configuration."
fi

echo ""

# =================================================
# Step 7: Create Verification Directories
# =================================================

print_status "Creating verification directories..."
echo ""

cd "$GO_FRONTEND"

# Create directories for screenshots and reports
mkdir -p verification/screenshots/react
mkdir -p verification/screenshots/go
mkdir -p verification/screenshots/diff
mkdir -p verification/reports

print_success "Verification directories created:"
echo "  - verification/screenshots/react/"
echo "  - verification/screenshots/go/"
echo "  - verification/screenshots/diff/"
echo "  - verification/reports/"

echo ""

# =================================================
# Step 8: Summary and Next Steps
# =================================================

print_status "Setup complete!"
echo ""

echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}    Next Steps for Visual Verification${NC}"
echo -e "${GREEN}==================================================${NC}"
echo ""

echo -e "${YELLOW}1. Start Backend API${NC}"
echo "   cd $BACKEND"
echo "   go run main.go"
echo "   (Backend will run on http://localhost:3001)"
echo ""

echo -e "${YELLOW}2. Start React Frontend${NC}"
echo "   cd $REACT_FRONTEND"
echo "   npm run dev"
echo "   (React will run on http://localhost:3000)"
echo ""

echo -e "${YELLOW}3. Start Go WASM Frontend${NC}"
echo "   cd $GO_FRONTEND"
echo "   ./bin/serve"
echo "   (Go WASM will run on http://localhost:$GO_PORT)"
echo ""

echo -e "${YELLOW}4. Open Side-by-Side Comparison${NC}"
echo "   firefox --new-window http://localhost:3000 &"
echo "   firefox --new-window http://localhost:$GO_PORT &"
echo ""

echo -e "${YELLOW}5. Follow Verification Checklist${NC}"
echo "   Open: $GO_FRONTEND/VERIFICATION_CHECKLIST.md"
echo "   Compare each screen systematically"
echo ""

echo -e "${YELLOW}6. Capture Screenshots (Optional)${NC}"
echo "   cd $GO_FRONTEND"
echo "   ./scripts/capture-screenshots.sh"
echo ""

echo -e "${YELLOW}7. Generate Comparison Report (Optional)${NC}"
echo "   cd $GO_FRONTEND"
echo "   ./scripts/compare-screenshots.sh"
echo ""

echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}    Quick Start Commands${NC}"
echo -e "${BLUE}==================================================${NC}"
echo ""

# Create a quick start script
cat > "$GO_FRONTEND/scripts/start-verification.sh" << 'EOF'
#!/bin/bash
# Quick start script for visual verification

PROJECT_ROOT="/home/zwelch/projects/inventory"
REACT_FRONTEND="$PROJECT_ROOT/nishiki-frontend"
GO_FRONTEND="$PROJECT_ROOT/frontend"
BACKEND="$PROJECT_ROOT"

# Start backend in background
cd "$BACKEND"
echo "Starting backend API..."
go run main.go > logs/backend.log 2>&1 &
BACKEND_PID=$!
echo "Backend PID: $BACKEND_PID"

# Wait for backend to start
sleep 3

# Start React frontend in background
cd "$REACT_FRONTEND"
echo "Starting React frontend..."
npm run dev > "$GO_FRONTEND/logs/react.log" 2>&1 &
REACT_PID=$!
echo "React PID: $REACT_PID"

# Wait for React to start
sleep 5

# Start Go frontend in background
cd "$GO_FRONTEND"
echo "Starting Go WASM frontend..."
./bin/serve > logs/go.log 2>&1 &
GO_PID=$!
echo "Go WASM PID: $GO_PID"

# Wait for Go frontend to start
sleep 3

echo ""
echo "All services started!"
echo "Backend PID: $BACKEND_PID (http://localhost:3001)"
echo "React PID: $REACT_PID (http://localhost:3000)"
echo "Go WASM PID: $GO_PID (http://localhost:PORT_FROM_CONFIG)"
echo ""
echo "To stop all services, run:"
echo "  kill $BACKEND_PID $REACT_PID $GO_PID"
echo ""
echo "Opening browsers for comparison..."
firefox --new-window "http://localhost:3000" &
sleep 2
firefox --new-window "http://localhost:$(grep '^port = ' config.toml | sed 's/port = "\(.*\)"/\1/')" &

# Save PIDs to file for cleanup script
echo "$BACKEND_PID $REACT_PID $GO_PID" > "$GO_FRONTEND/logs/verification.pids"
EOF

chmod +x "$GO_FRONTEND/scripts/start-verification.sh"

print_success "Created quick start script: scripts/start-verification.sh"
echo ""

# Create a cleanup script
cat > "$GO_FRONTEND/scripts/stop-verification.sh" << 'EOF'
#!/bin/bash
# Stop all verification services

GO_FRONTEND="/home/zwelch/projects/inventory/frontend"

if [ -f "$GO_FRONTEND/logs/verification.pids" ]; then
    PIDS=$(cat "$GO_FRONTEND/logs/verification.pids")
    echo "Stopping services: $PIDS"
    kill $PIDS 2>/dev/null
    rm "$GO_FRONTEND/logs/verification.pids"
    echo "All services stopped"
else
    echo "No PID file found. Manually kill processes on ports 3000, 3001, and your Go frontend port."
fi
EOF

chmod +x "$GO_FRONTEND/scripts/stop-verification.sh"

print_success "Created cleanup script: scripts/stop-verification.sh"
echo ""

# Create logs directory
mkdir -p "$GO_FRONTEND/logs"
print_success "Created logs directory: logs/"
echo ""

echo -e "${GREEN}Run this command to start everything at once:${NC}"
echo -e "  ${BLUE}cd $GO_FRONTEND && ./scripts/start-verification.sh${NC}"
echo ""

echo -e "${GREEN}Run this command to stop all services:${NC}"
echo -e "  ${BLUE}cd $GO_FRONTEND && ./scripts/stop-verification.sh${NC}"
echo ""

echo -e "${BLUE}==================================================${NC}"
print_success "Setup script complete!"
echo -e "${BLUE}==================================================${NC}"
