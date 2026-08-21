package structcodec

import "reflect"

// Pretouch compiles and caches the decode plan for the struct type behind v,
// mirroring sonic.Pretouch: calling it at boot removes the first-request
// compilation latency. Nil and non-struct values are ignored.
func Pretouch(v any) error {
	if v == nil {
		return nil
	}
	typ := reflect.TypeOf(v)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil
	}
	_, err := getPlan(typ)
	return err
}
