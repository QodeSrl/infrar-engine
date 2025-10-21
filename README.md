# Infrar Engine

**Version**: 1.0.0 (MVP)
**License**: Apache 2.0
**Language**: Go 1.21+

The core transformation engine for Infrar - converts provider-agnostic code (Infrar SDK) into native cloud provider SDK code at deployment time, enabling true multi-cloud portability with **zero runtime overhead**.

## 🚀 What is Infrar Engine?

Infrar Engine uses **compile-time code transformation** to convert your infrastructure-agnostic code into provider-specific implementations:

```
User Code (Infrar SDK)  →  Infrar Engine  →  Provider Code (Native SDK)
     (GitHub repo)        (Transformation)      (Deployed to cloud)
```

### Example Transformation

**Input** (Infrar SDK):
```python
from infrar.storage import upload

upload(bucket='my-bucket', source='file.txt', destination='backup/file.txt')
```

**Output** (AWS/boto3):
```python
import boto3

s3 = boto3.client('s3')
s3.upload_file('file.txt', 'my-bucket', 'backup/file.txt')
```

**Output** (GCP/Cloud Storage):
```python
from google.cloud import storage

storage_client = storage.Client()
bucket = storage_client.bucket('my-bucket')
blob = bucket.blob('backup/file.txt')
blob.upload_from_filename('file.txt')
```

## ✨ Key Features

- ✅ **Zero Runtime Overhead** - Code is transformed at deployment time, not runtime
- ✅ **AST-Based Transformation** - Accurate code parsing using language-native parsers
- ✅ **Plugin Architecture** - Extensible transformation rules via YAML
- ✅ **Multi-Provider Support** - AWS, GCP, Azure (MVP: AWS + GCP for storage)
- ✅ **Validation** - Generated code is validated for syntax correctness
- ✅ **Type-Safe** - Full type system for transformation pipeline

## 📦 Installation

### Prerequisites

- Go 1.21 or higher
- Python 3.8+ (for Python AST parsing)

### Build from Source

```bash
git clone https://github.com/QodeSrl/infrar-engine.git
cd infrar-engine
go build -o bin/transform ./cmd/transform
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run specific package
go test ./pkg/parser -v
```

## 🎯 Quick Start

### 1. Create Sample Python File

```python
# app.py
from infrar.storage import upload

def backup_data():
    upload(
        bucket='my-data-bucket',
        source='/tmp/data.csv',
        destination='backups/data.csv'
    )
```

### 2. Transform to AWS

```bash
./bin/transform -provider aws -input app.py -output app_aws.py
```

**Output** (`app_aws.py`):
```python
import boto3

s3 = boto3.client('s3')

def backup_data():
    s3.upload_file('/tmp/data.csv', 'my-data-bucket', 'backups/data.csv')
```

### 3. Transform to GCP

```bash
./bin/transform -provider gcp -input app.py -output app_gcp.py
```

## 🏗️ Architecture

The transformation pipeline consists of 6 core components:

```
┌──────────┐   ┌──────────┐   ┌───────────┐   ┌──────────┐   ┌───────────┐   ┌──────────┐
│  Parser  │──>│ Detector │──>│Transformer│──>│ Generator│──>│ Validator │──>│  Result  │
└──────────┘   └──────────┘   └───────────┘   └──────────┘   └───────────┘   └──────────┘
     │              │               │                │              │
     ▼              ▼               ▼                ▼              ▼
   AST          Infrar Calls   Transformed     Final Code      Validated
                                 Calls                           Code
```

### Components

1. **Parser** (`pkg/parser`) - Parses source code into AST using Python's native parser
2. **Detector** (`pkg/detector`) - Identifies Infrar SDK calls in the AST
3. **Plugin Loader** (`pkg/plugin`) - Loads transformation rules from YAML files
4. **Transformer** (`pkg/transformer`) - Applies rules to generate provider code
5. **Generator** (`pkg/generator`) - Combines transformed calls into final code
6. **Validator** (`pkg/validator`) - Validates generated code syntax

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed technical documentation.

## 🔌 Plugin System

Transformation rules are defined in YAML files:

```yaml
# storage/aws/rules.yaml
operations:
  - name: upload
    pattern: "infrar.storage.upload"
    target:
      provider: aws
      service: s3

    transformation:
      imports:
        - "import boto3"

      setup_code: |
        s3 = boto3.client('s3')

      code_template: |
        s3.upload_file(
            {{ .source }},
            {{ .bucket }},
            {{ .destination }}
        )

      parameter_mapping:
        bucket: bucket
        source: source
        destination: destination

    requirements:
      - package: boto3
        version: ">=1.28.0"
```

