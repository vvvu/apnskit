package apnskit

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_PushNotification_SuccessResponse(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Apns-Id":        []string{"123e4567-e89b-12d3-a456-4266554400a0"},
					"Apns-Unique-Id": []string{"unique-id"},
				},
				Body: io.NopCloser(strings.NewReader("")),
			}, nil
		}),
	}

	client, err := NewClient(WithHTTPClient(httpClient))
	require.NoError(t, err)

	result, err := client.pushNotification(context.Background(), &Notification{
		DeviceToken: "abc123",
		Topic:       "com.example.app",
		Payload:     []byte(`{"aps":{"alert":"hello"}}`),
	})
	require.NoError(t, err)
	require.Equal(t, "123e4567-e89b-12d3-a456-4266554400a0", result.APNSID)
	require.Equal(t, "unique-id", result.UniqueID)
	require.Equal(t, http.StatusOK, result.StatusCode)
	require.True(t, result.Success())
	require.Nil(t, result.Error)
}

func TestClient_PushNotification_ErrorResponse(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusGone,
				Header: http.Header{
					"Apns-Id": []string{"123e4567-e89b-12d3-a456-4266554400a0"},
				},
				Body: io.NopCloser(strings.NewReader(`{"reason":"Unregistered","timestamp":1715603200000}`)),
			}, nil
		}),
	}

	client, err := NewClient(WithHTTPClient(httpClient))
	require.NoError(t, err)

	result, err := client.pushNotification(context.Background(), &Notification{
		DeviceToken: "abc123",
		Topic:       "com.example.app",
		Payload:     []byte(`{"aps":{"alert":"hello"}}`),
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusGone, result.StatusCode)
	require.Equal(t, ErrorReasonUnregistered, result.Error.Reason)
	require.False(t, result.Success())

	require.NotNil(t, result.Error.Timestamp)
	require.Equal(t, time.UnixMilli(1715603200000).UTC(), *result.Error.Timestamp)
}

func TestResponseError_Retryable(t *testing.T) {
	require.True(t, (&ResponseError{Reason: ErrorReasonTooManyRequests}).Retryable())
	require.True(t, ErrorReasonInternalServerError.Retryable())
	require.False(t, ErrorReasonUnregistered.Retryable())
	require.False(t, (*ResponseError)(nil).Retryable())
}
