package apnskit

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	// maxCollapseIDBytes is the maximum allowed size of the collapse identifier.
	maxCollapseIDBytes = 64
)

var canonicalUUIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type PushType string

const (
	// PushTypeAlert is the push type for notifications that trigger a user interaction.
	PushTypeAlert PushType = "alert"

	// PushTypeBackground is the push type for notifications that deliver content in the background.
	PushTypeBackground PushType = "background"

	// PushTypeComplication is the push type for notifications that contain update information for a watchOS app's complications.
	PushTypeComplication PushType = "complication"

	// PushTypeControls is the push type to reload and update a control.
	PushTypeControls PushType = "controls"

	// PushTypeFileProvider is the push type to signal changes to a File Provider extension.
	PushTypeFileProvider PushType = "fileprovider"

	// PushTypeLiveActivity is the push type to signal changes to a Live Activity.
	PushTypeLiveActivity PushType = "liveactivity"

	// PushTypeLocation is the push type for notifications that request a user's location.
	PushTypeLocation PushType = "location"

	// PushTypeMDM is the push type for notifications that tell managed devices to contact the MDM server.
	PushTypeMDM PushType = "mdm"

	// PushTypePushToTalk is the push type for notifications that provide information about updates to your application's push to talk services.
	PushTypePushToTalk PushType = "pushtotalk"

	// PushTypeVoIP is the push type for notifications that provide information about an incoming Voice-over-IP call.
	PushTypeVoIP PushType = "voip"

	// PushTypeWidgets is the push type to signal changes to widgets and watch complications you create with WidgetKit.
	PushTypeWidgets PushType = "widgets"
)

func (p PushType) isValid() bool {
	switch p {
	case PushTypeAlert, PushTypeBackground, PushTypeComplication,
		PushTypeControls, PushTypeFileProvider, PushTypeLiveActivity,
		PushTypeLocation, PushTypeMDM, PushTypePushToTalk,
		PushTypeVoIP, PushTypeWidgets:
		return true
	default:
		return false
	}
}

func maximumPayloadLimit(pt PushType) int {
	if pt == PushTypeVoIP {
		return 5 * 1024
	}
	return 4 * 1024
}

type Priority int

const (
	// PriorityConservePower prioritizes the device's power considerations over all other factors for delivery, and prevents awakening the device.
	PriorityConservePower Priority = 1

	// PriorityPowerAware sends the notification based on power considerations on the user's device.
	PriorityPowerAware Priority = 5

	// PriorityImmediate sends the notification immediately.
	PriorityImmediate Priority = 10
)

func (p Priority) isValid() bool {
	switch p {
	case PriorityConservePower, PriorityPowerAware, PriorityImmediate:
		return true
	default:
		return false
	}
}

// Notification is an APNs notification request.
// Details: https://developer.apple.com/documentation/usernotifications/sending-notification-requests-to-apns
type Notification struct {
	// DeviceToken: path parameter <required>
	DeviceToken string

	// Payload: JSON payload body <required>
	Payload []byte

	// Authorization: authorization <optional> but required for token-based authentication.
	Authorization string

	// PushType: apns-push-type <required>
	PushType PushType

	// ID: apns-id <optional>
	ID string // <optional>

	// Expiration: apns-expiration <optional>
	Expiration *int64

	// Priority: apns-priority <optional>
	Priority *Priority

	// Topic: apns-topic <required>
	Topic string

	// CollapseID: apns-collapse-id <optional>
	CollapseID string
}

func (r *Notification) Validate() error {
	if strings.TrimSpace(r.DeviceToken) == "" {
		return errors.New("apnskit: device token is required")
	}

	if strings.TrimSpace(r.Topic) == "" {
		return errors.New("apnskit: topic is required")
	}

	if len(r.Payload) == 0 {
		return errors.New("apnskit: payload is required")
	}

	if limit := maximumPayloadLimit(r.PushType); len(r.Payload) > limit {
		return fmt.Errorf("apnskit: payload exceeds APNs limit of %d bytes", limit)
	}

	if r.PushType != "" && !r.PushType.isValid() {
		return fmt.Errorf("apnskit: invalid push type %q", r.PushType)
	}

	if r.ID != "" && !canonicalUUIDPattern.MatchString(r.ID) {
		return fmt.Errorf("apnskit: apns-id must be a canonical lowercase UUID: %q", r.ID)
	}

	if len(r.CollapseID) > 0 && len([]byte(r.CollapseID)) > maxCollapseIDBytes {
		return fmt.Errorf("apnskit: apns-collapse-id exceeds %d bytes", maxCollapseIDBytes)
	}

	if r.Priority != nil && !r.Priority.isValid() {
		return fmt.Errorf("apnskit: invalid priority %d", *r.Priority)
	}

	if r.PushType == PushTypeBackground && r.Priority != nil && *r.Priority != PriorityPowerAware {
		return errors.New("apnskit: background push type must use priority 5")
	}

	return nil
}
