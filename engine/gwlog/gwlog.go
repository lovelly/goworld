package gwlog

import (
	"runtime/debug"

	"io"

	"os"

	"strings"

	sublog "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"log"
	"fmt"
)

var (
	outputWriter io.Writer

	// DebugLevel level
	DebugLevel Level = Level(sublog.DebugLevel)
	// InfoLevel level
	InfoLevel Level = Level(sublog.InfoLevel)
	// WarnLevel level
	WarnLevel Level = Level(sublog.WarnLevel)
	// ErrorLevel level
	ErrorLevel Level = Level(sublog.ErrorLevel)
	// PanicLevel level
	PanicLevel Level = Level(sublog.PanicLevel)
	// FatalLevel level
	FatalLevel Level = Level(sublog.FatalLevel)
)

type logFormatFunc func(format string, args ...interface{})

// Level is type of log levels
type Level uint8

func init() {
	fmt.Println("111111111")
	outputWriter = os.Stdout
	sublog.SetOutput(outputWriter)
	sublog.SetLevel(sublog.DebugLevel)
	sublog.SetFormatter(&sublog.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02T15:04:05.000000000", Calldepth:8,Flag:log.Llongfile})
}

// SetSource sets the component name (dispatcher/gate/game) of gwlog module
func SetSource(comp string) {
	//logEntry := sublog.WithField("source", comp)
	//
	//Debugf = logEntry.Debugf
	//Infof = logEntry.Infof
	//Warnf = logEntry.Warnf
	//Errorf = logEntry.Errorf
}

// ParseLevel parses log level string to Level
func ParseLevel(lvl string) (Level, error) {
	lv, err := sublog.ParseLevel(lvl)
	return Level(lv), err
}

// SetLevel sets the log level
func SetLevel(lv Level) {
	sublog.SetLevel(sublog.Level(lv))
}

// TraceError prints the stack and error
func TraceError(format string, args ...interface{}) {
	outputWriter.Write(debug.Stack())
	Errorf(format, args...)
}

func Debugf(format string, args ...interface{}) {
	sublog.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	sublog.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	sublog.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	sublog.Errorf(format, args...)
}



// Fatalf prints formatted fatal message
func Fatalf(format string, args ...interface{}) {
	debug.PrintStack()
	sublog.Fatalf(format, args...)
}

// Panic panics
func Panic(v interface{}) {
	panic(v)
}

// Panicf prints formatted panic message
func Panicf(format string, args ...interface{}) {
	panic(errors.Errorf(format, args...))
}

// SetOutput sets the output writer
func SetOutput(out io.Writer) {
	outputWriter = out
	sublog.SetOutput(outputWriter)
}

// GetOutput returns the output writer
func GetOutput() io.Writer {
	return outputWriter
}

// StringToLevel converts string to Levels
func StringToLevel(s string) Level {
	if strings.ToLower(s) == "debug" {
		return DebugLevel
	} else if strings.ToLower(s) == "info" {
		return InfoLevel
	} else if strings.ToLower(s) == "warn" || strings.ToLower(s) == "warning" {
		return WarnLevel
	} else if strings.ToLower(s) == "error" {
		return ErrorLevel
	} else if strings.ToLower(s) == "panic" {
		return PanicLevel
	} else if strings.ToLower(s) == "fatal" {
		return FatalLevel
	}
	Errorf("StringToLevel: unknown level: %s", s)
	return DebugLevel
}
