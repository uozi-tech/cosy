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
	if inputValue.Type().AssignableTo(target.Type()) && !isContainerKind(target.Kind()) {
		target.Set(inputValue)
		return nil
	}
	// Decode into a copy so a failed decode never leaves a half-written
	// target (invariant I4). decodeReflect resolves registered converters and
	// built-in special types before falling back to the compiled struct plan.
	copy := cloneValue(target)
	if err := decodeReflect(input, copy, ""); err != nil {
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

// matchQuality records how an input key was matched to a plan field; an exact
// match always wins over a case-insensitive one.
type matchQuality uint8

const (
	matchNone matchQuality = iota
	matchFolded
	matchExact
)

type fieldInput struct {
	value   any
	quality matchQuality
}

// scratchFields is the number of fields a decode handles without allocating
// its per-decode input table.
const scratchFields = 32

func (plan *decodePlan) decode(input any, dst unsafe.Pointer, path string) error {
	source, ok := asStringMap(input)
	if !ok {
		return plan.decodeView(input, dst, path)
	}

	var scratch [scratchFields]fieldInput
	var inputs []fieldInput
	if len(plan.fields) <= scratchFields {
		inputs = scratch[:len(plan.fields)]
	} else {
		inputs = make([]fieldInput, len(plan.fields))
	}

	// Walk the input once: exact name first, lower-cased name second. This
	// replaces the two map copies newMapView used to build per decode.
	for key, value := range source {
		if index, found := plan.index[key]; found {
			for ; index >= 0; index = plan.fields[index].next {
				inputs[index] = fieldInput{value: value, quality: matchExact}
			}
			continue
		}
		index, found := plan.lowerIndex[strings.ToLower(key)]
		if !found {
			continue
		}
		for ; index >= 0; index = plan.fields[index].next {
			if inputs[index].quality != matchExact {
				inputs[index] = fieldInput{value: value, quality: matchFolded}
			}
		}
	}

	var decodeErrors []string
	var cloned []unsafe.Pointer
	for i := range plan.fields {
		if inputs[i].quality == matchNone {
			continue
		}
		field := &plan.fields[i]
		address := field.address(dst, &cloned)
		if err := field.decode(address, inputs[i].value, joinPath(path, field.name)); err != nil {
			decodeErrors = appendDecodeErrors(decodeErrors, err)
		}
	}
	return combineErrors(decodeErrors)
}

// decodeView handles inputs that are not a map[string]any: structs (and
// pointers to them) and maps with other value types.
func (plan *decodePlan) decodeView(input any, dst unsafe.Pointer, path string) error {
	view, err := newMapView(input)
	if err != nil {
		return namedExpectedMap(path, input)
	}

	var decodeErrors []string
	var cloned []unsafe.Pointer
	for i := range plan.fields {
		field := &plan.fields[i]
		inputValue, ok := view.lookup(field.name)
		if !ok {
			continue
		}
		address := field.address(dst, &cloned)
		if err := field.decode(address, inputValue, joinPath(path, field.name)); err != nil {
			decodeErrors = appendDecodeErrors(decodeErrors, err)
		}
	}
	return combineErrors(decodeErrors)
}

var (
	stringMapType = reflect.TypeFor[map[string]any]()
)

// asStringMap returns the input as a map[string]any when it is one (or a named
// type such as gin.H, or a pointer to either) without copying it.
func asStringMap(input any) (map[string]any, bool) {
	switch source := input.(type) {
	case map[string]any:
		return source, true
	case nil:
		return nil, false
	}
	value := reflect.ValueOf(input)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, false
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Map || !value.Type().ConvertibleTo(stringMapType) {
		return nil, false
	}
	return value.Convert(stringMapType).Interface().(map[string]any), true
}

// address resolves the field's storage, allocating or copy-on-write cloning
// embedded pointer structs on the way. cloned remembers the pointer slots
// already handled during this decode so each embedded struct is cloned once
// rather than once per promoted field.
func (field *fieldPlan) address(base unsafe.Pointer, cloned *[]unsafe.Pointer) unsafe.Pointer {
	address := base
	for _, step := range field.steps {
		address = unsafe.Add(address, step.offset)
		if step.ptrType != nil {
			pointerValue := reflect.NewAt(step.ptrType, address).Elem()
			if !alreadyCloned(*cloned, address) {
				if pointerValue.IsNil() {
					pointerValue.Set(reflect.New(step.ptrType.Elem()))
				} else {
					copy := reflect.New(step.ptrType.Elem())
					copy.Elem().Set(pointerValue.Elem())
					pointerValue.Set(copy)
				}
				*cloned = append(*cloned, address)
			}
			address = unsafe.Pointer(pointerValue.Pointer())
		}
	}
	return address
}

func alreadyCloned(cloned []unsafe.Pointer, slot unsafe.Pointer) bool {
	for _, done := range cloned {
		if done == slot {
			return true
		}
	}
	return false
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
