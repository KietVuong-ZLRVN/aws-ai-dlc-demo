#!/usr/bin/env bash
# demo.sh — starts both Go backends and the React dev server
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
PD_SRC="$ROOT/product_discovery/src"
WL_SRC="$ROOT/wishlist/src"
FE_SRC="$ROOT/frontend"

# ── Cleanup on exit ──────────────────────────────────────────────────────────
pids=()
cleanup() {
  echo ""
  echo "Stopping all processes..."
  for pid in "${pids[@]}"; do
    kill "$pid" 2>/dev/null || true
  done
  wait 2>/dev/null || true
  echo "Done."
}
trap cleanup EXIT INT TERM

# ── Prerequisite checks ───────────────────────────────────────────────────────
check() {
  command -v "$1" >/dev/null 2>&1 || { echo "Error: '$1' not found. Please install it."; exit 1; }
}
check go
check node
check npm

# ── Install frontend deps if needed ──────────────────────────────────────────
if [ ! -d "$FE_SRC/node_modules" ]; then
  echo "Installing frontend dependencies..."
  npm --prefix "$FE_SRC" install --silent
fi

# ── Start Product Discovery backend (port 8080) ───────────────────────────────
echo "[product-discovery] Starting on :8080 ..."
PORT=8080 go run -C "$PD_SRC" ./cmd/server &
pids+=($!)

# ── Start Wishlist backend (port 8081) ────────────────────────────────────────
echo "[wishlist]           Starting on :8081 ..."
PORT=8081 go run -C "$WL_SRC" ./cmd/server &
pids+=($!)

# ── Give backends a moment to start ──────────────────────────────────────────
sleep 2

# ── Start React frontend (port 5173) ─────────────────────────────────────────
echo "[frontend]           Starting on :5173 ..."
npm --prefix "$FE_SRC" run dev -- --open &
pids+=($!)

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Product Discovery API → http://localhost:8080/api/v1/products"
echo " Wishlist API          → http://localhost:8081/api/v1/wishlist"
echo " Frontend              → http://localhost:5173"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo " Press Ctrl+C to stop all services."
echo ""

# ── Wait for any child to exit ───────────────────────────────────────────────
wait -n 2>/dev/null || wait
