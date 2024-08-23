package main

import (
	"context"
	"fmt"
	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/traPtitech/traq-ws-bot/payload"
	"log/slog"
	"os"
)

func mustGetEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("environment variable %s must be set", key))
	}
	return v
}

var (
	systemMessageChannelID = mustGetEnv("BOT_SYSTEM_MESSAGE_CHANNEL_ID")
)

func main() {
	bot, err := traqwsbot.NewBot(&traqwsbot.Options{
		AccessToken: mustGetEnv("BOT_ACCESS_TOKEN"),
		Origin:      os.Getenv("TRAQ_ORIGIN"),
	})
	if err != nil {
		panic(err)
	}

	registerHandlers(bot)

	err = bot.Start()
	if err != nil {
		panic(err)
	}
}

func registerHandlers(bot *traqwsbot.Bot) {
	bot.OnUserCreated(func(p *payload.UserCreated) {
		slog.Info("USER_CREATED event received", "uid", p.User.ID)
		if !p.User.Bot {
			if err := sendMessage(bot, fmt.Sprintf(`%s がtraQに参加しました`, createUserMention(p.User))); err != nil {
				slog.Error("sendMessage failed", "err", err)
			}
		}
	})
	bot.OnUserActivated(func(p *payload.UserActivated) {
		slog.Info("USER_ACTIVATED event received", "uid", p.User.ID)
		if !p.User.Bot {
			if err := sendMessage(bot, fmt.Sprintf(`%s がtraQに帰ってきました`, createUserMention(p.User))); err != nil {
				slog.Error("sendMessage failed", "err", err)
			}
		}
	})

	bot.OnChannelCreated(func(p *payload.ChannelCreated) {
		slog.Info("CHANNEL_CREATED event received", "cid", p.Channel.ID)
		if err := sendMessage(bot, fmt.Sprintf(`%s がチャンネル %s を作成しました`, createUserMention(p.Channel.Creator), createChannelMention(p.Channel))); err != nil {
			slog.Error("sendMessage failed", "err", err)
		}
	})

	bot.OnStampCreated(func(p *payload.StampCreated) {
		slog.Info("STAMP_CREATED event received", "sid", p.ID)
		if err := sendMessage(bot, fmt.Sprintf("%s がスタンプ `:%s:` を作成しました\n:%s.ex-large:", createUserMention(p.Creator), p.Name, p.Name)); err != nil {
			slog.Error("sendMessage failed", "err", err)
		}
	})
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
