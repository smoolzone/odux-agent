package api

import (
	"fmt"
	"net/url"
)

// Event is one timeline record (audit or telemetry).
type Event struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	ServerID  *int   `json:"serverId"`
	Metadata  string `json:"metadata"`
	CreatedAt string `json:"createdAt"`
}

// ListEvents returns the unified audit + telemetry timeline.
func (c *Client) ListEvents(token string, limit, serverID int, action string) ([]Event, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if serverID > 0 {
		q.Set("serverId", fmt.Sprintf("%d", serverID))
	}
	if action != "" {
		q.Set("action", action)
	}
	path := "/api/events"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var out []Event
	if err := request(c.Base, token, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
