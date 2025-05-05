package model

import (
	"reflect"
	"strings"
	"sync"
)

var collection []any

// GenerateAllModel generate all models
func GenerateAllModel() []any {
	return collection
}

// RegisterModels register models
func RegisterModels(models ...any) {
	collection = append(collection, models...)
}

// ClearCollection clear collection for testing purpose
func ClearCollection() {
	collection = make([]any, 0)
}

type ResolvedModelField struct {
	Name         string
	Type         string
	JsonTag      string
	CosyTag      CosyTag
	Unique       bool
	DefaultValue string
}

type ResolvedModel struct {
	Name          string
	Fields        map[string]*ResolvedModelField
	OrderedFields []*ResolvedModelField
}

var (
	resolvedModelMap = make(map[string]*ResolvedModel)
	mu               sync.RWMutex
)

func deepResolve(r *ResolvedModel, m reflect.Type) {
	for i := 0; i < m.NumField(); i++ {
		field := m.Field(i)
		fieldType := field.Type

		// Check if the field is a pointer to a struct
		if fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct {
			// If it is, we want to resolve the struct it points to
			fieldType = fieldType.Elem()
		}

		// Continue with the existing logic for anonymous structs
		if fieldType.Kind() == reflect.Struct && field.Anonymous {
			deepResolve(r, fieldType)
			continue
		}

		jsonTag := field.Tag.Get("json")
		jsonTags := strings.Split(jsonTag, ",")
		if len(jsonTags) > 0 {
			jsonTag = jsonTags[0]
		} else {
			jsonTag = ""
		}

		resolvedField := &ResolvedModelField{
			Name:    field.Name,
			Type:    field.Type.String(),
			JsonTag: jsonTag,
			CosyTag: NewCosyTag(field.Tag.Get("cosy")),
		}

		gormTags := field.Tag.Get("gorm")

		if gormTags != "" {
			tags := strings.Split(gormTags, ";")
			for _, tag := range tags {
				// gorm:"uniqueIndex;type:varchar(255);default:0"
				if strings.Contains(tag, "default") {
					defaultValueTag := strings.Split(tag, ":")
					if len(defaultValueTag) != 2 {
						continue
					}
					resolvedField.DefaultValue = defaultValueTag[1]
				}
				if strings.Contains(tag, "unique") {
					resolvedField.Unique = true
				}
			}
		}

		// out-of-order, CamelCase as the key
		r.Fields[field.Name] = resolvedField
		// out-of-order, jsonTagName as the key
		r.Fields[jsonTag] = resolvedField
		// sorted
		r.OrderedFields = append(r.OrderedFields, resolvedField)
	}
}

// ResolvedModels resolved meta of models
func ResolvedModels() {
	for _, model := range collection {
		// resolve model meta
		m := reflect.TypeOf(model)
		r := &ResolvedModel{
			Name:          m.Name(),
			Fields:        make(map[string]*ResolvedModelField),
			OrderedFields: make([]*ResolvedModelField, 0),
		}

		deepResolve(r, m)

		mu.Lock()
		resolvedModelMap[r.Name] = r
		mu.Unlock()
	}
}

// GetResolvedModel get resolved model from resolvedModelMap
func GetResolvedModel[T any]() *ResolvedModel {
	name := reflect.TypeFor[T]().Name()

	mu.RLock()
	defer mu.RUnlock()
	return resolvedModelMap[name]
}
