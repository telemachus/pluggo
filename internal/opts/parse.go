package opts

import (
	"fmt"
	"strings"
)

// Parse scans args and sets option values defined in the [Group].
//
// Parse must be called after all options are defined and before option values
// are used. If Parse returns without error, the Group is considered parsed
// and any subsequent calls to Parse will return [ErrAlreadyParsed].
//
// Parsing stops at the first non-option argument. Any remaining arguments can
// be accessed via Args(). Both '-' and '--' are treated as non-option
// arguments, with '--' marking the end of parsing but not appearing in Args(),
// while '-' is preserved in Args(). (By convention, many programs treat '-' as
// stdin, but that is up to the calling program to decide and handle.)
//
// If Parse encounters an invalid or undefined option, it returns an error and
// the Group remains unparsed. The caller may retry with different arguments.
//
// Args should not include the program name. If using `os.Args` directly, the
// caller should pass `os.Args[1:]`.
func (g *Group) Parse(args []string) error {
	if g.parsed {
		return ErrAlreadyParsed
	}

	err := g.parse(args)
	if err == nil {
		g.parsed = true
	} else {
		g.args = []string{}
		// TODO: clear the options map?
	}

	return err
}

func (g *Group) parse(args []string) error {
	g.args = args

	for len(args) > 0 {
		arg := args[0]
		args = args[1:]

		var (
			isEmpty    = arg == ""
			noDash     = !isEmpty && arg[0] != '-'
			singleDash = arg == "-"
			doubleDash = arg == "--"
		)

		switch {
		case isEmpty, noDash, singleDash:
			// In all these cases, parsing is over and args[0]
			// should remain in g.args.
			return nil
		case doubleDash:
			// g.args should not include "--".
			g.args = args
			return nil
		}

		var (
			isLongFlag  = len(arg) > 2 && arg[0:2] == "--"
			isShortFlag = len(arg) > 1 && arg[0] == '-' && !isLongFlag
		)

		var parseErr error
		switch {
		case isShortFlag:
			args, parseErr = g.parseOpt(arg[1:], args)
		case isLongFlag:
			args, parseErr = g.parseOpt(arg[2:], args)
		}
		if parseErr != nil {
			return parseErr
		}

		g.args = args
	}

	return nil
}

func (g *Group) parseOpt(arg string, args []string) ([]string, error) {
	name, value, eqFound := strings.Cut(arg, "=")

	opt, ok := g.opts[name]
	if !ok {
		return nil, fmt.Errorf("unknown option --%s", name)
	}

	if eqFound {
		// TODO: immediately bail out if the opt is a boolean?

		if err := opt.value.set(value); err != nil {
			// Distinguish no value from a bad value.
			if value == "" {
				return nil, fmt.Errorf("--%s: missing value", name)
			}

			return nil, fmt.Errorf("--%s set %q: %w", name, value, err)
		}

		// A string option `--foo=` will not produce an error when
		// calling set above. However, for consistency with other option
		// types, we should return an error indicating that there is no
		// value.
		if value == "" && arg[len(arg)-1] == '=' {
			return nil, fmt.Errorf("--%s: missing value", name)
		}

		return args, nil
	}

	if value == "" {
		switch {
		case opt.isBool:
			value = "true"
		case len(args) > 0:
			value, args = args[0], args[1:]
		default:
			return nil, fmt.Errorf("--%s: missing value", name)
		}
	}

	if err := opt.value.set(value); err != nil {
		return nil, fmt.Errorf("--%s set %q: %w", name, value, err)
	}

	return args, nil
}
