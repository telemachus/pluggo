package opts

type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

// String creates a new string option with the default value and binds that
// option to s. String will panic if name is not a valid option name or if name
// repeats the name of an existing flag.
func (g *Group) String(s *string, name, defValue string) {
	if err := validateName("String", name); err != nil {
		panic(err)
	}

	sv := newStringValue(defValue, s)
	opt := &Opt{
		value:    sv,
		defValue: defValue,
		name:     name,
		isBool:   false,
	}

	if err := g.optAlreadySet(name); err != nil {
		panic(err)
	}
	g.opts[name] = opt
}

// Set assigns s to a stringValue. The method always returns a nil error since
// there is no parsing to fail.
func (s *stringValue) set(val string) error {
	*s = stringValue(val)
	return nil
}
