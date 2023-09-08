package logr

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fatih/color"
)

// LogLevel describes the log level.
type LogLevel int

// Constant block describes log levels.
const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

var logLevelNames = []string{
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
}

// var logLevelColors = []*color.Color{
// 	color.New(color.FgHiBlue),
// 	color.New(color.FgGreen),
// 	color.New(color.FgHiYellow),
// 	color.New(color.FgHiRed),
// }

var logLevelColors = []struct {
	accent *color.Color
	base   *color.Color
}{
	{
		color.New(color.FgHiBlue),
		color.New(color.FgBlue),
	},
	{
		color.New(color.FgGreen),
		color.New(color.FgGreen),
	},
	{
		color.New(color.FgHiYellow),
		color.New(color.FgYellow),
	},
	{
		color.New(color.FgRed),
		color.New(color.FgRed),
	},
}

// debug 0
// info  1
// warn  2
// error 3

// LogR simple logger.
type LogR struct {
	w      io.Writer
	prefix string
	level  LogLevel
}

// NewLogR creates a new logger.
func NewLogR(w io.Writer, prefix string, level LogLevel) *LogR {
	return &LogR{
		w:      w,
		prefix: prefix,
		level:  level,
	}
}

// Prefix creates a new logger adding the specified prefix to parent loggers
// prefix.
func (lr *LogR) Prefix(prefix string) *LogR {
	copyLr := *lr
	copyLr.prefix = fmt.Sprintf("%s.%s", lr.prefix, prefix)

	return &copyLr
}

// Debugf logs a debug message.
func (lr *LogR) Debugf(fmtStr string, args ...any) {
	lr.log(LogLevelDebug, fmtStr, args...)
}

// Infof logs an info message.
func (lr *LogR) Infof(fmtStr string, args ...any) {
	lr.log(LogLevelInfo, fmtStr, args...)
}

// Warnf logs a warn message.
func (lr *LogR) Warnf(fmtStr string, args ...any) {
	lr.log(LogLevelWarn, fmtStr, args...)
}

// Errorf logs an error message.
func (lr *LogR) Errorf(fmtStr string, args ...any) {
	lr.log(LogLevelError, fmtStr, args...)
}

func (lr *LogR) log(level LogLevel, fmtStr string, args ...any) {
	if level < lr.level {
		return
	}

	// header
	lr.w.Write([]byte(fmt.Sprintf( //nolint:errcheck
		"%s\n",
		fmt.Sprintf(
			"%s %s %s",
			logLevelColors[level].accent.Sprintf("%-5s", logLevelNames[level]),
			color.New(color.FgCyan).Sprintf(
				"%s", time.Now().Format("02-01-2006 15:04:05.00000"),
			),
			color.New(color.FgMagenta).Sprint(lr.prefix),
		),
	)))

	for _, part := range strings.Split(fmt.Sprintf(fmtStr, args...), "\n") {
		lr.w.Write([]byte(fmt.Sprintf( //nolint:errcheck
			"%s\n",
			fmt.Sprintf(
				"%s %s",
				logLevelColors[level].accent.Sprintf("▶"),
				part,
			),
		)))
	}
}
