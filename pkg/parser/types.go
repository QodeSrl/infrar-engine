package parser

import "github.com/QodeSrl/infrar-engine/pkg/types"

// PythonCall represents a function call from Python parser
// This is exported so detector can access it
type PythonCall struct {
	LineNumber   int                    `json:"lineno"`
	ColumnOffset int                    `json:"col_offset"`
	Function     string                 `json:"function"`
	Module       string                 `json:"module"`
	Arguments    map[string]types.Value `json:"arguments"`
	SourceCode   string                 `json:"source_code"`
}
