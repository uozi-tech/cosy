package settings

import (
	"github.com/spf13/cast"
	"gopkg.in/ini.v1"
	"log"
	"strings"
	"time"
)

var (
	Conf         *ini.File
	ConfPath     string
	buildTime    string
	LastModified string
)

type section struct {
	Name string
	Ptr  any
}

var sections = []section{
	{
		Name: "app",
		Ptr:  AppSettings,
	},
	{
		Name: "server",
		Ptr:  ServerSettings,
	},
	{
		Name: "database",
		Ptr:  DataBaseSettings,
	},
}

// init the settings package
func init() {
	t := time.Unix(cast.ToInt64(buildTime), 0)
	LastModified = strings.ReplaceAll(t.Format(time.RFC1123), "UTC", "GMT")
}

// Register the setting, this should be called before Init
func Register(name string, ptr any) {
	sections = append(sections, section{name, ptr})
}

// Init the settings
func Init(confPath string) {
	ConfPath = confPath
	setup()
}

// Set up the settings
func setup() {
	var err error
	Conf, err = ini.Load(ConfPath)
	if err != nil {
		log.Fatalf("setting.init, fail to parse 'app.ini': %v", err)
	}

	for _, s := range sections {
		mapTo(s.Name, s.Ptr)
	}
}

// MapTo the settings
func mapTo(section string, v any) {
	err := Conf.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("setting.mapTo %s err: %v", section, err)
	}
}

// ReflectFrom the settings
func reflectFrom(section string, v any) {
	err := Conf.Section(section).ReflectFrom(v)
	if err != nil {
		log.Fatalf("Cfg.ReflectFrom %s err: %v", section, err)
	}
}

// Save the settings
func Save() (err error) {

	for _, s := range sections {
		reflectFrom(s.Name, s.Ptr)
	}

	err = Conf.SaveTo(ConfPath)
	if err != nil {
		return
	}
	setup()
	return
}
