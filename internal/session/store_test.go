package session

import (
	"testing"

	"gocode/internal/llm"
)

func TestStoreSaveLoadList(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	sess := New("/tmp", "openai/gpt-5")
	sess.Title = "Fix the bug"
	sess.Messages = append(sess.Messages, llm.Message{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart("hi")}})
	if err := store.Save(sess); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load(sess.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Title != sess.Title || loaded.Messages[0].Text() != "hi" {
		t.Fatalf("loaded %+v", loaded)
	}
	list, err := store.List()
	if err != nil || len(list) != 1 || list[0].Title != "Fix the bug" {
		t.Fatalf("list = %v, %v", list, err)
	}
}
