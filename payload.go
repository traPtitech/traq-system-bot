package traq_system_bot

import "time"

type basePayload struct {
	EventTime time.Time `json:"eventTime"`
}

type userPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	IconID      string `json:"iconId"`
	Bot         bool   `json:"bot"`
}

type userCreatedPayload struct {
	basePayload
	User userPayload `json:"user"`
}

type userActivatedPayload struct {
	basePayload
	User userPayload `json:"user"`
}

type channelPayload struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	ParentID  string      `json:"parentId"`
	Creator   userPayload `json:"creator"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

type channelCreatedPayload struct {
	basePayload
	Channel channelPayload `json:"channel"`
}

type stampCreatedPayload struct {
	basePayload
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	FileID  string      `json:"fileId"`
	Creator userPayload `json:"creator"`
}
