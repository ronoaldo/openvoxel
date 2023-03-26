// package log provides a simple logging facility
package log

import (
	"fmt"
	"time"
)

const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARNING"
	LevelError = "ERROR"
)

// Debugf prints a log message with DEBUG level
func Debugf(message string, args ...interface{}) {
	printf(LevelDebug, message, args...)
}

// Infof prints a log message with INFO level
func Infof(message string, args ...interface{}) {
	printf(LevelInfo, message, args...)
}

// Warnf prints a log message with WARNING level
func Warnf(message string, args ...interface{}) {
	printf(LevelWarn, message, args...)
}

// Errorf prints a log message with ERROR level
func Errorf(message string, args ...interface{}) {
	printf(LevelError, message, args...)
}

func printf(level, message string, args ...interface{}) {
	fmt.Printf(time.Now().Format("2006-01-02T15:04:05 ")+level+": "+message+"\n", args...)
}
