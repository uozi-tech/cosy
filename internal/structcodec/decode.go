package structcodec

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

// Decode weakly decodes input into output using a cached plan for struct
// targets. Output must be a non-nil pointer.
func Decode(input, output any) error {
	value := reflect.ValueOf(output)
	if value.Kind() != reflect.Pointer {
		return errors.New("result must be a pointer")
	}
	if value.IsNil() {
		return errors.New("result must be addressable (a pointer)")
	}

	target := value.Elem()
	if input == nil || isNilPointer(input) {
		return nil
	}
	inputValue := reflect.ValueOf(input)
	if inputValue.Type().AssignableTo(target.Type()) {
		target.Set(inputValue)
		return nil
	}
	if target.Kind() != reflect.Struct {
		copy := cloneValue(target)
		if err := decodeReflect(input, copy, ""); err != nil {
			return err
		}
		target.Set(copy)
		return nil
	}

	plan, err := getPlan(target.Type())
	if err != nil {
		return err
	}
	copy := cloneValue(target)
	if err := plan.decode(input, unsafe.Pointer(copy.Addr().Pointer()), ""); err != nil {
		return err
	}
	target.Set(copy)
	return nil
}

func cloneValue(source reflect.Value) reflect.Value {
	clone := reflect.New(source.Type()).Elem()
	clone.Set(source)
	return clone
}

func (plan *decodePlan) decode(input any, dst unsafe.Pointer, path string) error {
	view, err := newMapView(input)
	if err != nil {
		return namedExpectedMap(path, input)
	}

	var decodeErrors []string
	for i := range plan.fields {
		field := &plan.fields[i]
		inputValue, ok := view.lookup(field.name)
		if !ok {
			continue
		}
		fieldPath := joinPath(path, field.name)
		address := field.address(dst)
		if err := field.decode(address, inputValue, fieldPath); err != nil {
			decodeErrors = appendDecodeErrors(decodeErrors, err)
		}
	}
	return combineErrors(decodeErrors)
}

func (field *fieldPlan) address(base unsafe.Pointer) unsafe.Pointer {
	address := base
	for _, step := range field.steps {
		address = unsafe.Add(address, step.offset)
		if step.ptrType != nil {
			pointerValue := reflect.NewAt(step.ptrType, address).Elem()
			if pointerValue.IsNil() {
				pointerValue.Set(reflect.New(step.ptrType.Elem()))
			} else {
				copy := reflect.New(step.ptrType.Elem())
				copy.Elem().Set(cloneValue(pointerValue.Elem()))
				pointerValue.Set(copy)
			}
			address = unsafe.Pointer(pointerValue.Pointer())
		}
	}
	return address
}

type mapView struct {
	exact map[string]any
	lower map[string]any
}

func newMapView(input any) (*mapView, error) {
	value := reflect.ValueOf(input)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return &mapView{exact: map[string]any{}, lower: map[string]any{}}, nil
		}
		value = value.Elem()
	}
	if value.Kind() == reflect.Struct {
		return newStructView(value), nil
	}
	if value.Kind() != reflect.Map {
		return nil, fmt.Errorf("not a map")
	}
	if value.Type().Key().Kind() != reflect.String && value.Type().Key().Kind() != reflect.Interface {
		return nil, fmt.Errorf("map key is not a string")
	}

	view := &mapView{
		exact: make(map[string]any, value.Len()),
		lower: make(map[string]any, value.Len()),
	}
	iterator := value.MapRange()
	for iterator.Next() {
		key, ok := iterator.Key().Interface().(string)
		if !ok {
			continue
		}
		item := iterator.Value()
		if item.Kind() == reflect.Interface && !item.IsNil() {
			item = item.Elem()
		}
		var decoded any
		if item.IsValid() {
			decoded = item.Interface()
		}
		view.exact[key] = decoded
		lowerKey := strings.ToLower(key)
		if _, exists := view.lower[lowerKey]; !exists {
			view.lower[lowerKey] = decoded
		}
	}
	return view, nil
}

func newStructView(value reflect.Value) *mapView {
	view := &mapView{
		exact: make(map[string]any, value.NumField()),
		lower: make(map[string]any, value.NumField()),
	}
	appendStructFields(view, value)
	return view
}

func appendStructFields(view *mapView, value reflect.Value) {
	typ := value.Type()
	for i := 0; i < value.NumField(); i++ {
		fieldType := typ.Field(i)
		fieldValue := value.Field(i)
		if fieldType.PkgPath != "" {
			continue
		}
		tag := fieldType.Tag.Get("json")
		parts := strings.Split(tag, ",")
		if len(parts) > 0 && parts[0] == "-" {
			continue
		}
		squash := fieldType.Anonymous
		for _, option := range parts[1:] {
			if option == "squash" {
				squash = true
			}
		}
		if squash {
			for fieldValue.Kind() == reflect.Pointer && !fieldValue.IsNil() {
				fieldValue = fieldValue.Elem()
			}
			if fieldValue.Kind() == reflect.Struct {
				appendStructFields(view, fieldValue)
				continue
			}
		}

		name := fieldType.Name
		if len(parts) > 0 && parts[0] != "" {
			name = parts[0]
		}
		decoded := fieldValue.Interface()
		view.exact[name] = decoded
		lowerName := strings.ToLower(name)
		if _, exists := view.lower[lowerName]; !exists {
			view.lower[lowerName] = decoded
		}
	}
}

func (view *mapView) lookup(name string) (any, bool) {
	if value, ok := view.exact[name]; ok {
		return value, true
	}
	value, ok := view.lower[strings.ToLower(name)]
	return value, ok
}

func isNilPointer(input any) bool {
	value := reflect.ValueOf(input)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func joinPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}

func namedExpectedMap(path string, input any) error {
	value := reflect.ValueOf(input)
	if value.Kind() == reflect.Map {
		return fmt.Errorf("'%s' needs a map with string keys, has '%s' keys", path, value.Type().Key().Kind())
	}
	return fmt.Errorf("'%s' expected a map, got '%s'", path, value.Kind())
}

func appendDecodeErrors(target []string, err error) []string {
	var aggregate *decodeError
	if errors.As(err, &aggregate) {
		return append(target, aggregate.errors...)
	}
	return append(target, err.Error())
}

func combineErrors(items []string) error {
	if len(items) == 0 {
		return nil
	}
	return &decodeError{errors: items}
}

type decodeError struct {
	errors []string
}

func (err *decodeError) Error() string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "%d error(s) decoding:\n\n", len(err.errors))
	for _, item := range err.errors {
		builder.WriteString("* ")
		builder.WriteString(item)
		builder.WriteByte('\n')
	}
	return strings.TrimSuffix(builder.String(), "\n")
}
