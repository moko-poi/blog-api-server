#!/bin/bash

# Build script for production deployment
# Creates optimized binaries for different platforms

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Build configuration
BINARY_NAME="blog-api-server"
BUILD_DIR="build"
CMD_DIR="./cmd/server"

# Version information
VERSION=${VERSION:-"$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=${GIT_COMMIT:-"$(git rev-parse HEAD 2>/dev/null || echo 'unknown')"}

# Build flags
LDFLAGS="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}"

echo -e "${BLUE}🔨 Building ${BINARY_NAME}...${NC}"
echo -e "${YELLOW}Version: ${VERSION}${NC}"
echo -e "${YELLOW}Build Time: ${BUILD_TIME}${NC}"
echo -e "${YELLOW}Git Commit: ${GIT_COMMIT}${NC}"

# Create build directory
mkdir -p ${BUILD_DIR}

# Clean previous builds
echo -e "${YELLOW}🧹 Cleaning previous builds...${NC}"
rm -rf ${BUILD_DIR}/*

# Run tests before building
echo -e "${YELLOW}🧪 Running tests before build...${NC}"
go test ./...

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Tests failed. Build aborted.${NC}"
    exit 1
fi

# Build for current platform
echo -e "${YELLOW}🔨 Building for current platform...${NC}"
CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME} ${CMD_DIR}

# Build for multiple platforms if requested
if [ "$1" = "all" ]; then
    echo -e "${YELLOW}🌐 Building for multiple platforms...${NC}"
    
    # Linux AMD64
    echo -e "${YELLOW}  📦 Building for Linux AMD64...${NC}"
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME}-linux-amd64 ${CMD_DIR}
    
    # Linux ARM64
    echo -e "${YELLOW}  📦 Building for Linux ARM64...${NC}"
    GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME}-linux-arm64 ${CMD_DIR}
    
    # macOS AMD64
    echo -e "${YELLOW}  📦 Building for macOS AMD64...${NC}"
    GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME}-darwin-amd64 ${CMD_DIR}
    
    # macOS ARM64 (Apple Silicon)
    echo -e "${YELLOW}  📦 Building for macOS ARM64...${NC}"
    GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME}-darwin-arm64 ${CMD_DIR}
    
    # Windows AMD64
    echo -e "${YELLOW}  📦 Building for Windows AMD64...${NC}"
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME}-windows-amd64.exe ${CMD_DIR}
fi

echo -e "${GREEN}✅ Build completed successfully!${NC}"
echo -e "${BLUE}📁 Binaries available in: ${BUILD_DIR}/${NC}"
ls -la ${BUILD_DIR}/

# Test the binary
echo -e "${YELLOW}🧪 Testing binary...${NC}"
if [ -f ${BUILD_DIR}/${BINARY_NAME} ]; then
    ${BUILD_DIR}/${BINARY_NAME} --version 2>/dev/null || echo "Binary created successfully"
fi