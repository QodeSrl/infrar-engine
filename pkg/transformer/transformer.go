package transformer

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/QodeSrl/infrar-engine/pkg/plugin"
	"github.com/QodeSrl/infrar-engine/pkg/types"
)

// Transformer applies transformation rules to Infrar calls
type Transformer struct {
	registry *plugin.Registry
}

// New creates a new transformer with a rule registry
func New(registry *plugin.Registry) *Transformer {
	return &Transformer{
		registry: registry,
	}
}

// Transform transforms a single Infrar call to provider-specific code
func (t *Transformer) Transform(call types.InfrarCall) (types.TransformedCall, error) {
	// Get transformation rule for this call
	rule, err := t.registry.GetRuleByCall(call)
	if err != nil {
		return types.TransformedCall{}, &types.TransformationError{
			Category:   types.ErrorCategoryTransformation,
			Message:    fmt.Sprintf("no transformation rule found for %s", call.FullName()),
			Line:       call.LineNumber,
			SourceCode: call.SourceCode,
			Suggestion: fmt.Sprintf("Check if plugin is loaded for %s on %s", call.Module, rule.Provider),
		}
	}

	// Validate required parameters
	if err := t.validateParameters(call, rule); err != nil {
		return types.TransformedCall{}, err
	}

	// Generate code from template
	code, err := t.generateCode(call, rule)
	if err != nil {
		return types.TransformedCall{}, &types.TransformationError{
			Category:   types.ErrorCategoryTransformation,
			Message:    fmt.Sprintf("failed to generate code: %v", err),
			Line:       call.LineNumber,
			SourceCode: call.SourceCode,
		}
	}

	return types.TransformedCall{
		OriginalCall:    call,
		TransformedCode: code,
		LineNumber:      call.LineNumber,
		ColumnOffset:    call.ColumnOffset,
	}, nil
}

// TransformMultiple transforms multiple Infrar calls
func (t *Transformer) TransformMultiple(calls []types.InfrarCall) ([]types.TransformedCall, error) {
	var transformed []types.TransformedCall
	var errors []error

	for _, call := range calls {
		tc, err := t.Transform(call)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		transformed = append(transformed, tc)
	}

	if len(errors) > 0 {
		// Return first error
		return transformed, errors[0]
	}

	return transformed, nil
}

// validateParameters checks if all required parameters are present
func (t *Transformer) validateParameters(call types.InfrarCall, rule types.TransformationRule) error {
	// Check if parameter mapping specifies required parameters
	for infraParam := range rule.ParameterMapping {
		if _, ok := call.Arguments[infraParam]; !ok {
			return &types.TransformationError{
				Category:   types.ErrorCategoryTransformation,
				Message:    fmt.Sprintf("missing required parameter: %s", infraParam),
				Line:       call.LineNumber,
				SourceCode: call.SourceCode,
				Suggestion: fmt.Sprintf("Add %s parameter to %s call", infraParam, call.Function),
			}
		}
	}

	return nil
}

// generateCode generates provider-specific code using template
func (t *Transformer) generateCode(call types.InfrarCall, rule types.TransformationRule) (string, error) {
	// Prepare template data - format all values as strings
	data := make(map[string]string)

	for infraParam, value := range call.Arguments {
		// Convert value to properly formatted string representation
		valueStr := t.formatValue(value)
		data[infraParam] = valueStr
	}

	// Parse and execute template
	tmpl, err := template.New("code").Parse(rule.CodeTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	code := buf.String()

	// Clean up code (remove extra whitespace, etc.)
	code = strings.TrimSpace(code)

	return code, nil
}

// formatValue formats a value for code generation
func (t *Transformer) formatValue(value types.Value) string {
	switch value.Type {
	case types.ValueTypeString:
		// String values should be quoted
		return fmt.Sprintf("'%v'", value.Value)

	case types.ValueTypeNumber:
		// Numbers are used as-is
		return fmt.Sprintf("%v", value.Value)

	case types.ValueTypeBool:
		// Booleans: True/False (Python)
		if b, ok := value.Value.(bool); ok {
			if b {
				return "True"
			}
			return "False"
		}
		return "False"

	case types.ValueTypeVariable:
		// Variables are used as-is (no quotes)
		return fmt.Sprintf("%v", value.Value)

	case types.ValueTypeNone:
		return "None"

	default:
		return fmt.Sprintf("%v", value.Value)
	}
}
