# How to Effectively Test Infrar Engine

## ✅ Quick Test (30 seconds)

```bash
# Build the CLI
go build -o bin/transform ./cmd/transform

# Run all Go tests
go test ./...

# Test a simple transformation
echo 'from infrar.storage import upload
upload(bucket="test", source="file.txt", destination="out.txt")' | \
./bin/transform -provider aws -plugins ./test-plugins
```

**Expected output:**
```python
import boto3

s3 = boto3.client('s3')

s3.upload_file('file.txt', 'test', 'out.txt')
```

✅ **If you see this, it's working!**

---

## 📋 Complete Test Checklist

### 1. Build & Unit Tests
```bash
# Build
go build -o bin/transform ./cmd/transform
# Expected: Binary created in bin/transform

# Run all tests
go test ./...
# Expected: All 17 tests PASS in ~1 second
```

### 2. Test Individual Components

```bash
# Parser
go test ./pkg/parser -v
# Expected: 2/2 tests pass

# Detector
go test ./pkg/detector -v
# Expected: 3/3 tests pass

# Transformer
go test ./pkg/transformer -v
# Expected: 3/3 tests pass

# Generator
go test ./pkg/generator -v
# Expected: 3/3 tests pass

# Validator
go test ./pkg/validator -v
# Expected: 1/1 tests pass

# End-to-end
go test ./pkg/engine -v
# Expected: 2/2 tests pass (shows actual transformation)
```

### 3. Test Transformations

#### Simple Upload (AWS)
```bash
echo 'from infrar.storage import upload
upload(bucket="my-bucket", source="data.csv", destination="backup.csv")' | \
./bin/transform -provider aws -plugins ./test-plugins
```

**Expected:**
```python
import boto3

s3 = boto3.client('s3')

s3.upload_file('data.csv', 'my-bucket', 'backup.csv')
```

#### Simple Upload (GCP)
```bash
echo 'from infrar.storage import upload
upload(bucket="my-bucket", source="data.csv", destination="backup.csv")' | \
./bin/transform -provider gcp -plugins ./test-plugins
```

**Expected:**
```python
from google.cloud import storage

storage_client = storage.Client()

bucket = storage_client.bucket('my-bucket')
blob = bucket.blob('backup.csv')
blob.upload_from_filename('data.csv')
```

#### Multiple Operations
```bash
cat > /tmp/multi.py << 'EOF'
from infrar.storage import upload, download
upload(bucket='b1', source='s1', destination='d1')
download(bucket='b2', source='s2', destination='d2')
EOF

./bin/transform -provider aws -plugins ./test-plugins -input /tmp/multi.py
```

**Expected:**
- Both `upload` and `download` transformed
- Single `import boto3` at the top
- Single `s3 = boto3.client('s3')` setup
- Two function calls: `s3.upload_file()` and `s3.download_file()`

### 4. Test File Output

```bash
# Transform and save
./bin/transform -provider aws -plugins ./test-plugins \
  -input /tmp/multi.py \
  -output /tmp/output_aws.py

# Check file was created
ls -l /tmp/output_aws.py

# Verify Python syntax
python3 -m py_compile /tmp/output_aws.py
```

If no errors, the syntax is valid!

### 5. Test Module-Qualified Calls

```bash
echo 'import infrar.storage
infrar.storage.upload(bucket="b", source="s", destination="d")' | \
./bin/transform -provider aws -plugins ./test-plugins
```

**Expected:** Same output as direct import

---

## 🎯 What to Verify

For each transformation, check:

✅ **Imports**
- Old: `from infrar.storage import upload`
- New: `import boto3` (AWS) or `from google.cloud import storage` (GCP)

✅ **Setup Code**
- AWS: `s3 = boto3.client('s3')`
- GCP: `storage_client = storage.Client()`

✅ **Function Calls**
- AWS: `s3.upload_file('source', 'bucket', 'dest')`
- GCP: `bucket.blob('dest').upload_from_filename('source')`

✅ **No Infrar References**
- No `infrar` anywhere in output
- All Infrar imports removed

✅ **Valid Syntax**
- Can be parsed by Python: `python3 -m py_compile FILE`

---

