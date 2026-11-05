package chat

import (
	"maps"
	"slices"
	"time"
)

// save writes the session in the background. It snapshots the session on
// the UI goroutine, so the writer never sees it change.
func (c *ChatTab) save() {
	if len(c.session.Messages) == 0 {
		return // an empty chat isn't worth a history entry
	}
	c.session.Updated = time.Now()
	snapshot := *c.session
	snapshot.Messages = slices.Clone(c.session.Messages)
	snapshot.ToolDisplays = maps.Clone(c.session.ToolDisplays)
	store, services := c.services.Sessions, c.services
	go func() {
		if err := store.Save(&snapshot); err != nil {
			services.Error("Couldn't save the chat: " + err.Error())
		}
	}()
}
