package apnskit

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_PushNotification_Request(t *testing.T) {
	expiration := int64(1_715_603_200)
	priority := PriorityImmediate

	var sentRequest *http.Request
	httpClient := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			sentRequest = req
			return &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Apns-Id": []string{"123e4567-e89b-12d3-a456-4266554400a0"},
				},
				Body: io.NopCloser(strings.NewReader("")),
			}, nil
		}),
	}

	client, err := NewClient(WithEnv(Production), WithHTTPClient(httpClient))
	require.NoError(t, err)

	notification := &Notification{
		DeviceToken:   "abc123",
		Payload:       []byte(`{"aps":{"alert":"hello"}}`),
		Authorization: "bearer test-token",
		PushType:      PushTypeAlert,
		ID:            "123e4567-e89b-12d3-a456-4266554400a0",
		Expiration:    &expiration,
		Priority:      &priority,
		Topic:         "com.example.app",
		CollapseID:    "order-42",
	}

	resp, err := client.pushNotification(context.Background(), notification)
	require.NoError(t, err)
	require.True(t, resp.Success())

	require.Equal(t, "POST", sentRequest.Method)
	require.Equal(t, "https://api.push.apple.com/3/device/abc123", sentRequest.URL.String())
	require.Equal(t, "application/json", sentRequest.Header.Get("content-type"))
	require.Equal(t, "bearer test-token", sentRequest.Header.Get("authorization"))
	require.Equal(t, "alert", sentRequest.Header.Get("apns-push-type"))
	require.Equal(t, "123e4567-e89b-12d3-a456-4266554400a0", sentRequest.Header.Get("apns-id"))
	require.Equal(t, "1715603200", sentRequest.Header.Get("apns-expiration"))
	require.Equal(t, "10", sentRequest.Header.Get("apns-priority"))
	require.Equal(t, "com.example.app", sentRequest.Header.Get("apns-topic"))
	require.Equal(t, "order-42", sentRequest.Header.Get("apns-collapse-id"))
}

func TestRequest_Validate(t *testing.T) {
	t.Run("missing device token", func(t *testing.T) {
		req := &Notification{
			Topic:   "com.example.app",
			Payload: []byte(`{"aps":{"alert":"hello"}}`),
		}

		err := req.Validate()
		require.EqualError(t, err, "apnskit: device token is required")
	})

	t.Run("missing topic", func(t *testing.T) {
		req := &Notification{
			DeviceToken: "abc123",
			Payload:     []byte(`{"aps":{"alert":"hello"}}`),
		}

		err := req.Validate()
		require.EqualError(t, err, "apnskit: topic is required")
	})

	t.Run("missing payload", func(t *testing.T) {
		req := &Notification{
			DeviceToken: "abc123",
			Topic:       "com.example.app",
		}

		err := req.Validate()
		require.EqualError(t, err, "apnskit: payload is required")
	})

	t.Run("invalid push type", func(t *testing.T) {
		req := &Notification{
			DeviceToken: "abc123",
			Topic:       "com.example.app",
			Payload:     []byte(`{"aps":{"alert":"hello"}}`),
			PushType:    PushType("invalid"),
		}

		err := req.Validate()
		require.EqualError(t, err, `apnskit: invalid push type "invalid"`)
	})

	t.Run("invalid apns id", func(t *testing.T) {
		req := &Notification{
			DeviceToken: "abc123",
			Topic:       "com.example.app",
			Payload:     []byte(`{"aps":{"alert":"hello"}}`),
			ID:          "123E4567-E89B-12D3-A456-4266554400A0",
		}

		err := req.Validate()
		require.EqualError(t, err, `apnskit: apns-id must be a canonical lowercase UUID: "123E4567-E89B-12D3-A456-4266554400A0"`)
	})

	t.Run("collapse id too long", func(t *testing.T) {
		req := &Notification{
			DeviceToken: "abc123",
			Topic:       "com.example.app",
			Payload:     []byte(`{"aps":{"alert":"hello"}}`),
			CollapseID:  strings.Repeat("a", 65),
		}

		err := req.Validate()
		require.EqualError(t, err, "apnskit: apns-collapse-id exceeds 64 bytes")
	})

	t.Run("invalid priority", func(t *testing.T) {
		priority := Priority(2)
		req := &Notification{
			DeviceToken: "abc123",
			Topic:       "com.example.app",
			Payload:     []byte(`{"aps":{"alert":"hello"}}`),
			Priority:    &priority,
		}

		err := req.Validate()
		require.EqualError(t, err, "apnskit: invalid priority 2")
	})

	t.Run("background priority must be 5", func(t *testing.T) {
		priority := PriorityImmediate
		req := &Notification{
			DeviceToken: "abc123",
			Topic:       "com.example.app",
			Payload:     []byte(`{"aps":{"content-available":1}}`),
			PushType:    PushTypeBackground,
			Priority:    &priority,
		}

		err := req.Validate()
		require.EqualError(t, err, "apnskit: background push type must use priority 5")
	})

	t.Run("regular payload too large", func(t *testing.T) {
		req := &Notification{
			DeviceToken: "abc123",
			Topic:       "com.example.app",
			Payload:     []byte(strings.Repeat("a", maximumPayloadLimit(PushTypeAlert)+1)),
		}

		err := req.Validate()
		require.EqualError(t, err, "apnskit: payload exceeds APNs limit of 4096 bytes")
	})

	t.Run("voip payload limit", func(t *testing.T) {
		req := &Notification{
			DeviceToken: "abc123",
			Topic:       "com.example.app.voip",
			Payload:     []byte(strings.Repeat("a", maximumPayloadLimit(PushTypeVoIP))),
			PushType:    PushTypeVoIP,
		}

		err := req.Validate()
		require.NoError(t, err)
	})
}
