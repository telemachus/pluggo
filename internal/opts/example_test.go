package opts

func Example_typical() {
	// This example demonstrates how to create an options group and define
	// options for it.
	cfg := struct {
		rcfile     string
		convention string
		level      uint
		dryRun     bool
		write      bool
	}{}

	og := NewGroup("caser")
	og.String(&cfg.rcfile, "rcfile", "caser.ini")
	og.String(&cfg.convention, "convention", "camel")
	og.Uint(&cfg.level, "strictness", 3)
	og.Bool(&cfg.dryRun, "dry-run")
	og.Bool(&cfg.write, "write")
	// Output:
}
