package transformer

import (
	"testing"

	"github.com/QodeSrl/infrar-engine/pkg/plugin"
	"github.com/QodeSrl/infrar-engine/pkg/types"
)

func TestTransformer_Transform(t *testing.T) {
	// Create registry and add a rule
	registry := plugin.NewRegistry()

	rule := types.TransformationRule{
		Name:     "upload",
		Pattern:  "infrar.storage.upload",
		Provider: types.ProviderAWS,
		Service:  "s3",
		Imports:  []string{"import boto3"},
		SetupCode: "s3 = boto3.client('s3')",
		CodeTemplate: `s3.upload_file(
    Filename={{ .source }},
    Bucket={{ .bucket }},
    Key={{ .destination }}
)`,
		ParameterMapping: map[string]string{
			"bucket":      "Bucket",
			"source":      "Filename",
			"destination": "Key",
		},
	}

	registry.Register(rule)

	// Create transformer
	transformer := New(registry)

	// Create an Infrar call
	call := types.InfrarCall{
		Module:   "infrar.storage",
		Function: "upload",
		Arguments: map[string]types.Value{
			"bucket":      {Type: types.ValueTypeString, Value: "my-bucket"},
			"source":      {Type: types.ValueTypeString, Value: "file.txt"},
			"destination": {Type: types.ValueTypeString, Value: "remote.txt"},
		},
		LineNumber: 5,
	}

	// Transform
	transformed, err := transformer.Transform(call)
	if err != nil {
		t.Fatalf("Transform() error = %v", err)
	}

	// Check result
	expectedCode := `s3.upload_file(
    Filename='file.txt',
    Bucket='my-bucket',
    Key='remote.txt'
)`

	if transformed.TransformedCode != expectedCode {
		t.Errorf("Transform() got:\n%s\n\nwant:\n%s", transformed.TransformedCode, expectedCode)
	}
}

func TestTransformer_MissingParameter(t *testing.T) {
	// Create registry and add a rule
	registry := plugin.NewRegistry()

	rule := types.TransformationRule{
		Pattern:  "infrar.storage.upload",
		Provider: types.ProviderAWS,
		ParameterMapping: map[string]string{
			"bucket":      "Bucket",
			"source":      "Filename",
			"destination": "Key",
		},
		CodeTemplate: "s3.upload_file(...)",
	}

	registry.Register(rule)

	transformer := New(registry)

	// Create call with missing parameter
	call := types.InfrarCall{
		Module:   "infrar.storage",
		Function: "upload",
		Arguments: map[string]types.Value{
			"bucket": {Type: types.ValueTypeString, Value: "my-bucket"},
			// Missing: source and destination
		},
	}

	// Should fail
	_, err := transformer.Transform(call)
	if err == nil {
		t.Error("Expected error for missing parameters, got nil")
	}
}

func TestTransformer_FormatValue(t *testing.T) {
	transformer := New(plugin.NewRegistry())

	tests := []struct {
		name  string
		value types.Value
		want  string
	}{
		{
			name:  "String value",
			value: types.Value{Type: types.ValueTypeString, Value: "hello"},
			want:  "'hello'",
		},
		{
			name:  "Number value",
			value: types.Value{Type: types.ValueTypeNumber, Value: "42"},
			want:  "42",
		},
		{
			name:  "Bool true",
			value: types.Value{Type: types.ValueTypeBool, Value: true},
			want:  "True",
		},
		{
			name:  "Bool false",
			value: types.Value{Type: types.ValueTypeBool, Value: false},
			want:  "False",
		},
		{
			name:  "Variable",
			value: types.Value{Type: types.ValueTypeVariable, Value: "my_var"},
			want:  "my_var",
		},
		{
			name:  "None",
			value: types.Value{Type: types.ValueTypeNone, Value: nil},
			want:  "None",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transformer.formatValue(tt.value)
			if got != tt.want {
				t.Errorf("formatValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
