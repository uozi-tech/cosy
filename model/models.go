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

type resolvedModelField struct {
	Name    string
	Type    string
	JsonTag string
	CosyTag CosyTag
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
		if field.Type.Kind() == reflect.Struct && field.Anonymous {
			deepResolve(r, field.Type)
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
