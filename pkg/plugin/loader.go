package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/QodeSrl/infrar-engine/pkg/types"
	"gopkg.in/yaml.v3"
)

// Loader loads transformation rules from plugin YAML files
type Loader struct {
	pluginDir string
}

// NewLoader creates a new plugin loader
func NewLoader(pluginDir string) *Loader {
	return &Loader{
		pluginDir: pluginDir,
	}
}

// LoadRules loads transformation rules for a specific provider
func (l *Loader) LoadRules(provider types.Provider, capability string) ([]types.TransformationRule, error) {
	// Construct path to rules file
	// Expected structure: pluginDir/capability/provider/rules.yaml
	// Example: ../infrar-plugins/packages/storage/aws/rules.yaml
	rulesPath := filepath.Join(l.pluginDir, capability, provider.String(), "rules.yaml")

	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("rules file not found: %s", rulesPath)
	}

	// Read YAML file
	data, err := os.ReadFile(rulesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file: %w", err)
	}

	// Parse YAML
	var pluginRules types.PluginRules
	if err := yaml.Unmarshal(data, &pluginRules); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Convert to TransformationRule
	var rules []types.TransformationRule
	for _, op := range pluginRules.Operations {
		rule := types.TransformationRule{
			Name:             op.Name,
			Pattern:          op.Pattern,
			Provider:         provider,
			Service:          op.Target.Service,
			Imports:          op.Transformation.Imports,
			SetupCode:        op.Transformation.SetupCode,
			CodeTemplate:     op.Transformation.CodeTemplate,
			ParameterMapping: op.Transformation.ParameterMapping,
			Requirements:     op.Requirements,
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// LoadAllRules loads all transformation rules for a provider (all capabilities)
func (l *Loader) LoadAllRules(provider types.Provider) (map[string][]types.TransformationRule, error) {
	allRules := make(map[string][]types.TransformationRule)

	// Walk through plugin directory
	entries, err := os.ReadDir(l.pluginDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		capability := entry.Name()

		// Try to load rules for this capability
		rules, err := l.LoadRules(provider, capability)
		if err != nil {
			// Skip if rules don't exist for this capability
			continue
		}

		allRules[capability] = rules
	}

	return allRules, nil
}
