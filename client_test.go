package apnskit

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClient_Push(t *testing.T) {
	priority := PriorityImmediate

	var sentRequest *http.Request
	httpClient := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			sentRequest = req
			require.Equal(t, "https://api.push.apple.com/3/device/abc123", req.URL.String())
			require.Equal(t, "application/json", req.Header.Get("content-type"))
			require.Equal(t, "com.example.app", req.Header.Get("apns-topic"))
			require.Equal(t, "alert", req.Header.Get("apns-push-type"))
			require.Equal(t, "10", req.Header.Get("apns-priority"))

			body, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			require.Equal(t, `{"aps":{"alert":"hello"}}`, string(body))

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

	resp, err := client.Push(context.Background(), &Notification{
		DeviceToken: "abc123",
		Topic:       "com.example.app",
		Payload:     []byte(`{"aps":{"alert":"hello"}}`),
		PushType:    PushTypeAlert,
		Priority:    &priority,
	})
	require.NoError(t, err)
	require.NotNil(t, sentRequest)
	require.True(t, resp.Success())
	require.Equal(t, "123e4567-e89b-12d3-a456-4266554400a0", resp.APNSID)
}

func TestClient_PushErrorResponse(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Header: http.Header{
					"Apns-Id": []string{"123e4567-e89b-12d3-a456-4266554400a0"},
				},
				Body: io.NopCloser(strings.NewReader(`{"reason":"BadDeviceToken"}`)),
			}, nil
		}),
	}

	client, err := NewClient(WithHTTPClient(httpClient))
	require.NoError(t, err)

	resp, err := client.Push(context.Background(), &Notification{
		DeviceToken: "abc123",
		Topic:       "com.example.app",
		Payload:     []byte(`{"aps":{"alert":"hello"}}`),
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, ErrorReasonBadDeviceToken, resp.Error.Reason)
}
