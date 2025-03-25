package opts_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/telemachus/pluggo/internal/opts"
)

func TestParseMultipleDifferentFlags(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		wantS    string
		args     []string
		postArgs []string
		wantI    int
		wantF    float64
		wantB    bool
	}{
		"Mixed single dash options": {
			args:     []string{"-v", "-n", "42", "-x", "3.14", "-f", "test.txt"},
			postArgs: []string{},
			wantB:    true,
			wantI:    42,
			wantF:    3.14,
			wantS:    "test.txt",
		},
		"Mixed double dash options": {
			args:     []string{"--verbose", "--number", "42", "--value", "3.14", "--file", "test.txt"},
			postArgs: []string{},
			wantB:    true,
			wantI:    42,
			wantF:    3.14,
			wantS:    "test.txt",
		},
		"Mixed single and double dash options": {
			args:     []string{"-v", "--number", "42", "-x", "3.14", "--file", "test.txt"},
			postArgs: []string{},
			wantB:    true,
			wantI:    42,
			wantF:    3.14,
			wantS:    "test.txt",
		},
		"Mixed single and double dash options with remaining args": {
			args:     []string{"-v", "-n=42", "--value=3.14", "-f", "test.txt", "foo", "bar"},
			postArgs: []string{"foo", "bar"},
			wantB:    true,
			wantI:    42,
			wantF:    3.14,
			wantS:    "test.txt",
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			var gotB bool
			var gotI int
			var gotF float64
			var gotS string

			og := opts.NewGroup("test-parsing")
			og.Bool(&gotB, "v")
			og.Bool(&gotB, "verbose")
			og.Int(&gotI, "n", 0)
			og.Int(&gotI, "number", 0)
			og.Float64(&gotF, "x", 0.0)
			og.Float64(&gotF, "value", 0.0)
			og.String(&gotS, "f", "")
			og.String(&gotS, "file", "")

			err := og.Parse(tc.args)
			if err != nil {
				t.Fatalf("after err := og.Parse(%+v), err == %v; want nil", tc.args, err)
			}

			if gotB != tc.wantB {
				t.Errorf("after og.Parse(%+v), bool = %t; want %t", tc.args, gotB, tc.wantB)
			}
			if gotI != tc.wantI {
				t.Errorf("after og.Parse(%+v), int = %d; want %d", tc.args, gotI, tc.wantI)
			}
			if gotF != tc.wantF {
				t.Errorf("after og.Parse(%+v), float = %g; want %g", tc.args, gotF, tc.wantF)
			}
			if gotS != tc.wantS {
				t.Errorf("after og.Parse(%+v), string = %q; want %q", tc.args, gotS, tc.wantS)
			}

			postArgs := og.Args()
			if diff := cmp.Diff(tc.postArgs, postArgs); diff != "" {
				t.Errorf("og.Parse(%+v); (-want +got):\n%s", tc.args, diff)
			}
		})
	}
}
