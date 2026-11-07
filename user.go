package main

import (
	"fmt"
	"log/slog"
	"slices"

	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/traPtitech/traq-ws-bot/payload"
)

func registerUserHandlers(bot *traqwsbot.Bot) {
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
}

func createUserMention(user payload.User) string {
	return fmt.Sprintf(`!{"type":"user","raw":"@%s","id":"%s"}`, user.Name, user.ID)
}
