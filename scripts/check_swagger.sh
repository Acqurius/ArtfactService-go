#!/bin/bash

# Script to clear browser cache and verify Swagger documentation
# This helps when Swagger UI shows old cached content

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}=== Swagger Cache Clear Helper ===${NC}\n"

# Check if server is running
if ! curl -s http://localhost:8080/swagger/doc.json > /dev/null 2>&1; then
    echo -e "${YELLOW}Warning: Server doesn't seem to be running on port 8080${NC}"
    echo "Please start the server first: ./server"
    exit 1
fi

echo -e "${BLUE}Checking current Swagger documentation...${NC}"

# Check if new endpoints are in the swagger doc
if curl -s http://localhost:8080/swagger/doc.json | grep -q "artifact-service/v1/artifacts"; then
    echo -e "${GREEN}✓ New API endpoints found in Swagger documentation${NC}\n"
    
    echo -e "${BLUE}Available endpoints:${NC}"
    curl -s http://localhost:8080/swagger/doc.json | grep -o '"/artifact-service[^"]*"' | sort -u
    echo ""
    
    echo -e "${YELLOW}If you still don't see the new endpoints in your browser:${NC}"
    echo "1. Hard refresh the page (Ctrl+Shift+R or Cmd+Shift+R)"
    echo "2. Clear browser cache for this site"
    echo "3. Try opening in incognito/private mode"
    echo "4. Add a timestamp to URL: http://localhost:8080/swagger/index.html?t=$(date +%s)"
    echo ""
    echo -e "${GREEN}Direct link with cache buster:${NC}"
    echo "http://10.188.157.24:8080/swagger/index.html?t=$(date +%s)"
else
    echo -e "${YELLOW}⚠ New endpoints not found in Swagger doc${NC}"
    echo "Running swagger update..."
    swag init
    echo ""
    echo -e "${YELLOW}Please restart the server to load the updated documentation${NC}"
fi
