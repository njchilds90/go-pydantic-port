package pydantic

// ToolDefinitionModel returns a baseline model for LLM tool call payloads.
func ToolDefinitionModel(name string) *Model {
	return NewModel(name).
		Field("tool", "string").Required().Description("Tool name").End().
		Field("arguments", "object").Required().Description("Tool arguments payload").End()
}

// AgentMemoryModel returns a baseline model for agent memory entries.
func AgentMemoryModel(name string) *Model {
	return NewModel(name).
		Field("key", "string").Required().Min(1).Description("Memory key").End().
		Field("value", "string").Required().Description("Memory value").End().
		Field("namespace", "string").Default("default").Description("Memory namespace").End()
}
