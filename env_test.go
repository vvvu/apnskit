package apnskit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvironment_Host_Production(t *testing.T) {
	require.Equal(t, "https://api.push.apple.com", Production.Host())
}

func TestEnvironment_Host_Sandbox(t *testing.T) {
	require.Equal(t, "https://api.sandbox.push.apple.com", Sandbox.Host())
}

func TestEnvironment_Host_Unknown(t *testing.T) {
	require.Panics(t, func() {
		_ = Environment("staging").Host()
	})
}
