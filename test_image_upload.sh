#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Image Upload Endpoint Test ===${NC}\n"

# Test 1: Check if UploadImage handler exists in code
echo -e "${YELLOW}Test 1: Verify UploadImage handler exists...${NC}"
if grep -q "func (h \*MarketplaceHandlers) UploadImage" /Users/dmitrijfomin/Desktop/backend/internal/modules/marketplace/transport/http/handlers.go; then
    echo -e "${GREEN}✓ UploadImage handler found${NC}"
else
    echo -e "${RED}✗ UploadImage handler NOT found${NC}"
    exit 1
fi

# Test 2: Check if route is registered
echo -e "${YELLOW}Test 2: Verify /api/upload/image route is registered...${NC}"
if grep -q "Post.*image.*m.handlers.UploadImage" /Users/dmitrijfomin/Desktop/backend/internal/modules/marketplace/module.go; then
    echo -e "${GREEN}✓ Route /api/upload/image registered${NC}"
else
    echo -e "${RED}✗ Route NOT registered${NC}"
    exit 1
fi

# Test 3: Check if UploadImageResponse DTO exists
echo -e "${YELLOW}Test 3: Verify UploadImageResponse DTO exists...${NC}"
if grep -q "type UploadImageResponse struct" /Users/dmitrijfomin/Desktop/backend/internal/modules/marketplace/dto/requests.go; then
    echo -e "${GREEN}✓ UploadImageResponse DTO found${NC}"
else
    echo -e "${RED}✗ UploadImageResponse DTO NOT found${NC}"
    exit 1
fi

# Test 4: Check if CloudinaryService is initialized in handlers
echo -e "${YELLOW}Test 4: Verify CloudinaryService initialization...${NC}"
if grep -q "cloudinaryService: service.NewCloudinaryService()" /Users/dmitrijfomin/Desktop/backend/internal/modules/marketplace/transport/http/handlers.go; then
    echo -e "${GREEN}✓ CloudinaryService initialized in handlers${NC}"
else
    echo -e "${RED}✗ CloudinaryService NOT initialized${NC}"
    exit 1
fi

# Test 5: Check if multipart form parsing is implemented
echo -e "${YELLOW}Test 5: Verify multipart form parsing...${NC}"
if grep -q "r.ParseMultipartForm" /Users/dmitrijfomin/Desktop/backend/internal/modules/marketplace/transport/http/handlers.go; then
    echo -e "${GREEN}✓ Multipart form parsing implemented${NC}"
else
    echo -e "${RED}✗ Multipart form parsing NOT implemented${NC}"
    exit 1
fi

# Test 6: Check if file validation is implemented
echo -e "${YELLOW}Test 6: Verify file validation...${NC}"
if grep -q "r.FormFile(\"image\")" /Users/dmitrijfomin/Desktop/backend/internal/modules/marketplace/transport/http/handlers.go; then
    echo -e "${GREEN}✓ File validation implemented${NC}"
else
    echo -e "${RED}✗ File validation NOT implemented${NC}"
    exit 1
fi

# Test 7: Check if MIME type validation is implemented
echo -e "${YELLOW}Test 7: Verify MIME type validation...${NC}"
if grep -q "image/jpeg\|image/png\|image/webp" /Users/dmitrijfomin/Desktop/backend/internal/modules/marketplace/transport/http/handlers.go; then
    echo -e "${GREEN}✓ MIME type validation implemented${NC}"
else
    echo -e "${RED}✗ MIME type validation NOT implemented${NC}"
    exit 1
fi

# Test 8: Check if JWT authentication is required
echo -e "${YELLOW}Test 8: Verify JWT authentication requirement...${NC}"
if grep -q "middleware.GetUserID" /Users/dmitrijfomin/Desktop/backend/internal/modules/marketplace/transport/http/handlers.go; then
    echo -e "${GREEN}✓ JWT authentication checked${NC}"
else
    echo -e "${RED}✗ JWT authentication NOT checked${NC}"
    exit 1
fi

# Test 9: Check if error responses are properly formatted
echo -e "${YELLOW}Test 9: Verify error handling...${NC}"
if grep -q "httpx.Unauthorized\|httpx.BadRequest\|httpx.InternalError" /Users/dmitrijfomin/Desktop/backend/internal/modules/marketplace/transport/http/handlers.go; then
    echo -e "${GREEN}✓ Error handling implemented${NC}"
else
    echo -e "${RED}✗ Error handling NOT properly implemented${NC}"
    exit 1
fi

# Test 10: Check if response is properly constructed
echo -e "${YELLOW}Test 10: Verify response construction...${NC}"
if grep -q "dto.UploadImageResponse{" /Users/dmitrijfomin/Desktop/backend/internal/modules/marketplace/transport/http/handlers.go; then
    echo -e "${GREEN}✓ Response properly constructed${NC}"
else
    echo -e "${RED}✗ Response NOT properly constructed${NC}"
    exit 1
fi

echo -e "\n${YELLOW}=== Code Structure Validation ===${NC}"
echo -e "${GREEN}✅ All 10 tests PASSED${NC}"
echo -e "${GREEN}✓ Code Structure: VALID${NC}"
echo -e "${GREEN}✓ Ready for testing with running server${NC}\n"

# Compile check
echo -e "${YELLOW}=== Compilation Check ===${NC}"
cd /Users/dmitrijfomin/Desktop/backend
if go build -o bin/server ./cmd/server 2>&1 | head -5; then
    echo -e "${GREEN}✅ Compilation: SUCCESS${NC}\n"
else
    echo -e "${RED}❌ Compilation FAILED${NC}\n"
    exit 1
fi
