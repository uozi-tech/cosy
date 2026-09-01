package settings

import "time"

type Server struct {
	Host        string `json:"host"`
	Port        uint   `json:"port"`
	RunMode     string `json:"run_mode"`
	BaseUrl     string `json:"base_url"`
	EnableHTTPS bool   `json:"enable_https"`
	SSLCert     string `json:"ssl_cert"`
	SSLKey      string `json:"ssl_key"`
	EnableH2    bool   `json:"enable_h2"`
	EnableH3    bool   `json:"enable_h3"`
	// PayloadMaxBytes limits the size of request bodies; the router applies it
	// to every route and the CRUD pipeline / BindAndValid enforce it again.
	// 0 uses the built-in default (10 MiB), a negative value disables the
	// limit.
	PayloadMaxBytes int64 `json:"payload_max_bytes"`
	// ShutdownTimeoutSeconds caps how long graceful shutdown waits for in-flight
	// requests after the interrupt signal, before the server is torn down under
	// them. 0 uses the built-in default (15s); so does a negative value, which
	// is not a way to wait forever.
	ShutdownTimeoutSeconds int `json:"shutdown_timeout_seconds"`
}

var ServerSettings = &Server{
	RunMode: "debug",
}

// DefaultPayloadMaxBytes is the request-body cap applied when
// Server.PayloadMaxBytes is 0.
const DefaultPayloadMaxBytes int64 = 10 << 20 // 10 MiB

// DefaultShutdownTimeout is the graceful-shutdown budget applied when
// Server.ShutdownTimeoutSeconds is 0. It has to stay well under the 30s
// terminationGracePeriodSeconds Kubernetes defaults to, because a preStop hook
// spends part of that budget before SIGTERM is even delivered; anything the
// grace period does not cover is taken by SIGKILL instead.
const DefaultShutdownTimeout = 15 * time.Second

// PayloadLimit returns the effective request-body cap in bytes: PayloadMaxBytes
// when set, DefaultPayloadMaxBytes when 0, and a negative value (no limit)
// when PayloadMaxBytes is negative.
func (s *Server) PayloadLimit() int64 {
	if s.PayloadMaxBytes != 0 {
		return s.PayloadMaxBytes
	}
	return DefaultPayloadMaxBytes
}

// ShutdownTimeout returns the effective graceful-shutdown budget:
// ShutdownTimeoutSeconds when positive, DefaultShutdownTimeout otherwise.
// Negative deliberately does not mean "unbounded" the way it does for
// PayloadLimit: one hung stream would then keep Shutdown from ever returning,
// and only Kubernetes sends the SIGKILL that would end that wait. A deployment
// that needs a long drain says so with a large positive value.
func (s *Server) ShutdownTimeout() time.Duration {
	if s.ShutdownTimeoutSeconds > 0 {
		return time.Duration(s.ShutdownTimeoutSeconds) * time.Second
	}
	return DefaultShutdownTimeout
}
