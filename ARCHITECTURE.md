# Infrar Engine Architecture

**Version**: 1.0.0
**Status**: Phase 1 - MVP Development
**Language**: Go
**License**: GPL v3

## Overview

The Infrar Engine is the core transformation component that converts provider-agnostic code (using Infrar SDK) into native cloud provider SDK code at deployment time. This enables true multi-cloud portability with **zero runtime overhead**.

## Core Principle

**Compile-Time Transformation** - Unlike traditional abstraction layers that add runtime overhead, Infrar transforms code **before deployment**:

```
User Code (infrar SDK)  →  Transformation Engine  →  Provider Code (native SDK)
     (GitHub repo)              (infrar-engine)           (deployed to cloud)
```

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Transformation Pipeline                  │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────┐   ┌──────────┐   ┌───────────┐   ┌────────┐  │
│  │  Parser  │──>│ Detector │──>│Transformer│──>│Generator│  │
│  └──────────┘   └──────────┘   └───────────┘   └────────┘  │
│       │                              │              │        │
│       ▼                              ▼              ▼        │
│  ┌──────────┐   ┌──────────┐   ┌─────────┐   ┌────────┐    │
│  │   AST    │   │  Plugin  │   │  Rules  │   │ Native │    │
│  │  Tree    │   │  Loader  │   │ Engine  │   │  Code  │    │
│  └──────────┘   └──────────┘   └─────────┘   └────────┘    │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Validator (syntax check)                 │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Component Breakdown

### 1. Parser

**Purpose**: Parse source code into an Abstract Syntax Tree (AST) for analysis

**Input**:
- Source code files (Python initially)
- Language type

**Output**:
- AST representation (JSON format for cross-language compatibility)
- Parse metadata (imports, function calls, line numbers)

**Implementation Strategy**:
Since we're writing in Go but need to parse Python:

**Option A - Python subprocess** (Recommended for MVP):
```go
// Use Python's own ast module via subprocess
// Most accurate, leverages Python's official parser
cmd := exec.Command("python3", "-m", "json.tool")
// Parse output JSON
```

**Option B - Tree-sitter** (Future):
```go
// Universal parser library
// Better performance, no Python dependency
// More complex setup
```

**Files**:
- `parser/python.go` - Python AST parser
- `parser/ast.go` - AST data structures
- `parser/interface.go` - Parser interface for extensibility

### 2. Detector

**Purpose**: Identify Infrar SDK usage in the AST

**What it detects**:
```python
# Imports to detect:
from infrar.storage import upload, download
import infrar.storage

# Function calls to detect:
infrar.storage.upload(bucket='data', source='file.txt', destination='file.txt')
upload(bucket='data', source='file.txt', destination='file.txt')
```

**Detection Algorithm**:
1. Walk the AST tree
2. Find `ImportFrom` nodes matching `infrar.*`
3. Track imported symbols
4. Find `Call` nodes using those symbols
5. Extract function names and arguments

**Output**:
```go
type InfrarCall struct {
    Module       string            // "infrar.storage"
    Function     string            // "upload"
    Arguments    map[string]Value  // {bucket: "data", source: "file.txt", ...}
    LineNumber   int
    ColumnOffset int
    SourceCode   string            // Original code snippet
}
```

**Files**:
- `detector/detector.go` - Main detection logic
- `detector/patterns.go` - Pattern matching for Infrar calls

### 3. Plugin Loader

**Purpose**: Load and parse transformation rules from infrar-plugins repository

**Plugin Structure** (from infrar-plugins):
```yaml
# packages/storage/aws/rules.yaml
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
            Filename={{ .source }},
            Bucket={{ .bucket }},
            Key={{ .destination }}
        )

      parameter_mapping:
        bucket: Bucket
        source: Filename
        destination: Key
```

**Files**:
- `plugin/loader.go` - Load YAML plugin files
- `plugin/rule.go` - Transformation rule data structures
- `plugin/registry.go` - Plugin registry and lookup

### 4. Transformer

**Purpose**: Core transformation engine that applies rules to generate provider-specific code

**Transformation Flow**:
```
1. Receive InfrarCall objects from Detector
2. For each call:
   a. Look up transformation rule (provider + function)
   b. Extract parameter values
   c. Validate required parameters exist
   d. Generate code using template
   e. Track imports needed
3. Return TransformationResult
```

