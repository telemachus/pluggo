package git

import (
	"testing"
)

func TestDigestFrom(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		headFile string
		rhString string
	}{
		"refs/head/main": {
			headFile: "testdata/mainHeadFile",
			rhString: "refs/heads/main",
		},
		"refs/head/master": {
			headFile: "testdata/masterHeadFile",
			rhString: "refs/heads/master",
		},
		"refs/head/somebranch": {
			headFile: "testdata/branchHeadFile",
			rhString: "refs/heads/somebranch",
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			got, err := branchRef(tc.headFile)
			if err != nil {
				t.Fatalf("%s: %s", tc.headFile, err)
			}

			if got != tc.rhString {
				t.Errorf("branchRef(%q) = %q; want %q", tc.headFile, got, tc.rhString)
			}
		})
	}
}

func TestDigestEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		digestBefore string
		digestAfter  string
		expected     bool
	}{
		"original should equal iteself": {
			digestBefore: "testdata/originalDigest",
			digestAfter:  "testdata/originalDigest",
			expected:     true,
		},
		"original should equal an identical digest in another file": {
			digestBefore: "testdata/originalDigest",
			digestAfter:  "testdata/identicalDigest",
			expected:     true,
		},
		"original should not equal a different digest": {
			digestBefore: "testdata/originalDigest",
			digestAfter:  "testdata/differentDigest",
			expected:     false,
		},
		"original should not equal an empty digest": {
			digestBefore: "testdata/originalDigest",
			digestAfter:  "testdata/emptyDigest",
			expected:     false,
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			digestBefore, err := digestFrom(tc.digestBefore)
			if err != nil {
				t.Fatalf("%s: %s", tc.digestBefore, err)
			}

			digestAfter, err := digestFrom(tc.digestAfter)
			if err != nil {
				t.Fatalf("%s: %s", tc.digestAfter, err)
			}

			got := digestBefore.Equals(digestAfter)
			if got != tc.expected {
				t.Errorf("digestBefore.Equals(digestAfter) = %v; want %v", got, tc.expected)
			}
		})
	}
}
