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
// It will skip the gorm internal files and the logger files in the project
func fileWithLineNum() string {
	// Get the current file directory, used to skip the logger package internal calls
	_, currentFile, _, _ := runtime.Caller(0)
	loggerDir := filepath.Dir(currentFile)

	// Get the gorm source code directory (used to skip the gorm internal calls)
	gormSourceDir := getGormSourceDir()

	pcs := make([]uintptr, 15)
	// Start capturing from the first caller (skipping fileWithLineNum itself)
	depth := runtime.Callers(1, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

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
			return frame.File + ":" + strconv.Itoa(frame.Line)
		}

		if !more {
			break
		}
	}

	return ""
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
