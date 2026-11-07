package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/traPtitech/traq-ws-bot/payload"
)

func mustGetEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("environment variable %s must be set", key))
	}
	return v
}

func getEnvWithDefault(key, defaultValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return v
}

type eventType int

const (
	eventTypeUserCreated eventType = iota + 1
	eventTypeUserActivated
	eventTypeChannelCreated
	eventTypeStampCreated
	eventTypeStampUpdated
	eventTypeStampDeleted
)

func mustParseSubscribingEvents(s string) []eventType {
	var events []eventType
	for _, name := range strings.Split(s, ",") {
		switch name {
		case "USER_CREATED":
			events = append(events, eventTypeUserCreated)
		case "USER_ACTIVATED":
			events = append(events, eventTypeUserActivated)
		case "CHANNEL_CREATED":
			events = append(events, eventTypeChannelCreated)
		case "STAMP_CREATED":
			events = append(events, eventTypeStampCreated)
		case "STAMP_UPDATED":
			events = append(events, eventTypeStampUpdated)
		case "STAMP_DELETED":
			events = append(events, eventTypeStampDeleted)
		default:
			panic(fmt.Sprintf("unknown event type: %s", name))
		}
	}
	return events
}

var (
	systemMessageChannelID string
	subscribingEvents      []eventType
)

func main() {
	systemMessageChannelID = mustGetEnv("BOT_SYSTEM_MESSAGE_CHANNEL_ID")
	subscribingEvents = mustParseSubscribingEvents(getEnvWithDefault("BOT_SYSTEM_SUBSCRIBING_EVENTS", "USER_CREATED,USER_ACTIVATED,CHANNEL_CREATED,STAMP_CREATED,STAMP_UPDATED,STAMP_DELETED"))

	bot, err := traqwsbot.NewBot(&traqwsbot.Options{
		AccessToken: mustGetEnv("BOT_ACCESS_TOKEN"),
		Origin:      os.Getenv("TRAQ_ORIGIN"),
	})
	if err != nil {
		panic(err)
	}

	registerHandlers(bot)
	if err := registerStampHandlers(bot); err != nil {
		panic(err)
	}

	err = bot.Start()
	if err != nil {
		panic(err)
	}
}

func registerHandlers(bot *traqwsbot.Bot) {
	if slices.Contains(subscribingEvents, eventTypeUserCreated) {
		bot.OnUserCreated(func(p *payload.UserCreated) {
			slog.Info("USER_CREATED event received", "uid", p.User.ID)
			if !p.User.Bot {
				if err := sendMessage(bot, fmt.Sprintf(`%s がtraQに参加しました`, createUserMention(p.User))); err != nil {
					slog.Error("sendMessage failed", "err", err)
				}
			}
		})
	}

	if slices.Contains(subscribingEvents, eventTypeUserActivated) {
		bot.OnUserActivated(func(p *payload.UserActivated) {
			slog.Info("USER_ACTIVATED event received", "uid", p.User.ID)
			if !p.User.Bot {
				if err := sendMessage(bot, fmt.Sprintf(`%s がtraQに帰ってきました`, createUserMention(p.User))); err != nil {
					slog.Error("sendMessage failed", "err", err)
				}
			}
		})
	}

	if slices.Contains(subscribingEvents, eventTypeChannelCreated) {
		bot.OnChannelCreated(func(p *payload.ChannelCreated) {
			slog.Info("CHANNEL_CREATED event received", "cid", p.Channel.ID)
			if err := sendMessage(bot, fmt.Sprintf(`%s がチャンネル %s を作成しました`, createUserMention(p.Channel.Creator), createChannelMention(p.Channel))); err != nil {
				slog.Error("sendMessage failed", "err", err)
			}
		})
	}
}

func createUserMention(user payload.User) string {
	return fmt.Sprintf(`!{"type":"user","raw":"@%s","id":"%s"}`, user.Name, user.ID)
}

func createChannelMention(channel payload.Channel) string {
	return fmt.Sprintf(`!{"type":"channel","raw":"%s","id":"%s"}`, channel.Path, channel.ID)
}

func sendMessage(bot *traqwsbot.Bot, text string) error {
	_, _, err := bot.API().
		ChannelApi.
		PostMessage(context.Background(), systemMessageChannelID).
		PostMessageRequest(traq.PostMessageRequest{
			Content: text,
			Embed:   nil,
		}).
		Execute()
	return err
}
