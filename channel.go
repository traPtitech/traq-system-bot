package main

import (
	"fmt"
	"log/slog"
	"slices"

	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/traPtitech/traq-ws-bot/payload"
)

func registerChannelHandlers(bot *traqwsbot.Bot) {
	if slices.Contains(subscribingEvents, eventTypeChannelCreated) {
		bot.OnChannelCreated(func(p *payload.ChannelCreated) {
			slog.Info("CHANNEL_CREATED event received", "cid", p.Channel.ID)
			if err := sendMessage(bot, fmt.Sprintf(`%s がチャンネル %s を作成しました`, createUserMention(p.Channel.Creator), createChannelMention(p.Channel))); err != nil {
				slog.Error("sendMessage failed", "err", err)
			}
		})
	}
}

func createChannelMention(channel payload.Channel) string {
	return fmt.Sprintf(`!{"type":"channel","raw":"%s","id":"%s"}`, channel.Path, channel.ID)
}
