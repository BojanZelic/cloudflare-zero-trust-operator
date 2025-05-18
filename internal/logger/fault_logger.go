package logger

import (
	"github.com/Southclaws/fault/fctx"
	"github.com/go-logr/logr"
)

type FaultLogger struct {
	logr.LogSink

	inner logr.Logger
}

func NewFaultLogger(logger logr.Logger) logr.Logger {
	sink := FaultLogger{inner: logger}
	return logr.New(sink)
}

func (cl FaultLogger) Init(info logr.RuntimeInfo) {
	return
}

func (cl FaultLogger) Enabled(level int) bool {
	return cl.inner.Enabled()
}

// Info logs an info message with the prefix and a custom field
func (cl FaultLogger) Info(level int, msg string, keysAndValues ...any) {
	// cl.inner.Info(msg, keysAndValues...)
}

// Error logs an error message with the prefix
func (cl FaultLogger) Error(err error, msg string, keysAndValues ...any) {
	//
	ctx := fctx.Unwrap(err)
	if ctx != nil {
		keysAndValues = append(keysAndValues, "context")
		keysAndValues = append(keysAndValues, fctx.Unwrap(err))
	}

	//
	cl.inner.Error(err, msg, keysAndValues...)
}

// WithValues returns a new FaultLogger with additional key-value pairs
func (cl FaultLogger) WithValues(keysAndValues ...any) logr.LogSink {
	newLogger := cl.inner.WithValues(keysAndValues...)
	return NewFaultLogger(newLogger).GetSink()
}

// WithName returns a new FaultLogger with a name
func (cl FaultLogger) WithName(name string) logr.LogSink {
	newLogger := cl.inner.WithName(name)
	return NewFaultLogger(newLogger).GetSink()
}
