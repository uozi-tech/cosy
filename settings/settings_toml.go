//go:build toml_settings

package settings

import (
	"log"
	"os"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/elliotchance/orderedmap/v3"
)

var (
	ConfPath string
)

var sections = orderedmap.NewOrderedMap[string, any]()

func init() {
	sections.Set("app", AppSettings)
	sections.Set("server", ServerSettings)
	sections.Set("database", DataBaseSettings)
	sections.Set("redis", RedisSettings)
	sections.Set("sonyflake", SonyflakeSettings)
	sections.Set("log", LogSettings)
	sections.Set("sls", SLSSettings)
}

// Register the setting, this should be called before Init
func Register(name string, ptr any) {
	sections.Set(name, ptr)
}

// Init the settings
func Init(confPath string) {
	ConfPath = confPath
	setup()
	parseAllEnv()
}

// Load the settings
func load() (err error) {
	data, err := os.ReadFile(ConfPath)
	if err != nil {
		return err
	}

	// Create a map to temporarily hold the parsed data
	tmpConfig := make(map[string]interface{})
	err = toml.Unmarshal(data, &tmpConfig)
	if err != nil {
		return err
	}

	// Map each section to its corresponding struct
	for name, sectionData := range tmpConfig {
		ptr, ok := sections.Get(name)
		if ok {
			// Convert section data to TOML again
			bytes, err := toml.Marshal(sectionData)
			if err != nil {
				return err
			}

			// Unmarshal into the target struct
			err = toml.Unmarshal(bytes, ptr)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Reload the settings
func Reload() error {
	return load()
}

// Set up the settings
func setup() {
	err := load()
	if err != nil {
		log.Fatalf("setting.init, fail to parse TOML file: %v", err)
	}
}

// MapTo the settings (kept for backward compatibility)
func MapTo(section string, v any) error {
	// No need to do anything as load() already mapped to the structs
	return nil
}

// ReflectFrom the settings (kept for backward compatibility)
func ReflectFrom(section string, v any) {
	// Nothing to do as we're directly modifying the struct pointers
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
	// Create a map to hold all sections for saving
	configToSave := make(map[string]interface{})

	for name, ptr := range sections.AllFromFront() {
		configToSave[name] = ptr
	}

	f, err := os.Create(ConfPath)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	err = encoder.Encode(configToSave)
	if err != nil {
		return err
	}

	return nil
}

// WithoutRedis remove the redis settings
func WithoutRedis() {
	sections.Delete("redis")
}

// WithoutSonyflake remove the sonyflake settings
func WithoutSonyflake() {
	sections.Delete("sonyflake")
}
