package validator

import (
	"context"
	"fmt"
	"time"

	"github.com/QodeSrl/infrar-engine/internal/util"
	"github.com/QodeSrl/infrar-engine/pkg/types"
)

// Validator validates generated code
type Validator struct {
	pythonExecutable string
	timeout          time.Duration
}

// NewValidator creates a new code validator
func NewValidator() (*Validator, error) {
	// Find Python executable
	pythonExec, err := util.FindPythonExecutable()
	if err != nil {
		return nil, fmt.Errorf("failed to find Python executable: %w", err)
	}

	return &Validator{
		pythonExecutable: pythonExec,
		timeout:          5 * time.Second,
	}, nil
}

// Validate validates Python code syntax
func (v *Validator) Validate(code string) error {
	return v.ValidatePython(code)
}

// ValidatePython validates Python code using Python's compile function
func (v *Validator) ValidatePython(code string) error {
	ctx, cancel := context.WithTimeout(context.Background(), v.timeout)
	defer cancel()

	// Use Python's compile function to check syntax
	pythonCode := `
import sys
try:
    compile(sys.stdin.read(), '<string>', 'exec')
    sys.exit(0)
except SyntaxError as e:
    print(f"SyntaxError: {e}", file=sys.stderr)
    sys.exit(1)
`

	stdout, stderr, err := util.ExecuteCommandWithStdin(
		ctx,
		code,
		v.pythonExecutable,
		"-c",
		pythonCode,
	)

	if err != nil {
		return &types.TransformationError{
			Category:   types.ErrorCategoryValidation,
			Message:    fmt.Sprintf("invalid Python syntax: %s", stderr),
			Suggestion: "Check the generated code for syntax errors",
		}
	}

	if stderr != "" {
		return &types.TransformationError{
			Category:   types.ErrorCategoryValidation,
			Message:    stderr,
			Suggestion: "Check the generated code for syntax errors",
		}
	}

	_ = stdout // Not used, but keep for potential future use

	return nil
}
