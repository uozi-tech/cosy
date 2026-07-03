//go:build json_settings && !toml_settings && !yaml_settings

package settings

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"reflect"
	"strings"
	"unicode"

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

	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}

	sectionMessages := make(map[string]json.RawMessage)
	if err = json.Unmarshal(data, &sectionMessages); err != nil {
		return err
	}

	for name, rawSectionMessage := range sectionMessages {
		ptr, ok := sections.Get(name)
		if !ok {
			continue
		}

		if err = decodeJSONSection(rawSectionMessage, ptr); err != nil {
			return err
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
		log.Fatalf("setting.init, fail to parse JSON file: %v", err)
	}
}

// MapTo the settings (kept for backward compatibility)
func MapTo(section string, v any) error {
	return nil
}

// ReflectFrom the settings (kept for backward compatibility)
func ReflectFrom(section string, v any) {
}

// ProtectedFill fill the target settings with new settings
func ProtectedFill(targetSettings any, newSettings any) {
	settingsMu.Lock()
	defer settingsMu.Unlock()

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
	settingsMu.Lock()
	defer settingsMu.Unlock()

	configToSave := make(map[string]any)

	for name, ptr := range sections.AllFromFront() {
		configToSave[name] = encodeJSONSection(ptr)
	}

	return writeAtomically(ConfPath, func(f *os.File) error {
		encoder := json.NewEncoder(f)
		encoder.SetIndent("", "  ")
		return encoder.Encode(configToSave)
	})
}

func decodeJSONSection(rawSectionMessage json.RawMessage, ptr any) error {
	ptrValue := reflect.ValueOf(ptr)
	if ptrValue.Kind() != reflect.Pointer || ptrValue.IsNil() {
		return json.Unmarshal(rawSectionMessage, ptr)
	}

	sectionValue := ptrValue.Elem()
	if sectionValue.Kind() != reflect.Struct {
		return json.Unmarshal(rawSectionMessage, ptr)
	}

	if err := json.Unmarshal(rawSectionMessage, ptr); err != nil {
		return err
	}

	sectionFields := make(map[string]json.RawMessage)
	if err := json.Unmarshal(rawSectionMessage, &sectionFields); err != nil {
		return err
	}

	sectionType := sectionValue.Type()
	for fieldIndex := 0; fieldIndex < sectionType.NumField(); fieldIndex++ {
		structField := sectionType.Field(fieldIndex)
		fieldValue := sectionValue.Field(fieldIndex)

		if !structField.IsExported() || !fieldValue.CanAddr() || !fieldValue.CanSet() {
			continue
		}

		rawFieldMessage, ok := findJSONFieldMessage(structField, sectionFields)
		if !ok {
			continue
		}

		if err := json.Unmarshal(rawFieldMessage, fieldValue.Addr().Interface()); err != nil {
			return err
		}
	}

	return nil
}

func encodeJSONSection(ptr any) any {
	ptrValue := reflect.ValueOf(ptr)
	if ptrValue.Kind() != reflect.Pointer || ptrValue.IsNil() {
		return ptr
	}

	sectionValue := ptrValue.Elem()
	if sectionValue.Kind() != reflect.Struct {
		return ptr
	}

	sectionType := sectionValue.Type()
	encodedSection := make(map[string]any)

	for fieldIndex := 0; fieldIndex < sectionType.NumField(); fieldIndex++ {
		structField := sectionType.Field(fieldIndex)
		fieldValue := sectionValue.Field(fieldIndex)

		if !structField.IsExported() || !fieldValue.CanInterface() {
			continue
		}

		encodedSection[structField.Name] = fieldValue.Interface()
	}

	return encodedSection
}

func findJSONFieldMessage(structField reflect.StructField, sectionFields map[string]json.RawMessage) (json.RawMessage, bool) {
	candidateFieldNames := getJSONFieldNameCandidates(structField)

	for _, candidateFieldName := range candidateFieldNames {
		if rawFieldMessage, ok := sectionFields[candidateFieldName]; ok {
			return rawFieldMessage, true
		}
	}

	for sectionFieldName, rawFieldMessage := range sectionFields {
		for _, candidateFieldName := range candidateFieldNames {
			if normalizeJSONFieldName(sectionFieldName) == normalizeJSONFieldName(candidateFieldName) {
				return rawFieldMessage, true
			}
		}
	}

	return nil, false
}

func getJSONFieldNameCandidates(structField reflect.StructField) []string {
	candidateFieldNames := []string{structField.Name}

	jsonTagName := strings.Split(structField.Tag.Get("json"), ",")[0]
	if jsonTagName != "" && jsonTagName != "-" {
		candidateFieldNames = append(candidateFieldNames, jsonTagName)
	}

	return candidateFieldNames
}

func normalizeJSONFieldName(name string) string {
	var normalizedName strings.Builder

	for _, nameRune := range name {
		if !unicode.IsLetter(nameRune) && !unicode.IsDigit(nameRune) {
			continue
		}

		normalizedName.WriteRune(unicode.ToLower(nameRune))
	}

	return normalizedName.String()
}

// WithoutRedis remove the redis settings
func WithoutRedis() {
	sections.Delete("redis")
}

// WithoutSonyflake remove the sonyflake settings
func WithoutSonyflake() {
	sections.Delete("sonyflake")
}
