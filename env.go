package apnskit

import "log/slog"

type Environment string

const (
	// Production is the APNs production environment.
	Production Environment = "production"
	// Sandbox is the APNs development environment.
	Sandbox Environment = "sandbox"
)

func (e Environment) Host() string {
	switch e {
	case Production:
		return "https://api.push.apple.com"
	case Sandbox:
		return "https://api.sandbox.push.apple.com"
	}

	slog.Error("Unexpected environment", slog.String("environemnt", string(e)))
	panic("unknown environment")
}
