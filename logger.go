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

func NewLogger(config Config) Logger {
	return &StdLogger{
		logger: log.New(os.Stderr, "", 0),
		debug:  config.Debug,
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
