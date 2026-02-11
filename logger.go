package main

import (
	"log"
	"os"
)

type Logger interface {
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
}

type StdLogger struct {
	logger *log.Logger
	debug  bool
}

func NewLogger(debug bool) Logger {
	return &StdLogger{
		logger: log.New(os.Stderr, "", 0),
		debug:  debug,
	}
}

func (l *StdLogger) Debug(msg string, args ...any) {
	if l.debug {
		l.logger.Printf(msg, args...)
	}
}

func (l *StdLogger) Error(msg string, args ...any) {
	l.logger.Printf(msg, args...)
}
