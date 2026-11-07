package main

import (
	"fmt"
	"os"
	"strings"

	traqwsbot "github.com/traPtitech/traq-ws-bot"
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

	if err := registerHandlers(bot); err != nil {
		panic(err)
	}

	err = bot.Start()
	if err != nil {
		panic(err)
	}
}

func registerHandlers(bot *traqwsbot.Bot) error {
	registerUserHandlers(bot)
	registerChannelHandlers(bot)
	return registerStampHandlers(bot)
}
