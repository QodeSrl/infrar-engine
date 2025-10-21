package plugin

import (
	"fmt"
	"sync"

	"github.com/QodeSrl/infrar-engine/pkg/types"
)

// Registry manages transformation rules
type Registry struct {
	mu    sync.RWMutex
	rules map[string]types.TransformationRule // pattern -> rule
}

// NewRegistry creates a new rule registry
func NewRegistry() *Registry {
	return &Registry{
		rules: make(map[string]types.TransformationRule),
	}
}

// Register registers a transformation rule
func (r *Registry) Register(rule types.TransformationRule) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rules[rule.Pattern] = rule
}

// RegisterMultiple registers multiple transformation rules
func (r *Registry) RegisterMultiple(rules []types.TransformationRule) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, rule := range rules {
		r.rules[rule.Pattern] = rule
	}
}

// GetRule retrieves a transformation rule by pattern
func (r *Registry) GetRule(pattern string) (types.TransformationRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rule, ok := r.rules[pattern]
	if !ok {
		return types.TransformationRule{}, fmt.Errorf("no rule found for pattern: %s", pattern)
	}

	return rule, nil
}

// GetRuleByCall retrieves a transformation rule for an Infrar call
func (r *Registry) GetRuleByCall(call types.InfrarCall) (types.TransformationRule, error) {
	pattern := call.FullName() // e.g., "infrar.storage.upload"
	return r.GetRule(pattern)
}

// HasRule checks if a rule exists for a pattern
func (r *Registry) HasRule(pattern string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.rules[pattern]
	return ok
}

// AllRules returns all registered rules
func (r *Registry) AllRules() []types.TransformationRule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rules := make([]types.TransformationRule, 0, len(r.rules))
	for _, rule := range r.rules {
		rules = append(rules, rule)
	}

	return rules
}

// Clear clears all rules from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rules = make(map[string]types.TransformationRule)
}
