package main

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/traPtitech/traq-ws-bot/payload"
)

func TestStampLifecycle(t *testing.T) {
	var messages []string
	n := &stampNotifier{
		names:  map[string]string{"existing": "original"},
		events: []eventType{eventTypeStampCreated, eventTypeStampUpdated, eventTypeStampDeleted},
		send: func(text string) error {
			messages = append(messages, text)
			return nil
		},
	}

	// Decode the actual wire format, including the custom SDK payloads.
	var updated stampUpdated
	if err := json.Unmarshal([]byte(`{"id":"existing","name":"renamed","fileId":"image","creator":{"id":"user","name":"alice"}}`), &updated); err != nil {
		t.Fatal(err)
	}
	n.updated(&updated)
	n.updated(&updated) // Image-only updates retain the same name.
	var deleted stampDeleted
	if err := json.Unmarshal([]byte(`{"id":"existing"}`), &deleted); err != nil {
		t.Fatal(err)
	}
	n.deleted(&deleted)
	n.created(&payload.StampCreated{ID: "new", Name: "added", Creator: payload.User{ID: "user", Name: "alice"}})
	n.deleted(&stampDeleted{ID: "new"})

	want := []string{
		"スタンプ `:original:` が `:renamed:` に更新されました\n:renamed.ex-large:",
		"スタンプ `:renamed:` が更新されました\n:renamed.ex-large:",
		"スタンプ `:renamed:` が削除されました",
		"!{\"type\":\"user\",\"raw\":\"@alice\",\"id\":\"user\"} がスタンプ `:added:` を作成しました\n:added.ex-large:",
		"スタンプ `:added:` が削除されました",
	}
	if !slices.Equal(messages, want) {
		t.Fatalf("messages = %q, want %q", messages, want)
	}
	if len(n.names) != 0 {
		t.Fatalf("deleted stamps remain cached: %v", n.names)
	}
}

func TestStampCacheTracksDisabledNotifications(t *testing.T) {
	var messages []string
	n := &stampNotifier{
		names:  make(map[string]string),
		events: []eventType{eventTypeStampDeleted},
		send: func(text string) error {
			messages = append(messages, text)
			return nil
		},
	}
	n.created(&payload.StampCreated{ID: "stamp", Name: "original"})
	n.updated(&stampUpdated{ID: "stamp", Name: "renamed"})
	n.deleted(&stampDeleted{ID: "stamp"})
	want := []string{"スタンプ `:renamed:` が削除されました"}
	if !slices.Equal(messages, want) {
		t.Fatalf("messages = %q, want %q", messages, want)
	}
}

func TestStampCacheSurvivesNotificationFailure(t *testing.T) {
	n := &stampNotifier{
		names:  map[string]string{"stamp": "original"},
		events: []eventType{eventTypeStampUpdated, eventTypeStampDeleted},
		send:   func(string) error { return errors.New("send failed") },
	}
	n.updated(&stampUpdated{ID: "stamp", Name: "renamed"})
	if n.names["stamp"] != "renamed" {
		t.Fatal("failed notification prevented cache update")
	}
	n.deleted(&stampDeleted{ID: "stamp"})
	if _, ok := n.names["stamp"]; ok {
		t.Fatal("failed notification prevented cache deletion")
	}
}

func TestUnknownStampNotification(t *testing.T) {
	var messages []string
	n := &stampNotifier{
		names:  make(map[string]string),
		events: []eventType{eventTypeStampUpdated, eventTypeStampDeleted},
		send: func(text string) error {
			messages = append(messages, text)
			return nil
		},
	}
	n.updated(&stampUpdated{ID: "unknown-update", Name: "recovered"})
	n.deleted(&stampDeleted{ID: "unknown-delete"})
	if !strings.Contains(messages[0], "unknown-update") || !strings.Contains(messages[1], "unknown-delete") {
		t.Fatalf("unknown stamps must be identified by ID: %q", messages)
	}
	if n.names["unknown-update"] != "recovered" {
		t.Fatal("updated stamp was not cached")
	}
}

func TestConcurrentStampEvents(t *testing.T) {
	n := &stampNotifier{names: make(map[string]string)}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n.created(&payload.StampCreated{ID: "stamp", Name: "original"})
			n.updated(&stampUpdated{ID: "stamp", Name: "renamed"})
			n.deleted(&stampDeleted{ID: "stamp"})
		}()
	}
	wg.Wait()
}

func TestParseStampEvents(t *testing.T) {
	got := mustParseSubscribingEvents("STAMP_CREATED,STAMP_UPDATED,STAMP_DELETED")
	want := []eventType{eventTypeStampCreated, eventTypeStampUpdated, eventTypeStampDeleted}
	if !slices.Equal(got, want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
}
