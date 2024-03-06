package settings

type Server struct {
	Host    string `json:"host"`
	Port    uint   `json:"port"`
	RunMode string `json:"run_mode"`
}

var ServerSettings = &Server{}
