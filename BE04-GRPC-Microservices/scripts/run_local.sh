#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PID_DIR="/tmp/be4_pids"

mkdir -p "${PID_DIR}"

echo "==> Ensuring PostgreSQL databases exist..."
psql -d postgres -c "SELECT 1 FROM pg_database WHERE datname = 'auth_db'" | grep -q 1 || psql -d postgres -c "CREATE DATABASE auth_db"
psql -d postgres -c "SELECT 1 FROM pg_database WHERE datname = 'store_db'" | grep -q 1 || psql -d postgres -c "CREATE DATABASE store_db"

echo "==> Building services..."
(cd "${SRC_DIR}/auth_service" && go build -o server main.go)
(cd "${SRC_DIR}/store_service" && go build -o server main.go)

# Stop existing services if running
"${SCRIPT_DIR}/stop_local.sh" >/dev/null 2>&1 || true

echo "==> Starting auth_service (gRPC on port 50051)..."
AUTH_LOG="/tmp/be4_auth.log"
export GRPC_PORT=50051
export DB_NAME=auth_db
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER="${USER:-artempak}"
export DB_PASSWORD=""
export DB_SSLMODE=disable
export JWT_SECRET="school21-secret-jwt-token-key-2026"

nohup "${SRC_DIR}/auth_service/server" > "${AUTH_LOG}" 2>&1 &
AUTH_PID=$!
echo "${AUTH_PID}" > "${PID_DIR}/auth.pid"
disown "${AUTH_PID}" 2>/dev/null || true
echo "    auth_service started with PID ${AUTH_PID} (logs: ${AUTH_LOG})"

sleep 1

echo "==> Starting store_service (HTTP on port 8080)..."
STORE_LOG="/tmp/be4_store.log"
export PORT=8080
export DB_NAME=store_db
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER="${USER:-artempak}"
export DB_PASSWORD=""
export DB_SSLMODE=disable
export AUTH_SERVICE_ADDR="localhost:50051"

nohup "${SRC_DIR}/store_service/server" > "${STORE_LOG}" 2>&1 &
STORE_PID=$!
echo "${STORE_PID}" > "${PID_DIR}/store.pid"
disown "${STORE_PID}" 2>/dev/null || true
echo "    store_service started with PID ${STORE_PID} (logs: ${STORE_LOG})"

sleep 2

echo "==> Checking service status..."
if curl -s "http://localhost:8080/health" | grep -q "ok"; then
  echo "==> Store service is healthy and running on http://localhost:8080"
  echo "==> Auth service is running on gRPC localhost:50051"
  echo "==> Services launched successfully."
else
  echo "==> Warning: store_service healthcheck failed. Check ${STORE_LOG}"
fi
