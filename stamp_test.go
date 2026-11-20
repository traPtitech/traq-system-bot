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
		stamps: map[string]stampInfo{"existing": {name: "original", owner: payload.User{ID: "user", Name: "alice"}}},
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
	if len(n.stamps) != 0 {
		t.Fatalf("deleted stamps remain cached: %v", n.stamps)
	}
}

func TestStampCacheTracksDisabledNotifications(t *testing.T) {
	var messages []string
	n := &stampNotifier{
		stamps: make(map[string]stampInfo),
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
		stamps: map[string]stampInfo{"stamp": {name: "original"}},
		events: []eventType{eventTypeStampUpdated, eventTypeStampDeleted},
		send:   func(string) error { return errors.New("send failed") },
	}
	n.updated(&stampUpdated{ID: "stamp", Name: "renamed"})
	if n.stamps["stamp"].name != "renamed" {
		t.Fatal("failed notification prevented cache update")
	}
	n.deleted(&stampDeleted{ID: "stamp"})
	if _, ok := n.stamps["stamp"]; ok {
		t.Fatal("failed notification prevented cache deletion")
	}
}

func TestUnknownStampNotification(t *testing.T) {
	var messages []string
	n := &stampNotifier{
		stamps: make(map[string]stampInfo),
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
	if n.stamps["unknown-update"].name != "recovered" {
		t.Fatal("updated stamp was not cached")
	}
}

func TestStampOwnerChange(t *testing.T) {
	alice := payload.User{ID: "alice-id", Name: "alice"}
	bob := payload.User{ID: "bob-id", Name: "bob"}
	var messages []string
	n := &stampNotifier{
		stamps: map[string]stampInfo{
			"owned":    {name: "owned", owner: alice},
			"system":   {name: "system", owner: payload.User{ID: nilUUID}},
			"nameless": {name: "nameless", owner: payload.User{ID: "ghost-id"}},
		},
		events: []eventType{eventTypeStampUpdated},
		send: func(text string) error {
			messages = append(messages, text)
			return nil
		},
	}

	n.updated(&stampUpdated{ID: "owned", Name: "owned", Creator: bob})
	n.updated(&stampUpdated{ID: "owned", Name: "renamed", Creator: bob}) // Owner unchanged.
	n.updated(&stampUpdated{ID: "owned", Name: "renamed", Creator: payload.User{ID: nilUUID}})
	n.updated(&stampUpdated{ID: "system", Name: "system", Creator: alice})
	n.updated(&stampUpdated{ID: "nameless", Name: "nameless", Creator: alice})
	n.updated(&stampUpdated{ID: "unknown", Name: "unknown", Creator: alice})

	want := []string{
		"スタンプ `:owned:` が更新されました\n所有者: !{\"type\":\"user\",\"raw\":\"@alice\",\"id\":\"alice-id\"} → !{\"type\":\"user\",\"raw\":\"@bob\",\"id\":\"bob-id\"}\n:owned.ex-large:",
		"スタンプ `:owned:` が `:renamed:` に更新されました\n:renamed.ex-large:",
		"スタンプ `:renamed:` が更新されました\n所有者: !{\"type\":\"user\",\"raw\":\"@bob\",\"id\":\"bob-id\"} → なし\n:renamed.ex-large:",
		"スタンプ `:system:` が更新されました\n所有者: なし → !{\"type\":\"user\",\"raw\":\"@alice\",\"id\":\"alice-id\"}\n:system.ex-large:",
		"スタンプ `:nameless:` が更新されました\n所有者: (ID: `ghost-id`) → !{\"type\":\"user\",\"raw\":\"@alice\",\"id\":\"alice-id\"}\n:nameless.ex-large:",
		"スタンプ (ID: `unknown`) が `:unknown:` に更新されました\n:unknown.ex-large:",
	}
	if !slices.Equal(messages, want) {
		t.Fatalf("messages = %q, want %q", messages, want)
	}
	if n.stamps["system"].owner != alice {
		t.Fatal("new owner was not cached")
	}
}

func TestStampOwnerTrackedWhileNotificationDisabled(t *testing.T) {
	alice := payload.User{ID: "alice-id", Name: "alice"}
	bob := payload.User{ID: "bob-id", Name: "bob"}
	var messages []string
	n := &stampNotifier{
		stamps: make(map[string]stampInfo),
		events: []eventType{eventTypeStampUpdated},
		send: func(text string) error {
			messages = append(messages, text)
			return nil
		},
	}
	n.created(&payload.StampCreated{ID: "stamp", Name: "stamp", Creator: alice})
	n.updated(&stampUpdated{ID: "stamp", Name: "stamp", Creator: bob})
	if len(messages) != 1 || !strings.Contains(messages[0], "@alice") || !strings.Contains(messages[0], "@bob") {
		t.Fatalf("owner from STAMP_CREATED was not cached: %q", messages)
	}
}

func TestConcurrentStampEvents(t *testing.T) {
	n := &stampNotifier{stamps: make(map[string]stampInfo)}
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
