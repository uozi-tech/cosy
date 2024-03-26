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

		for i := 0; i < m.NumField(); i++ {
			field := m.Field(i)
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

		resolvedModelMap[r.Name] = r
	}
}

func GetResolvedModel[T any]() *ResolvedModel {
	name := reflect.TypeFor[T]().Name()
	return resolvedModelMap[name]
}
