package logger

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

func getMessageln(fmtArgs ...any) string {
	msg := fmt.Sprintln(fmtArgs...)
	msg = msg[:len(msg)-1]
	return msg
}

func getMessagef(format string, args ...any) string {
	msg := fmt.Sprintf(format, args...)
	return msg
}

var (
	// loggerDir is this package's source directory: gorm_logger.go, session.go
	// and the other logging wrappers are never the code that issued a query.
	loggerDir = func() string {
		_, file, _, _ := runtime.Caller(0)
		return filepath.Dir(file)
	}()
	// cosyDir is the cosy module root: every package under it is a wrapper
	// around gorm, not the application code that issued the query.
	cosyDir = filepath.Dir(loggerDir) + string(filepath.Separator)
)

// gormPathMarker matches gorm and its drivers wherever they are built from
// (module cache, vendor/ or a replace directive).
const gormPathMarker = "gorm.io/"

// frameKind is how fileWithLineNum treats one stack frame.
type frameKind uint8

const (
	frameSkipped  frameKind = iota // gorm, the logger package, tests, generated code
	frameFallback                  // other cosy packages and runtime/testing entry points
	frameCaller                    // application code
)

type callerFrame struct {
	location string
	kind     frameKind
}

// callerFrames caches the classified frames of each program counter. A PC
// always resolves to the same frames, and the set of PCs is bounded by the
// binary's code, so the cache never needs eviction. One PC can expand to
// several frames when calls were inlined.
var callerFrames sync.Map // uintptr -> []callerFrame

func framesForPC(pc uintptr) []callerFrame {
	if cached, ok := callerFrames.Load(pc); ok {
		return cached.([]callerFrame)
	}
	var out []callerFrame
	frames := runtime.CallersFrames([]uintptr{pc})
	for {
		frame, more := frames.Next()
		out = append(out, classifyFrame(frame))
		if !more {
			break
		}
	}
	actual, _ := callerFrames.LoadOrStore(pc, out)
	return actual.([]callerFrame)
}

func classifyFrame(frame runtime.Frame) callerFrame {
	if strings.Contains(frame.File, gormPathMarker) ||
		strings.Contains(frame.File, loggerDir) ||
		strings.HasSuffix(frame.File, "_test.go") ||
		strings.HasSuffix(frame.File, ".gen.go") {
		return callerFrame{kind: frameSkipped}
	}
	location := frame.File + ":" + strconv.Itoa(frame.Line)
	if strings.HasPrefix(frame.File, cosyDir) || isEntryPointFrame(frame.Function) {
		return callerFrame{location: location, kind: frameFallback}
	}
	return callerFrame{location: location, kind: frameCaller}
}

// fileWithLineNum returns the file name and line number of the caller
// It will skip the gorm internal files and the files of the cosy module
// itself (the logger package as well as the CRUD helpers in the root and
// sub packages) so the reported location is the application code that
// issued the query. When every frame belongs to cosy (cosy's own tests),
// it falls back to the first frame outside gorm and the logger package.
func fileWithLineNum() string {
	var pcs [32]uintptr
	// Skip runtime.Callers and fileWithLineNum itself.
	depth := runtime.Callers(2, pcs[:])

	fallback := ""
	for _, pc := range pcs[:depth] {
		for _, frame := range framesForPC(pc) {
			switch frame.kind {
			case frameCaller:
				return frame.location
			case frameFallback:
				if fallback == "" {
					fallback = frame.location
				}
			}
		}
	}
	return fallback
}

// isEntryPointFrame reports whether the function belongs to the runtime or
// testing packages, which never are the code that issued a query.
func isEntryPointFrame(function string) bool {
	return strings.HasPrefix(function, "runtime.") || strings.HasPrefix(function, "testing.")
}

// splitCaller splits a "file:line" location produced by fileWithLineNum.
func splitCaller(caller string) (string, int, bool) {
	i := strings.LastIndexByte(caller, ':')
	if i <= 0 {
		return "", 0, false
	}
	line, err := strconv.Atoi(caller[i+1:])
	if err != nil || line <= 0 {
		return "", 0, false
	}
	return caller[:i], line, true
}
