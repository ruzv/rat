package logger

import (
	"fmt"
	"log"
	"os"

	"private/rat/errors"
)

var defaultLogger *Logger

func NewDefault(logPath string) error {
	l, err := NewLogger(logPath)
	if err != nil {
		return errors.Wrap(err, "failed to create logger")
	}

	defaultLogger = l

	return nil
}

type Logger struct {
	logger  *log.Logger
	logFile *os.File
}

func NewLogger(logPath string) (*Logger, error) {
	logFile, err := os.OpenFile(
		logPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		os.ModePerm,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create log file")
	}

	return &Logger{
		logger: log.New(
			logFile,
			"",
			log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile,
		),
		logFile: logFile,
	}, nil
}

func (l *Logger) log(level, format string, args ...interface{}) {
	l.logger.Printf("%s: %s\n", level, fmt.Sprintf(format, args...))
	l.logFile.Sync()
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log("DEBUG", format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.log("INFO", format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log("WARN", format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log("ERROR", format, args...)
}

func (l *Logger) Close() error {
	return l.logFile.Close()
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

func Close() error {
	return defaultLogger.Close()
}
