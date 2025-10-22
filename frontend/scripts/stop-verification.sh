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
