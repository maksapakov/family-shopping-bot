package matrix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/maksapakov/family-shopping-bot/internal/app"
	"github.com/maksapakov/family-shopping-bot/internal/domain"
)

type Listener struct {
	client     *http.Client
	homeserver string
	token      string
	roomID     string
	addItem    *app.AddItem
}

func NewListener(
	client *http.Client,
	homeserver string,
	token string,
	roomID string,
	addItem *app.AddItem,
) *Listener {
	return &Listener{
		client:     client,
		homeserver: homeserver,
		token:      token,
		roomID:     roomID,
		addItem:    addItem,
	}
}

func (l *Listener) Run(ctx context.Context) error {
	userID, err := l.whoami(ctx)
	if err != nil {
		return fmt.Errorf("whoami: %w", err)
	}
	slog.Info("matrix listener started", "user_id", userID)
	chatID := os.Getenv("SHOPPING_CHAT_ID")
	since := ""
	first := true

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		next, rooms, err := l.sync(ctx, since)
		if err != nil {
			return fmt.Errorf("sync: %w", err)
		}
		if !first {
			room, ok := rooms.Join[l.roomID]
			keys := make([]string, 0, len(rooms.Join))
			for id := range rooms.Join {
				keys = append(keys, id)
			}
			slog.Info("sync tick",
				"want_room", l.roomID,
				"join_keys", keys,
				"next", next,
			)
			if ok {
				for _, ev := range room.Timeline.Events {
					if ev.Type != "m.room.message" {
						continue
					}
					if ev.Content.MsgType != "m.text" {
						continue
					}
					if ev.Sender == userID {
						continue
					}
					slog.Info("incoming message",
						"sender", ev.Sender,
						"body", ev.Content.Body,
					)
					// парсим строку ввода
					names := app.ParseItemNames(ev.Content.Body)
					err := l.addItem.ExecuteMany(
						ctx,
						domain.ChatID(chatID),
						names,
						ev.Sender,
					)
					if err != nil {
						slog.Error("failed to execute", "error", err)
						continue
					}
					if err = l.redact(ctx, ev.EventID); err != nil {
						slog.Error("redact failed", "error", err, "event_id", ev.EventID)
					}
				}
			}
		}
		first = false
		since = next
	}
}

func (l *Listener) whoami(ctx context.Context) (string, error) {
	urlPath := fmt.Sprintf("%s/_matrix/client/v3/account/whoami",
		l.homeserver,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlPath, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", l.token))

	resp, err := l.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("failed to close response body")
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("matrix server responded with %d: %s", resp.StatusCode, string(b))
	}

	var out struct {
		UserID string `json:"user_id"`
	}
	err = json.Unmarshal(b, &out)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	return out.UserID, nil
}

func (l *Listener) redact(ctx context.Context, eventID string) error {
	txnId := fmt.Sprintf("%d", time.Now().UnixNano())
	u := fmt.Sprintf("%s/_matrix/client/v3/rooms/%s/redact/%s/%s",
		l.homeserver,
		url.PathEscape(l.roomID),
		url.PathEscape(eventID),
		txnId,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader([]byte("{}")))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", l.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("failed to close response body")
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("matrix server responded with %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

type SyncResp struct {
	NextBatch string `json:"next_batch"`
	Rooms     Rooms  `json:"rooms"`
}
type Rooms struct {
	Join map[string]struct {
		Timeline struct {
			Events []struct {
				EventID string `json:"event_id"`
				Type    string `json:"type"`
				Sender  string `json:"sender"`
				Content struct {
					MsgType string `json:"msgtype"`
					Body    string `json:"body"`
				} `json:"content"`
			} `json:"events"`
		} `json:"timeline"`
	} `json:"join"`
}

func (l *Listener) sync(ctx context.Context, since string) (string, Rooms, error) {
	u := fmt.Sprintf("%s/_matrix/client/v3/sync?timeout=30000",
		l.homeserver)
	if since != "" {
		u += "&since=" + url.QueryEscape(since)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", Rooms{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", l.token))

	resp, err := l.client.Do(req)
	if err != nil {
		return "", Rooms{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("failed to close response body")
		}
	}()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", Rooms{}, fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", Rooms{}, fmt.Errorf("matrix server responded with %d: %s", resp.StatusCode, string(b))
	}
	var out SyncResp
	err = json.Unmarshal(b, &out)
	if err != nil {
		return "", Rooms{}, fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	return out.NextBatch, out.Rooms, nil
}
