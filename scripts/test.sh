#!/bin/bash
set -e

echo "🧪 Running AI Router Test Suite"
echo "================================"
echo

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    exit 1
fi

echo -e "${YELLOW}📦 Downloading dependencies...${NC}"
go mod download
go mod verify
echo

echo -e "${YELLOW}🔨 Building project...${NC}"
make build
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Build successful${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi
echo

echo -e "${YELLOW}🧪 Running unit tests...${NC}"
go test ./pkg/... -v -cover
UNIT_TEST_EXIT=$?
echo

if [ $UNIT_TEST_EXIT -eq 0 ]; then
    echo -e "${GREEN}✓ Unit tests passed${NC}"
else
    echo -e "${RED}✗ Unit tests failed${NC}"
    exit 1
fi
echo

echo -e "${YELLOW}🔗 Running integration tests...${NC}"
go test ./test/... -v
INTEGRATION_TEST_EXIT=$?
echo

if [ $INTEGRATION_TEST_EXIT -eq 0 ]; then
    echo -e "${GREEN}✓ Integration tests passed${NC}"
else
    echo -e "${RED}✗ Integration tests failed${NC}"
    exit 1
fi
echo

echo -e "${YELLOW}📊 Generating coverage report...${NC}"
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -1
echo

echo -e "${GREEN}✅ All tests passed!${NC}"
echo
echo "To view detailed coverage:"
echo "  go tool cover -html=coverage.out"
echo
