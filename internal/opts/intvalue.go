package opts

import (
	"fmt"
	"strconv"
)

type intValue int

func newIntValue(val int, i *int) *intValue {
	*i = val
	return (*intValue)(i)
}

// Int creates a new integer option and binds its default value to i. Int will
// panic if name is not a valid option name or if name repeats the name of an
// existing flag.
func (g *Group) Int(i *int, name string, defValue int) {
	if err := validateName("Int", name); err != nil {
		panic(err)
	}

	iv := newIntValue(defValue, i)
	opt := &Opt{
		value:    iv,
		defValue: strconv.Itoa(defValue),
		name:     name,
		isBool:   false,
	}

	if err := g.optAlreadySet(name); err != nil {
		panic(err)
	}
	g.opts[name] = opt
}

// Set assigns s to an intValue and returns an error if s cannot be parsed as
// an int.
func (i *intValue) set(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("parsing %q: %w", s, err)
	}

	*i = intValue(v)

	return nil
}
