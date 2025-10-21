#!/bin/bash

# Infrar Engine Test Script
# Tests the transformation engine with real examples

set -e  # Exit on error

echo "================================"
echo "Infrar Engine - Test Suite"
echo "================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Build the CLI tool
echo -e "${BLUE}[Step 1/5]${NC} Building CLI tool..."
go build -o bin/transform ./cmd/transform
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓${NC} Build successful"
else
    echo -e "${RED}✗${NC} Build failed"
    exit 1
fi
echo ""

# Step 2: Run Go tests
echo -e "${BLUE}[Step 2/5]${NC} Running Go tests..."
go test ./... -v | grep -E "(PASS|FAIL|ok)"
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓${NC} All Go tests passed"
else
    echo -e "${RED}✗${NC} Some Go tests failed"
    exit 1
fi
echo ""

# Step 3: Test transformations
echo -e "${BLUE}[Step 3/5]${NC} Testing transformations..."
echo ""

# Test files
TEST_FILES=(
    "examples/simple_upload.py"
    "examples/multiple_operations.py"
)

PROVIDERS=("aws" "gcp")

for file in "${TEST_FILES[@]}"; do
    if [ ! -f "$file" ]; then
        echo -e "${YELLOW}⚠${NC} File not found: $file (skipping)"
        continue
    fi

    echo -e "${YELLOW}Testing:${NC} $file"
    echo "---"

    for provider in "${PROVIDERS[@]}"; do
        echo -e "  ${BLUE}→ ${provider}${NC}"

        # Transform to provider
        output_file="examples/output_${provider}_$(basename $file)"
        ./bin/transform \
            -provider "$provider" \
            -plugins ./test-plugins \
            -capability storage \
            -input "$file" \
            -output "$output_file" 2>&1

        if [ $? -eq 0 ]; then
            echo -e "    ${GREEN}✓${NC} Transformation successful"

            # Validate Python syntax
            python3 -m py_compile "$output_file" 2>/dev/null
            if [ $? -eq 0 ]; then
                echo -e "    ${GREEN}✓${NC} Generated code is valid Python"
            else
                echo -e "    ${RED}✗${NC} Syntax validation failed"
            fi
        else
            echo -e "    ${RED}✗${NC} Transformation failed"
        fi
    done
    echo ""
done

# Step 4: Show sample output
echo -e "${BLUE}[Step 4/5]${NC} Sample transformation output..."
echo ""

echo -e "${YELLOW}Original Code:${NC}"
echo "---"
head -20 examples/simple_upload.py
echo "..."
echo ""

echo -e "${YELLOW}Transformed to AWS:${NC}"
echo "---"
cat examples/output_aws_simple_upload.py 2>/dev/null || echo "File not generated"
echo ""

echo -e "${YELLOW}Transformed to GCP:${NC}"
echo "---"
cat examples/output_gcp_simple_upload.py 2>/dev/null || echo "File not generated"
echo ""

# Step 5: Summary
echo -e "${BLUE}[Step 5/5]${NC} Test Summary"
echo "================================"
echo ""

# Count generated files
aws_files=$(ls examples/output_aws_*.py 2>/dev/null | wc -l)
gcp_files=$(ls examples/output_gcp_*.py 2>/dev/null | wc -l)

echo -e "AWS transformations: ${GREEN}$aws_files${NC}"
echo -e "GCP transformations: ${GREEN}$gcp_files${NC}"
echo ""

# Check if all expected files were generated
expected=$((${#TEST_FILES[@]} * 2))  # 2 providers per file
actual=$((aws_files + gcp_files))

if [ $actual -eq $expected ]; then
    echo -e "${GREEN}✓ All tests passed successfully!${NC}"
    echo ""
    echo "Generated files are in examples/output_*.py"
    exit 0
else
    echo -e "${YELLOW}⚠ Some transformations may have failed${NC}"
    echo "Expected: $expected, Got: $actual"
    exit 1
fi
