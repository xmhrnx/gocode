package chat

import (
	"errors"
	"fmt"
	"slices"

	"gocode/internal/secrets"
	"gocode/internal/skills"
	"gocode/internal/tools"
	"gocode/internal/websearch"
)

// toolRegistry returns the tools for a run: the built-in ones, web search
// when a service is set up, the skill loader, and the tools of the
// connected MCP servers.
func (c *ChatTab) toolRegistry() *tools.Registry {
	all := append(tools.Builtin(), skills.NewTool(c.skillLister()))
	if search, ok := c.webSearchTool(); ok {
		all = append(all, search)
	}
	all = append(all, c.services.MCP.Tools()...)
	return tools.NewRegistry(all...)
}

// webSearchTool returns the web_search tool for the service set in
// Settings → Web search, if one is.
func (c *ChatTab) webSearchTool() (tools.Tool, bool) {
	cfg := c.services.Settings.WebSearch
	if cfg.Service == "" {
		return nil, false
	}
	keys, client := c.services.Keys, c.services.HTTP
	searcher := websearch.Lazy(func() (websearch.Searcher, error) {
		key := ""
		if websearch.NeedsKey(cfg.Service) {
			var err error
			if key, err = keys.Get(websearch.KeyName(cfg.Service)); err != nil && !errors.Is(err, secrets.ErrNotFound) {
				return nil, fmt.Errorf("couldn't read the search key: %w", err)
			}
		}
		return websearch.New(websearch.Config{Service: cfg.Service, Key: key, URL: cfg.SearXNGURL, Client: client})
	})
	return tools.WebSearch{Searcher: searcher, Results: cfg.Results}, true
}

// skillLister returns a function listing the enabled skills for this
// chat's folder. It reads the disk, so it runs on the run's goroutine.
func (c *ChatTab) skillLister() func() []skills.Skill {
	userDir, workDir := c.services.SkillsDir, c.session.WorkDir
	disabled := slices.Clone(c.services.Settings.Skills.Disabled)
	return func() []skills.Skill {
		found, _ := skills.Discover(userDir, workDir) // broken skills are shown in Settings → Skills
		return skills.Enabled(found, disabled)
	}
}
