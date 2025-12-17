#!/bin/bash

# Script to update Swagger documentation
# This script regenerates the Swagger docs from Go annotations

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Updating Swagger Documentation ===${NC}\n"

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo -e "${YELLOW}swag command not found. Installing...${NC}"
    go install github.com/swaggo/swag/cmd/swag@latest
    echo -e "${GREEN}✓ swag installed${NC}\n"
fi

# Navigate to project root (in case script is run from elsewhere)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

echo -e "${BLUE}Generating Swagger documentation...${NC}"

# Run swag init to generate docs
swag init

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Swagger documentation updated successfully!${NC}\n"
    echo -e "${BLUE}Updated files:${NC}"
    echo "  - docs/docs.go"
    echo "  - docs/swagger.json"
    echo "  - docs/swagger.yaml"
    echo ""
    
    # Rebuild the server to include updated docs
    echo -e "${BLUE}Rebuilding server with updated documentation...${NC}"
    go build -o server .
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Server rebuilt successfully!${NC}\n"
        echo -e "${GREEN}You can now view the updated documentation at:${NC}"
        echo "  http://localhost:8080/swagger/index.html"
        echo ""
        echo -e "${YELLOW}Note: Restart the server to see the changes${NC}"
        echo "  ./server"
    else
        echo -e "${RED}✗ Failed to rebuild server${NC}"
        exit 1
    fi
else
    echo -e "${RED}✗ Failed to generate Swagger documentation${NC}"
    exit 1
fi
