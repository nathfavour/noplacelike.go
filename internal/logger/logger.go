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

type zapLogger struct {
	*zap.SugaredLogger
}

// Ensure zapLogger implements the Logger interface
func (l *zapLogger) Debug(msg string, fields ...interface{}) {
	l.SugaredLogger.Debugw(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...interface{}) {
	l.SugaredLogger.Infow(msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...interface{}) {
	l.SugaredLogger.Warnw(msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...interface{}) {
	l.SugaredLogger.Errorw(msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...interface{}) {
	l.SugaredLogger.Fatalw(msg, fields...)
}

// New creates a new structured logger
func New() Logger {
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

	return &zapLogger{
		SugaredLogger: logger.Sugar(),
	}
}

// NewDevelopment creates a development logger with pretty printing
func NewDevelopment() Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	return &zapLogger{
		SugaredLogger: logger.Sugar(),
	}
}

// WithFields adds structured fields to the logger
func (l *zapLogger) WithFields(fields map[string]interface{}) Logger {
	var zapFields []interface{}
	for k, v := range fields {
		zapFields = append(zapFields, k, v)
	}
	
	return &zapLogger{
		SugaredLogger: l.SugaredLogger.With(zapFields...),
	}
}