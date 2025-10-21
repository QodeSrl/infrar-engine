package types

// AST represents parsed source code
type AST struct {
	Language   Language          `json:"language"`
	Nodes      []Node            `json:"nodes"`
	Imports    []Import          `json:"imports"`
	SourceCode string            `json:"source_code"`
	Filepath   string            `json:"filepath"`
	Metadata   map[string]any    `json:"metadata,omitempty"`
}

// Node represents a node in the AST
type Node struct {
	Type         string         `json:"type"`          // "ImportFrom", "Call", "FunctionDef", etc.
	LineNumber   int            `json:"lineno"`
	ColumnOffset int            `json:"col_offset"`
	Attributes   map[string]any `json:"attributes,omitempty"`
	Children     []Node         `json:"children,omitempty"`
}

// Import represents an import statement
type Import struct {
	Module string   `json:"module"` // "infrar.storage"
	Names  []string `json:"names"`  // ["upload", "download"]
	Alias  string   `json:"alias,omitempty"` // Optional alias
	LineNumber int  `json:"lineno"`
}

// Value represents a value in function arguments
type Value struct {
	Type  ValueType `json:"type"`
	Value any       `json:"value"`
}

// String returns the string representation of a value
func (v Value) String() string {
	if v.Value == nil {
		return ""
	}

	switch v.Type {
	case ValueTypeString:
		return v.Value.(string)
	case ValueTypeNumber:
		return v.Value.(string)
	case ValueTypeBool:
		if v.Value.(bool) {
			return "True"
		}
		return "False"
	case ValueTypeVariable:
		return v.Value.(string)
	case ValueTypeNone:
		return "None"
	default:
		return ""
	}
}
