package settings

type Server struct {
	Host    string `json:"host"`
	Port    uint   `json:"port"`
	RunMode string `json:"run_mode"`
	BaseUrl string `json:"base_url"`
}

var ServerSettings = &Server{
	RunMode: "debug",
}
