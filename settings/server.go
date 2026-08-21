package settings

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
}

var ServerSettings = &Server{
	RunMode: "debug",
}

// DefaultPayloadMaxBytes is the request-body cap applied when
// Server.PayloadMaxBytes is 0.
const DefaultPayloadMaxBytes int64 = 10 << 20 // 10 MiB

// PayloadLimit returns the effective request-body cap in bytes: PayloadMaxBytes
// when set, DefaultPayloadMaxBytes when 0, and a negative value (no limit)
// when PayloadMaxBytes is negative.
func (s *Server) PayloadLimit() int64 {
	if s.PayloadMaxBytes != 0 {
		return s.PayloadMaxBytes
	}
	return DefaultPayloadMaxBytes
}
