package detector

import (
	"fmt"
	"strings"

	"github.com/QodeSrl/infrar-engine/pkg/parser"
	"github.com/QodeSrl/infrar-engine/pkg/types"
)

// Detector identifies Infrar SDK usage in parsed code
type Detector struct {
	infraPrefix string // "infrar"
}

// NewDetector creates a new Infrar call detector
func NewDetector() *Detector {
	return &Detector{
		infraPrefix: "infrar",
	}
}

// DetectCalls detects Infrar SDK calls in an AST
func (d *Detector) DetectCalls(ast *types.AST) ([]types.InfrarCall, error) {
	if ast == nil {
		return nil, fmt.Errorf("AST is nil")
	}

	// Get the raw calls from metadata (populated by parser)
	rawCalls, ok := ast.Metadata["calls"]
	if !ok {
		return []types.InfrarCall{}, nil
	}

	// Type assertion based on parser type
	var infraCalls []types.InfrarCall

	switch ast.Language {
	case types.LanguagePython:
		pythonCalls, ok := rawCalls.([]parser.PythonCall)
		if !ok {
			return nil, fmt.Errorf("invalid call type in metadata")
		}
		infraCalls = d.filterPythonCalls(pythonCalls, ast.Imports)

	default:
		return nil, fmt.Errorf("unsupported language: %s", ast.Language)
	}

	return infraCalls, nil
}

// filterPythonCalls filters calls to find Infrar SDK usage
func (d *Detector) filterPythonCalls(calls []parser.PythonCall, imports []types.Import) []types.InfrarCall {
	var infraCalls []types.InfrarCall

	// Build a map of imported Infrar symbols
	infraImports := d.buildInfrarImportMap(imports)

	for _, call := range calls {
		infraCall := d.matchInfrarCall(call, infraImports)
		if infraCall != nil {
			infraCalls = append(infraCalls, *infraCall)
		}
	}

	return infraCalls
}

// buildInfrarImportMap builds a map of imported Infrar symbols
// Key: symbol name (e.g., "upload")
// Value: module path (e.g., "infrar.storage")
func (d *Detector) buildInfrarImportMap(imports []types.Import) map[string]string {
	importMap := make(map[string]string)

	for _, imp := range imports {
		// Check if this is an infrar import
		if !strings.HasPrefix(imp.Module, d.infraPrefix) {
			continue
		}

		// Map each imported name to its module
		for _, name := range imp.Names {
			if name == "*" {
				// Handle star imports (import all)
				// We'll need to check module prefix for calls
				continue
			}
			importMap[name] = imp.Module
		}

		// If it's a direct module import (import infrar.storage)
		if len(imp.Names) == 0 || (len(imp.Names) == 1 && imp.Names[0] == imp.Module) {
			// Store the module itself
			parts := strings.Split(imp.Module, ".")
			if len(parts) > 0 {
				lastPart := parts[len(parts)-1]
				importMap[lastPart] = imp.Module
			}
		}
	}

	return importMap
}

// matchInfrarCall checks if a call is an Infrar SDK call and converts it
func (d *Detector) matchInfrarCall(call parser.PythonCall, infraImports map[string]string) *types.InfrarCall {
	var module string

	// Case 1: Direct function call with imported symbol
	// from infrar.storage import upload
	// upload(...)
	if call.Module == "" && call.Function != "" {
		if mod, ok := infraImports[call.Function]; ok {
			module = mod
		} else {
			return nil // Not an Infrar call
		}
	}

	// Case 2: Module.function call
	// import infrar.storage
	// infrar.storage.upload(...)
	if call.Module != "" {
		if strings.HasPrefix(call.Module, d.infraPrefix) {
			module = call.Module
		} else {
			// Check if the first part matches an imported module
			parts := strings.Split(call.Module, ".")
			if len(parts) > 0 {
				if mod, ok := infraImports[parts[0]]; ok {
					// Reconstruct full module path
					module = mod + "." + strings.Join(parts[1:], ".")
				} else {
					return nil
				}
			} else {
				return nil
			}
		}
	}

	// If we didn't find an Infrar module, this isn't an Infrar call
	if module == "" {
		return nil
	}

	// Convert to InfrarCall
	return &types.InfrarCall{
		Module:       module,
		Function:     call.Function,
		Arguments:    call.Arguments,
		LineNumber:   call.LineNumber,
		ColumnOffset: call.ColumnOffset,
		SourceCode:   call.SourceCode,
	}
}

// DetectFromSource is a convenience method that parses and detects in one call
func (d *Detector) DetectFromSource(sourceCode string, language types.Language) ([]types.InfrarCall, error) {
	var p parser.Parser
	var err error

	switch language {
	case types.LanguagePython:
		p, err = parser.NewPythonParser()
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	ast, err := p.Parse(sourceCode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}

	return d.DetectCalls(ast)
}
