package apnskit

import (
	"net/http"
	"time"
)

type ErrorReason string

const (
	// ErrorReasonBadCollapseID means the collapse identifier exceeds the maximum allowed size.
	ErrorReasonBadCollapseID ErrorReason = "BadCollapseId"
	// ErrorReasonBadDeviceToken means the specified device token is invalid.
	ErrorReasonBadDeviceToken ErrorReason = "BadDeviceToken"
	// ErrorReasonBadExpirationDate means the apns-expiration value is invalid.
	ErrorReasonBadExpirationDate ErrorReason = "BadExpirationDate"
	// ErrorReasonBadMessageID means the apns-id value is invalid.
	ErrorReasonBadMessageID ErrorReason = "BadMessageId"
	// ErrorReasonBadPriority means the apns-priority value is invalid.
	ErrorReasonBadPriority ErrorReason = "BadPriority"
	// ErrorReasonBadTopic means the apns-topic value is invalid.
	ErrorReasonBadTopic ErrorReason = "BadTopic"
	// ErrorReasonDeviceTokenNotForTopic means the device token doesn't match the specified topic.
	ErrorReasonDeviceTokenNotForTopic ErrorReason = "DeviceTokenNotForTopic"
	// ErrorReasonDuplicateHeaders means one or more headers are repeated.
	ErrorReasonDuplicateHeaders ErrorReason = "DuplicateHeaders"
	// ErrorReasonIdleTimeout means APNs closed the request because of idle timeout.
	ErrorReasonIdleTimeout ErrorReason = "IdleTimeout"
	// ErrorReasonInvalidPushType means the apns-push-type value is invalid.
	ErrorReasonInvalidPushType ErrorReason = "InvalidPushType"
	// ErrorReasonMissingDeviceToken means the device token isn't specified in the request path.
	ErrorReasonMissingDeviceToken ErrorReason = "MissingDeviceToken"
	// ErrorReasonMissingTopic means the apns-topic header isn't specified and is required.
	ErrorReasonMissingTopic ErrorReason = "MissingTopic"
	// ErrorReasonPayloadEmpty means the message payload is empty.
	ErrorReasonPayloadEmpty ErrorReason = "PayloadEmpty"
	// ErrorReasonTopicDisallowed means pushing to this topic is not allowed.
	ErrorReasonTopicDisallowed ErrorReason = "TopicDisallowed"
	// ErrorReasonBadCertificate means the certificate is invalid.
	ErrorReasonBadCertificate ErrorReason = "BadCertificate"
	// ErrorReasonBadCertificateEnv means the client certificate doesn't match the environment.
	ErrorReasonBadCertificateEnv ErrorReason = "BadCertificateEnvironment"
	// ErrorReasonExpiredProviderToken means the provider token is stale and a new token should be generated.
	ErrorReasonExpiredProviderToken ErrorReason = "ExpiredProviderToken"
	// ErrorReasonForbidden means the specified action is not allowed.
	ErrorReasonForbidden ErrorReason = "Forbidden"
	// ErrorReasonInvalidProviderToken means the provider token is not valid, or the token signature can't be verified.
	ErrorReasonInvalidProviderToken ErrorReason = "InvalidProviderToken"
	// ErrorReasonMissingProviderToken means the authorization header is missing or no provider token is specified.
	ErrorReasonMissingProviderToken ErrorReason = "MissingProviderToken"
	// ErrorReasonUnrelatedKeyIDInToken means the key ID in the provider token isn't related to the key ID of the token used in the first push of this connection.
	ErrorReasonUnrelatedKeyIDInToken ErrorReason = "UnrelatedKeyIdInToken"
	// ErrorReasonBadEnvironmentKeyIDToken means the key ID in the provider token doesn't match the environment.
	ErrorReasonBadEnvironmentKeyIDToken ErrorReason = "BadEnvironmentKeyIdInToken"
	// ErrorReasonBadPath means the request contained an invalid path value.
	ErrorReasonBadPath ErrorReason = "BadPath"
	// ErrorReasonMethodNotAllowed means the specified method value isn't POST.
	ErrorReasonMethodNotAllowed ErrorReason = "MethodNotAllowed"
	// ErrorReasonExpiredToken means the device token has expired.
	ErrorReasonExpiredToken ErrorReason = "ExpiredToken"
	// ErrorReasonUnregistered means the device token is inactive for the specified topic.
	ErrorReasonUnregistered ErrorReason = "Unregistered"
	// ErrorReasonPayloadTooLarge means the message payload is too large.
	ErrorReasonPayloadTooLarge ErrorReason = "PayloadTooLarge"
	// ErrorReasonTooManyProviderTokenUpdates means the provider's authentication token is being updated too often.
	ErrorReasonTooManyProviderTokenUpdates ErrorReason = "TooManyProviderTokenUpdates"
	// ErrorReasonTooManyRequests means too many requests were made consecutively to the same device token.
	ErrorReasonTooManyRequests ErrorReason = "TooManyRequests"
	// ErrorReasonInternalServerError means an internal server error occurred.
	ErrorReasonInternalServerError ErrorReason = "InternalServerError"
	// ErrorReasonServiceUnavailable means the service is unavailable.
	ErrorReasonServiceUnavailable ErrorReason = "ServiceUnavailable"
	// ErrorReasonShutdown means the APNs server is shutting down.
	ErrorReasonShutdown ErrorReason = "Shutdown"
)

// Response is an APNs notification response.
// Details: https://developer.apple.com/documentation/usernotifications/handling-notification-responses-from-apns
type Response struct {
	APNSID     string
	UniqueID   string
	StatusCode int
	Error      *ResponseError
}

type ResponseError struct {
	Reason    ErrorReason
	Timestamp *time.Time
}

func (r *Response) Success() bool {
	return r != nil && r.StatusCode == http.StatusOK && r.Error == nil
}

func (e *ResponseError) Retryable() bool {
	if e == nil {
		return false
	}

	return e.Reason.Retryable()
}

func (r ErrorReason) Retryable() bool {
	switch r {
	case ErrorReasonTooManyRequests,
		ErrorReasonTooManyProviderTokenUpdates,
		ErrorReasonInternalServerError,
		ErrorReasonServiceUnavailable,
		ErrorReasonShutdown:
		return true
	default:
		return false
	}
}
