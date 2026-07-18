package structcodec

import (
	"reflect"
	"sync"
	"unsafe"
)

type valueDecoder func(dst unsafe.Pointer, input any, path string) error

type addressStep struct {
	offset  uintptr
	ptrType reflect.Type
}

type fieldPlan struct {
	name   string
	steps  []addressStep
	decode valueDecoder
}

type decodePlan struct {
	typ    reflect.Type
	fields []fieldPlan
}

var (
	planCache sync.Map
	planMu    sync.RWMutex
)

func getPlan(typ reflect.Type) (*decodePlan, error) {
	planMu.RLock()
	defer planMu.RUnlock()
	if cached, ok := planCache.Load(typ); ok {
		return cached.(*decodePlan), nil
	}
	plan, err := compilePlan(typ)
	if err != nil {
		return nil, err
	}
	actual, _ := planCache.LoadOrStore(typ, plan)
	return actual.(*decodePlan), nil
}
