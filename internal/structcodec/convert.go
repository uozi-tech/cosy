package structcodec

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/guregu/null/v6"
	"github.com/jackc/pgtype"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
)

var (
	timeType       = reflect.TypeFor[time.Time]()
	timePtrType    = reflect.TypeFor[*time.Time]()
	decimalType    = reflect.TypeFor[decimal.Decimal]()
	nullStringType = reflect.TypeFor[null.String]()
	pgDateType     = reflect.TypeFor[pgtype.Date]()
	pgDatePtrType  = reflect.TypeFor[*pgtype.Date]()
)

func compileValueDecoder(typ reflect.Type) valueDecoder {
	if converter := registeredConverter(typ); converter != nil {
		return converterDecoder(typ, converter)
	}

	if decoder := builtinSpecialDecoder(typ); decoder != nil {
		return decoder
	}

	generic := func(dst unsafe.Pointer, input any, path string) error {
		return decodeReflect(input, reflect.NewAt(typ, dst).Elem(), path)
	}

	// Fast paths for the input shapes JSON produces (string, float64, bool,
	// object). Anything else takes the generic reflective conversion, so the
	// semantics stay identical; only the common case skips reflection.
	switch typ.Kind() {
	case reflect.String:
		return func(dst unsafe.Pointer, input any, path string) error {
			if text, ok := input.(string); ok {
				*(*string)(dst) = text
				return nil
			}
			return generic(dst, input, path)
		}
	case reflect.Bool:
		return func(dst unsafe.Pointer, input any, path string) error {
			if value, ok := input.(bool); ok {
				*(*bool)(dst) = value
				return nil
			}
			return generic(dst, input, path)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(dst unsafe.Pointer, input any, path string) error {
			switch number := input.(type) {
			case float64:
				reflect.NewAt(typ, dst).Elem().SetInt(floatToInt64(number))
				return nil
			case int:
				reflect.NewAt(typ, dst).Elem().SetInt(int64(number))
				return nil
			case int64:
				reflect.NewAt(typ, dst).Elem().SetInt(number)
				return nil
			}
			return generic(dst, input, path)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return func(dst unsafe.Pointer, input any, path string) error {
			if number, ok := input.(float64); ok {
				reflect.NewAt(typ, dst).Elem().SetUint(uint64(number))
				return nil
			}
			return generic(dst, input, path)
		}
	case reflect.Float32, reflect.Float64:
		return func(dst unsafe.Pointer, input any, path string) error {
			if number, ok := input.(float64); ok {
				reflect.NewAt(typ, dst).Elem().SetFloat(number)
				return nil
			}
			return generic(dst, input, path)
		}
	case reflect.Struct:
		// Nested struct: resolve the plan once, on first use (lazily so that
		// recursive types compile), instead of looking it up per value.
		var (
			once   sync.Once
			nested *decodePlan
			err    error
		)
		return func(dst unsafe.Pointer, input any, path string) error {
			if input == nil || isNilPointer(input) {
				return nil
			}
			if _, ok := asStringMap(input); ok {
				once.Do(func() { nested, err = getPlan(typ) })
				if err != nil {
					return err
				}
				return nested.decode(input, dst, path)
			}
			return generic(dst, input, path)
		}
	}
	return generic
}

func converterDecoder(typ reflect.Type, converter Converter) valueDecoder {
	return func(dst unsafe.Pointer, input any, path string) error {
		if input == nil {
			return nil
		}
		converted, err := converter(input)
		if err != nil {
			return fmt.Errorf("error decoding '%s': %s", path, err)
		}
		value := reflect.ValueOf(converted)
		if !value.IsValid() || !value.Type().AssignableTo(typ) {
			return fmt.Errorf("error decoding '%s': converter returned %T, want %s", path, converted, typ)
		}
		reflect.NewAt(typ, dst).Elem().Set(value)
		return nil
	}
}

func decodeReflect(input any, target reflect.Value, path string) error {
	if input == nil || isNilPointer(input) {
		return nil
	}

	if converter := registeredConverter(target.Type()); converter != nil {
		return converterDecoder(target.Type(), converter)(unsafe.Pointer(target.Addr().Pointer()), input, path)
	}
	if decoder := builtinSpecialDecoder(target.Type()); decoder != nil {
		return decoder(unsafe.Pointer(target.Addr().Pointer()), input, path)
	}

	inputValue := reflect.ValueOf(input)
	if inputValue.Type().AssignableTo(target.Type()) && !isContainerKind(target.Kind()) {
		target.Set(inputValue)
		return nil
	}
	if inputValue.Kind() == reflect.Pointer && inputValue.Type().Elem() == target.Type() && !isContainerKind(target.Kind()) {
		target.Set(inputValue.Elem())
		return nil
	}
	// F14: weak conversions see through one non-nil pointer level, as
	// mapstructure's reflect.Indirect did; a typed nil pointer leaves the
	// target untouched.
	if inputValue.Kind() == reflect.Pointer {
		if inputValue.IsNil() {
			return nil
		}
		inputValue = inputValue.Elem()
		input = inputValue.Interface()
	}

	switch target.Kind() {
	case reflect.Interface:
		// an assignable input was handled above; anything else cannot satisfy
		// a non-empty interface target.
		return unconvertible(path, target.Type(), inputValue)
	case reflect.String:
		return decodeStringValue(inputValue, target, path)
	case reflect.Bool:
		return decodeBoolValue(inputValue, target, path)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return decodeIntValue(inputValue, target, path)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return decodeUintValue(inputValue, target, path)
	case reflect.Float32, reflect.Float64:
		return decodeFloatValue(inputValue, target, path)
	case reflect.Pointer:
		return decodePointerValue(input, target, path)
	case reflect.Struct:
		plan, err := getPlan(target.Type())
		if err != nil {
			return err
		}
		return plan.decode(input, unsafe.Pointer(target.Addr().Pointer()), path)
	case reflect.Slice:
		return decodeSliceValue(input, inputValue, target, path)
	case reflect.Array:
		return decodeArrayValue(input, inputValue, target, path)
	case reflect.Map:
		return decodeMapValue(input, inputValue, target, path)
	default:
		return fmt.Errorf("%s: unsupported type: %s", path, target.Kind())
	}
}

// isContainerKind reports whether values of this kind must be copied element
// by element: storing them by reference would alias the caller's payload and
// replace, rather than merge into, an existing map.
func isContainerKind(kind reflect.Kind) bool {
	return kind == reflect.Slice || kind == reflect.Map
}

func builtinSpecialDecoder(typ reflect.Type) valueDecoder {
	switch typ {
	case timeType:
		return decodeTime
	case timePtrType:
		return decodeTimePointer
	case decimalType:
		return decodeDecimal
	case nullStringType:
		return decodeNullString
	case pgDateType:
		return decodePGDate
	case pgDatePtrType:
		return decodePGDatePointer
	default:
		return nil
	}
}

func decodeStringValue(input, target reflect.Value, path string) error {
	switch input.Kind() {
	case reflect.String:
		target.SetString(input.String())
	case reflect.Bool:
		if input.Bool() {
			target.SetString("1")
		} else {
			target.SetString("0")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		target.SetString(strconv.FormatInt(input.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		target.SetString(strconv.FormatUint(input.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		target.SetString(strconv.FormatFloat(input.Float(), 'f', -1, 64))
	case reflect.Slice, reflect.Array:
		if input.Type().Elem().Kind() != reflect.Uint8 {
			return unconvertible(path, target.Type(), input)
		}
		bytes := make([]byte, input.Len())
		reflect.Copy(reflect.ValueOf(bytes), input)
		target.SetString(string(bytes))
	default:
		return unconvertible(path, target.Type(), input)
	}
	return nil
}

func decodeBoolValue(input, target reflect.Value, path string) error {
	switch input.Kind() {
	case reflect.Bool:
		target.SetBool(input.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		target.SetBool(input.Int() != 0)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		target.SetBool(input.Uint() != 0)
	case reflect.Float32, reflect.Float64:
		target.SetBool(input.Float() != 0)
	case reflect.String:
		if input.String() == "" {
			target.SetBool(false)
			return nil
		}
		value, err := strconv.ParseBool(input.String())
		if err != nil {
			return fmt.Errorf("cannot parse '%s' as bool: %s", path, err)
		}
		target.SetBool(value)
	default:
		return unconvertible(path, target.Type(), input)
	}
	return nil
}

func decodeIntValue(input, target reflect.Value, path string) error {
	switch input.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		target.SetInt(input.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		target.SetInt(int64(input.Uint()))
	case reflect.Float32, reflect.Float64:
		target.SetInt(floatToInt64(input.Float()))
	case reflect.Bool:
		if input.Bool() {
			target.SetInt(1)
		} else {
			target.SetInt(0)
		}
	case reflect.String:
		text := input.String()
		if text == "" {
			text = "0"
		}
		value, err := strconv.ParseInt(text, 0, target.Type().Bits())
		if err != nil {
			return fmt.Errorf("cannot parse '%s' as int: %s", path, err)
		}
		target.SetInt(value)
	default:
		return unconvertible(path, target.Type(), input)
	}
	return nil
}

// floatToInt64 makes the otherwise implementation-specific conversion of
// out-of-range floating-point values deterministic. In particular,
// float64(math.MaxInt64) rounds to 1<<63, so converting it directly can yield
// math.MinInt64 on amd64 instead of the saturating result callers expect.
func floatToInt64(value float64) int64 {
	const limit = float64(uint64(1) << 63)
	switch {
	case math.IsNaN(value):
		return 0
	case value >= limit:
		return math.MaxInt64
	case value <= -limit:
		return math.MinInt64
	default:
		return int64(value)
	}
}

func decodeUintValue(input, target reflect.Value, path string) error {
	switch input.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		target.SetUint(uint64(input.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		target.SetUint(input.Uint())
	case reflect.Float32, reflect.Float64:
		target.SetUint(uint64(input.Float()))
	case reflect.Bool:
		if input.Bool() {
			target.SetUint(1)
		} else {
			target.SetUint(0)
		}
	case reflect.String:
		text := input.String()
		if text == "" {
			text = "0"
		}
		value, err := strconv.ParseUint(text, 0, target.Type().Bits())
		if err != nil {
			return fmt.Errorf("cannot parse '%s' as uint: %s", path, err)
		}
		target.SetUint(value)
	default:
		return unconvertible(path, target.Type(), input)
	}
	return nil
}

func decodeFloatValue(input, target reflect.Value, path string) error {
	switch input.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		target.SetFloat(float64(input.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		target.SetFloat(float64(input.Uint()))
	case reflect.Float32, reflect.Float64:
		target.SetFloat(input.Float())
	case reflect.Bool:
		if input.Bool() {
			target.SetFloat(1)
		} else {
			target.SetFloat(0)
		}
	case reflect.String:
		text := input.String()
		if text == "" {
			text = "0"
		}
		value, err := strconv.ParseFloat(text, target.Type().Bits())
		if err != nil {
			return fmt.Errorf("cannot parse '%s' as float: %s", path, err)
		}
		target.SetFloat(value)
	default:
		return unconvertible(path, target.Type(), input)
	}
	return nil
}

func decodePointerValue(input any, target reflect.Value, path string) error {
	if input == nil || isNilPointer(input) {
		return nil
	}
	copy := reflect.New(target.Type().Elem())
	if !target.IsNil() {
		copy.Elem().Set(cloneValue(target.Elem()))
	}
	target.Set(copy)
	return decodeReflect(input, target.Elem(), path)
}

func decodeSliceValue(raw any, input, target reflect.Value, path string) error {
	if input.Kind() != reflect.Slice && input.Kind() != reflect.Array {
		if input.Kind() == reflect.Map && input.Len() == 0 {
			target.Set(reflect.MakeSlice(target.Type(), 0, 0))
			return nil
		}
		if input.Kind() == reflect.String && target.Type().Elem().Kind() == reflect.Uint8 {
			input = reflect.ValueOf([]byte(input.String()))
		} else {
			input = reflect.ValueOf([]any{raw})
		}
	}
	if input.Kind() == reflect.Slice && input.IsNil() {
		return nil
	}

	var result reflect.Value
	if target.IsNil() {
		result = reflect.MakeSlice(target.Type(), input.Len(), input.Len())
	} else {
		result = reflect.MakeSlice(target.Type(), target.Len(), target.Cap())
		reflect.Copy(result, target)
	}
	for result.Len() < input.Len() {
		result = reflect.Append(result, reflect.Zero(target.Type().Elem()))
	}
	var decodeErrors []string
	for i := 0; i < input.Len(); i++ {
		if err := decodeReflect(input.Index(i).Interface(), result.Index(i), indexPath(path, i)); err != nil {
			decodeErrors = appendDecodeErrors(decodeErrors, err)
		}
	}
	target.Set(result)
	return combineErrors(decodeErrors)
}

func decodeArrayValue(raw any, input, target reflect.Value, path string) error {
	if input.Kind() != reflect.Slice && input.Kind() != reflect.Array {
		if input.Kind() == reflect.Map && input.Len() == 0 {
			target.SetZero()
			return nil
		}
		input = reflect.ValueOf([]any{raw})
	}
	if input.Len() > target.Len() {
		return fmt.Errorf("'%s': expected source data to have length less or equal to %d, got %d", path, target.Len(), input.Len())
	}
	var decodeErrors []string
	for i := 0; i < input.Len(); i++ {
		if err := decodeReflect(input.Index(i).Interface(), target.Index(i), indexPath(path, i)); err != nil {
			decodeErrors = appendDecodeErrors(decodeErrors, err)
		}
	}
	return combineErrors(decodeErrors)
}

func decodeMapValue(raw any, input, target reflect.Value, path string) error {
	if input.Kind() == reflect.Slice || input.Kind() == reflect.Array {
		if input.Len() == 0 {
			if target.IsNil() {
				target.Set(reflect.MakeMap(target.Type()))
			}
			return nil
		}
		for i := 0; i < input.Len(); i++ {
			element := input.Index(i).Interface()
			if err := decodeMapValue(element, reflect.ValueOf(element), target, indexPath(path, i)); err != nil {
				return err
			}
		}
		return nil
	}
	if input.Kind() == reflect.Pointer && !input.IsNil() && input.Elem().Kind() == reflect.Struct {
		input = input.Elem()
	}
	if input.Kind() == reflect.Struct {
		// struct -> map, as mapstructure's decodeMapFromStruct: the struct's
		// exported fields (json names, squash honoured) become the source map.
		view, err := newMapView(raw)
		if err != nil {
			return namedExpectedMap(path, raw)
		}
		return decodeMapValue(view.exact, reflect.ValueOf(view.exact), target, path)
	}
	if input.Kind() != reflect.Map {
		return fmt.Errorf("'%s' expected a map, got '%s'", path, input.Kind())
	}
	if input.IsNil() {
		return nil
	}
	if input.Len() == 0 {
		// nothing to merge; a pre-populated target keeps its entries
		if target.IsNil() {
			target.Set(reflect.MakeMap(target.Type()))
		}
		return nil
	}
	if target.IsNil() {
		target.Set(reflect.MakeMap(target.Type()))
	} else {
		copy := reflect.MakeMapWithSize(target.Type(), target.Len()+input.Len())
		iterator := target.MapRange()
		for iterator.Next() {
			copy.SetMapIndex(iterator.Key(), iterator.Value())
		}
		target.Set(copy)
	}

	var decodeErrors []string
	iterator := input.MapRange()
	for iterator.Next() {
		key := reflect.New(target.Type().Key()).Elem()
		value := reflect.New(target.Type().Elem()).Elem()
		itemPath := keyPath(path, iterator.Key())
		if err := decodeReflect(iterator.Key().Interface(), key, itemPath); err != nil {
			decodeErrors = appendDecodeErrors(decodeErrors, err)
			continue
		}
		if err := decodeReflect(iterator.Value().Interface(), value, itemPath); err != nil {
			decodeErrors = appendDecodeErrors(decodeErrors, err)
			continue
		}
		target.SetMapIndex(key, value)
	}
	return combineErrors(decodeErrors)
}

func decodeTime(dst unsafe.Pointer, input any, path string) error {
	if input == nil {
		return nil
	}
	if result, ok := input.(time.Time); ok {
		*(*time.Time)(dst) = result
		return nil
	}
	if pointer, ok := input.(*time.Time); ok {
		if pointer != nil {
			*(*time.Time)(dst) = *pointer
		}
		return nil
	}
	value := reflect.ValueOf(input)
	var result time.Time
	var err error
	switch value.Kind() {
	case reflect.String:
		result, err = parseTime(value.String())
	case reflect.Float64:
		result = time.Unix(0, int64(value.Float())*int64(time.Millisecond))
	case reflect.Int64:
		result = time.Unix(0, value.Int()*int64(time.Millisecond))
	default:
		return namedExpectedMap(path, input)
	}
	if err != nil {
		return fmt.Errorf("error decoding '%s': %s", path, err)
	}
	*(*time.Time)(dst) = result
	return nil
}

func decodeTimePointer(dst unsafe.Pointer, input any, path string) error {
	if input == nil {
		return nil
	}
	if result, ok := input.(*time.Time); ok {
		*(**time.Time)(dst) = result
		return nil
	}
	if result, ok := input.(time.Time); ok {
		*(**time.Time)(dst) = &result
		return nil
	}
	value := reflect.ValueOf(input)
	if value.Kind() == reflect.String && value.String() == "" {
		*(**time.Time)(dst) = nil
		return nil
	}
	var result time.Time
	var err error
	switch value.Kind() {
	case reflect.String:
		result, err = parseTime(value.String())
	case reflect.Float64:
		result = time.Unix(0, int64(value.Float())*int64(time.Millisecond))
	case reflect.Int64:
		result = time.Unix(0, value.Int()*int64(time.Millisecond))
	default:
		return namedExpectedMap(path, input)
	}
	if err != nil {
		return fmt.Errorf("error decoding '%s': %s", path, err)
	}
	*(**time.Time)(dst) = &result
	return nil
}

func decodeDecimal(dst unsafe.Pointer, input any, path string) error {
	if input == nil {
		return nil
	}
	var result decimal.Decimal
	var err error
	switch value := input.(type) {
	case float64:
		result = decimal.NewFromFloat(value)
	case string:
		if value != "" {
			result, err = decimal.NewFromString(value)
		}
	default:
		return fmt.Errorf("error decoding '%s': expected string or float64 for decimal.Decimal, got %T", path, input)
	}
	if err != nil {
		return fmt.Errorf("error decoding '%s': %s", path, err)
	}
	*(*decimal.Decimal)(dst) = result
	return nil
}

func decodeNullString(dst unsafe.Pointer, input any, path string) error {
	if input == nil {
		return nil
	}
	value, ok := input.(string)
	if !ok {
		return fmt.Errorf("error decoding '%s': expected string for null.String, got %T", path, input)
	}
	*(*null.String)(dst) = null.StringFrom(value)
	return nil
}

func decodePGDate(dst unsafe.Pointer, input any, _ string) error {
	if input == nil {
		return nil
	}
	result := pgtype.Date{}
	_ = result.Set(input)
	*(*pgtype.Date)(dst) = result
	return nil
}

func decodePGDatePointer(dst unsafe.Pointer, input any, _ string) error {
	if input == nil {
		return nil
	}
	result := &pgtype.Date{}
	_ = result.Set(input)
	*(**pgtype.Date)(dst) = result
	return nil
}

// indexPath formats "path[i]" without fmt; it is built per element, so keep
// it to a single concatenation.
func indexPath(path string, index int) string {
	return path + "[" + strconv.Itoa(index) + "]"
}

func keyPath(path string, key reflect.Value) string {
	if key.Kind() == reflect.String {
		return path + "[" + key.String() + "]"
	}
	return fmt.Sprintf("%s[%v]", path, key.Interface())
}

// parseTime accepts RFC 3339 (with or without fractional seconds) directly —
// the form JSON clients send — and falls back to cast's layout list for
// everything else, which is what the legacy hook always did.
func parseTime(text string) (time.Time, error) {
	if result, err := time.Parse(time.RFC3339Nano, text); err == nil {
		return result, nil
	}
	return cast.ToTimeInDefaultLocationE(text, nil)
}

func unconvertible(path string, target reflect.Type, input reflect.Value) error {
	return fmt.Errorf("'%s' expected type '%s', got unconvertible type '%s', value: '%v'", path, target, input.Type(), input.Interface())
}
