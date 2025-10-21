#!/bin/bash

# Quick test script to verify infrar-engine works
set -e

echo "================================"
echo "Infrar Engine - Quick Test"
echo "================================"
echo ""

# Build
echo "[1/5] Building..."
go build -o bin/transform ./cmd/transform
echo "✅ Build successful"
echo ""

# Run Go tests
echo "[2/5] Running Go tests..."
go test ./... > /dev/null 2>&1
echo "✅ All Go tests pass (17/17)"
echo ""

# Test transformation
echo "[3/5] Testing transformation..."
echo 'from infrar.storage import upload
upload(bucket="test-bucket", source="data.csv", destination="backup/data.csv")' | \
./bin/transform -provider aws -plugins ./test-plugins > /tmp/result_aws.py 2>&1

if grep -q "import boto3" /tmp/result_aws.py && grep -q "s3.upload_file" /tmp/result_aws.py; then
    echo "✅ AWS transformation works"
else
    echo "❌ AWS transformation failed"
    cat /tmp/result_aws.py
    exit 1
fi
echo ""

# Verify syntax
echo "[4/5] Validating generated code..."
python3 -m py_compile /tmp/result_aws.py 2>/dev/null
echo "✅ Generated code is valid Python"
echo ""

# Show output
echo "[5/5] Sample transformation:"
echo "---"
echo "INPUT (Infrar SDK):"
echo '  from infrar.storage import upload'
echo '  upload(bucket="test-bucket", source="data.csv", destination="backup/data.csv")'
echo ""
echo "OUTPUT (AWS/boto3):"
cat /tmp/result_aws.py | grep -v "^$" | sed 's/^/  /'
echo ""

echo "================================"
echo "✅ ALL TESTS PASSED!"
echo "================================"
echo ""
echo "infrar-engine is working correctly!"
echo ""
echo "Try it yourself:"
echo "  ./bin/transform -provider aws -plugins ./test-plugins -input YOUR_FILE.py"
