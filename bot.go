package traq_system_bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

var (
	verificationToken      string
	accessToken            string
	systemMessageChannelID string
	traqOrigin             string
)

func init() {
	verificationToken = os.Getenv("BOT_VERIFICATION_TOKEN")
	accessToken = os.Getenv("BOT_ACCESS_TOKEN")
	systemMessageChannelID = os.Getenv("BOT_SYSTEM_MESSAGE_CHANNEL_ID")
	traqOrigin = os.Getenv("TRAQ_ORIGIN")
	loggerInit()
}

func BotEndpoint(w http.ResponseWriter, r *http.Request) {
	defer logger.Flush()
	if r.Header.Get("X-TRAQ-BOT-TOKEN") != verificationToken {
		infoL(r, "Wrong X-TRAQ-BOT-TOKEN request was received")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	event := r.Header.Get("X-TRAQ-BOT-EVENT")
	switch event {
	case "PING":
		infoL(r, "PING was received")
		w.WriteHeader(http.StatusNoContent)
	case "USER_CREATED":
		var req userCreatedPayload
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		infoL(r, fmt.Sprintf("USER_CREATED(UID:%s) was received", req.User.ID))

		if !req.User.Bot {
			if err := sendMessage(systemMessageChannelID, fmt.Sprintf(`%s がtraQに参加しました`, createUserMention(req.User))); err != nil {
				errorL(r, fmt.Sprintf("sendMessage failed: %v", err))
			}
		}
		w.WriteHeader(http.StatusNoContent)
	case "CHANNEL_CREATED":
		var req channelCreatedPayload
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		infoL(r, fmt.Sprintf("CHANNEL_CREATED(UID:%s) was received", req.Channel.ID))

		if err := sendMessage(systemMessageChannelID, fmt.Sprintf(`%s がチャンネル %s を作成しました`, createUserMention(req.Channel.Creator), createChannelMention(req.Channel))); err != nil {
			errorL(r, fmt.Sprintf("sendMessage failed: %v", err))
		}
		w.WriteHeader(http.StatusNoContent)
	case "STAMP_CREATED":
		var req stampCreatedPayload
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		infoL(r, fmt.Sprintf("STAMP_CREATED(SID:%s) was received", req.ID))

		if err := sendMessage(systemMessageChannelID, fmt.Sprintf("%s がスタンプ `:%s:` :%s: を作成しました", createUserMention(req.Creator), req.Name, req.Name)); err != nil {
			errorL(r, fmt.Sprintf("sendMessage failed: %v", err))
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		infoL(r, fmt.Sprintf("Unknown X-TRAQ-BOT-EVENT was received: %s", event))
		w.WriteHeader(http.StatusBadRequest)
	}
	return
}

func createUserMention(user userPayload) string {
	return fmt.Sprintf(`!{"type":"user","raw":"@%s","id":"%s"}`, user.Name, user.ID)
}

func createChannelMention(channel channelPayload) string {
	return fmt.Sprintf(`!{"type":"channel","raw":"%s","id":"%s"}`, channel.Path, channel.ID)
}

func sendMessage(channelID string, text string) error {
	b, _ := json.Marshal(map[string]string{"text": text})
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/1.0/channels/%s/messages", traqOrigin, channelID), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return errors.New(res.Status)
	}
	return nil
}