**Plugin Locations**:
- **Production plugins**: [infrar-plugins](https://github.com/QodeSrl/infrar-plugins) repository (`../infrar-plugins/packages`)
- **Test plugins**: `./test-plugins` directory (for local development and testing)

Use the production plugins repository for actual transformations. The test-plugins directory is kept for development convenience.

## 📚 Usage

### As a Library

```go
package main

import (
    "fmt"
    "github.com/QodeSrl/infrar-engine/pkg/engine"
    "github.com/QodeSrl/infrar-engine/pkg/types"
)

func main() {
    // Create engine
    eng, err := engine.New()
    if err != nil {
        panic(err)
    }

    // Load transformation rules
    err = eng.LoadRules("../infrar-plugins/packages", types.ProviderAWS, "storage")
    if err != nil {
        panic(err)
    }

    // Transform code
    sourceCode := `
from infrar.storage import upload

upload(bucket='data', source='file.txt', destination='file.txt')
    `

    result, err := eng.Transform(sourceCode, types.ProviderAWS)
    if err != nil {
        panic(err)
    }

    fmt.Println(result.TransformedCode)
}
```

### CLI Tool

```bash
# Transform from stdin
echo "from infrar.storage import upload" | ./bin/transform -provider aws

# Transform file
./bin/transform -provider aws -input app.py -output app_aws.py

# Transform to GCP
./bin/transform -provider gcp -input app.py -output app_gcp.py

# Specify plugin directory
./bin/transform -provider aws -plugins ./custom-plugins -input app.py
```

### CLI Options

```
-provider string
    Target cloud provider (aws, gcp, azure) (default "aws")

-plugins string
    Path to plugins directory (default "../infrar-plugins/packages")

-capability string
    Capability to transform (storage, database, etc.) (default "storage")

-input string
    Input file to transform (or use stdin)

-output string
    Output file (or use stdout)
```

## 🧪 Testing

### Test Coverage

Current test coverage (MVP):

- ✅ Parser: 100% (all tests passing)
- ✅ Detector: 100% (all tests passing)
- ✅ Plugin Loader: 100% (all tests passing)
- ✅ Transformer: 100% (all tests passing)
- ✅ Generator: 100% (all tests passing)
- ✅ Validator: 100% (all tests passing)
- ✅ Engine (E2E): 100% (all tests passing)

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 🛠️ Development

### Project Structure

```
infrar-engine/
├── cmd/
│   └── transform/          # CLI tool
├── pkg/
│   ├── types/              # Core type definitions
│   ├── parser/             # AST parsing (Python)
│   ├── detector/           # Infrar call detection
│   ├── plugin/             # Plugin loader & registry
│   ├── transformer/        # Core transformation logic
│   ├── generator/          # Code generation
│   ├── validator/          # Code validation
│   └── engine/             # Main engine (public API)
├── internal/
│   └── util/               # Internal utilities
├── tests/
│   ├── integration/        # Integration tests
│   └── fixtures/           # Test fixtures
├── go.mod
├── go.sum
├── README.md               # This file
├── ARCHITECTURE.md         # Technical architecture
└── LICENSE
```

## 📊 Performance

Target performance metrics (MVP):

- Transform 100 lines in < 100ms ✅
- Support files up to 10,000 lines ✅
- Cache parsed ASTs for repeated transformations 🚧
- Parallel transformation of multiple files 🚧

## 🗺️ Roadmap

### Phase 1 (MVP) - ✅ COMPLETED
- [x] Python AST parser
- [x] Infrar call detector
- [x] Plugin system with YAML rules
- [x] Transformation engine
- [x] Code generator
- [x] Code validator
- [x] AWS S3 transformations
- [x] GCP Cloud Storage transformations
- [x] CLI tool
- [x] Comprehensive test suite

### Phase 2 (Next)
- [ ] Node.js/TypeScript support
- [ ] Database capability (RDS, Cloud SQL)
- [ ] Messaging capability (SQS, Pub/Sub)
- [ ] Azure support
- [ ] Performance optimizations (caching, parallelization)

### Phase 3 (Future)
- [ ] Go language support
- [ ] Multi-file project transformation
- [ ] IDE integration (VS Code extension)
- [ ] Language Server Protocol (LSP)
- [ ] Advanced optimizations

## 📝 License

Apache License 2.0 - see [LICENSE](LICENSE) file for details.

## 📧 Support

- **Issues**: https://github.com/QodeSrl/infrar-engine/issues
- **Docs**: https://docs.infrar.io
- **Email**: support@infrar.io

---

**Made with ❤️ by the Infrar Team**

[Website](https://infrar.io) • [Documentation](https://docs.infrar.io) • [GitHub](https://github.com/QodeSrl/infrar-engine)
