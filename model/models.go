package model

import (
	"reflect"
	"strings"
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

type resolvedModelField struct {
	Name    string
	Type    string
	JsonTag string
	CosyTag CosyTag
	Unique  bool
}

type ResolvedModel struct {
	Name          string
	Fields        map[string]*resolvedModelField
	OrderedFields []*resolvedModelField
}

var resolvedModelMap = make(map[string]*ResolvedModel)

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

		resolvedField := &resolvedModelField{
			Name:    field.Name,
			Type:    field.Type.String(),
			JsonTag: jsonTag,
			CosyTag: NewCosyTag(field.Tag.Get("cosy")),
		}

		if field.Tag.Get("gorm") != "" {
			if strings.Contains(field.Tag.Get("gorm"), "unique") {
				resolvedField.Unique = true
			}
		}

		// out-of-order
		r.Fields[field.Name] = resolvedField
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
			Fields:        make(map[string]*resolvedModelField),
			OrderedFields: make([]*resolvedModelField, 0),
		}

		deepResolve(r, m)

		resolvedModelMap[r.Name] = r
	}
}

func GetResolvedModel[T any]() *ResolvedModel {
	name := reflect.TypeFor[T]().Name()
	return resolvedModelMap[name]
}
