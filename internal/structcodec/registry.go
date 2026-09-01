package structcodec

import (
	"fmt"
	"maps"
	"reflect"
	"sync"
	"sync/atomic"
)

// Converter converts an input value into one registered concrete type.
type Converter func(input any) (any, error)

var (
	// converters is a copy-on-write snapshot so the per-value lookup on the
	// decode path is a single atomic load; nil means nothing is registered.
	converters   atomic.Pointer[map[reflect.Type]Converter]
	convertersMu sync.Mutex
)

// RegisterConverter registers a converter for the concrete type represented by
// sample. Existing compiled plans are discarded so later decodes see it.
func RegisterConverter(sample any, converter Converter) error {
	if sample == nil {
		return fmt.Errorf("structcodec: converter sample must not be nil")
	}
	if converter == nil {
		return fmt.Errorf("structcodec: converter must not be nil")
	}
	updateConverters(func(registry map[reflect.Type]Converter) {
		registry[reflect.TypeOf(sample)] = converter
	})
	return nil
}

// UnregisterConverter removes the converter registered for the concrete type
// represented by sample and discards compiled plans, mirroring RegisterConverter.
func UnregisterConverter(sample any) {
	if sample == nil {
		return
	}
	updateConverters(func(registry map[reflect.Type]Converter) {
		delete(registry, reflect.TypeOf(sample))
	})
}

func updateConverters(mutate func(map[reflect.Type]Converter)) {
	convertersMu.Lock()
	defer convertersMu.Unlock()
	registry := make(map[reflect.Type]Converter)
	if current := converters.Load(); current != nil {
		registry = maps.Clone(*current)
	}
	mutate(registry)
	if len(registry) == 0 {
		converters.Store(nil)
	} else {
		converters.Store(&registry)
	}
	invalidatePlans()
}

func registeredConverter(typ reflect.Type) Converter {
	registry := converters.Load()
	if registry == nil {
		return nil
	}
	return (*registry)[typ]
}
