package opts

import (
	"strconv"
)

type uintValue uint

func newUintValue(val uint, p *uint) *uintValue {
	*p = val
	return (*uintValue)(p)
}

// Uint creates a new uint option and binds its default value to u. Uint will
// panic if name is not a valid option name or if name repeats the name of an
// existing flag.
func (g *Group) Uint(u *uint, name string, defValue uint) {
	if err := validateName("Uint", name); err != nil {
		panic(err)
	}

	uv := newUintValue(defValue, u)
	opt := &Opt{
		value:    uv,
		defValue: strconv.FormatUint(uint64(defValue), 10),
		name:     name,
		isBool:   false,
	}

	if err := g.optAlreadySet(name); err != nil {
		panic(err)
	}
	g.opts[name] = opt
}

// Set assigns s to a uintValue and returns an error if s cannot be parsed as
// a uint.
func (u *uintValue) set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err != nil {
		return numError(err, s)
	}

	*u = uintValue(v)

	return nil
}
