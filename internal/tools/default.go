package tools

// Builtin returns every built-in tool.
func Builtin() []Tool {
	return []Tool{
		ReadFile{},
		WriteFile{},
		EditFile{},
		ListDir{},
		Glob{},
		Grep{},
		Bash{},
		WebFetch{},
		TodoWrite{},
	}
}

// Default returns a registry with every built-in tool.
func Default() *Registry {
	return NewRegistry(Builtin()...)
}

// Tools offering previews must keep satisfying Previewer.
var (
	_ Previewer = WriteFile{}
	_ Previewer = EditFile{}
)
