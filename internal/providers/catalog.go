package providers

import (
	"slices"

	"gocode/internal/llm/openai"
)

// Catalog returns the built-in providers in display order: model labs,
// then routers and inference hosts, then local servers. Each call returns
// a fresh slice the caller may modify.
func Catalog() []Definition {
	return slices.Concat(labProviders(), hostedProviders(), localProviders())
}

// Lookup finds a built-in provider by ID.
func Lookup(id string) (Definition, bool) {
	for _, def := range Catalog() {
		if def.ID == id {
			return def, true
		}
	}
	return Definition{}, false
}

const reasoningContent = "reasoning_content"

// labProviders are the model makers' own APIs.
func labProviders() []Definition {
	return []Definition{
		{ID: "anthropic", Name: "Anthropic", Style: StyleAnthropic, Category: CategoryCloud,
			BaseURL: "https://api.anthropic.com/v1", KeyURL: "https://console.anthropic.com/settings/keys",
			SuggestedModels: []string{"claude-opus-5-5", "claude-sonnet-5-5", "claude-haiku-5-5", "claude-fable-5-1"}},
		{ID: "openai", Name: "OpenAI", Style: StyleOpenAI, Category: CategoryCloud,
			BaseURL: "https://api.openai.com/v1", KeyURL: "https://platform.openai.com/api-keys",
			MaxTokensField: openai.MaxCompletionTokens, SendReasoningEffort: true, FileInput: true,
			SuggestedModels: []string{"gpt-6-astra", "gpt-6.1-sol", "gpt-6.1-terra", "gpt-6-luna"}},
		{ID: "gemini", Name: "Google Gemini", Style: StyleGemini, Category: CategoryCloud,
			BaseURL: "https://generativelanguage.googleapis.com/v1beta", KeyURL: "https://aistudio.google.com/apikey",
			SuggestedModels: []string{"gemini-4-argon", "gemini-3.8-flash", "gemini-3.5-flash-lite"}},
		{ID: "xai", Name: "xAI", Style: StyleOpenAI, Category: CategoryCloud,
			BaseURL: "https://api.x.ai/v1", KeyURL: "https://console.x.ai",
			SuggestedModels: []string{"grok-4.7", "grok-4.6", "grok-build-0.1"}},
		{ID: "deepseek", Name: "DeepSeek", Style: StyleOpenAI, Category: CategoryCloud,
			BaseURL: "https://api.deepseek.com/v1", KeyURL: "https://platform.deepseek.com/api_keys",
			EchoReasoningField: reasoningContent,
			SuggestedModels:    []string{"deepseek-v4-pro", "deepseek-v4-flash"}},
		{ID: "mistral", Name: "Mistral", Style: StyleOpenAI, Category: CategoryCloud,
			BaseURL: "https://api.mistral.ai/v1", KeyURL: "https://console.mistral.ai/api-keys",
			MaxTokensField: openai.MaxTokens},
		{ID: "moonshot", Name: "Moonshot AI (Kimi)", Style: StyleOpenAI, Category: CategoryCloud,
			BaseURL: "https://api.moonshot.ai/v1", EchoReasoningField: reasoningContent,
			SuggestedModels: []string{"kimi-k3", "kimi-k3.1-pro", "kimi-k2.7-code"}},
		{ID: "zai", Name: "Z.ai (GLM)", Style: StyleOpenAI, Category: CategoryCloud,
			BaseURL: "https://api.z.ai/api/paas/v4", EchoReasoningField: reasoningContent,
			SuggestedModels: []string{"glm-5.3"}},
		{ID: "dashscope", Name: "Alibaba Qwen (DashScope)", Style: StyleOpenAI, Category: CategoryCloud,
			BaseURL:         "https://dashscope-intl.aliyuncs.com/compatible-mode/v1",
			SuggestedModels: []string{"qwen3.7-plus", "qwen3.7-max"}},
		{ID: "minimax", Name: "MiniMax", Style: StyleOpenAI, Category: CategoryCloud,
			BaseURL: "https://api.minimax.io/v1", EchoReasoningField: reasoningContent,
			SuggestedModels: []string{"MiniMax-M3", "MiniMax-M3.1-Flash"}},
		{ID: "cohere", Name: "Cohere", Style: StyleOpenAI, Category: CategoryCloud,
			BaseURL: "https://api.cohere.ai/compatibility/v1", KeyURL: "https://dashboard.cohere.com/api-keys"},
	}
}
