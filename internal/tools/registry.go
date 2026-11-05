package tools

import (
	"sort"

	"gocode/internal/llm"
)

// Registry holds the available tools by name.
type Registry struct {
	byName map[string]Tool
}

// NewRegistry returns a registry holding tools.
func NewRegistry(tools ...Tool) *Registry {
	r := &Registry{byName: make(map[string]Tool, len(tools))}
	for _, tool := range tools {
		r.byName[tool.Spec().Name] = tool
	}
	return r
}

// Lookup returns the tool with the given name.
func (r *Registry) Lookup(name string) (Tool, bool) {
	tool, ok := r.byName[name]
	return tool, ok
}

// Specs returns the specs of the tools, sorted by name, leaving out those
// for which include returns false.
func (r *Registry) Specs(include func(Tool) bool) []llm.ToolSpec {
	specs := make([]llm.ToolSpec, 0, len(r.byName))
	for _, tool := range r.byName {
		if include == nil || include(tool) {
			specs = append(specs, tool.Spec())
		}
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].Name < specs[j].Name })
	return specs
}
