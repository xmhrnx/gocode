package settingsview

import (
	"strings"

	"gocode/internal/providers"
	"gocode/ui/core"
)

func providerEntries(s *core.Services) []entry {
	var names []string
	for _, def := range providers.Catalog() {
		names = append(names, def.Name, def.ID)
	}
	return []entry{{
		title:       "Providers",
		description: "Turn on the services you use and add their keys. Local servers (LM Studio, Ollama, llama.cpp…) need no key; start the server, then switch it on.",
		keywords:    "api key provider local custom endpoint base url " + strings.Join(names, " "),
		wide:        true,
		control:     newProvidersPage(s),
	}}
}
