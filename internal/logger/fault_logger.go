package logger

import (
	"slices"
	"strings"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/go-logr/logr"
)

type FaultLoggerOptions struct {
	// Until https://github.com/uber-go/zap/pull/1487 gets merged, setting this to `true` prevents printing `errorVerbose` in the log output,
	// which can be undesired (in testing scenarii notably)
	DismissErrorVerbose bool
}

// Special wrapper to adapt Fault errors rendering on log
type FaultLogger struct {
	logr.LogSink

	options FaultLoggerOptions
	inner   logr.Logger
}

func NewFaultLogger(logger logr.Logger, options *FaultLoggerOptions) logr.Logger {
	if options == nil {
		options = &FaultLoggerOptions{}
	}
	sink := FaultLogger{inner: logger, options: *options}
	return logr.New(sink)
}

func (cl FaultLogger) Init(info logr.RuntimeInfo) {
	// nothing to init !
}

func (cl FaultLogger) Enabled(level int) bool {
	return cl.inner.GetSink().Enabled(level)
}

// Info logs an info message with the prefix and a custom field
func (cl FaultLogger) Info(level int, msg string, keysAndValues ...any) {
	cl.inner.GetSink().Info(level, msg, keysAndValues...)
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
	if cl.options.DismissErrorVerbose {
		// reconstruct error stringification of chain - which is normally done automatically when passing err to log.Error(err)
		chain := fault.Flatten(err)
		if chain != nil {
			//
			messages := []string{}
			for _, link := range slices.Backward(chain) {
				messages = append(messages, link.Message)
			}

			//
			cl.inner.Error(fault.New(strings.Join(messages, ": ")), msg, keysAndValues...)
			return
		}
	}

	//
	cl.inner.GetSink().Error(err, msg, keysAndValues...)
}

// WithValues returns a new FaultLogger with additional key-value pairs
func (cl FaultLogger) WithValues(keysAndValues ...any) logr.LogSink {
	// we need to re-wrap here
	newLogger := cl.inner.WithValues(keysAndValues...)
	return NewFaultLogger(newLogger, &cl.options).GetSink()
}

// WithName returns a new FaultLogger with a name
func (cl FaultLogger) WithName(name string) logr.LogSink {
	// we need to re-wrap here
	newLogger := cl.inner.WithName(name)
	return NewFaultLogger(newLogger, &cl.options).GetSink()
}
