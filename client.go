package apnskit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vvvu/apnskit/internal/transport"
)

type Client struct {
	client *http.Client
	env    Environment
}

// NewClient creates a new APNs client with the given options.
// If no HTTP client is provided via WithHTTPClient, an HTTP/2-capable
// client is used by default. If no environment is provided via WithEnv,
// Sandbox is used by default.
func NewClient(opts ...Option) (*Client, error) {
	cfg := &option{
		env: Sandbox,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.httpClient == nil {
		cfg.httpClient = transport.NewHTTP2Client()
	}

	return &Client{
		client: cfg.httpClient,
		env:    cfg.env,
	}, nil
}

// Push a notification request to APNs.
func (c *Client) Push(ctx context.Context, n *Notification) (*Response, error) {
	return c.pushNotification(ctx, n)
}

func (c *Client) pushNotification(ctx context.Context, n *Notification) (*Response, error) {
	if n == nil {
		return nil, errors.New("apnskit: request is nil")
	}

	if err := n.Validate(); err != nil {
		return nil, err
	}

	url := strings.TrimRight(c.env.Host(), "/") + "/3/device/" + n.DeviceToken
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(n.Payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("apns-topic", n.Topic)

	if n.Authorization != "" {
		req.Header.Set("authorization", n.Authorization)
	}

	if n.PushType != "" {
		req.Header.Set("apns-push-type", string(n.PushType))
	}

	if n.ID != "" {
		req.Header.Set("apns-id", n.ID)
	}

	if n.Expiration != nil {
		req.Header.Set("apns-expiration", strconv.FormatInt(*n.Expiration, 10))
	}

	if n.Priority != nil {
		req.Header.Set("apns-priority", strconv.Itoa(int(*n.Priority)))
	}

	if n.CollapseID != "" {
		req.Header.Set("apns-collapse-id", n.CollapseID)
	}

	httpResp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer httpResp.Body.Close()

	result := &Response{
		APNSID:     httpResp.Header.Get("apns-id"),
		UniqueID:   httpResp.Header.Get("apns-unique-id"),
		StatusCode: httpResp.StatusCode,
	}

	if httpResp.StatusCode == http.StatusOK {
		_, err := io.Copy(io.Discard, httpResp.Body)
		return result, err
	}

	var payload struct {
		Reason    ErrorReason `json:"reason"`
		Timestamp *int64      `json:"timestamp"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	result.Error = &ResponseError{
		Reason: payload.Reason,
	}

	if payload.Timestamp != nil {
		timestamp := time.UnixMilli(*payload.Timestamp).UTC()
		result.Error.Timestamp = &timestamp
	}

	return result, nil
}
