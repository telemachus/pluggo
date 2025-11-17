package cli

import "sync/atomic"

type stats struct {
	warnings atomic.Int64
	errors   atomic.Int64
}

func (s *stats) incWarn() {
	s.warnings.Add(1)
}

func (s *stats) incError() {
	s.errors.Add(1)
}

func (s *stats) snapshot() (warns, errs int64) {
	return s.warnings.Load(), s.errors.Load()
}
