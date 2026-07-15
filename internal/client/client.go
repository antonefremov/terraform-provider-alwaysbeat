// Package client is a thin HTTP client for the Alwaysbeat JSON API
// (/api/v1/*). It is intentionally self-contained (no dependency on the closed
// core's internal packages) so it can later graduate into a public alwaysbeat-go SDK.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultEndpoint is the production API front door (CloudFront). Overridable
// via the provider's `endpoint` argument or ALWAYSBEAT_ENDPOINT for staging/local.
const DefaultEndpoint = "https://dn8kvceixrafj.cloudfront.net"

// ErrNotFound is returned by Get/Update/Delete/SetPaused when the check no
// longer exists (HTTP 404), so the resource layer can drop it from state.
var ErrNotFound = errors.New("check not found")

// Client talks to one Alwaysbeat API endpoint with one API key.
type Client struct {
	endpoint string // no trailing slash
	apiKey   string
	http     *http.Client
}

// New builds a Client. endpoint may carry a trailing slash (trimmed).
func New(endpoint, apiKey string) *Client {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	return &Client{
		endpoint: strings.TrimRight(endpoint, "/"),
		apiKey:   apiKey,
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

// Schedule is the nested schedule block in a create/update request.
type Schedule struct {
	Kind     string `json:"kind,omitempty"`
	Interval string `json:"interval,omitempty"` // Go duration, e.g. "1h"
	CronExpr string `json:"cron_expr,omitempty"`
	TZ       string `json:"tz,omitempty"`
}

// CheckInput is the create/update request body. Durations are Go duration
// strings (the API parses them); the API stores and returns seconds.
type CheckInput struct {
	Name         string   `json:"name,omitempty"`
	Schedule     Schedule `json:"schedule"`
	Grace        string   `json:"grace,omitempty"`
	Channels     []string `json:"channels,omitempty"`
	FlapCooldown string   `json:"flap_cooldown,omitempty"`
	NagInterval  string   `json:"nag_interval,omitempty"`
	MaxRun       string   `json:"max_run,omitempty"`
	MaxRunMode   string   `json:"max_run_mode,omitempty"`
}

// Check is the API response shape. Durations come back as *_s seconds.
type Check struct {
	CheckID       string   `json:"check_id"`
	AccountID     string   `json:"account_id"`
	Name          string   `json:"name"`
	Status        string   `json:"status"`
	ScheduleKind  string   `json:"schedule_kind"`
	IntervalS     int64    `json:"interval_s,omitempty"`
	CronExpr      string   `json:"cron_expr,omitempty"`
	TZ            string   `json:"tz"`
	GraceS        int64    `json:"grace_s"`
	Channels      []string `json:"channels,omitempty"`
	FlapCooldownS int64    `json:"flap_cooldown_s,omitempty"`
	NagIntervalS  int64    `json:"nag_interval_s,omitempty"`
	MaxRunS       int64    `json:"max_run_s,omitempty"`
	MaxRunMode    string   `json:"max_run_mode,omitempty"`
	PingURL       string   `json:"ping_url"`
}

// apiError is the API's {"error": "..."} body.
type apiError struct {
	Error string `json:"error"`
}

// CreateCheck POSTs a new check.
func (c *Client) CreateCheck(ctx context.Context, in CheckInput) (*Check, error) {
	var out Check
	if err := c.do(ctx, http.MethodPost, "/api/v1/checks", in, &out, http.StatusCreated); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCheck reads one check by id. Returns ErrNotFound on 404.
func (c *Client) GetCheck(ctx context.Context, id string) (*Check, error) {
	var out Check
	if err := c.do(ctx, http.MethodGet, "/api/v1/checks/"+id, nil, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateCheck PATCHes a check. The API treats a full body as a merge patch.
func (c *Client) UpdateCheck(ctx context.Context, id string, in CheckInput) (*Check, error) {
	var out Check
	if err := c.do(ctx, http.MethodPatch, "/api/v1/checks/"+id, in, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteCheck removes a check. A 404 is treated as success (already gone).
func (c *Client) DeleteCheck(ctx context.Context, id string) error {
	err := c.do(ctx, http.MethodDelete, "/api/v1/checks/"+id, nil, nil, http.StatusNoContent)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}

// SetPaused pauses or resumes a check via the dedicated actions.
func (c *Client) SetPaused(ctx context.Context, id string, paused bool) (*Check, error) {
	action := "resume"
	if paused {
		action = "pause"
	}
	var out Check
	if err := c.do(ctx, http.MethodPost, "/api/v1/checks/"+id+"/"+action, nil, &out, http.StatusOK); err != nil {
		return nil, err
	}
	return &out, nil
}

// do performs one request, sending body (if non-nil) as JSON and decoding a
// JSON response into out (if non-nil). A status other than wantStatus is an
// error; 404 maps to ErrNotFound; the API's {"error"} message is surfaced.
func (c *Client) do(ctx context.Context, method, path string, body, out any, wantStatus int) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.endpoint+path, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	// Behind CloudFront the Authorization header is consumed by OAC SigV4, so
	// the API key travels in X-DMF-Token.
	req.Header.Set("X-DMF-Token", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != wantStatus {
		return fmt.Errorf("%s %s: unexpected status %d: %s", method, path, resp.StatusCode, readAPIError(resp.Body))
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// readAPIError best-effort extracts the API's error message from a failed
// response body, falling back to the raw text.
func readAPIError(r io.Reader) string {
	b, _ := io.ReadAll(io.LimitReader(r, 8<<10))
	var ae apiError
	if json.Unmarshal(b, &ae) == nil && ae.Error != "" {
		return ae.Error
	}
	return strings.TrimSpace(string(b))
}
