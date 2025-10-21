package types

// Provider represents a cloud provider
type Provider string

const (
	ProviderAWS   Provider = "aws"
	ProviderGCP   Provider = "gcp"
	ProviderAzure Provider = "azure"
)

// String returns the string representation of the provider
func (p Provider) String() string {
	return string(p)
}

// IsValid checks if the provider is supported
func (p Provider) IsValid() bool {
	switch p {
	case ProviderAWS, ProviderGCP, ProviderAzure:
		return true
	default:
		return false
	}
}

// Language represents a programming language
type Language string

const (
	LanguagePython Language = "python"
	LanguageNodeJS Language = "nodejs"
	LanguageGo     Language = "go"
)

// String returns the string representation of the language
func (l Language) String() string {
	return string(l)
}

// ValueType represents the type of a value in the AST
type ValueType string

const (
	ValueTypeString   ValueType = "string"
	ValueTypeNumber   ValueType = "number"
	ValueTypeBool     ValueType = "bool"
	ValueTypeVariable ValueType = "variable"
	ValueTypeNone     ValueType = "none"
)
