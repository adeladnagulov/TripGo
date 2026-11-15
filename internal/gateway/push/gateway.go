package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	pushapi "github.com/adeladnagulov/TripGo/api/push"
	"github.com/google/uuid"
)

type Gateway struct {
	client  *http.Client
	baseURL string
}

func NewGateway(baseURl string, timeout time.Duration) *Gateway {
	return &Gateway{
		baseURL: baseURl,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (g *Gateway) ReuestPosition(ctx context.Context, tripId, driverId uuid.UUID) error {
	reqBody := pushapi.PushRequest{
		RecipientId: driverId,
		RequestId:   uuid.NewString(),
		Kind:        pushapi.REQUESTPOSITION,
		Data: &map[string]string{
			"trip_id": tripId.String(),
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal push request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/api/v1/push", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("transport error: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusAccepted:
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	case http.StatusBadRequest:
		return fmt.Errorf("push service bad request (400)")
	case http.StatusTooManyRequests:
		return fmt.Errorf("push service rate limited (429)")
	case http.StatusServiceUnavailable:
		return fmt.Errorf("push service	 unavailable (503)")
	default:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("push service unexpected status %d: %s", resp.StatusCode, string(body))
	}
}