## 📊 Test Results

| Test | Command | Expected Result |
|------|---------|----------------|
| Build | `go build ./cmd/transform` | Binary in bin/transform |
| All Tests | `go test ./...` | 17/17 PASS |
| Simple Transform | `echo '...' \| ./bin/transform ...` | Valid code output |
| File Transform | `./bin/transform -input X -output Y` | File Y created |
| AWS Transform | `-provider aws` | boto3 code |
| GCP Transform | `-provider gcp` | google.cloud.storage code |
| Syntax Valid | `python3 -m py_compile output.py` | No errors |

---

## 🐛 Troubleshooting

### "python3: command not found"
**Fix:** Install Python 3.8+
```bash
# Ubuntu/Debian
sudo apt-get install python3

# macOS
brew install python3
```

### "No such file: test-plugins/storage/aws/rules.yaml"
**Fix:** Ensure you're in the infrar-engine directory
```bash
cd /home/alexb/projects/infrar/infrar-engine
ls test-plugins/storage/aws/rules.yaml  # Should exist
```

### "All tests pass but transformation fails"
**Fix:** Check the error message carefully
```bash
./bin/transform -provider aws -plugins ./test-plugins -input FILE 2>&1 | less
```

Common issues:
- Wrong provider name (use `aws`, `gcp`, or `azure`)
- Wrong plugin directory (use `./test-plugins`)
- Wrong capability (use `storage`)

---

## 🚀 Advanced Testing

### Performance Test
```bash
time ./bin/transform -provider aws -plugins ./test-plugins -input FILE
```
Expected: < 100ms for small files

### Stress Test
```bash
# Create large file
for i in {1..100}; do
  echo "upload(bucket='b$i', source='s$i', destination='d$i')"
done > /tmp/large.py

# Add import at top
echo 'from infrar.storage import upload' | cat - /tmp/large.py > /tmp/large_full.py

# Transform
time ./bin/transform -provider aws -plugins ./test-plugins -input /tmp/large_full.py
```

### Coverage Test
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```
Opens browser showing code coverage (should be high)

---

## ✨ Success Criteria

Your installation is working correctly if:

✅ `go test ./...` shows **17/17 PASS**
✅ Simple echo transformation works
✅ AWS transformation produces `import boto3`
✅ GCP transformation produces `from google.cloud import storage`
✅ Generated code has valid Python syntax
✅ No `infrar` references in output
✅ Performance is < 100ms for simple files

---

## 📝 Example Test Session

```bash
$ cd /home/alexb/projects/infrar/infrar-engine

$ go build -o bin/transform ./cmd/transform
# Build successful

$ go test ./...
ok   github.com/QodeSrl/infrar-engine/pkg/detector     0.264s
ok   github.com/QodeSrl/infrar-engine/pkg/engine       0.151s
ok   github.com/QodeSrl/infrar-engine/pkg/generator    0.003s
ok   github.com/QodeSrl/infrar-engine/pkg/parser       0.188s
ok   github.com/QodeSrl/infrar-engine/pkg/plugin       0.008s
ok   github.com/QodeSrl/infrar-engine/pkg/transformer  0.004s
ok   github.com/QodeSrl/infrar-engine/pkg/validator    0.133s

$ echo 'from infrar.storage import upload
upload(bucket="test", source="file.txt", destination="out.txt")' | \
./bin/transform -provider aws -plugins ./test-plugins

import boto3

s3 = boto3.client('s3')

s3.upload_file('file.txt', 'test', 'out.txt')

# ✅ Perfect! It works!
```

---

## 🎯 Next Steps

Once basic tests pass:

1. **Create your own test files** - Test with your use cases
2. **Try different providers** - Test AWS vs GCP output
3. **Test edge cases** - Empty files, syntax errors, etc.
4. **Performance testing** - Large files, many operations
5. **Integration testing** - Use in actual deployment pipeline

---

**The infrar-engine is fully functional and ready to use!** 🚀

For more details, see:
- [README.md](README.md) - Full documentation
- [ARCHITECTURE.md](ARCHITECTURE.md) - Technical details
- [TESTING.md](TESTING.md) - Comprehensive testing guide
