package opts_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/telemachus/pluggo/internal/opts"
)

func TestParseNoFlags(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		args     []string
		postArgs []string
	}{
		"Empty args": {args: []string{}, postArgs: []string{}},
		"-- should not be in fs.Args()": {
			args:     []string{"--", "foo", "bar"},
			postArgs: []string{"foo", "bar"},
		},
		"- should be in fs.Args()": {
			args:     []string{"-", "foo", "bar"},
			postArgs: []string{"-", "foo", "bar"},
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			g := opts.NewGroup("test-parsing")

			err := g.Parse(tc.args)
			if err != nil {
				t.Fatalf("after err := g.Parse(%+v), err == %v; want nil", tc.args, err)
			}

			postArgs := g.Args()
			if diff := cmp.Diff(tc.postArgs, postArgs); diff != "" {
				t.Errorf("g.Parse(%+v); (-want +got):\n%s", tc.args, diff)
			}
		})
	}
}

func TestParseInvalidFlags(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		args []string
	}{
		"Unknown short flag": {args: []string{"-x"}},
		"Unknown long flag":  {args: []string{"--unknown"}},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()
			g := opts.NewGroup("test-parsing")
			err := g.Parse(tc.args)
			if err == nil {
				t.Errorf("after g.Parse(%+v), err == nil; want error", tc.args)
			}
		})
	}
}
