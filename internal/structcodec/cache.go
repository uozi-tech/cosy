package structcodec

import (
	"reflect"
	"sync"
	"sync/atomic"
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
	// next is the index of the next field sharing the same json name (for
	// squashed structs that repeat a name), or -1.
	next int
}

type decodePlan struct {
	typ    reflect.Type
	fields []fieldPlan
	// index maps the exact json name to the first field carrying it;
	// lowerIndex does the same for the lower-cased name so that input keys
	// can be matched case-insensitively without materialising a lowered copy
	// of the input map per decode.
	index      map[string]int
	lowerIndex map[string]int
}

var (
	planCache sync.Map
	// planGeneration is bumped by invalidatePlans; a compile that straddles a
	// bump is used once but not cached, so a converter registered mid-compile
	// is never shadowed by a stale plan.
	planGeneration atomic.Uint64
)

func getPlan(typ reflect.Type) (*decodePlan, error) {
	if cached, ok := planCache.Load(typ); ok {
		return cached.(*decodePlan), nil
	}
	generation := planGeneration.Load()
	plan, err := compilePlan(typ)
	if err != nil {
		return nil, err
	}
	if planGeneration.Load() != generation {
		return plan, nil
	}
	actual, _ := planCache.LoadOrStore(typ, plan)
	return actual.(*decodePlan), nil
}

// invalidatePlans discards every compiled plan; called when the converter
// registry changes.
func invalidatePlans() {
	planGeneration.Add(1)
	planCache.Clear()
}
