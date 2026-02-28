package category

import "github.com/mcphailtom/healthanalyzer/internal/agent"

// registry holds all registered sub-agents keyed by category name.
var registry = map[string]agent.SubAgent{}

// Register adds a sub-agent to the registry. Call this from init() in
// each category package to make it available to the orchestrator.
func Register(s agent.SubAgent) {
	registry[s.Category()] = s
}

// All returns every registered sub-agent.
func All() map[string]agent.SubAgent {
	out := make(map[string]agent.SubAgent, len(registry))
	for k, v := range registry {
		out[k] = v
	}
	return out
}

// Get returns the sub-agent for a given category, or nil if not registered.
func Get(name string) agent.SubAgent {
	return registry[name]
}

// Categories returns the names of all registered categories.
func Categories() []string {
	names := make([]string, 0, len(registry))
	for k := range registry {
		names = append(names, k)
	}
	return names
}
