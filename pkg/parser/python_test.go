package parser

import (
	"testing"

	"github.com/QodeSrl/infrar-engine/pkg/types"
)

func TestPythonParser_Parse(t *testing.T) {
	parser, err := NewPythonParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tests := []struct {
		name        string
		code        string
		wantErr     bool
		wantImports int
	}{
		{
			name: "Simple import",
			code: `from infrar.storage import upload`,
			wantErr: false,
			wantImports: 1,
		},
		{
			name: "Multiple imports",
			code: `
from infrar.storage import upload, download
import infrar.database
`,
			wantErr: false,
			wantImports: 2,
		},
		{
			name: "Function call",
			code: `
from infrar.storage import upload

upload(bucket='test', source='file.txt', destination='file.txt')
`,
			wantErr: false,
			wantImports: 1,
		},
		{
			name: "Syntax error",
			code: `
from infrar.storage import upload
def invalid syntax here
`,
			wantErr: true,
			wantImports: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := parser.Parse(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if ast == nil {
				t.Error("Expected AST but got nil")
				return
			}

			if ast.Language != types.LanguagePython {
				t.Errorf("Expected language %s, got %s", types.LanguagePython, ast.Language)
			}

			if len(ast.Imports) != tt.wantImports {
				t.Errorf("Expected %d imports, got %d", tt.wantImports, len(ast.Imports))
			}
		})
	}
}

func TestPythonParser_ExtractCalls(t *testing.T) {
	parser, err := NewPythonParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	code := `
from infrar.storage import upload

upload(bucket='my-bucket', source='file.txt', destination='remote.txt')
`

	ast, err := parser.Parse(code)
	if err != nil {
		t.Fatalf("Failed to parse code: %v", err)
	}

	calls, ok := ast.Metadata["calls"].([]pythonCall)
	if !ok {
		t.Fatal("No calls found in metadata")
	}

	if len(calls) == 0 {
		t.Error("Expected to find function calls but found none")
	}

	// Check that we found the upload call
	foundUpload := false
	for _, call := range calls {
		if call.Function == "upload" {
			foundUpload = true

			// Check arguments
			if call.Arguments["bucket"].Value != "my-bucket" {
				t.Errorf("Expected bucket='my-bucket', got %v", call.Arguments["bucket"].Value)
			}

			if call.Arguments["source"].Value != "file.txt" {
				t.Errorf("Expected source='file.txt', got %v", call.Arguments["source"].Value)
			}
		}
	}

	if !foundUpload {
		t.Error("Did not find upload() call")
	}
}
