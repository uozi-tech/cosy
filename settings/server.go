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
	// PayloadMaxBytes limits the size of JSON request bodies accepted by the
	// CRUD pipeline and BindAndValid. 0 uses the built-in default (10 MiB),
	// a negative value disables the limit.
	PayloadMaxBytes int64 `json:"payload_max_bytes"`
}

var ServerSettings = &Server{
	RunMode: "debug",
}
