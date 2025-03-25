/*
Package opts creates and parses command line flags.

An opts [Group] stores a set of options and gives the user access to their
methods. The user gets access to a Group by calling [NewGroup]. The parameter
passed to NewGroup becomes the Group's [Name]. As an example, we can imagine
a tool that validates and optionally corrects the naming conventions used in one
or more files.

	og := opts.NewGroup("caser")

The user adds options (aka, flags) to the Group by calling typely named methods on the Group.  The examples below should make this clear.

	cfg := struct {
		rcfile     string
		convention string
		level      uint
		dryRun     bool
		write      bool
	}{}

	og.String(&cfg.rcfile, "rcfile", "caser.ini")
	og.String(&cfg.convention, "convention", "camel")
	og.Uint(&cfg.level, "strictness", 3)
	og.Bool(&cfg.dryRun, "dry-run")
	og.Bool(&cfg.write, "write")

Once the user has defined all necessary options, they can call [Parse]. Parse
returns an error or nil. If Parse returns without error, then the variables
defined for options are ready for use however the programmer chooses. If Parse
does return an error, then no options are defined and none of the variables
associated with those options can be trusted for use. Parse should receive
a slice of strings containing everything on the command line other than the
name of the program. In other words, you want to pass os.Args[1:], though you
may want to tweak those values before you pass them.
*/
package opts
