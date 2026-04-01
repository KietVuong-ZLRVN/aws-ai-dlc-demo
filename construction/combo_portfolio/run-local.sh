#!/usr/bin/env bash
# run-local.sh — start the mock backend and Vite dev server together
# Usage: ./run-local.sh
# Ctrl-C will stop both processes.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR/src"
FRONTEND_DIR="$SCRIPT_DIR/frontend"

cleanup() {
  echo ""
  echo "Stopping servers..."
  kill "$BACKEND_PID" "$FRONTEND_PID" 2>/dev/null || true
  wait "$BACKEND_PID" "$FRONTEND_PID" 2>/dev/null || true
  echo "Done."
}
trap cleanup INT TERM

echo "==> Clearing port 8081 if in use..."
lsof -ti :8081 | xargs kill -9 2>/dev/null || true

echo "==> Starting mock backend (port 8081)..."
cd "$BACKEND_DIR"
go run ./mockserver &
BACKEND_PID=$!

echo "==> Starting Vite dev server (port 5173)..."
cd "$FRONTEND_DIR"
npm run dev &
FRONTEND_PID=$!

echo ""
echo "  Backend  : http://localhost:8081/api/v1/combos"
echo "  Frontend : http://localhost:5173"
echo ""
echo "Press Ctrl-C to stop."

wait "$BACKEND_PID" "$FRONTEND_PID"
