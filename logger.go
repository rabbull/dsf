package dsf

import "context"

type Logger interface {
	Context(ctx context.Context) Logger
	Debugf(fmt string, args ...interface{})
	Logf(fmt string, args ...interface{})
	Errorf(fmt string, args ...interface{})
}

type loggerWrapper struct {
	logger Logger
}

func (w *loggerWrapper) D(fmt string, args ...interface{}) {
	if w.logger != nil {
		w.logger.Debugf(fmt, args...)
	}
}

func (w *loggerWrapper) I(fmt string, args ...interface{}) {
	if w.logger != nil {
		w.logger.Logf(fmt, args...)
	}
}

func (w *loggerWrapper) E(fmt string, args ...interface{}) {
	if w.logger != nil {
		w.logger.Errorf(fmt, args...)
	}
}
