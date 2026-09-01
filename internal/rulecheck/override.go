package rulecheck

import "sync"

var overridden sync.Map

// Override marks tag names whose built-in implementation must be bypassed in
// favour of the *validator.Validate handed to ValidateMap, typically because a
// custom validation was registered under that name. Compiled rules are
// discarded so rule strings seen before the call pick the override up.
func Override(tags ...string) {
	for _, tag := range tags {
		overridden.Store(tag, struct{}{})
	}
	ruleCache.Clear()
}

func isOverridden(name string) bool {
	_, ok := overridden.Load(name)
	return ok
}

// clearOverrides resets Override state; tests only.
func clearOverrides() {
	overridden.Range(func(key, _ any) bool {
		overridden.Delete(key)
		return true
	})
	ruleCache.Clear()
}
