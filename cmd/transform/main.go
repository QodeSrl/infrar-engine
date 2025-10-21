package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/QodeSrl/infrar-engine/pkg/engine"
	"github.com/QodeSrl/infrar-engine/pkg/types"
)

func main() {
	// Define flags
	provider := flag.String("provider", "aws", "Target cloud provider (aws, gcp, azure)")
	pluginDir := flag.String("plugins", "../infrar-plugins/packages", "Path to plugins directory")
	capability := flag.String("capability", "storage", "Capability to transform (storage, database, etc.)")
	inputFile := flag.String("input", "", "Input file to transform (or use stdin)")
	outputFile := flag.String("output", "", "Output file (or use stdout)")

	flag.Parse()

	// Parse provider
	var targetProvider types.Provider
	switch *provider {
	case "aws":
		targetProvider = types.ProviderAWS
	case "gcp":
		targetProvider = types.ProviderGCP
	case "azure":
		targetProvider = types.ProviderAzure
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown provider '%s'\n", *provider)
		os.Exit(1)
	}

	// Create engine
	eng, err := engine.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating engine: %v\n", err)
		os.Exit(1)
	}

	// Load rules
	if err := eng.LoadRules(*pluginDir, targetProvider, *capability); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading rules: %v\n", err)
		os.Exit(1)
	}

	// Read input
	var sourceCode string
	if *inputFile != "" {
		content, err := os.ReadFile(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
			os.Exit(1)
		}
		sourceCode = string(content)
	} else {
		// Read from stdin
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}
		sourceCode = string(content)
	}

	// Transform
	result, err := eng.Transform(sourceCode, targetProvider)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Transformation error: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, []byte(result.TransformedCode), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "✓ Transformed code written to %s\n", *outputFile)

		// Show metadata only when writing to file
		showMetadata(result)
	} else {
		// Write ONLY code to stdout (for piping)
		fmt.Print(result.TransformedCode)
		// Metadata goes to stderr
		showMetadata(result)
	}
}

// showMetadata displays warnings and metadata to stderr
func showMetadata(result *types.TransformationResult) {
	// Show warnings
	if len(result.Warnings) > 0 {
		fmt.Fprintln(os.Stderr, "\nWarnings:")
		for _, w := range result.Warnings {
			fmt.Fprintf(os.Stderr, "  - %s\n", w.Message)
		}
	}

	// Show metadata
	if len(result.Imports) > 0 {
		fmt.Fprintln(os.Stderr, "\nImports added:")
		for _, imp := range result.Imports {
			fmt.Fprintf(os.Stderr, "  - %s\n", imp)
		}
	}

	if len(result.Requirements) > 0 {
		fmt.Fprintln(os.Stderr, "\nDependencies required:")
		for _, req := range result.Requirements {
			fmt.Fprintf(os.Stderr, "  - %s %s\n", req.Package, req.Version)
		}
	}
}
