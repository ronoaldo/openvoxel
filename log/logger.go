// package log provides a simple logging facility
package log

import (
	"fmt"
	"time"
)

const (
	LevelInfo = "INFO"
	LevelWarn = "WARN"
)

func Warnf(message string, args ...interface{}) {
	printf(LevelWarn, message, args...)
}

func Infof(message string, args ...interface{}) {
	printf(LevelInfo, message, args...)
}

func printf(level, message string, args ...interface{}) {
	fmt.Printf(time.Now().Format("2006-01-02T15:04:05 ")+level+": "+message+"\n", args...)
}
