#!/bin/bash
set -e

# This script sets up the development environment for Tea API on macOS.
# It installs required dependencies and prepares environment files.

# Ensure Homebrew is installed
if ! command -v brew >/dev/null 2>&1; then
  echo "Homebrew not found. Installing..."
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
else
  echo "Homebrew already installed"
fi

# Install Go and Node.js
brew install go node

# Install frontend dependencies
if [ -d "$(dirname "$0")/../web" ]; then
  echo "Installing frontend dependencies..."
  cd "$(dirname "$0")/../web"
  npm install
  cd - >/dev/null
fi

# Copy example environment file if .env does not exist
ENV_FILE="$(dirname "$0")/../.env"
if [ ! -f "$ENV_FILE" ]; then
  cp "$(dirname "$0")/../.env.example" "$ENV_FILE"
  echo "Created .env from .env.example"
fi

echo "Environment setup completed."
