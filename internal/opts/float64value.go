package opts

import (
	"errors"
	"fmt"
	"strconv"
)

type float64Value float64

func numError(err error, s string) error {
	var ne *strconv.NumError
	if !errors.As(err, &ne) {
		return fmt.Errorf("%w", err)
	}

	if errors.Is(ne.Err, strconv.ErrSyntax) {
		return fmt.Errorf("parsing %q: %w", s, strconv.ErrSyntax)
	}

	if errors.Is(ne.Err, strconv.ErrRange) {
		return fmt.Errorf("parsing %q: %w", s, strconv.ErrRange)
	}

	return err
}

func newFloat64Value(val float64, f *float64) *float64Value {
	*f = val
	return (*float64Value)(f)
}

// Float64 creates a new float64 option and binds its default value to f.
// Float64 will panic if name is not a valid option name or if name repeats the
// name of an existing flag.
func (g *Group) Float64(f *float64, name string, defValue float64) {
	if err := validateName("Float64", name); err != nil {
		panic(err)
	}

	fv := newFloat64Value(defValue, f)
	opt := &Opt{
		value:    fv,
		defValue: strconv.FormatFloat(defValue, 'g', -1, 64),
		name:     name,
		isBool:   false,
	}

	if err := g.optAlreadySet(name); err != nil {
		panic(err)
	}
	g.opts[name] = opt
}

// Set assigns s to an float64Value and returns an error if s cannot be parsed
// as a float64.
func (f *float64Value) set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return numError(err, s)
	}

	*f = float64Value(v)

	return nil
}
