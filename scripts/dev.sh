#!/bin/bash

# Development server start script
# This script starts the development server with proper environment setup

set -e

echo "🚀 Starting Blog API Server in development mode..."

# Check if .env file exists
if [ ! -f .env ]; then
    echo "⚠️  .env file not found. Running setup first..."
    chmod +x ./scripts/setup.sh
    ./scripts/setup.sh
fi

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Create tmp directory if it doesn't exist
mkdir -p tmp

# Check if Air is available for hot reloading
if command -v air &> /dev/null; then
    echo "🔄 Starting with hot reload using Air..."
    air -c .air.toml
else
    echo "📝 Air not found. Starting without hot reload..."
    echo "   Tip: Install Air with 'go install github.com/air-verse/air@latest' for hot reloading"
    go run ./cmd/server/main.go
fi