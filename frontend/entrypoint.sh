#!/bin/sh
set -e

cd /app

# Ensure package manifests exist (bind mount may replace image layer)
if [ ! -f package.json ]; then
  echo "Error: package.json not found in /app. Is the volume mounted correctly?"
  ls -la
  exit 1
fi

# Install deps if missing or incomplete
if [ ! -d node_modules ] || [ ! -f node_modules/.bin/next ]; then
  echo "Installing dependencies (node_modules missing or Next binary not found)..."
  npm ci --include=dev || npm install
fi

# Avoid running proto generation during dev container start
export SKIP_PROTO=1

echo "Starting Next.js dev server..."
exec ./node_modules/.bin/next dev --turbopack --port 3002 --hostname 0.0.0.0
