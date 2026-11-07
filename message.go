package main

import (
	"context"

	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
)

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
