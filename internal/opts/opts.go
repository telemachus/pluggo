package opts

import (
	"iter"
	"maps"
)

// value allows
type value interface {
	set(string) error
}

// Opt encapsulates a single option.
type Opt struct {
	value    value
	defValue string
	name     string
	isBool   bool
}

// DefValue returns an option's default value as a string.
func (o *Opt) DefValue() string {
	return o.defValue
}

// Name returns the expected command line name of the option.
func (o *Opt) Name() string {
	return o.name
}

// Group stores and manages a set of options.
type Group struct {
	opts   map[string]*Opt
	name   string
	args   []string // arguments remaining after flag parsing
	parsed bool
}

// NewGroup returns a pointer to an option Group ready to use.
func NewGroup(name string) *Group {
	return &Group{
		name: name,
		opts: make(map[string]*Opt, 10),
	}
}

// Name returns the name of the group.
func (g *Group) Name() string {
	return g.name
}

// All iterates over all opts in a group.
func (g *Group) All() iter.Seq[*Opt] {
	return maps.Values(g.opts)
}

// Args returns non-flag arguments from the command line. This method should
// only be called after Parse has finished without error.
func (g *Group) Args() []string {
	return g.args
}
