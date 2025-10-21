package types

// PluginManifest represents metadata about a plugin
type PluginManifest struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Provides    []string `yaml:"provides"` // Capabilities provided
}

// OperationRule represents a transformation rule for a single operation
type OperationRule struct {
	Name             string                 `yaml:"name"`
	Pattern          string                 `yaml:"pattern"`
	Target           TargetConfig           `yaml:"target"`
	Transformation   TransformationConfig   `yaml:"transformation"`
	Requirements     []Requirement          `yaml:"requirements,omitempty"`
}

// TargetConfig describes the target provider configuration
type TargetConfig struct {
	Provider string `yaml:"provider"` // "aws", "gcp", "azure"
	Service  string `yaml:"service"`  // "s3", "cloud_storage"
	Operation string `yaml:"operation,omitempty"` // Optional: specific operation name
}

// TransformationConfig describes how to perform the transformation
type TransformationConfig struct {
	Imports          []string          `yaml:"imports"`
	SetupCode        string            `yaml:"setup_code,omitempty"`
	CodeTemplate     string            `yaml:"code_template"`
	ParameterMapping map[string]string `yaml:"parameter_mapping"`
}

// PluginRules represents all transformation rules from a plugin
type PluginRules struct {
	Operations []OperationRule `yaml:"operations"`
}
