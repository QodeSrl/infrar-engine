package validator

import (
	"testing"
)

func TestValidator_ValidatePython(t *testing.T) {
	validator, err := NewValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name: "Valid code",
			code: `
import boto3

s3 = boto3.client('s3')
s3.upload_file('file.txt', 'bucket', 'key')
`,
			wantErr: false,
		},
		{
			name: "Valid code with function",
			code: `
def upload_file():
    import boto3
    s3 = boto3.client('s3')
    s3.upload_file('file.txt', 'bucket', 'key')
`,
			wantErr: false,
		},
		{
			name: "Invalid syntax - missing colon",
			code: `
def upload_file()
    print('hello')
`,
			wantErr: true,
		},
		{
			name: "Invalid syntax - bad indentation",
			code: `
def upload_file():
print('hello')
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePython(tt.code)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
