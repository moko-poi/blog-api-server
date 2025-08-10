#!/bin/bash

# Blog API Server Development Setup Script
# This script sets up the local development environment

set -e  # Exit on any error

echo "🚀 Setting up Blog API Server development environment..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.24 or later."
    echo "   Visit: https://golang.org/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
REQUIRED_VERSION="1.24"
if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ Go version $GO_VERSION is too old. Please upgrade to Go $REQUIRED_VERSION or later."
    exit 1
fi

echo "✅ Go version $GO_VERSION is installed"

# Check if Docker is installed (optional)
if command -v docker &> /dev/null; then
    echo "✅ Docker is installed"
    DOCKER_AVAILABLE=true
else
    echo "⚠️  Docker is not installed (optional for development)"
    DOCKER_AVAILABLE=false
fi

# Create .env file from example if it doesn't exist
if [ ! -f .env ]; then
    if [ -f .env.example ]; then
        cp .env.example .env
        echo "✅ Created .env file from .env.example"
    else
        echo "⚠️  .env.example not found, creating basic .env file"
        cat > .env << EOF
HOST=localhost
PORT=8080
LOG_LEVEL=debug
READ_TIMEOUT=10s
WRITE_TIMEOUT=10s
IDLE_TIMEOUT=120s
DEV_MODE=true
EOF
    fi
else
    echo "✅ .env file already exists"
fi

# Install Go dependencies
echo "📦 Installing Go dependencies..."
go mod download
go mod tidy

# Install development tools
echo "🔧 Installing development tools..."

# Install Air for hot reloading (optional)
if command -v air &> /dev/null; then
    echo "✅ Air is already installed"
else
    echo "📦 Installing Air for hot reloading..."
    go install github.com/air-verse/air@latest
    if [ $? -eq 0 ]; then
        echo "✅ Air installed successfully"
    else
        echo "⚠️  Failed to install Air. Hot reloading may not work."
    fi
fi

# Create tmp directory for Air
mkdir -p tmp

# Test build
echo "🔨 Testing build..."
if go build -o ./tmp/test-build ./cmd/server; then
    echo "✅ Build successful"
    rm -f ./tmp/test-build
else
    echo "❌ Build failed"
    exit 1
fi

# Run tests
echo "🧪 Running tests..."
if go test ./...; then
    echo "✅ All tests passed"
else
    echo "❌ Some tests failed"
    exit 1
fi

echo ""
echo "🎉 Development environment setup complete!"
echo ""
echo "📋 Next steps:"
echo "   • Run 'make dev' to start the development server with hot reload"
echo "   • Run 'make test' to run all tests"
echo "   • Run 'make audit' for code quality checks"
if [ "$DOCKER_AVAILABLE" = true ]; then
echo "   • Run 'docker-compose -f docker-compose.dev.yml up' for Docker development"
fi
echo "   • Visit http://localhost:8080/healthz to check if the server is running"
echo ""
echo "📖 Available commands:"
echo "   • make help - Show all available make targets"
echo "   • make build - Build production binary"
echo "   • make test-cover - Run tests with coverage report"