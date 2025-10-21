package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/QodeSrl/infrar-engine/pkg/types"
)

func TestLoader_LoadRules(t *testing.T) {
	// Create temporary plugin directory structure
	tmpDir := t.TempDir()

	// Create storage/aws directory
	awsDir := filepath.Join(tmpDir, "storage", "aws")
	if err := os.MkdirAll(awsDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a sample rules.yaml file
	rulesYAML := `operations:
  - name: upload
    pattern: "infrar.storage.upload"
    target:
      provider: aws
      service: s3
      operation: upload_file
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
    requirements:
      - package: boto3
        version: ">=1.28.0"
`

	rulesPath := filepath.Join(awsDir, "rules.yaml")
	if err := os.WriteFile(rulesPath, []byte(rulesYAML), 0644); err != nil {
		t.Fatalf("Failed to write rules file: %v", err)
	}

	// Test loading
	loader := NewLoader(tmpDir)
	rules, err := loader.LoadRules(types.ProviderAWS, "storage")
	if err != nil {
		t.Fatalf("LoadRules() error = %v", err)
	}

	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	rule := rules[0]

	if rule.Name != "upload" {
		t.Errorf("Expected name 'upload', got '%s'", rule.Name)
	}

	if rule.Pattern != "infrar.storage.upload" {
		t.Errorf("Expected pattern 'infrar.storage.upload', got '%s'", rule.Pattern)
	}

	if rule.Provider != types.ProviderAWS {
		t.Errorf("Expected provider AWS, got %s", rule.Provider)
	}

	if rule.Service != "s3" {
		t.Errorf("Expected service 's3', got '%s'", rule.Service)
	}

	if len(rule.Imports) != 1 || rule.Imports[0] != "import boto3" {
		t.Errorf("Expected import 'import boto3', got %v", rule.Imports)
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	registry := NewRegistry()

	rule := types.TransformationRule{
		Name:     "upload",
		Pattern:  "infrar.storage.upload",
		Provider: types.ProviderAWS,
		Service:  "s3",
	}

	registry.Register(rule)

	// Test retrieval
	retrieved, err := registry.GetRule("infrar.storage.upload")
	if err != nil {
		t.Fatalf("GetRule() error = %v", err)
	}

	if retrieved.Name != rule.Name {
		t.Errorf("Expected name '%s', got '%s'", rule.Name, retrieved.Name)
	}

	// Test non-existent rule
	_, err = registry.GetRule("non.existent.pattern")
	if err == nil {
		t.Error("Expected error for non-existent pattern, got nil")
	}
}

func TestRegistry_HasRule(t *testing.T) {
	registry := NewRegistry()

	rule := types.TransformationRule{
		Pattern: "infrar.storage.upload",
	}

	registry.Register(rule)

	if !registry.HasRule("infrar.storage.upload") {
		t.Error("Expected HasRule to return true")
	}

	if registry.HasRule("non.existent.pattern") {
		t.Error("Expected HasRule to return false for non-existent pattern")
	}
}
