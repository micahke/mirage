package clients

import (
	"go.uber.org/zap"
)

type Logger interface {
	Named(scopes map[string]string) Logger
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
}

type LoggingClient struct {
	scopes map[string]string
	sugar  *zap.SugaredLogger
}

// NewLogClient initializes a new LoggingClient with optional scopes
func NewLogClient(scopes map[string]string) *LoggingClient {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	return &LoggingClient{
		scopes: scopes,
		sugar:  sugar,
	}
}

// Named creates a new Logger with additional or updated scopes
func (l *LoggingClient) Named(scopes map[string]string) Logger {
	// Merge existing scopes with new ones
	newScopes := make(map[string]string)
	for k, v := range l.scopes {
		newScopes[k] = v
	}
	for k, v := range scopes {
		newScopes[k] = v
	}
	return &LoggingClient{
		scopes: newScopes,
		sugar:  l.sugar,
	}
}

// Info logs an informational message
func (l *LoggingClient) Info(msg string, keysAndValues ...interface{}) {
	l.sugar.Infow(msg, append(l.scopeFields(), keysAndValues...)...)
}

// Warn logs a warning message
func (l *LoggingClient) Warn(msg string, keysAndValues ...interface{}) {
	l.sugar.Warnw(msg, append(l.scopeFields(), keysAndValues...)...)
}

// Error logs an error message
func (l *LoggingClient) Error(msg string, keysAndValues ...interface{}) {
	l.sugar.Errorw(msg, append(l.scopeFields(), keysAndValues...)...)
}

// Debug logs a debug message
func (l *LoggingClient) Debug(msg string, keysAndValues ...interface{}) {
	l.sugar.Debugw(msg, append(l.scopeFields(), keysAndValues...)...)
}

// Fatal logs a fatal message and exits
func (l *LoggingClient) Fatal(msg string, keysAndValues ...interface{}) {
	l.sugar.Fatalw(msg, append(l.scopeFields(), keysAndValues...)...)
}

// scopeFields converts the scope map into structured log fields
func (l *LoggingClient) scopeFields() []interface{} {
	var fields []interface{}
	for k, v := range l.scopes {
		fields = append(fields, k, v)
	}
	return fields
}
