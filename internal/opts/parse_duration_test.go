// parse_duration_test.go
package opts_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/telemachus/pluggo/internal/opts"
)

func TestParseShortDurationFlag(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		args     []string
		postArgs []string
		want     time.Duration
	}{
		"Basic seconds; one dash": {
			args:     []string{"-duration", "10s"},
			postArgs: []string{},
			want:     10 * time.Second,
		},
		"Zero; one dash": {
			args:     []string{"-duration", "0s"},
			postArgs: []string{},
			want:     0,
		},
		"Minutes; one dash": {
			args:     []string{"-duration", "5m"},
			postArgs: []string{},
			want:     5 * time.Minute,
		},
		"Hours; one dash": {
			args:     []string{"-duration", "2h"},
			postArgs: []string{},
			want:     2 * time.Hour,
		},
		"Complex duration; one dash": {
			args:     []string{"-duration", "2h30m"},
			postArgs: []string{},
			want:     2*time.Hour + 30*time.Minute,
		},
		"Milliseconds; one dash": {
			args:     []string{"-duration", "1500ms"},
			postArgs: []string{},
			want:     1500 * time.Millisecond,
		},
		"Args after value; one dash": {
			args:     []string{"-duration", "1h", "foo", "bar"},
			postArgs: []string{"foo", "bar"},
			want:     time.Hour,
		},
		"Space separated; two dashes": {
			args:     []string{"--duration", "10s"},
			postArgs: []string{},
			want:     10 * time.Second,
		},
		"With equals; two dashes": {
			args:     []string{"--duration=10s"},
			postArgs: []string{},
			want:     10 * time.Second,
		},
		"Zero; two dashes": {
			args:     []string{"--duration", "0s"},
			postArgs: []string{},
			want:     0,
		},
		"Minutes; two dashes": {
			args:     []string{"--duration=5m"},
			postArgs: []string{},
			want:     5 * time.Minute,
		},
		"Hours; two dashes": {
			args:     []string{"--duration", "2h"},
			postArgs: []string{},
			want:     2 * time.Hour,
		},
		"Complex duration; two dashes": {
			args:     []string{"--duration=2h30m"},
			postArgs: []string{},
			want:     2*time.Hour + 30*time.Minute,
		},
		"Milliseconds; two dashes": {
			args:     []string{"--duration", "1500ms"},
			postArgs: []string{},
			want:     1500 * time.Millisecond,
		},
		"Args after value; two dashes": {
			args:     []string{"--duration", "1h", "foo", "bar"},
			postArgs: []string{"foo", "bar"},
			want:     time.Hour,
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			var got time.Duration
			g := opts.NewGroup("test-parsing")
			g.Duration(&got, "duration", 0)

			err := g.Parse(tc.args)
			if err != nil {
				t.Fatalf("after err := g.Parse(%+v), err == %v; want nil", tc.args, err)
			}

			if got != tc.want {
				t.Errorf("after g.Parse(%+v), got = %v; want %v", tc.args, got, tc.want)
			}

			postArgs := g.Args()
			if diff := cmp.Diff(tc.postArgs, postArgs); diff != "" {
				t.Errorf("g.Parse(%+v); (-want +got):\n%s", tc.args, diff)
			}
		})
	}
}

func TestParseDurationErrors(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		args []string
	}{
		"No value; single dash": {
			args: []string{"-duration"},
		},
		"No value; double dash": {
			args: []string{"--duration"},
		},
		"Invalid value; single dash": {
			args: []string{"-duration", "xyz"},
		},
		"Invalid value; double dash": {
			args: []string{"--duration", "xyz"},
		},
		"Equals without value": {
			args: []string{"--duration="},
		},
		"Equals with invalid value": {
			args: []string{"--duration=xyz"},
		},
		"Missing unit; single dash": {
			args: []string{"-duration", "42"},
		},
		"Missing unit; double dash": {
			args: []string{"--duration=42"},
		},
		"Invalid unit": {
			args: []string{"-duration", "42y"},
		},
		"Negative without number": {
			args: []string{"-duration", "-s"},
		},
		"Long multiple equals": {
			args: []string{"--duration=1h=2h"},
		},
		"Unregistered option": {
			args: []string{"-foo=1h"},
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			var got time.Duration
			g := opts.NewGroup("test-parsing")
			g.Duration(&got, "d", 0)
			g.Duration(&got, "duration", 0)

			err := g.Parse(tc.args)
			if err == nil {
				t.Errorf("after g.Parse(%+v), err == nil; want error", tc.args)
			}
		})
	}
}
