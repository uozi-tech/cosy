package settings

import (
	"gopkg.in/ini.v1"
	"log"
	"reflect"
)

var (
	Conf     *ini.File
	ConfPath string
)

var sections = map[string]interface{}{
	"app":       AppSettings,
	"server":    ServerSettings,
	"database":  DataBaseSettings,
	"redis":     RedisSettings,
	"sonyflake": SonyflakeSettings,
}

// Register the setting, this should be called before Init
func Register(name string, ptr any) {
	sections[name] = ptr
}

// Init the settings
func Init(confPath string) {
	ConfPath = confPath
	setup()
}

// Load the settings
func load() (err error) {
	Conf, err = ini.LoadSources(ini.LoadOptions{
		Loose:        true,
		AllowShadows: true,
	}, ConfPath)

	return
}

// Reload the settings
func Reload() error {
	return load()
}

// Set up the settings
func setup() {
	err := load()
	if err != nil {
		log.Fatalf("setting.init, fail to parse 'app.ini': %v", err)
	}
	for name, ptr := range sections {
		err = MapTo(name, ptr)
		if err != nil {
			log.Fatalf("setting.MapTo %s err: %v", name, err)
		}
	}
}

// MapTo the settings
func MapTo(section string, v any) error {
	return Conf.Section(section).MapTo(v)
}

// ReflectFrom the settings
func ReflectFrom(section string, v any) {
	err := Conf.Section(section).ReflectFrom(v)
	if err != nil {
		log.Fatalf("Cfg.ReflectFrom %s err: %v", section, err)
	}
}

// ProtectedFill fill the target settings with new settings
func ProtectedFill(targetSettings interface{}, newSettings interface{}) {
	s := reflect.TypeOf(targetSettings).Elem()
	vt := reflect.ValueOf(targetSettings).Elem()
	vn := reflect.ValueOf(newSettings).Elem()

	// copy the values from new to target settings if it is not protected
	for i := 0; i < s.NumField(); i++ {
		if s.Field(i).Tag.Get("protected") != "true" {
			vt.Field(i).Set(vn.Field(i))
		}
	}
}

// Save the settings
func Save() (err error) {
	for name, ptr := range sections {
		ReflectFrom(name, ptr)
	}
	err = Conf.SaveTo(ConfPath)
	if err != nil {
		return
	}
	setup()
	return
}

// WithoutRedis remove the redis settings
func WithoutRedis() {
	delete(sections, "redis")
}

// WithoutSonyflake remove the sonyflake settings
func WithoutSonyflake() {
	delete(sections, "sonyflake")
}
