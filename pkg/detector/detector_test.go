package detector

import (
	"testing"

	"github.com/QodeSrl/infrar-engine/pkg/types"
)

func TestDetector_DetectFromSource(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name      string
		code      string
		wantCalls int
		wantFuncs []string
	}{
		{
			name: "Simple upload call",
			code: `
from infrar.storage import upload

upload(bucket='data', source='file.txt', destination='file.txt')
`,
			wantCalls: 1,
			wantFuncs: []string{"upload"},
		},
		{
			name: "Multiple operations",
			code: `
from infrar.storage import upload, download, delete

upload(bucket='data', source='file.txt', destination='file.txt')
download(bucket='data', source='file.txt', destination='local.txt')
delete(bucket='data', path='old.txt')
`,
			wantCalls: 3,
			wantFuncs: []string{"upload", "download", "delete"},
		},
		{
			name: "Module-qualified call",
			code: `
import infrar.storage

infrar.storage.upload(bucket='data', source='file.txt', destination='file.txt')
`,
			wantCalls: 1,
			wantFuncs: []string{"upload"},
		},
		{
			name: "Mixed with non-Infrar calls",
			code: `
from infrar.storage import upload
import os

upload(bucket='data', source='file.txt', destination='file.txt')
os.path.exists('file.txt')
print('hello')
`,
			wantCalls: 1,
			wantFuncs: []string{"upload"},
		},
		{
			name: "No Infrar calls",
			code: `
import os

os.path.exists('file.txt')
print('hello')
`,
			wantCalls: 0,
			wantFuncs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls, err := detector.DetectFromSource(tt.code, types.LanguagePython)
			if err != nil {
				t.Fatalf("DetectFromSource() error = %v", err)
			}

			if len(calls) != tt.wantCalls {
				t.Errorf("DetectFromSource() got %d calls, want %d", len(calls), tt.wantCalls)
			}

			// Check function names
			for i, want := range tt.wantFuncs {
				if i >= len(calls) {
					t.Errorf("Missing call for function %s", want)
					continue
				}
				if calls[i].Function != want {
					t.Errorf("Call %d: got function %s, want %s", i, calls[i].Function, want)
				}
			}
		})
	}
}

func TestDetector_DetectArguments(t *testing.T) {
	detector := NewDetector()

	code := `
from infrar.storage import upload

upload(bucket='my-bucket', source='local.txt', destination='remote.txt')
`

	calls, err := detector.DetectFromSource(code, types.LanguagePython)
	if err != nil {
		t.Fatalf("DetectFromSource() error = %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}

	call := calls[0]

	// Check module and function
	if call.Module != "infrar.storage" {
		t.Errorf("Expected module 'infrar.storage', got '%s'", call.Module)
	}

	if call.Function != "upload" {
		t.Errorf("Expected function 'upload', got '%s'", call.Function)
	}

	// Check arguments
	expectedArgs := map[string]string{
		"bucket":      "my-bucket",
		"source":      "local.txt",
		"destination": "remote.txt",
	}

	for key, expectedValue := range expectedArgs {
		arg, ok := call.Arguments[key]
		if !ok {
			t.Errorf("Missing argument: %s", key)
			continue
		}

		if arg.Value != expectedValue {
			t.Errorf("Argument %s: got '%v', want '%v'", key, arg.Value, expectedValue)
		}
	}
}

func TestDetector_ModuleQualifiedCalls(t *testing.T) {
	detector := NewDetector()

	code := `
import infrar.storage

infrar.storage.upload(bucket='data', source='file.txt', destination='file.txt')
`

	calls, err := detector.DetectFromSource(code, types.LanguagePython)
	if err != nil {
		t.Fatalf("DetectFromSource() error = %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}

	call := calls[0]

	if call.Module != "infrar.storage" {
		t.Errorf("Expected module 'infrar.storage', got '%s'", call.Module)
	}

	if call.Function != "upload" {
		t.Errorf("Expected function 'upload', got '%s'", call.Function)
	}
}
