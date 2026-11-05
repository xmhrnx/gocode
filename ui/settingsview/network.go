package settingsview

import "gocode/ui/core"

func networkEntries(s *core.Services) []entry {
	n := &s.Settings.Network
	return []entry{
		{title: "Response timeout (seconds)", description: "How long to wait for a provider to start answering. Applies after a restart.",
			keywords: "timeout network request slow", control: number(s, &n.RequestTimeoutSeconds, 10, 3600)},
		{title: "Retries", description: "How often a failed request (rate limit, overload, network error) is retried, with growing delays.",
			keywords: "retry rate limit 429 overloaded", control: number(s, &n.MaxRetries, 0, 10)},
	}
}
