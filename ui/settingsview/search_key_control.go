package settingsview

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"gocode/internal/secrets"
	"gocode/internal/websearch"
	"gocode/ui/core"

	"gocode/ui/components"
)

// searchKeyControl stores the API key of the chosen search service in the
// keychain and tests the service with a search.
type searchKeyControl struct {
	services *core.Services
	key      *input.Input
	save     *giode.ButtonElement
	test     *giode.ButtonElement
}

// searchTestTimeout bounds the test search.
const searchTestTimeout = 20 * time.Second

func newSearchKeyControl(s *core.Services) *searchKeyControl {
	c := &searchKeyControl{services: s}
	c.key = giode.Input.Text("Paste the key", fieldStyle(220)).Password(true).Submit(func(string) { c.store() })
	components.AddEditMenu(s.App, c.key, components.EditOptions{Secret: true})
	c.save = smallButton("Save key", true).OnClick(c.store)
	c.test = smallButton("Test", false).OnClick(c.runTest)
	return c
}

func (c *searchKeyControl) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter}, c.key, c.save, c.test).Layout(gtx)
}

// store saves the key for the service chosen now.
func (c *searchKeyControl) store() {
	s, service := c.services, c.services.Settings.WebSearch.Service
	value := strings.TrimSpace(c.key.Text())
	if value == "" {
		return
	}
	if !websearch.NeedsKey(service) {
		s.Error("Choose a search service that takes a key first.")
		return
	}
	c.key.SetText("")
	go func() {
		err := s.Keys.Set(websearch.KeyName(service), value)
		s.App.Post(func() {
			if err != nil {
				s.Error("Couldn't save the key: " + err.Error())
				return
			}
			s.Info("Key saved for " + serviceName(service) + ".")
		})
	}()
}

// runTest searches for something once and reports how it went.
func (c *searchKeyControl) runTest() {
	s, cfg := c.services, c.services.Settings.WebSearch
	if cfg.Service == "" {
		s.Error("Choose a search service first.")
		return
	}
	go func() {
		err := testSearch(s, cfg.Service, cfg.SearXNGURL)
		s.App.Post(func() {
			if err != nil {
				s.Error("Search failed: " + err.Error())
				return
			}
			s.Info(serviceName(cfg.Service) + " works.")
		})
	}()
}

func testSearch(s *core.Services, service, url string) error {
	key := ""
	if websearch.NeedsKey(service) {
		var err error
		if key, err = s.Keys.Get(websearch.KeyName(service)); err != nil {
			if errors.Is(err, secrets.ErrNotFound) {
				return fmt.Errorf("no key saved for %s", serviceName(service))
			}
			return err
		}
	}
	searcher, err := websearch.New(websearch.Config{Service: service, Key: key, URL: url, Client: s.HTTP})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), searchTestTimeout)
	defer cancel()
	results, err := searcher.Search(ctx, "Go programming language", 1)
	if err == nil && len(results) == 0 {
		err = errors.New("it answered, but with no results")
	}
	return err
}
