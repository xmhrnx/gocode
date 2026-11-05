package providers

import "gocode/internal/llm/openai"

// hostedProviders are routers and inference hosts serving many models
// through an OpenAI-compatible API.
func hostedProviders() []Definition {
	return openAICompatible(CategoryRouter, []Definition{
		{ID: "openrouter", Name: "OpenRouter", BaseURL: "https://openrouter.ai/api/v1", KeyURL: "https://openrouter.ai/keys",
			// OpenRouter attributes traffic to apps by these headers.
			Headers: map[string]string{"HTTP-Referer": "https://github.com/gocode", "X-OpenRouter-Title": "GoCode"}},
		{ID: "groq", Name: "Groq", BaseURL: "https://api.groq.com/openai/v1", KeyURL: "https://console.groq.com/keys"},
		{ID: "together", Name: "Together AI", BaseURL: "https://api.together.ai/v1"},
		{ID: "fireworks", Name: "Fireworks AI", BaseURL: "https://api.fireworks.ai/inference/v1"},
		{ID: "cerebras", Name: "Cerebras", BaseURL: "https://api.cerebras.ai/v1"},
		{ID: "deepinfra", Name: "DeepInfra", BaseURL: "https://api.deepinfra.com/v1/openai"},
		{ID: "huggingface", Name: "Hugging Face", BaseURL: "https://router.huggingface.co/v1", KeyURL: "https://huggingface.co/settings/tokens"},
		{ID: "nebius", Name: "Nebius Token Factory", BaseURL: "https://api.tokenfactory.nebius.com/v1"},
		{ID: "sambanova", Name: "SambaNova", BaseURL: "https://api.sambanova.ai/v1"},
		{ID: "novita", Name: "Novita AI", BaseURL: "https://api.novita.ai/openai/v1"},
		{ID: "nvidia", Name: "NVIDIA NIM", BaseURL: "https://integrate.api.nvidia.com/v1"},
		{ID: "perplexity", Name: "Perplexity", BaseURL: "https://api.perplexity.ai"},
		{ID: "azure-openai", Name: "Azure OpenAI", BaseURL: "https://YOUR-RESOURCE.openai.azure.com/openai/v1",
			AuthHeader: "api-key", MaxTokensField: openai.MaxCompletionTokens,
			Notes: "Replace YOUR-RESOURCE with your Azure resource name; the model is your deployment name."},
		{ID: "bedrock", Name: "Amazon Bedrock", BaseURL: "https://bedrock-runtime.us-east-1.amazonaws.com/openai/v1",
			Notes: "Uses a Bedrock API key. Change the region in the URL to match your account."},
		{ID: "ollama-cloud", Name: "Ollama Cloud", BaseURL: "https://ollama.com/v1"},
	})
}

// localProviders are servers on this machine. They need no key and expect
// the older "max_tokens" field.
func localProviders() []Definition {
	defs := []Definition{
		{ID: "lmstudio", Name: "LM Studio", BaseURL: "http://localhost:1234/v1"},
		{ID: "ollama", Name: "Ollama", BaseURL: "http://localhost:11434/v1"},
		{ID: "llamacpp", Name: "llama.cpp", BaseURL: "http://localhost:9931/v1",
			Notes: "Older llama.cpp builds listen on port 8080."},
		{ID: "vllm", Name: "vLLM", BaseURL: "http://localhost:8000/v1"},
		{ID: "jan", Name: "Jan", BaseURL: "http://127.0.0.1:1337/v1"},
		{ID: "localai", Name: "LocalAI", BaseURL: "http://localhost:8080/v1"},
		{ID: "koboldcpp", Name: "KoboldCpp", BaseURL: "http://localhost:5001/v1"},
		{ID: "textgen-webui", Name: "text-generation-webui", BaseURL: "http://127.0.0.1:5000/v1"},
	}
	for i := range defs {
		defs[i].KeyOptional, defs[i].MaxTokensField = true, openai.MaxTokens
	}
	return openAICompatible(CategoryLocal, defs)
}

// openAICompatible marks defs as OpenAI-style providers in category.
func openAICompatible(category string, defs []Definition) []Definition {
	for i := range defs {
		defs[i].Style, defs[i].Category = StyleOpenAI, category
	}
	return defs
}
