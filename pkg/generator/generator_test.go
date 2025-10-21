package generator

import (
	"strings"
	"testing"

	"github.com/QodeSrl/infrar-engine/pkg/plugin"
	"github.com/QodeSrl/infrar-engine/pkg/types"
)

func TestGenerator_Generate(t *testing.T) {
	// Create registry with a rule
	registry := plugin.NewRegistry()

	rule := types.TransformationRule{
		Pattern:  "infrar.storage.upload",
		Provider: types.ProviderAWS,
		Imports:  []string{"import boto3"},
		SetupCode: "s3 = boto3.client('s3')",
	}

	registry.Register(rule)

	// Create AST
	ast := &types.AST{
		Language: types.LanguagePython,
		SourceCode: `from infrar.storage import upload

upload(bucket='data', source='file.txt', destination='file.txt')
`,
		Imports: []types.Import{
			{Module: "infrar.storage", Names: []string{"upload"}, LineNumber: 1},
		},
	}

	// Create transformed call
	transformedCalls := []types.TransformedCall{
		{
			OriginalCall: types.InfrarCall{
				Module:   "infrar.storage",
				Function: "upload",
			},
			TransformedCode: "s3.upload_file('file.txt', 'data', 'file.txt')",
			LineNumber:      3,
		},
	}

	// Generate
	generator := New(types.ProviderAWS, registry)
	result, err := generator.Generate(ast, transformedCalls)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check that boto3 is imported
	if !strings.Contains(result.TransformedCode, "import boto3") {
		t.Error("Expected 'import boto3' in transformed code")
	}

	// Check that infrar import is removed
	if strings.Contains(result.TransformedCode, "from infrar.storage") {
		t.Error("Old infrar import should be removed")
	}

	// Check that transformed call is present
	if !strings.Contains(result.TransformedCode, "s3.upload_file") {
		t.Error("Expected transformed call in code")
	}

	// Check that setup code is present
	if !strings.Contains(result.TransformedCode, "s3 = boto3.client('s3')") {
		t.Error("Expected setup code in transformed output")
	}
}

func TestGenerator_NoTransformations(t *testing.T) {
	registry := plugin.NewRegistry()
	generator := New(types.ProviderAWS, registry)

	ast := &types.AST{
		SourceCode: "print('hello world')",
	}

	result, err := generator.Generate(ast, []types.TransformedCall{})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if result.TransformedCode != ast.SourceCode {
		t.Error("Expected original code when no transformations")
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warning about no transformations")
	}
}

func TestGetIndentation(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"    code", "    "},
		{"\t\tcode", "\t\t"},
		{"code", ""},
		{"  ", ""}, // Empty line returns no indentation
	}

	for _, tt := range tests {
		got := getIndentation(tt.line)
		if got != tt.want {
			t.Errorf("getIndentation(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}
