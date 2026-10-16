#!/bin/bash

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_DIR="/tmp/store_pids"

echo "==> Stopping Nginx..."
nginx -s stop -c "$DIR/nginx/nginx.local.conf" 2>/dev/null || true
pkill -f "nginx: master process.*nginx.local.conf" 2>/dev/null || true

echo "==> Stopping Backend instances..."
if [ -d "$PID_DIR" ]; then
    for pidfile in "$PID_DIR"/*.pid; do
        if [ -f "$pidfile" ]; then
            PID=$(cat "$pidfile")
            kill -9 "$PID" 2>/dev/null || true
            rm -f "$pidfile"
        fi
    done
fi

pkill -9 -f "go run.*/src/backend/main.go" 2>/dev/null || true
pkill -9 -f "src/backend/server" 2>/dev/null || true

echo "==> All local services stopped."
