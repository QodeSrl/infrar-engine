package parser

import "github.com/QodeSrl/infrar-engine/pkg/types"

// Parser is the interface for language parsers
type Parser interface {
	// Parse parses source code and returns an AST
	Parse(sourceCode string) (*types.AST, error)

	// ParseFile parses a file and returns an AST
	ParseFile(filepath string) (*types.AST, error)

	// Language returns the language this parser supports
	Language() types.Language
}
