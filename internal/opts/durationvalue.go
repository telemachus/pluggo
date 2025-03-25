package opts

import (
	"time"
)

type durationValue time.Duration

func newDurationValue(val time.Duration, d *time.Duration) *durationValue {
	*d = val
	return (*durationValue)(d)
}

// Duration creates a new time.Duration option with the default value and binds
// that option to b. Duration will panic if name is not a valid option name or
// if name repeats the name of an existing flag.
func (g *Group) Duration(d *time.Duration, name string, defValue time.Duration) {
	if err := validateName("Duration", name); err != nil {
		panic(err)
	}

	dv := newDurationValue(defValue, d)
	opt := &Opt{
		value:    dv,
		defValue: defValue.Round(time.Second).String(),
		name:     name,
		isBool:   false,
	}

	if err := g.optAlreadySet(name); err != nil {
		panic(err)
	}
	g.opts[name] = opt
}

func (d *durationValue) set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = durationValue(v)

	return nil
}
