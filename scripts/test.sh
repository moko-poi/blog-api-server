#!/bin/bash

# Test runner script with comprehensive coverage
# Runs different types of tests with proper reporting

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Running comprehensive test suite...${NC}"

# Run unit tests with coverage
echo -e "${YELLOW}📋 Running unit tests...${NC}"
go test -v -race -buildvcs -coverprofile=coverage.out ./...

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Unit tests passed${NC}"
else
    echo -e "${RED}❌ Unit tests failed${NC}"
    exit 1
fi

# Generate coverage report
echo -e "${YELLOW}📊 Generating coverage report...${NC}"
go tool cover -html=coverage.out -o coverage.html
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo -e "${GREEN}📈 Total coverage: $COVERAGE${NC}"

# Run integration tests if they exist
if find . -name "*_test.go" -exec grep -l "integration" {} \; | grep -q .; then
    echo -e "${YELLOW}🔗 Running integration tests...${NC}"
    go test -v -race -buildvcs -tags=integration ./...
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Integration tests passed${NC}"
    else
        echo -e "${RED}❌ Integration tests failed${NC}"
        exit 1
    fi
fi

# Run benchmarks if they exist
if find . -name "*_test.go" -exec grep -l "Benchmark" {} \; | grep -q .; then
    echo -e "${YELLOW}⚡ Running benchmarks...${NC}"
    go test -bench=. -benchmem ./...
fi

echo -e "${GREEN}🎉 All tests completed successfully!${NC}"
echo -e "${BLUE}📋 Coverage report available at: coverage.html${NC}"