**Template Engine**:
Uses Go's `text/template` for code generation:
```go
template := `s3.upload_file(
    Filename={{ .source }},
    Bucket={{ .bucket }},
    Key={{ .destination }}
)`

data := map[string]string{
    "source": "'local.txt'",
    "bucket": "'my-bucket'",
    "destination": "'remote.txt'",
}

// Renders to:
// s3.upload_file(
//     Filename='local.txt',
//     Bucket='my-bucket',
//     Key='remote.txt'
// )
```

**Files**:
- `transformer/transformer.go` - Main transformation logic
- `transformer/template.go` - Template engine
- `transformer/context.go` - Transformation context and state

### 5. Generator

**Purpose**: Generate final provider-specific source code files

**Responsibilities**:
- Collect all imports needed
- Generate provider client initialization code
- Insert transformed function calls
- Preserve non-Infrar code as-is
- Format code properly

**Example Transformation**:

**Input** (user code):
```python
from infrar.storage import upload

def backup_data():
    upload(
        bucket='backups',
        source='/tmp/data.csv',
        destination='data/2024/backup.csv'
    )
```

**Output** (AWS):
```python
import boto3

s3 = boto3.client('s3')

def backup_data():
    s3.upload_file(
        Filename='/tmp/data.csv',
        Bucket='backups',
        Key='data/2024/backup.csv'
    )
```

**Output** (GCP):
```python
from google.cloud import storage

storage_client = storage.Client()

def backup_data():
    bucket = storage_client.bucket('backups')
    blob = bucket.blob('data/2024/backup.csv')
    blob.upload_from_filename('/tmp/data.csv')
```

**Code Rewriting Strategy**:
1. Parse original code to AST
2. Identify Infrar calls and their positions
3. Replace Infrar calls with provider code
4. Replace imports
5. Unparse modified AST back to source code

**Files**:
- `generator/generator.go` - Main code generation
- `generator/aws.go` - AWS-specific code patterns
- `generator/gcp.go` - GCP-specific code patterns
- `generator/imports.go` - Import management
- `generator/rewriter.go` - AST rewriting

### 6. Validator

**Purpose**: Validate generated code is syntactically correct

