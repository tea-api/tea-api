#!/bin/bash
set -e

# This script builds Tea API for macOS for local development testing.
# It compiles the frontend and backend and outputs a macOS binary.

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)

# Build frontend
cd "$REPO_ROOT/web"
if [ ! -d node_modules ]; then
  npm install
fi
DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat "$REPO_ROOT/VERSION") npm run build
cd "$REPO_ROOT"

# Determine architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)
    GOARCH=amd64
    ;;
  arm64|aarch64)
    GOARCH=arm64
    ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

# Build backend
GOOS=darwin GOARCH=$GOARCH go build -ldflags "-X 'tea-api/common.Version=$(cat VERSION)'" -o tea-api-macos

echo "Build complete: tea-api-macos"
