package rpc

import "log"

type Logger interface {
	Logf(format string, args ...interface{})
}

type nopLogger struct{}

func (n nopLogger) Logf(format string, args ...interface{}) {
}

type stdLogger struct{}

func (n stdLogger) Logf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

var StdLogger = stdLogger{}
