package opts_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/telemachus/pluggo/internal/opts"
)

func TestParseFloat64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		args     []string
		postArgs []string
		want     float64
	}{
		"Integer value; single dash": {
			args:     []string{"-x", "42"},
			postArgs: []string{},
			want:     42.0,
		},
		"Decimal value; single dash": {
			args:     []string{"-x", "3.14"},
			postArgs: []string{},
			want:     3.14,
		},
		"Scientific notation; single dash": {
			args:     []string{"-x", "1e-2"},
			postArgs: []string{},
			want:     0.01,
		},
		"Negative value; single dash": {
			args:     []string{"-x", "-3.14"},
			postArgs: []string{},
			want:     -3.14,
		},
		"Args after value; single dash": {
			args:     []string{"-x", "3.14", "foo", "bar"},
			postArgs: []string{"foo", "bar"},
			want:     3.14,
		},
		"Space separated decimal; double dash": {
			args:     []string{"--value", "3.14"},
			postArgs: []string{},
			want:     3.14,
		},
		"With equals decimal; double dash": {
			args:     []string{"--value=3.14"},
			postArgs: []string{},
			want:     3.14,
		},
		"Scientific notation; double dash": {
			args:     []string{"--value=1e-2"},
			postArgs: []string{},
			want:     0.01,
		},
		"Integer value; double dash": {
			args:     []string{"--value", "42"},
			postArgs: []string{},
			want:     42.0,
		},
		"Negative value; double dash": {
			args:     []string{"--value", "-3.14"},
			postArgs: []string{},
			want:     -3.14,
		},
		"Args after value; double dash": {
			args:     []string{"--value", "3.14", "foo", "bar"},
			postArgs: []string{"foo", "bar"},
			want:     3.14,
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			var got float64
			g := opts.NewGroup("test-parsing")
			g.Float64(&got, "x", 0.0)
			g.Float64(&got, "value", 0.0)

			err := g.Parse(tc.args)
			if err != nil {
				t.Fatalf("after err := g.Parse(%+v), err == %v; want nil", tc.args, err)
			}

			if got != tc.want {
				t.Errorf("after g.Parse(%+v), got = %g; want %g", tc.args, got, tc.want)
			}

			postArgs := g.Args()
			if diff := cmp.Diff(tc.postArgs, postArgs); diff != "" {
				t.Errorf("g.Parse(%+v); (-want +got):\n%s", tc.args, diff)
			}
		})
	}
}

func TestParseFloat64Errors(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		args []string
	}{
		"Short no value": {
			args: []string{"-x"},
		},
		"Long no value": {
			args: []string{"--value"},
		},
		"Short invalid value": {
			args: []string{"-x", "xyz"},
		},
		"Long invalid value": {
			args: []string{"--value", "xyz"},
		},
		"Long equals no value": {
			args: []string{"--value="},
		},
		"Long equals invalid": {
			args: []string{"--value=xyz"},
		},
		"Invalid scientific notation": {
			args: []string{"--value=1e"},
		},
		"Long multiple equals": {
			args: []string{"--value=3.14=2.718"},
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			var got float64
			g := opts.NewGroup("test-parsing")
			g.Float64(&got, "x", 0.0)
			g.Float64(&got, "value", 0.0)

			err := g.Parse(tc.args)
			if err == nil {
				t.Errorf("after g.Parse(%+v), err == nil; want error", tc.args)
			}
		})
	}
}
