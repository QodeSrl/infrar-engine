package types

// InfrarCall represents a detected Infrar SDK usage
type InfrarCall struct {
	Module       string           `json:"module"`        // "infrar.storage"
	Function     string           `json:"function"`      // "upload"
	Arguments    map[string]Value `json:"arguments"`     // {bucket: "data", source: "file.txt", ...}
	LineNumber   int              `json:"lineno"`
	ColumnOffset int              `json:"col_offset"`
	SourceCode   string           `json:"source_code"`   // Original code snippet
}

// FullName returns the full qualified name of the call
func (c InfrarCall) FullName() string {
	return c.Module + "." + c.Function
}

// TransformationRule defines how to transform an Infrar call
type TransformationRule struct {
	Name             string            `yaml:"name"`
	Pattern          string            `yaml:"pattern"`          // "infrar.storage.upload"
	Provider         Provider          `yaml:"provider"`
	Service          string            `yaml:"service"`          // "s3", "cloud_storage"
	Imports          []string          `yaml:"imports"`
	SetupCode        string            `yaml:"setup_code"`       // Client initialization
	CodeTemplate     string            `yaml:"code_template"`    // Go template
	ParameterMapping map[string]string `yaml:"parameter_mapping"`
	Requirements     []Requirement     `yaml:"requirements"`
}

// Requirement represents a package dependency requirement
type Requirement struct {
	Package string `yaml:"package"` // "boto3"
	Version string `yaml:"version"` // ">=1.28.0"
}

// TransformedCall represents a transformed function call
type TransformedCall struct {
	OriginalCall     InfrarCall
	TransformedCode  string
	LineNumber       int
	ColumnOffset     int
}

// TransformationResult is the output of transformation
type TransformationResult struct {
	Provider        Provider      `json:"provider"`
	TransformedCode string        `json:"transformed_code"`
	Imports         []string      `json:"imports"`
	Requirements    []Requirement `json:"requirements"`
	Warnings        []Warning     `json:"warnings,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// Warning represents a transformation warning
type Warning struct {
	Message    string `json:"message"`
	LineNumber int    `json:"lineno,omitempty"`
	Category   string `json:"category,omitempty"`
}

// ErrorCategory represents the category of an error
type ErrorCategory string

const (
	ErrorCategoryParse          ErrorCategory = "parse"
	ErrorCategoryDetection      ErrorCategory = "detection"
	ErrorCategoryTransformation ErrorCategory = "transformation"
	ErrorCategoryGeneration     ErrorCategory = "generation"
	ErrorCategoryValidation     ErrorCategory = "validation"
)

// TransformationError represents an error during transformation
type TransformationError struct {
	Category   ErrorCategory `json:"category"`
	Message    string        `json:"message"`
	Line       int           `json:"line,omitempty"`
	Column     int           `json:"column,omitempty"`
	SourceCode string        `json:"source_code,omitempty"`
	Suggestion string        `json:"suggestion,omitempty"`
}

// Error implements the error interface
func (e *TransformationError) Error() string {
	if e.Line > 0 {
		return e.Category.String() + " error at line " + string(rune(e.Line)) + ": " + e.Message
	}
	return e.Category.String() + " error: " + e.Message
}

// String returns the string representation of an error category
func (ec ErrorCategory) String() string {
	return string(ec)
}
