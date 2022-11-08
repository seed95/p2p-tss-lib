package log

import (
	"github.com/logrusorgru/aurora"
	golog "log"
)

type Level uint

const (
	InfoLevel Level = iota + 1
	ErrorLevel
	NoLog
)

var (
	level Level
)

func init() {
	level = NoLog
	golog.SetFlags(0)
}

func SetLevel(lvl Level) {
	level = lvl
}

func Info(caller string, message any) {
	if level <= InfoLevel {
		printLog(aurora.Blue("INFO"), caller, message)
	}
}

func Error(caller string, message any) {
	if level <= ErrorLevel {
		printLog(aurora.Red("ERROR"), caller, aurora.Sprintf(aurora.Red(message)))
	}
}

func printLog(level aurora.Value, caller string, message any) {
	golog.Printf("[%s] [%s]: %s", level, aurora.White(caller), message)
}
