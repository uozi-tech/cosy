package logger

import "fmt"

func getMessageln(fmtArgs ...any) string {
	msg := fmt.Sprintln(fmtArgs...)
	msg = msg[:len(msg)-1]
	return msg
}

func getMessagef(format string, args ...any) string {
	msg := fmt.Sprintf(format, args...)
	return msg
}
