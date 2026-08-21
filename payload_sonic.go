//go:build !go1.27

package cosy

import "github.com/bytedance/sonic"

// jsonDecoder is the frozen sonic configuration for decoding request bodies on
// toolchains where sonic's native JIT is available.
//
// Config discipline (PERF_REFACTOR_PLAN.md §3.3, invariant I3): CopyString
// must stay true so decoded strings never alias the request buffer, and
// ValidateString must stay true so raw control characters in string values
// are rejected as encoding/json does. Do not switch to ConfigDefault or
// ConfigFastest for speed — ConfigDefault has CopyString=false.
var jsonDecoder = sonic.Config{
	CopyString:     true,
	ValidateString: true,
}.Froze()
