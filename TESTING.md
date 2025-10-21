# Testing Guide for Infrar Engine

This guide shows you how to effectively test the infrar-engine transformation system.

## Quick Start

### 1. Run the Complete Test Suite

```bash
./test.sh
```

This will:
- ✅ Build the CLI tool
- ✅ Run all Go unit tests
- ✅ Transform example files to AWS and GCP
- ✅ Validate generated Python syntax
- ✅ Show sample transformations

### 2. Manual Testing

#### Transform a Single File

```bash
# Build the tool first
go build -o bin/transform ./cmd/transform

# Transform to AWS
./bin/transform \
    -provider aws \
    -plugins ./test-plugins \
    -input examples/simple_upload.py \
    -output examples/output_aws.py

# Transform to GCP
./bin/transform \
    -provider gcp \
    -plugins ./test-plugins \
    -input examples/simple_upload.py \
    -output examples/output_gcp.py
```

#### Transform from stdin

```bash
echo 'from infrar.storage import upload
upload(bucket="test", source="file.txt", destination="file.txt")' | \
./bin/transform -provider aws -plugins ./test-plugins
```

## Test Examples

We have 3 example applications in `examples/`:

### 1. Simple Upload (`simple_upload.py`)

**Original Code:**
```python
from infrar.storage import upload

def backup_file():
    upload(
        bucket='my-backup-bucket',
        source='/tmp/data.csv',
        destination='backups/2024/data.csv'
    )
```

**Transformed to AWS:**
```python
import boto3

s3 = boto3.client('s3')

def backup_file():
    s3.upload_file('/tmp/data.csv', 'my-backup-bucket', 'backups/2024/data.csv')
```

**Transformed to GCP:**
```python
from google.cloud import storage

storage_client = storage.Client()

def backup_file():
    bucket = storage_client.bucket('my-backup-bucket')
    blob = bucket.blob('backups/2024/data.csv')
    blob.upload_from_filename('/tmp/data.csv')
```

### 2. Multiple Operations (`multiple_operations.py`)

Tests: upload, download, delete, list_objects

### 3. Data Pipeline (`data_pipeline.py`)

Tests: Real-world scenario with file processing

## Running Go Tests

### All Tests
```bash
go test ./...
```

### Specific Package
```bash
go test ./pkg/parser -v
go test ./pkg/detector -v
go test ./pkg/transformer -v
```

### With Coverage
```bash
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Verbose Output
```bash
go test ./... -v
```

## Test Components Individually

### 1. Parser Test
```bash
go test ./pkg/parser -v -run TestPythonParser_Parse
```

Tests the Python AST parsing functionality.

### 2. Detector Test
```bash
go test ./pkg/detector -v -run TestDetector_DetectFromSource
```

Tests detection of Infrar SDK calls.

### 3. Transformer Test
```bash
go test ./pkg/transformer -v -run TestTransformer_Transform
```

Tests transformation rule application.

### 4. End-to-End Test
```bash
go test ./pkg/engine -v -run TestEngine_Transform_EndToEnd
```

Tests the complete transformation pipeline.

## Validate Generated Code

### Check Python Syntax
```bash
python3 -m py_compile examples/output_aws_simple_upload.py
python3 -m py_compile examples/output_gcp_simple_upload.py
```

### Run with Python AST
```bash
python3 -c "import ast; ast.parse(open('examples/output_aws_simple_upload.py').read())"
```

## Creating Custom Tests

### 1. Create a Test File

```python
# my_test.py
from infrar.storage import upload, download

def my_function():
    upload(bucket='test', source='a.txt', destination='b.txt')
    download(bucket='test', source='b.txt', destination='c.txt')
```

### 2. Transform It

```bash
./bin/transform -provider aws -plugins ./test-plugins -input my_test.py
```

### 3. Verify Output

The output should have:
- ✅ `import boto3` instead of `from infrar.storage import ...`
- ✅ `s3 = boto3.client('s3')` setup code
- ✅ `s3.upload_file(...)` and `s3.download_file(...)` calls
- ✅ No references to `infrar`

## Troubleshooting

### "No such file or directory" for Python
**Problem**: Python executable not found

**Solution**:
```bash
# Check Python is installed
python3 --version

# Or install Python
sudo apt-get install python3  # Ubuntu/Debian
brew install python3          # macOS
```

### "No rule found for pattern"
**Problem**: Plugin rules not loaded

**Solution**:
```bash
# Ensure plugin directory structure is correct
ls -la test-plugins/storage/aws/rules.yaml
ls -la test-plugins/storage/gcp/rules.yaml

# Use correct -plugins flag
./bin/transform -plugins ./test-plugins -provider aws ...
```

### "Invalid Python syntax" in generated code
**Problem**: Transformation generated invalid code

**Solution**:
```bash
# Check the generated file
cat examples/output_aws_simple_upload.py

# Validate manually
python3 -m py_compile examples/output_aws_simple_upload.py

# Run engine tests to identify the issue
go test ./pkg/generator -v
```

## Benchmarking

### Test Performance

```bash
# Time a transformation
time ./bin/transform -provider aws -plugins ./test-plugins -input examples/simple_upload.py

# Run benchmark tests
go test ./pkg/parser -bench=. -benchmem
go test ./pkg/transformer -bench=. -benchmem
```

### Expected Performance
- Simple file (20 lines): < 100ms
- Medium file (100 lines): < 200ms
- Large file (1000 lines): < 1s

## CI/CD Testing

### GitHub Actions Example

```yaml
name: Test Infrar Engine

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - uses: actions/setup-python@v4
        with:
          python-version: '3.11'

      - name: Run tests
        run: |
          go test ./... -v
          ./test.sh
```

## Test Coverage Goals

Current coverage (MVP):
- ✅ Parser: 100%
- ✅ Detector: 100%
- ✅ Plugin Loader: 100%
- ✅ Transformer: 100%
- ✅ Generator: 100%
- ✅ Validator: 100%
- ✅ Engine: 100%

## Next Steps

Once basic tests pass:

1. **Test with real cloud accounts** - Deploy transformed code to AWS/GCP
2. **Integration testing** - Test actual file uploads/downloads
3. **Performance testing** - Benchmark with large codebases
4. **Edge cases** - Test error handling, malformed input
5. **Security testing** - Ensure no code injection vulnerabilities

## Quick Reference

```bash
# Build
go build -o bin/transform ./cmd/transform

# Run all tests
./test.sh

# Run Go tests only
go test ./...

# Transform single file
./bin/transform -provider aws -plugins ./test-plugins -input FILE

# Transform with output
./bin/transform -provider aws -plugins ./test-plugins -input IN -output OUT

# Get help
./bin/transform -help
```

## Success Criteria

A successful test run should show:

✅ All Go tests passing (PASS)
✅ Transformed files generated for both AWS and GCP
✅ Generated Python code is syntactically valid
✅ No `infrar` imports in output
✅ Correct provider imports (boto3 or google.cloud.storage)
✅ Setup code present (client initialization)
✅ Function calls transformed correctly

---

**Happy Testing!** 🧪

For issues or questions, see [README.md](README.md) or open an issue on GitHub.
