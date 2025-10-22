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
