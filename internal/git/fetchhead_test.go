package git_test

import (
	"testing"

	"github.com/telemachus/pluggo/internal/git"
)

func TestFetchHeadEquality(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fhBefore string
		fhAfter  string
		expected bool
	}{
		"original should equal identical": {
			fhBefore: "testdata/originalFetchHead",
			fhAfter:  "testdata/identicalFetchHead",
			expected: true,
		},
		"original should not equal different": {
			fhBefore: "testdata/originalFetchHead",
			fhAfter:  "testdata/differentFetchHead",
			expected: false,
		},
		"original should not equal longer": {
			fhBefore: "testdata/originalFetchHead",
			fhAfter:  "testdata/longerFetchHead",
			expected: false,
		},
		"original should not equal shorter": {
			fhBefore: "testdata/originalFetchHead",
			fhAfter:  "testdata/shorterFetchHead",
			expected: false,
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			fhBefore, err := git.NewFetchHead(tc.fhBefore)
			if err != nil {
				t.Fatalf("%s: %s", tc.fhBefore, err)
			}

			fhAfter, err := git.NewFetchHead(tc.fhAfter)
			if err != nil {
				t.Fatalf("%s: %s", tc.fhAfter, err)
			}

			got := fhBefore.Equals(fhAfter)
			if got != tc.expected {
				t.Errorf("fhBefore.Equals(fhAfter) = %v; want %v", got, tc.expected)
			}
		})
	}
}
