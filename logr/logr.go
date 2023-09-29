package logr

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Constant block describes log levels.
const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// debug 0
// info  1
// warn  2
// error 3

// LogLevel describes the log level.
type LogLevel int

// LogR simple logger.
type LogR struct {
	w      io.Writer
	prefix string
	level  LogLevel
	colors [5]*color.Color
}

// LogGroup represents a log group.
type LogGroup struct {
	lr    *LogR
	level LogLevel
	parts []string
}

// NewLogR creates a new logger.
func NewLogR(w io.Writer, prefix string, level LogLevel) *LogR {
	return &LogR{
		w:      w,
		prefix: prefix,
		level:  level,
		colors: [5]*color.Color{
			color.New(color.FgHiBlue),
			color.New(color.FgGreen),
			color.New(color.FgHiYellow),
			color.New(color.FgRed),
		},
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

// Group creates a new log group.
func (lr *LogR) Group(level LogLevel) *LogGroup {
	return &LogGroup{
		lr:    lr,
		level: level,
	}
}

// Log adds a log to the group.
func (lg *LogGroup) Log(fmtStr string, args ...any) {
	lg.parts = append(lg.parts, fmt.Sprintf(fmtStr, args...))
}

// Close closes the log group writeing all grouped logs.
func (lg *LogGroup) Close() {
	lg.lr.log(lg.level, strings.Join(lg.parts, "\n"))
}

// String returns a string representation of log level.
func (lvl LogLevel) String() string {
	return [5]string{"DEBUG", "INFO", "WARN", "ERROR"}[lvl]
}

func (lr *LogR) log(level LogLevel, fmtStr string, args ...any) {
	if level < lr.level {
		return
	}

	// header
	fmt.Fprintf(
		lr.w,
		"%s\n",
		fmt.Sprintf(
			"%s %s %s",
			lr.colors[level].Sprintf("%-5s", level.String()),
			color.New(color.FgCyan).Sprintf(
				"%s", time.Now().Format("02-01-2006 15:04:05.00000"),
			),
			color.New(color.FgMagenta).Sprint(lr.prefix),
		),
	)

	for _, part := range strings.Split(fmt.Sprintf(fmtStr, args...), "\n") {
		fmt.Fprintf(
			lr.w,
			"%s\n",
			fmt.Sprintf(
				"%s %s",
				lr.colors[level].Sprintf("▶"),
				part,
			),
		)
	}
}
