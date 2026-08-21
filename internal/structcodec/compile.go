package structcodec

import (
	"fmt"
	"reflect"
	"strings"
)

func compilePlan(typ reflect.Type) (*decodePlan, error) {
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("structcodec: cannot compile %s", typ)
	}

	plan := &decodePlan{typ: typ}
	walkFields(typ, nil, &plan.fields, map[reflect.Type]bool{typ: true})
	return plan, nil
}

// walkFields flattens typ into the plan. ancestors holds the struct types
// currently being expanded: an embedded pointer to one of them (a recursive
// type such as `type Node struct{ *Node }`) is compiled as a plain named
// field instead of being expanded again, so compilation always terminates.
func walkFields(typ reflect.Type, prefix []addressStep, fields *[]fieldPlan, ancestors map[reflect.Type]bool) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("json")
		parts := strings.Split(tag, ",")
		name := field.Name
		if len(parts) > 0 && parts[0] != "" {
			name = parts[0]
		}

		squash := field.Anonymous
		for _, option := range parts[1:] {
			if option == "squash" {
				squash = true
			}
		}

		steps := appendSteps(prefix, addressStep{offset: field.Offset})
		fieldType := field.Type
		if squash && fieldType.Kind() == reflect.Struct && !ancestors[fieldType] {
			ancestors[fieldType] = true
			walkFields(fieldType, steps, fields, ancestors)
			delete(ancestors, fieldType)
			continue
		}
		if squash && fieldType.Kind() == reflect.Pointer && fieldType.Elem().Kind() == reflect.Struct && !ancestors[fieldType.Elem()] {
			steps[len(steps)-1].ptrType = fieldType
			ancestors[fieldType.Elem()] = true
			walkFields(fieldType.Elem(), steps, fields, ancestors)
			delete(ancestors, fieldType.Elem())
			continue
		}

		if field.PkgPath != "" {
			continue
		}
		*fields = append(*fields, fieldPlan{
			name:   name,
			steps:  steps,
			decode: compileValueDecoder(fieldType),
		})
	}
}

func appendSteps(prefix []addressStep, step addressStep) []addressStep {
	steps := make([]addressStep, len(prefix)+1)
	copy(steps, prefix)
	steps[len(prefix)] = step
	return steps
}
