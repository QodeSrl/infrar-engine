package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/QodeSrl/infrar-engine/pkg/types"
)

func TestEngine_Transform_EndToEnd(t *testing.T) {
	// Create engine
	eng, err := New()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Create temporary plugin directory with rules
	tmpDir := t.TempDir()
	awsDir := filepath.Join(tmpDir, "storage", "aws")
	if err := os.MkdirAll(awsDir, 0755); err != nil {
		t.Fatalf("Failed to create plugin directory: %v", err)
	}

	// Create AWS S3 rules
	rulesYAML := `operations:
  - name: upload
    pattern: "infrar.storage.upload"
    target:
      provider: aws
      service: s3
    transformation:
      imports:
        - "import boto3"
      setup_code: "s3 = boto3.client('s3')"
      code_template: "s3.upload_file({{ .source }}, {{ .bucket }}, {{ .destination }})"
      parameter_mapping:
        bucket: bucket
        source: source
        destination: destination
`

	rulesPath := filepath.Join(awsDir, "rules.yaml")
	if err := os.WriteFile(rulesPath, []byte(rulesYAML), 0644); err != nil {
		t.Fatalf("Failed to write rules: %v", err)
	}

	// Load rules
	if err := eng.LoadRules(tmpDir, types.ProviderAWS, "storage"); err != nil {
		t.Fatalf("Failed to load rules: %v", err)
	}

	// Test transformation
	sourceCode := `from infrar.storage import upload

def backup_data():
    upload(bucket='my-bucket', source='file.txt', destination='backup/file.txt')
`

	result, err := eng.Transform(sourceCode, types.ProviderAWS)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	// Verify transformation
	if !strings.Contains(result.TransformedCode, "import boto3") {
		t.Error("Expected 'import boto3' in transformed code")
	}

	if !strings.Contains(result.TransformedCode, "s3 = boto3.client('s3')") {
		t.Error("Expected client setup in transformed code")
	}

	if !strings.Contains(result.TransformedCode, "s3.upload_file") {
		t.Error("Expected s3.upload_file call in transformed code")
	}

	if strings.Contains(result.TransformedCode, "from infrar.storage") {
		t.Error("Infrar import should be removed")
	}

	// Print for manual verification
	t.Logf("Transformed code:\n%s", result.TransformedCode)
}

func TestEngine_Transform_NoInfrarCalls(t *testing.T) {
	eng, err := New()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	sourceCode := `
def hello():
    print('Hello, World!')
`

	result, err := eng.Transform(sourceCode, types.ProviderAWS)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	// Should return original code with a warning
	if !strings.Contains(result.TransformedCode, "Hello, World!") {
		t.Error("Expected original code to be preserved")
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warning about no Infrar calls")
	}
}
