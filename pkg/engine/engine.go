package engine

import (
	"fmt"

	"github.com/QodeSrl/infrar-engine/pkg/detector"
	"github.com/QodeSrl/infrar-engine/pkg/generator"
	"github.com/QodeSrl/infrar-engine/pkg/parser"
	"github.com/QodeSrl/infrar-engine/pkg/plugin"
	"github.com/QodeSrl/infrar-engine/pkg/transformer"
	"github.com/QodeSrl/infrar-engine/pkg/types"
	"github.com/QodeSrl/infrar-engine/pkg/validator"
)

// Engine is the main transformation engine
type Engine struct {
	parser    parser.Parser
	detector  *detector.Detector
	registry  *plugin.Registry
	validator *validator.Validator
}

// New creates a new transformation engine
func New() (*Engine, error) {
	// Create Python parser
	pythonParser, err := parser.NewPythonParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	// Create detector
	det := detector.NewDetector()

	// Create registry
	reg := plugin.NewRegistry()

	// Create validator
	val, err := validator.NewValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	return &Engine{
		parser:    pythonParser,
		detector:  det,
		registry:  reg,
		validator: val,
	}, nil
}

// LoadRules loads transformation rules from a plugin directory
func (e *Engine) LoadRules(pluginDir string, provider types.Provider, capability string) error {
	loader := plugin.NewLoader(pluginDir)

	rules, err := loader.LoadRules(provider, capability)
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	e.registry.RegisterMultiple(rules)

	return nil
}

// Transform transforms source code from Infrar SDK to provider SDK
func (e *Engine) Transform(sourceCode string, targetProvider types.Provider) (*types.TransformationResult, error) {
	// Step 1: Parse source code
	ast, err := e.parser.Parse(sourceCode)
	if err != nil {
		return nil, err
	}

	// Step 2: Detect Infrar calls
	calls, err := e.detector.DetectCalls(ast)
	if err != nil {
		return nil, err
	}

	// Step 3: Transform calls
	trans := transformer.New(e.registry)
	transformedCalls, err := trans.TransformMultiple(calls)
	if err != nil {
		return nil, err
	}

	// Step 4: Generate final code
	gen := generator.New(targetProvider, e.registry)
	result, err := gen.Generate(ast, transformedCalls)
	if err != nil {
		return nil, err
	}

	// Step 5: Validate generated code
	if err := e.validator.Validate(result.TransformedCode); err != nil {
		return nil, err
	}

	return result, nil
}

// TransformFile transforms a file
func (e *Engine) TransformFile(filepath string, targetProvider types.Provider) (*types.TransformationResult, error) {
	content, err := e.parser.ParseFile(filepath)
	if err != nil {
		return nil, err
	}

	return e.Transform(content.SourceCode, targetProvider)
}

// GetRegistry returns the rule registry (for advanced usage)
func (e *Engine) GetRegistry() *plugin.Registry {
	return e.registry
}
