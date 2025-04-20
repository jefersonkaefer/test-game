package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func init() {
	Log = logrus.New()

	Log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		Log.SetLevel(logrus.InfoLevel)
	} else {
		Log.SetLevel(level)
	}

	Log.SetOutput(os.Stdout)
}

// Info logs a message at level Info
func Info(args ...interface{}) {
	Log.Info(args...)
}

// Infof logs a formatted message at level Info
func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

// Error logs a message at level Error
func Error(args ...interface{}) {
	Log.Error(args...)
}

// Errorf logs a formatted message at level Error
func Errorf(format string, args ...interface{}) {
	Log.Errorf(format, args...)
}

// Debug logs a message at level Debug
func Debug(args ...interface{}) {
	Log.Debug(args...)
}

// Debugf logs a formatted message at level Debug
func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
}

// Warn logs a message at level Warn
func Warn(args ...interface{}) {
	Log.Warn(args...)
}

// Warnf logs a formatted message at level Warn
func Warnf(format string, args ...interface{}) {
	Log.Warnf(format, args...)
}

// WithFields returns a new entry with fields added
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Log.WithFields(fields)
}
