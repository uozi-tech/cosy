package settings

type DataBase struct {
    Host        string `json:"host"`
    Port        uint   `json:"port"`
    User        string `json:"user"`
    Password    string `json:"-,omitempty"`
    Name        string `json:"name"`
    TablePrefix string `json:"table_prefix"`
}

var DataBaseSettings = &DataBase{}

func (d *DataBase) GetName() string {
    return d.Name
}

func (d *DataBase) GetHost() string {
    return d.Host
}

func (d *DataBase) GetPassword() string {
    return d.Password
}

func (d *DataBase) GetPort() uint {
    return d.Port
}

func (d *DataBase) GetUser() string {
    return d.User
}
