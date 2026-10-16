#!/bin/bash
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_DIR="/tmp/store_pids"
mkdir -p "$PID_DIR" /tmp/nginx_cache

echo "==> Ensuring SSL certificates exist..."
if [ ! -f "$DIR/ssl/server.crt" ] || [ ! -f "$DIR/ssl/server.key" ]; then
    "$DIR/ssl/generate_certs.sh"
fi

echo "==> Building Go Backend binary..."
(cd "$DIR/backend" && go build -o server main.go)

echo "==> Starting Backend Instance 1 (Primary Read-Write on :8081)..."
(
    cd "$DIR/backend"
    PORT=8081 \
    SERVER_PORT=8081 \
    INSTANCE_NAME="backend-main-8081" \
    DB_HOST=localhost \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_NAME=store_db \
    READ_ONLY=false \
    ./server > /tmp/backend_8081.log 2>&1
) &
echo $! > "$PID_DIR/backend_8081.pid"

echo "==> Starting Backend Instance 2 (Replica 1 Read-Only on :8082)..."
(
    cd "$DIR/backend"
    PORT=8082 \
    SERVER_PORT=8082 \
    INSTANCE_NAME="backend-replica1-8082" \
    DB_HOST=localhost \
    DB_PORT=5432 \
    DB_USER=store_readonly \
    DB_PASSWORD=readonly_pass \
    DB_NAME=store_db \
    READ_ONLY=true \
    ./server > /tmp/backend_8082.log 2>&1
) &
echo $! > "$PID_DIR/backend_8082.pid"

echo "==> Starting Backend Instance 3 (Replica 2 Read-Only on :8083)..."
(
    cd "$DIR/backend"
    PORT=8083 \
    SERVER_PORT=8083 \
    INSTANCE_NAME="backend-replica2-8083" \
    DB_HOST=localhost \
    DB_PORT=5432 \
    DB_USER=store_readonly \
    DB_PASSWORD=readonly_pass \
    DB_NAME=store_db \
    READ_ONLY=true \
    ./server > /tmp/backend_8083.log 2>&1
) &
echo $! > "$PID_DIR/backend_8083.pid"

echo "==> Waiting for backend instances to be ready..."
sleep 3

echo "==> Testing backend instances..."
curl -s http://localhost:8081/health || echo "Waiting for 8081..."
curl -s http://localhost:8082/health || echo "Waiting for 8082..."
curl -s http://localhost:8083/health || echo "Waiting for 8083..."

echo "==> Starting Nginx locally..."
if pgrep -f "nginx: master process.*nginx.local.conf" > /dev/null; then
    nginx -s stop -c "$DIR/nginx/nginx.local.conf" || true
    sleep 1
fi
nginx -c "$DIR/nginx/nginx.local.conf"

echo ""
echo "=========================================================="
echo "  All services started successfully in native local mode!  "
echo "  - Backend 1 (Main RW):   http://localhost:8081"
echo "  - Backend 2 (RO Repl 1): http://localhost:8082"
echo "  - Backend 3 (RO Repl 2): http://localhost:8083"
echo "  - Nginx HTTP:            http://localhost:8080"
echo "  - Nginx HTTPS:           https://localhost:8443"
echo "=========================================================="
