package logr

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"rat/graph/util"
)

// Constant block describes log levels.
const (
	LogLevelDebug LogLevel = iota // 0 (default)
	LogLevelInfo                  // 1
	LogLevelWarn                  // 2
	LogLevelError                 // 3
)

// Config describes logr configuration.
type Config struct {
	DefaultLevel LogLevel            `yaml:"defaultLevel"`
	PrefixLevels map[string]LogLevel `yaml:"prefixLevels"`
}

// LogLevel describes the log level.
//
//nolint:recvcheck
type LogLevel int

// LogR simple logger.
type LogR struct {
	w           io.Writer
	prefix      string
	timeFormat  string
	level       LogLevel
	levelCached bool
	config      *Config
	colors      [5]*color.Color
}

// LogGroup represents a log group.
type LogGroup struct {
	lr    *LogR
	level LogLevel
	parts []string
}

// NewLogR creates a new logger.
func NewLogR(w io.Writer, prefix string, config *Config) *LogR {
	return &LogR{
		w:          w,
		prefix:     prefix,
		timeFormat: "02-01-2006 15:04:05.00000",
		config:     config,
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
	copyLr.levelCached = false

	return &copyLr
}

// Debugf logs a debug message.
func (lr *LogR) Debugf(fmtStr string, args ...any) {
	// TODO: handle error by saving it on the logger, create a service that
	// would be able to report it FE or status.
	lr.log(LogLevelDebug, fmtStr, args...) //nolint:errcheck
}

// Infof logs an info message.
func (lr *LogR) Infof(fmtStr string, args ...any) {
	lr.log(LogLevelInfo, fmtStr, args...) //nolint:errcheck
}

// Warnf logs a warn message.
func (lr *LogR) Warnf(fmtStr string, args ...any) {
	lr.log(LogLevelWarn, fmtStr, args...) //nolint:errcheck
}

// Errorf logs an error message.
func (lr *LogR) Errorf(fmtStr string, args ...any) {
	lr.log(LogLevelError, fmtStr, args...) //nolint:errcheck
}

// Group creates a new log group.
func (lr *LogR) Group(level LogLevel) *LogGroup {
	return &LogGroup{
		lr:    lr,
		level: level,
	}
}

// Log adds a log to the group.
func (lg *LogGroup) Log(fmtStr string, args ...any) *LogGroup {
	lg.parts = append(lg.parts, fmt.Sprintf(fmtStr, args...))

	return lg
}

// Close closes the log group writeing all grouped logs.
func (lg *LogGroup) Close() {
	lg.lr.log(lg.level, "%s", strings.Join(lg.parts, "\n")) //nolint:errcheck
}

// String returns a string representation of log level.
func (lvl LogLevel) String() string {
	return [4]string{"DEBUG", "INFO", "WARN", "ERROR"}[lvl]
}

// Preview prepares long string for preview, so that they fit into a 80 column
// long terminal window. with pretty log format, adding a newline at the
// beginning. should be used as %s format string.
func Preview(s string) string {
	const maxLen = 63

	if len(s) > maxLen {
		return fmt.Sprintf("\npreview:\n%s\n---cut---", s[:maxLen])
	}

	return "\npreview:\n" + s
}

// UnmarshalYAML implements yaml.Unmarshaler for LogLevel type to check for
// valid values.
func (lvl *LogLevel) UnmarshalYAML(unmarshal func(any) error) error {
	var raw string

	err := unmarshal(&raw)
	if err != nil {
		return errors.Wrap(err, "failed to YAML unmarshal log level")
	}

	switch strings.ToLower(raw) {
	case "debug":
		*lvl = LogLevelDebug
	case "info":
		*lvl = LogLevelInfo
	case "warn":
		*lvl = LogLevelWarn
	case "error":
		*lvl = LogLevelError
	default:
		v, err := strconv.Atoi(raw)
		if err != nil {
			return errors.Wrap(err, "failed to YAML unmarshal log level as int")
		}

		if v < 0 || v > 3 {
			return errors.Errorf(
				"invalid log level %d, accepted values are 0, 1, 2, 3",
				v,
			)
		}

		*lvl = LogLevel(v)
	}

	return nil
}

func (lr *LogR) minLevel(prefix string) LogLevel {
	if lr.levelCached {
		return lr.level
	}

	lr.level = lr.config.DefaultLevel
	lr.levelCached = true

	if len(lr.config.PrefixLevels) == 0 {
		return lr.level
	}

	prefixes := util.Filter(
		util.Keys(lr.config.PrefixLevels),
		func(k string) bool { return strings.HasPrefix(prefix, k) },
	)

	if len(prefixes) == 0 {
		return lr.level
	}

	sort.SliceStable(
		prefixes,
		func(i, j int) bool {
			return len(prefixes[i]) > len(prefixes[j])
		},
	)

	lr.level = lr.config.PrefixLevels[prefixes[0]]

	return lr.level
}

func (lr *LogR) log(level LogLevel, fmtStr string, args ...any) error {
	if level < lr.minLevel(lr.prefix) {
		return nil
	}

	// header
	_, err := fmt.Fprintf(
		lr.w,
		"%s\n",
		fmt.Sprintf(
			"%s %s %s",
			lr.colors[level].Sprintf("%-5s", level.String()),
			color.New(color.FgCyan).Sprintf(
				"%s", time.Now().Format(lr.timeFormat),
			),
			color.New(color.FgMagenta).Sprint(lr.prefix),
		),
	)
	if err != nil {
		return errors.Wrap(err, "failed to write header to log")
	}

	for _, part := range strings.Split(fmt.Sprintf(fmtStr, args...), "\n") {
		_, err = fmt.Fprintf(
			lr.w,
			"%s\n",
			fmt.Sprintf(
				"%s %s",
				lr.colors[level].Sprintf("▶"),
				part,
			),
		)
		if err != nil {
			return errors.Wrap(err, "failed to write part to log")
		}
	}

	return nil
}
