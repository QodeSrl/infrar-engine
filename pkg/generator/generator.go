package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/QodeSrl/infrar-engine/pkg/plugin"
	"github.com/QodeSrl/infrar-engine/pkg/types"
)

// Generator generates final provider-specific code
type Generator struct {
	provider types.Provider
	registry *plugin.Registry
}

// New creates a new code generator
func New(provider types.Provider, registry *plugin.Registry) *Generator {
	return &Generator{
		provider: provider,
		registry: registry,
	}
}

// Generate generates final code from AST and transformed calls
func (g *Generator) Generate(ast *types.AST, transformedCalls []types.TransformedCall) (*types.TransformationResult, error) {
	if len(transformedCalls) == 0 {
		// No transformations needed, return original code
		return &types.TransformationResult{
			Provider:        g.provider,
			TransformedCode: ast.SourceCode,
			Warnings: []types.Warning{
				{
					Message:  "No Infrar SDK calls found - returning original code",
					Category: "info",
				},
			},
		}, nil
	}

	// Collect all imports and requirements
	imports := make(map[string]bool)
	var requirements []types.Requirement
	var setupCodes []string

	for _, tc := range transformedCalls {
		rule, err := g.registry.GetRuleByCall(tc.OriginalCall)
		if err != nil {
			continue
		}

		// Collect imports
		for _, imp := range rule.Imports {
			imports[imp] = true
		}

		// Collect setup code (deduplicated)
		if rule.SetupCode != "" && !contains(setupCodes, rule.SetupCode) {
			setupCodes = append(setupCodes, rule.SetupCode)
		}

		// Collect requirements
		requirements = append(requirements, rule.Requirements...)
	}

	// Generate the transformed code
	code, err := g.replaceCallsInSource(ast.SourceCode, transformedCalls)
	if err != nil {
		return nil, &types.TransformationError{
			Category: types.ErrorCategoryGeneration,
			Message:  fmt.Sprintf("failed to replace calls: %v", err),
		}
	}

	// Remove old infrar imports and add new provider imports
	code = g.replaceImports(code, ast.Imports, imports)

	// Add setup code after imports
	if len(setupCodes) > 0 {
		code = g.addSetupCode(code, setupCodes)
	}

	return &types.TransformationResult{
		Provider:        g.provider,
		TransformedCode: code,
		Imports:         mapKeysToSlice(imports),
		Requirements:    requirements,
		Metadata: map[string]any{
			"transformed_calls": len(transformedCalls),
		},
	}, nil
}

// replaceCallsInSource replaces Infrar calls with transformed code
func (g *Generator) replaceCallsInSource(sourceCode string, transformedCalls []types.TransformedCall) (string, error) {
	lines := strings.Split(sourceCode, "\n")

	// Sort by line number in reverse order (bottom to top)
	// This prevents line number shifts when replacing
	sort.Slice(transformedCalls, func(i, j int) bool {
		return transformedCalls[i].LineNumber > transformedCalls[j].LineNumber
	})

	for _, tc := range transformedCalls {
		lineIdx := tc.LineNumber - 1 // Convert to 0-indexed

		if lineIdx < 0 || lineIdx >= len(lines) {
			continue
		}

		// Get the indentation of the original line
		originalLine := lines[lineIdx]
		indent := getIndentation(originalLine)

		// Apply indentation to transformed code
		transformedLines := strings.Split(tc.TransformedCode, "\n")
		for i, line := range transformedLines {
			if i == 0 {
				transformedLines[i] = indent + line
			} else {
				transformedLines[i] = indent + line
			}
		}

		// Replace the line
		lines[lineIdx] = strings.Join(transformedLines, "\n")
	}

	return strings.Join(lines, "\n"), nil
}

// replaceImports removes Infrar imports and adds provider imports
func (g *Generator) replaceImports(code string, oldImports []types.Import, newImports map[string]bool) string {
	lines := strings.Split(code, "\n")
	var result []string

	// Track which lines to skip (old infrar imports)
	skipLines := make(map[int]bool)

	for _, imp := range oldImports {
		if strings.HasPrefix(imp.Module, "infrar") {
			skipLines[imp.LineNumber-1] = true // Mark for removal
		}
	}

	// Remove old imports
	for i, line := range lines {
		if !skipLines[i] {
			result = append(result, line)
		}
	}

	// Add new imports at the top
	if len(newImports) > 0 {
		// Find where to insert imports (after any docstrings/comments at the top)
		insertIdx := 0
		for i, line := range result {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "\"\"\"") || strings.HasPrefix(trimmed, "'''") {
				insertIdx = i + 1
			} else {
				break
			}
		}

		// Insert imports
		importLines := mapKeysToSlice(newImports)
		sort.Strings(importLines) // Sort for consistency

		var newResult []string
		newResult = append(newResult, result[:insertIdx]...)
		newResult = append(newResult, importLines...)
		newResult = append(newResult, "")
		newResult = append(newResult, result[insertIdx:]...)

		return strings.Join(newResult, "\n")
	}

	return strings.Join(result, "\n")
}

// addSetupCode adds setup code after imports
func (g *Generator) addSetupCode(code string, setupCodes []string) string {
	lines := strings.Split(code, "\n")

	// Find where to insert setup code (after imports)
	insertIdx := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "from ") {
			insertIdx = i + 1
		} else {
			break
		}
	}

	// Insert setup code
	var newResult []string
	newResult = append(newResult, lines[:insertIdx]...)
	newResult = append(newResult, "")
	newResult = append(newResult, setupCodes...)
	newResult = append(newResult, "")
	newResult = append(newResult, lines[insertIdx:]...)

	return strings.Join(newResult, "\n")
}

// Helper functions

func getIndentation(line string) string {
	for i, char := range line {
		if char != ' ' && char != '\t' {
			return line[:i]
		}
	}
	return ""
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func mapKeysToSlice(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
