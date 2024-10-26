package settings

type App struct {
	PageSize  int    `json:"page_size"`
	JwtSecret string `json:"jwt_secret"`
}

var AppSettings = &App{
	PageSize: 20,
}
