#!/bin/bash

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Task Service Smoke Test${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

echo -e "${YELLOW}Go version:${NC}"
go version
echo ""

# Format check
echo -e "${YELLOW}Checking code formatting...${NC}"
if ! gofmt -l ./... | grep -q "^"; then
    echo -e "${GREEN}✓ Code formatting is correct${NC}"
else
    echo -e "${RED}✗ Code formatting issues found:${NC}"
    gofmt -l ./...
    exit 1
fi
echo ""

# Vet check
echo -e "${YELLOW}Running go vet...${NC}"
if go vet ./... 2>&1; then
    echo -e "${GREEN}✓ No vet issues found${NC}"
else
    echo -e "${RED}✗ Vet issues found${NC}"
    exit 1
fi
echo ""

# Run unit tests
echo -e "${YELLOW}Running unit tests...${NC}"
if go test -v ./... 2>&1; then
    echo -e "${GREEN}✓ All tests passed${NC}"
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
echo ""

# Test coverage
echo -e "${YELLOW}Generating test coverage report...${NC}"
go test -coverprofile=coverage.out ./... > /dev/null 2>&1
echo ""
echo -e "${YELLOW}Coverage summary:${NC}"
go tool cover -func=coverage.out | tail -5
echo ""

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
echo -e "${GREEN}✓ Coverage report generated: coverage.html${NC}"
echo ""

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Smoke test completed successfully!${NC}"
echo -e "${BLUE}========================================${NC}"
