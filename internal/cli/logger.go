package cli

import (
	"log"
	"os"
)

type logger struct {
	l     *log.Logger
	debug bool
	quiet bool
}

func newLogger(debug, quiet bool) *logger {
	return &logger{
		l:     log.New(os.Stderr, "", 0),
		debug: debug,
		quiet: quiet,
	}
}

func (lg *logger) debugf(format string, args ...any) {
	if !lg.debug {
		return
	}
	lg.l.Printf(format, args...)
}

func (lg *logger) warnf(format string, args ...any) {
	if lg.quiet {
		return
	}
	lg.l.Printf(format, args...)
}

func (lg *logger) errorf(format string, args ...any) {
	lg.l.Printf(format, args...)
}
