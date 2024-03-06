package settings

type DataBase struct {
	Host     string `json:"host"`
	Port     uint   `json:"port"`
	User     string `json:"user"`
	Password string `json:"-,omitempty"`
	Name     string `json:"name"`
}

var DataBaseSettings = &DataBase{}
