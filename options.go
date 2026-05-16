package apnskit

import "net/http"

// Option is a functional option for configuring a Client.
type Option func(*option)

// option holds the configuration for creating a Client.
type option struct {
	httpClient *http.Client
	env        Environment
}

// WithHTTPClient sets a custom HTTP client.
// If not set, an HTTP/2-capable client is used by default.
func WithHTTPClient(client *http.Client) Option {
	return func(cfg *option) {
		cfg.httpClient = client
	}
}

// WithEnv sets the APNs environment.
// If not set, Sandbox is used by default.
func WithEnv(env Environment) Option {
	return func(cfg *option) {
		cfg.env = env
	}
}
