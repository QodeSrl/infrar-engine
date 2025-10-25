package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/QodeSrl/infrar-engine/internal/util"
	"github.com/QodeSrl/infrar-engine/pkg/types"
)

// PythonParser parses Python source code using Python's ast module
type PythonParser struct {
	pythonExecutable string
	parserScriptPath string
	timeout          time.Duration
}

// pythonParseResult represents the JSON output from the Python parser
type pythonParseResult struct {
	Language   string                   `json:"language"`
	Imports    []types.Import           `json:"imports"`
	Calls      []pythonCall             `json:"calls"`
	SourceCode string                   `json:"source_code"`
	Success    bool                     `json:"success"`
	Error      *pythonError             `json:"error,omitempty"`
}

// pythonCall is an alias for the exported PythonCall type
type pythonCall = PythonCall

// pythonError represents an error from Python parser
type pythonError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	LineNumber int `json:"lineno,omitempty"`
	Offset  int    `json:"offset,omitempty"`
	Text    string `json:"text,omitempty"`
}

// NewPythonParser creates a new Python parser
func NewPythonParser() (*PythonParser, error) {
	// Find Python executable
	pythonExec, err := util.FindPythonExecutable()
	if err != nil {
		return nil, fmt.Errorf("failed to find Python executable: %w", err)
	}

	// Get the parser script path
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}

	parserScriptPath := filepath.Join(filepath.Dir(currentFile), "ast_parser.py")

	if !util.FileExists(parserScriptPath) {
		return nil, fmt.Errorf("parser script not found at %s", parserScriptPath)
	}

	return &PythonParser{
		pythonExecutable: pythonExec,
		parserScriptPath: parserScriptPath,
		timeout:          30 * time.Second,
	}, nil
}

// Parse implements the Parser interface
func (p *PythonParser) Parse(sourceCode string) (*types.AST, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	// Execute Python parser script
	stdout, stderr, err := util.ExecuteCommandWithStdin(
		ctx,
		sourceCode,
		p.pythonExecutable,
		p.parserScriptPath,
	)

	if err != nil {
		return nil, &types.TransformationError{
			Category: types.ErrorCategoryParse,
			Message:  fmt.Sprintf("failed to execute Python parser: %v\nstderr: %s", err, stderr),
		}
	}

	// Parse JSON output
	var result pythonParseResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		return nil, &types.TransformationError{
			Category: types.ErrorCategoryParse,
			Message:  fmt.Sprintf("failed to parse JSON output: %v\noutput: %s", err, stdout),
		}
	}

	// Check for parsing errors
	if !result.Success {
		if result.Error != nil {
			return nil, &types.TransformationError{
				Category:   types.ErrorCategoryParse,
				Message:    result.Error.Message,
				Line:       result.Error.LineNumber,
				Column:     result.Error.Offset,
				SourceCode: result.Error.Text,
				Suggestion: "Check Python syntax",
			}
		}
		return nil, &types.TransformationError{
			Category: types.ErrorCategoryParse,
			Message:  "unknown parsing error",
		}
	}

	// Convert to types.AST
	ast := &types.AST{
		Language:   types.LanguagePython,
		Imports:    result.Imports,
		SourceCode: sourceCode,
		Metadata: map[string]any{
			"calls": result.Calls,
		},
	}

	return ast, nil
}

// ParseFile implements the Parser interface
func (p *PythonParser) ParseFile(filepath string) (*types.AST, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, &types.TransformationError{
			Category: types.ErrorCategoryParse,
			Message:  fmt.Sprintf("failed to read file %s: %v", filepath, err),
		}
	}

	ast, err := p.Parse(string(content))
	if err != nil {
		return nil, err
	}

	ast.Filepath = filepath
	return ast, nil
}

// Language implements the Parser interface
func (p *PythonParser) Language() types.Language {
	return types.LanguagePython
}