**Validation Steps**:
1. Parse generated code (ensure it's valid Python)
2. Check imports are valid
3. Verify indentation
4. Run basic linting

**Implementation**:
```go
func ValidatePythonCode(code string) error {
    // Use Python's ast.parse to validate
    cmd := exec.Command("python3", "-m", "py_compile", "-")
    cmd.Stdin = strings.NewReader(code)

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("invalid Python syntax: %w", err)
    }

    return nil
}
```

**Files**:
- `validator/validator.go` - Validation logic
- `validator/python.go` - Python-specific validation

## Data Structures

### Core Types

```go
// pkg/types/ast.go

// AST represents parsed source code
type AST struct {
    Language   Language
    Nodes      []Node
    Imports    []Import
    SourceCode string
    Filepath   string
}

type Node struct {
    Type         string          // "ImportFrom", "Call", "FunctionDef", etc.
    LineNumber   int
    ColumnOffset int
    Attributes   map[string]any
    Children     []Node
}

type Import struct {
    Module  string   // "infrar.storage"
    Names   []string // ["upload", "download"]
    Alias   string   // Optional alias
}
```

```go
// pkg/types/transformation.go

// InfrarCall represents detected Infrar SDK usage
type InfrarCall struct {
    Module       string            // "infrar.storage"
    Function     string            // "upload"
    Arguments    map[string]Value
    LineNumber   int
    ColumnOffset int
    SourceCode   string
}

type Value struct {
    Type  ValueType // String, Number, Bool, Variable
    Value any
}

// TransformationRule defines how to transform an Infrar call
type TransformationRule struct {
    Name             string
    Pattern          string
    Provider         Provider
    Service          string
    Imports          []string
    SetupCode        string
    CodeTemplate     string
    ParameterMapping map[string]string
    Requirements     []Requirement
}

type Requirement struct {
    Package string
    Version string
}

// TransformationResult is the output of transformation
type TransformationResult struct {
    Provider       Provider
    TransformedCode string
    Imports        []string
    Requirements   []Requirement
    Warnings       []Warning
    Metadata       map[string]any
}
```

## Transformation Algorithm

### Main Pipeline

```go
func Transform(sourceCode string, targetProvider Provider) (*TransformationResult, error) {
    // 1. Parse source code
    ast, err := parser.ParsePython(sourceCode)
    if err != nil {
        return nil, fmt.Errorf("parse error: %w", err)
    }

    // 2. Detect Infrar calls
    calls, err := detector.DetectInfrarCalls(ast)
    if err != nil {
        return nil, fmt.Errorf("detection error: %w", err)
    }

    if len(calls) == 0 {
        // No Infrar calls found, return original code
        return &TransformationResult{
            Provider:        targetProvider,
            TransformedCode: sourceCode,
        }, nil
    }

    // 3. Load transformation rules for target provider
    rules, err := plugin.LoadRules(targetProvider)
    if err != nil {
        return nil, fmt.Errorf("plugin load error: %w", err)
    }

    // 4. Transform each call
    transformer := transformer.New(rules)
    transformedCalls := []transformer.TransformedCall{}

    for _, call := range calls {
        transformed, err := transformer.Transform(call)
        if err != nil {
            return nil, fmt.Errorf("transformation error for %s: %w", call.Function, err)
        }
        transformedCalls = append(transformedCalls, transformed)
    }

    // 5. Generate final code
    generator := generator.New(targetProvider)
    result, err := generator.Generate(ast, transformedCalls)
    if err != nil {
        return nil, fmt.Errorf("generation error: %w", err)
    }

    // 6. Validate
    if err := validator.Validate(result.TransformedCode); err != nil {
        return nil, fmt.Errorf("validation error: %w", err)
    }

    return result, nil
}
```

## Error Handling

### Error Categories

1. **Parse Errors**: Invalid source code syntax
2. **Detection Errors**: Malformed Infrar SDK usage
3. **Transformation Errors**: Missing rules or invalid parameters
4. **Generation Errors**: Failed code generation
5. **Validation Errors**: Generated code is invalid

### Error Response

```go
type TransformationError struct {
    Category   ErrorCategory
    Message    string
    Line       int
    Column     int
    SourceCode string
    Suggestion string
}

func (e *TransformationError) Error() string {
    return fmt.Sprintf("[%s] %s at line %d:%d\n%s\nSuggestion: %s",
        e.Category, e.Message, e.Line, e.Column, e.SourceCode, e.Suggestion)
}
```

## Testing Strategy

### Unit Tests

Each component has isolated unit tests:

```go
// parser/python_test.go
func TestParsePythonCode(t *testing.T) {
    code := `
from infrar.storage import upload

upload(bucket='test', source='file.txt', destination='file.txt')
    `

    ast, err := ParsePython(code)
    assert.NoError(t, err)
    assert.NotNil(t, ast)
    assert.Equal(t, "infrar.storage", ast.Imports[0].Module)
}
```

### Integration Tests

End-to-end transformation tests:

```go
// integration_test.go
func TestTransformStorageUploadToAWS(t *testing.T) {
    input := `
from infrar.storage import upload

upload(bucket='data', source='file.txt', destination='file.txt')
    `

    result, err := Transform(input, ProviderAWS)
    assert.NoError(t, err)

    expected := `
import boto3

s3 = boto3.client('s3')
s3.upload_file('file.txt', 'data', 'file.txt')
    `

    assert.Contains(t, result.TransformedCode, "import boto3")
    assert.Contains(t, result.TransformedCode, "s3.upload_file")
}
```

### Fixture Tests

Use input/output fixtures:

```
tests/
├── fixtures/
│   ├── input/
│   │   ├── simple_upload.py
│   │   ├── multiple_operations.py
│   │   └── complex_app.py
│   └── expected/
│       ├── aws/
│       │   ├── simple_upload.py
│       │   └── multiple_operations.py
│       └── gcp/
│           ├── simple_upload.py
│           └── multiple_operations.py
```

## Performance Considerations

### Optimization Goals

- Transform 100 lines of code in < 100ms
- Support files up to 10,000 lines
- Cache parsed ASTs for repeated transformations
- Parallel transformation of multiple files

### Caching Strategy

```go
type TransformationCache struct {
    mu    sync.RWMutex
    cache map[string]*TransformationResult
}

func (c *TransformationCache) Get(key string) (*TransformationResult, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    result, ok := c.cache[key]
    return result, ok
}

func cacheKey(sourceCode string, provider Provider) string {
    hash := sha256.Sum256([]byte(sourceCode + string(provider)))
    return hex.EncodeToString(hash[:])
}
```

## Configuration

### Engine Configuration

```yaml
# engine.yaml
engine:
  version: "1.0.0"

  parser:
    python_executable: "python3"
    timeout_seconds: 30

  plugins:
    directory: "../infrar-plugins/packages"
    auto_reload: true

  generator:
    format_code: true
    preserve_comments: true

  validator:
    strict_mode: true

  cache:
    enabled: true
    max_size_mb: 100
```

## Directory Structure

```
infrar-engine/
├── cmd/
│   └── transform/
│       └── main.go              # CLI for testing
│
├── pkg/
│   ├── types/
│   │   ├── ast.go               # AST types
│   │   ├── transformation.go    # Transformation types
│   │   └── provider.go          # Provider enum
│   │
│   ├── parser/
│   │   ├── interface.go         # Parser interface
│   │   ├── python.go            # Python parser
│   │   ├── ast.go               # AST utilities
│   │   └── python_test.go
│   │
│   ├── detector/
│   │   ├── detector.go          # Main detector
│   │   ├── patterns.go          # Pattern matching
│   │   └── detector_test.go
│   │
│   ├── plugin/
│   │   ├── loader.go            # Load YAML rules
│   │   ├── rule.go              # Rule structures
│   │   ├── registry.go          # Rule registry
│   │   └── loader_test.go
│   │
│   ├── transformer/
│   │   ├── transformer.go       # Core transformation
│   │   ├── template.go          # Template engine
│   │   ├── context.go           # Transformation context
│   │   └── transformer_test.go
│   │
│   ├── generator/
│   │   ├── generator.go         # Main generator
│   │   ├── aws.go               # AWS specifics
│   │   ├── gcp.go               # GCP specifics
│   │   ├── imports.go           # Import management
│   │   ├── rewriter.go          # AST rewriting
│   │   └── generator_test.go
│   │
│   ├── validator/
│   │   ├── validator.go         # Main validator
│   │   ├── python.go            # Python validation
│   │   └── validator_test.go
│   │
│   └── engine/
│       ├── engine.go            # Public API
│       ├── config.go            # Configuration
│       ├── cache.go             # Caching
│       └── engine_test.go
│
├── internal/
│   └── util/
│       ├── exec.go              # Subprocess execution
│       ├── hash.go              # Hashing utilities
│       └── files.go             # File operations
│
├── tests/
│   ├── integration/
│   │   └── transform_test.go
│   └── fixtures/
│       ├── input/
│       └── expected/
│
├── go.mod
├── go.sum
├── README.md
├── ARCHITECTURE.md              # This file
├── LICENSE
└── .gitignore
```

## Future Enhancements

### Phase 2 Additions

1. **Node.js Support**:
   - `parser/nodejs.go`
   - Parse JavaScript/TypeScript AST
   - Detect `import { upload } from 'infrar/storage'`

2. **More Capabilities**:
   - Database operations
   - Messaging (queues, pub/sub)
   - Compute (beyond just code transformation)

3. **Optimization**:
   - Dead code elimination
   - Provider-specific optimizations
   - Cost-aware transformations

4. **Advanced Features**:
   - Type inference
   - Error handling transformation
   - Async/await transformation

### Phase 3 Additions

1. **Multi-file Projects**:
   - Project-wide analysis
   - Cross-file dependency resolution
   - Incremental transformation

2. **IDE Integration**:
   - Language server protocol
   - Real-time transformation preview
   - Inline error detection

3. **Advanced Validation**:
   - Static analysis
   - Security scanning
   - Performance profiling

## Conclusion

This architecture provides:

✅ **Clear separation of concerns** - Each component has a well-defined responsibility
✅ **Extensibility** - Easy to add new languages and providers
✅ **Testability** - Each component can be tested independently
✅ **Performance** - Caching and optimization strategies
✅ **Reliability** - Comprehensive error handling and validation

The transformation engine is designed to be the solid foundation for Infrar's multi-cloud platform, enabling true infrastructure portability with zero runtime overhead.

---

**Next Steps**: Begin implementation starting with:
1. Go module setup
2. Basic types and interfaces
3. Python parser
4. Plugin loader
5. Simple transformation for storage.upload()
