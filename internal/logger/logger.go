package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger interface for structured logging
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	WithFields(fields map[string]interface{}) Logger
}

type Logger struct {
	*zap.SugaredLogger
}

// Ensure Logger implements the core.Logger interface
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.SugaredLogger.Debugw(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	l.SugaredLogger.Infow(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.SugaredLogger.Warnw(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...interface{}) {
	l.SugaredLogger.Errorw(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...interface{}) {
	l.SugaredLogger.Fatalw(msg, fields...)
}

// New creates a new structured logger
func New() *Logger {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	
	// Set log level from environment
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		if parsedLevel, err := zapcore.ParseLevel(level); err == nil {
			config.Level = zap.NewAtomicLevelAt(parsedLevel)
		}
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}
}

// NewDevelopment creates a development logger with pretty printing
func NewDevelopment() *Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}
}

// WithFields adds structured fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) core.Logger {
	var zapFields []interface{}
	for k, v := range fields {
		zapFields = append(zapFields, k, v)
	}
	
	return &Logger{
		SugaredLogger: l.SugaredLogger.With(zapFields...),
	}
}