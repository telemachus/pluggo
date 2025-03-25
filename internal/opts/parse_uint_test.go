package opts_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/telemachus/pluggo/internal/opts"
)

func TestParseShortUintFlag(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		args     []string
		postArgs []string
		want     uint
	}{
		"Basic value; one dash": {
			args:     []string{"-n", "42"},
			postArgs: []string{},
			want:     42,
		},
		"Zero; one dash": {
			args:     []string{"-n", "0"},
			postArgs: []string{},
			want:     0,
		},
		"Hex value; one dash": {
			args:     []string{"-n", "0xff"},
			postArgs: []string{},
			want:     255,
		},
		"Octal value; one dash": {
			args:     []string{"-n", "0644"},
			postArgs: []string{},
			want:     420,
		},
		"Args after value; one dash": {
			args:     []string{"-n", "42", "foo", "bar"},
			postArgs: []string{"foo", "bar"},
			want:     42,
		},
		"Space separated; double dash": {
			args:     []string{"--number", "42"},
			postArgs: []string{},
			want:     42,
		},
		"With equals; double dash": {
			args:     []string{"--number=42"},
			postArgs: []string{},
			want:     42,
		},
		"Zero; double dash": {
			args:     []string{"--number", "0"},
			postArgs: []string{},
			want:     0,
		},
		"Hex value; double dash": {
			args:     []string{"--number=0xff"},
			postArgs: []string{},
			want:     255,
		},
		"Octal value; double dash": {
			args:     []string{"--number=0644"},
			postArgs: []string{},
			want:     420,
		},
		"Args after value; double dash": {
			args:     []string{"--number", "42", "foo", "bar"},
			postArgs: []string{"foo", "bar"},
			want:     42,
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			var got uint
			g := opts.NewGroup("test-parsing")
			g.Uint(&got, "n", 0)
			g.Uint(&got, "number", 0)

			err := g.Parse(tc.args)
			if err != nil {
				t.Fatalf("after err := g.Parse(%+v), err == %v; want nil", tc.args, err)
			}

			if got != tc.want {
				t.Errorf("after g.Parse(%+v), got = %d; want %d", tc.args, got, tc.want)
			}

			postArgs := g.Args()
			if diff := cmp.Diff(tc.postArgs, postArgs); diff != "" {
				t.Errorf("g.Parse(%+v); (-want +got):\n%s", tc.args, diff)
			}
		})
	}
}

func TestParseUintErrors(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		args []string
	}{
		"Short no value": {
			args: []string{"-n"},
		},
		"Long no value": {
			args: []string{"--number"},
		},
		"Short invalid value": {
			args: []string{"-n", "xyz"},
		},
		"Long invalid value": {
			args: []string{"--number", "xyz"},
		},
		"Long equals no value": {
			args: []string{"--number="},
		},
		"Long equals invalid": {
			args: []string{"--number=xyz"},
		},
		"Short negative value": {
			args: []string{"-n", "-42"},
		},
		"Long negative value": {
			args: []string{"--number=-42"},
		},
		"Short float value": {
			args: []string{"-n", "3.14"},
		},
		"Long float value": {
			args: []string{"--number=3.14"},
		},
		"Long multiple equals": {
			args: []string{"--number=42=13"},
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			var got uint
			g := opts.NewGroup("test-parsing")
			g.Uint(&got, "n", 0)
			g.Uint(&got, "number", 0)

			err := g.Parse(tc.args)
			if err == nil {
				t.Errorf("after g.Parse(%+v), err == nil; want error", tc.args)
			}
		})
	}
}
