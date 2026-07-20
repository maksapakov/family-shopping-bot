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
	"time"
)

func (m *Messenger) putRoomMessage(ctx context.Context, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return []byte(""), fmt.Errorf("marshal payload: %w", err)
	}

	txnID := fmt.Sprintf("%d", time.Now().UnixNano())
	urlPath := fmt.Sprintf("%s/_matrix/client/v3/rooms/%s/send/m.room.message/%s",
		m.homeserver,
		url.PathEscape(m.roomID),
		txnID,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, urlPath, bytes.NewReader(body))
	if err != nil {
		return []byte(""), fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.accessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return []byte(""), fmt.Errorf("send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}()

	var out struct {
		EventID string `json:"event_id"`
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return []byte(""), fmt.Errorf("matrix send: status code %d: %s", resp.StatusCode, string(b))
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return []byte(""), fmt.Errorf("unmarshal response body: %w", err)
	}
	return []byte(out.EventID), nil
}
