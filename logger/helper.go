package logger

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
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

// fileWithLineNum returns the file name and line number of the caller
// It will skip the gorm internal files and the files of the cosy module
// itself (the logger package as well as the CRUD helpers in the root and
// sub packages) so the reported location is the application code that
// issued the query. When every frame belongs to cosy (cosy's own tests),
// it falls back to the first frame outside gorm and the logger package.
func fileWithLineNum() string {
	// Get the current file directory, used to skip the logger package internal calls
	_, currentFile, _, _ := runtime.Caller(0)
	loggerDir := filepath.Dir(currentFile)
	// The cosy module root: every package under it is a wrapper around gorm,
	// not the application code that issued the query.
	cosyDir := filepath.Dir(loggerDir) + string(filepath.Separator)

	// Get the gorm source code directory (used to skip the gorm internal calls)
	gormSourceDir := getGormSourceDir()

	pcs := make([]uintptr, 32)
	// Start capturing from the first caller (skipping fileWithLineNum itself)
	depth := runtime.Callers(1, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	fallback := ""
	for i := 0; i < depth; i++ {
		frame, more := frames.Next()

		// Skip the following files:
		// 1. The files in the gorm source code directory
		// 2. The files in the logger directory of the project (gorm_logger.go, logger.go, etc.)
		// 3. Test files
		// 4. Generated files
		if !strings.Contains(frame.File, gormSourceDir) &&
			!strings.Contains(frame.File, loggerDir) &&
			!strings.HasSuffix(frame.File, "_test.go") &&
			!strings.HasSuffix(frame.File, ".gen.go") {
			location := frame.File + ":" + strconv.Itoa(frame.Line)
			// 5. The other packages of the cosy module (list.go, delete.go,
			//    model/, filter/, valid/ ...) and the runtime/testing entry
			//    points that sit on top of every stack (runtime.goexit,
			//    testing.tRunner) are only remembered as a fallback, so the
			//    application frame wins when there is one
			if !strings.HasPrefix(frame.File, cosyDir) && !isEntryPointFrame(frame.Function) {
				return location
			}
			if fallback == "" {
				fallback = location
			}
		}

		if !more {
			break
		}
	}

	return fallback
}

// isEntryPointFrame reports whether the function belongs to the runtime or
// testing packages, which never are the code that issued a query.
func isEntryPointFrame(function string) bool {
	return strings.HasPrefix(function, "runtime.") || strings.HasPrefix(function, "testing.")
}

// getGormSourceDir returns the gorm source code directory
func getGormSourceDir() string {
	pcs := make([]uintptr, 10)
	depth := runtime.Callers(0, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	for i := 0; i < depth; i++ {
		frame, more := frames.Next()
		// Find the gorm.io path
		if strings.Contains(frame.File, "gorm.io") {
			// Get the root directory of gorm.io
			idx := strings.Index(frame.File, "gorm.io")
			if idx > 0 {
				return frame.File[:idx+len("gorm.io")]
			}
		}
		if !more {
			break
		}
	}

	return "gorm.io"
}
