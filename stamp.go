package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/traPtitech/traq-ws-bot/payload"
)

type stampNotifier struct {
	mu     sync.Mutex
	stamps map[string]stampInfo
	events []eventType
	send   func(string) error
}

type stampInfo struct {
	name  string
	owner payload.User // A zero ID or the nil UUID means the stamp has no owner.
}

const nilUUID = "00000000-0000-0000-0000-000000000000"

// The SDK does not yet define STAMP_UPDATED or STAMP_DELETED payloads.
// STAMP_UPDATED has the same fields as STAMP_CREATED.
type stampUpdated payload.StampCreated

type stampDeleted struct {
	payload.Base
	ID string `json:"id"`
}

func registerStampHandlers(bot *traqwsbot.Bot) error {
	if !slices.Contains(subscribingEvents, eventTypeStampCreated) &&
		!slices.Contains(subscribingEvents, eventTypeStampUpdated) &&
		!slices.Contains(subscribingEvents, eventTypeStampDeleted) {
		return nil
	}

	stamps, _, err := bot.API().StampApi.GetStamps(context.Background()).Execute()
	if err != nil {
		return fmt.Errorf("load stamps: %w", err)
	}
	// Stamps only carry the owner's ID, so resolve names for owner mentions.
	// Suspended users can still own stamps.
	users, _, err := bot.API().UserApi.GetUsers(context.Background()).IncludeSuspended(true).Execute()
	if err != nil {
		return fmt.Errorf("load users: %w", err)
	}
	userNames := make(map[string]string, len(users))
	for _, user := range users {
		userNames[user.Id] = user.Name
	}

	n := &stampNotifier{
		stamps: make(map[string]stampInfo, len(stamps)),
		events: subscribingEvents,
		send:   func(text string) error { return sendMessage(bot, text) },
	}
	for _, stamp := range stamps {
		n.stamps[stamp.Id] = stampInfo{
			name:  stamp.Name,
			owner: payload.User{ID: stamp.CreatorId, Name: userNames[stamp.CreatorId]},
		}
	}

	// Keep the cache current even when a stamp event's notification is disabled.
	bot.OnStampCreated(n.created)
	bot.OnEvent("STAMP_UPDATED", func(raw json.RawMessage) {
		var p stampUpdated
		if err := json.Unmarshal(raw, &p); err != nil {
			slog.Error("invalid STAMP_UPDATED payload", "err", err)
			return
		}
		n.updated(&p)
	})
	bot.OnEvent("STAMP_DELETED", func(raw json.RawMessage) {
		var p stampDeleted
		if err := json.Unmarshal(raw, &p); err != nil {
			slog.Error("invalid STAMP_DELETED payload", "err", err)
			return
		}
		n.deleted(&p)
	})
	return nil
}

func (n *stampNotifier) created(p *payload.StampCreated) {
	n.mu.Lock()
	defer n.mu.Unlock()

	slog.Info("STAMP_CREATED event received", "sid", p.ID)
	n.stamps[p.ID] = stampInfo{name: p.Name, owner: p.Creator}
	if slices.Contains(n.events, eventTypeStampCreated) {
		n.notify(fmt.Sprintf("%s がスタンプ `:%s:` を作成しました\n:%s.ex-large:", createUserMention(p.Creator), p.Name, p.Name))
	}
}

func (n *stampNotifier) updated(p *stampUpdated) {
	n.mu.Lock()
	defer n.mu.Unlock()

	slog.Info("STAMP_UPDATED event received", "sid", p.ID)
	before := n.description(p.ID)
	old, known := n.stamps[p.ID]
	n.stamps[p.ID] = stampInfo{name: p.Name, owner: p.Creator}
	if slices.Contains(n.events, eventTypeStampUpdated) {
		text := fmt.Sprintf("%s が更新されました", before)
		if old.name != p.Name {
			text = fmt.Sprintf("%s が `:%s:` に更新されました", before, p.Name)
		}
		// The payload's creator is the owner after the update. Skip unknown
		// stamps because their previous owner cannot be determined.
		if known && ownerID(old.owner) != ownerID(p.Creator) {
			text += fmt.Sprintf("\n所有者: %s → %s", ownerMention(old.owner), ownerMention(p.Creator))
		}
		n.notify(fmt.Sprintf("%s\n:%s.ex-large:", text, p.Name))
	}
}

func (n *stampNotifier) deleted(p *stampDeleted) {
	n.mu.Lock()
	defer n.mu.Unlock()

	slog.Info("STAMP_DELETED event received", "sid", p.ID)
	before := n.description(p.ID)
	delete(n.stamps, p.ID)
	if slices.Contains(n.events, eventTypeStampDeleted) {
		n.notify(fmt.Sprintf("%s が削除されました", before))
	}
}

// description must be called with mu held. Unknown stamps can occur if events
// were missed while the WebSocket was disconnected.
func (n *stampNotifier) description(id string) string {
	if stamp, ok := n.stamps[id]; ok {
		return fmt.Sprintf("スタンプ `:%s:`", stamp.name)
	}
	return fmt.Sprintf("スタンプ (ID: `%s`)", id)
}

func (n *stampNotifier) notify(text string) {
	if err := n.send(text); err != nil {
		slog.Error("sendMessage failed", "err", err)
	}
}

// ownerID normalizes the representations of "no owner" to "".
func ownerID(u payload.User) string {
	if u.ID == nilUUID {
		return ""
	}
	return u.ID
}

func ownerMention(u payload.User) string {
	switch {
	case ownerID(u) == "":
		return "なし"
	case u.Name == "":
		return fmt.Sprintf("(ID: `%s`)", u.ID)
	default:
		return createUserMention(u)
	}
}
