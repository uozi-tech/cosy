package structcodec

import (
	"fmt"
	"reflect"
	"sync"
)

// Converter converts an input value into one registered concrete type.
type Converter func(input any) (any, error)

var converterRegistry = struct {
	sync.RWMutex
	values map[reflect.Type]Converter
}{values: make(map[reflect.Type]Converter)}

// RegisterConverter registers a converter for the concrete type represented by
// sample. Existing compiled plans are discarded so later decodes see it.
func RegisterConverter(sample any, converter Converter) error {
	if sample == nil {
		return fmt.Errorf("structcodec: converter sample must not be nil")
	}
	if converter == nil {
		return fmt.Errorf("structcodec: converter must not be nil")
	}
	typ := reflect.TypeOf(sample)
	planMu.Lock()
	converterRegistry.Lock()
	converterRegistry.values[typ] = converter
	converterRegistry.Unlock()
	planCache.Clear()
	planMu.Unlock()
	return nil
}

// UnregisterConverter removes the converter registered for the concrete type
// represented by sample and discards compiled plans, mirroring RegisterConverter.
func UnregisterConverter(sample any) {
	if sample == nil {
		return
	}
	typ := reflect.TypeOf(sample)
	planMu.Lock()
	converterRegistry.Lock()
	delete(converterRegistry.values, typ)
	converterRegistry.Unlock()
	planCache.Clear()
	planMu.Unlock()
}

func registeredConverter(typ reflect.Type) Converter {
	converterRegistry.RLock()
	converter := converterRegistry.values[typ]
	converterRegistry.RUnlock()
	return converter
}
