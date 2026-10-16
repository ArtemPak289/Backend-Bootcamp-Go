#!/usr/bin/env bash

PID_DIR="/tmp/be4_pids"

echo "==> Stopping local BE4 services..."

if [ -f "${PID_DIR}/store.pid" ]; then
  STORE_PID=$(cat "${PID_DIR}/store.pid")
  if kill -0 "${STORE_PID}" 2>/dev/null; then
    echo "    Stopping store_service (PID ${STORE_PID})..."
    kill -15 "${STORE_PID}" 2>/dev/null || true
  fi
  rm -f "${PID_DIR}/store.pid"
fi

if [ -f "${PID_DIR}/auth.pid" ]; then
  AUTH_PID=$(cat "${PID_DIR}/auth.pid")
  if kill -0 "${AUTH_PID}" 2>/dev/null; then
    echo "    Stopping auth_service (PID ${AUTH_PID})..."
    kill -15 "${AUTH_PID}" 2>/dev/null || true
  fi
  rm -f "${PID_DIR}/auth.pid"
fi

# Fallback kill by port/name
pkill -f "auth_service/server" 2>/dev/null || true
pkill -f "store_service/server" 2>/dev/null || true

echo "==> All local services stopped."
