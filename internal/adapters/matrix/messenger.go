package matrix

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/maksapakov/family-shopping-bot/internal/domain"
	"github.com/maksapakov/family-shopping-bot/internal/port"
)

type Messenger struct {
	homeserver  string
	accessToken string
	httpClient  *http.Client
	// roomID: как маппить ChatID -> room? на MVP - один env MATRIX_ROOM_ID
	roomID string
}

func (m *Messenger) SendList(ctx context.Context, _ domain.ChatID, rendered domain.RenderedList) (port.MessageRef, error) {
	plain, html := formatList(rendered)

	type matrixMessage struct {
		MsgType       string `json:"msgtype"`
		Body          string `json:"body"`
		Format        string `json:"format"`
		FormattedBody string `json:"formatted_body"`
	}
	payload := matrixMessage{
		MsgType:       "m.text",
		Body:          plain,
		Format:        "org.matrix.custom.html",
		FormattedBody: html,
	}

	out, err := m.putRoomMessage(ctx, payload)
	if err != nil {
		return "", fmt.Errorf("failed to put room message: %w", err)
	}
	return port.MessageRef(out), nil
}

func (m *Messenger) UpdateList(ctx context.Context, _ domain.ChatID, ref port.MessageRef, rendered domain.RenderedList) error {
	plain, html := formatList(rendered)
	type NewContent struct {
		MsgType       string `json:"msgtype"`
		Body          string `json:"body"`
		Format        string `json:"format"`
		FormattedBody string `json:"formatted_body"`
	}
	type RelatesTo struct {
		RelType string `json:"rel_type"`
		EventID string `json:"event_id"`
	}
	type matrixMessage struct {
		MsgType       string     `json:"msgtype"`
		Body          string     `json:"body"`
		Format        string     `json:"format"`
		FormattedBody string     `json:"formatted_body"`
		MNewContent   NewContent `json:"m.new_content"`
		MRelatesTo    RelatesTo  `json:"m.relates_to"`
	}

	payload := matrixMessage{
		MsgType:       "m.text",
		Body:          "* " + plain,
		Format:        "org.matrix.custom.html",
		FormattedBody: html,
		MNewContent: NewContent{
			MsgType:       "m.text",
			Body:          plain,
			Format:        "org.matrix.custom.html",
			FormattedBody: html,
		},
		MRelatesTo: RelatesTo{
			RelType: "m.replace",
			EventID: string(ref),
		},
	}
	_, err := m.putRoomMessage(ctx, payload)
	if err != nil {
		return fmt.Errorf("failed to put room message: %w", err)
	}
	return nil
}

func NewMessenger(homeserver string, accessToken string, roomID string) (*Messenger, error) {
	if homeserver == "" || accessToken == "" || roomID == "" {
		return nil, fmt.Errorf("messenger: homeserver, accessToken, roomID must not be empty")
	}
	return &Messenger{
		homeserver:  homeserver,
		accessToken: accessToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		roomID: roomID,
	}, nil
}

var _ port.ChatMessenger = (*Messenger)(nil)
