package git

import (
	"testing"
)

func TestBranchRef(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		headFile  string
		branchRef string
	}{
		"refs/head/main": {
			headFile:  "testdata/mainHeadFile",
			branchRef: "refs/heads/main",
		},
		"refs/head/master": {
			headFile:  "testdata/masterHeadFile",
			branchRef: "refs/heads/master",
		},
		"refs/head/someBranch": {
			headFile:  "testdata/branchHeadFile",
			branchRef: "refs/heads/someBranch",
		},
		"refs/head/detachedHead": {
			headFile:  "testdata/detachedHeadFile",
			branchRef: "9fe4d9792bb5aac4d5ec60ff8a37e8160f3de631",
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			got, err := BranchRef(tc.headFile)
			if err != nil {
				t.Fatalf("%s: %s", tc.headFile, err)
			}

			if got != tc.branchRef {
				t.Errorf("BranchRef(%q) = %q; want %q", tc.headFile, got, tc.branchRef)
			}
		})
	}
}

func TestBranchName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		headFile   string
		branchName string
	}{
		"refs/head/main": {
			headFile:   "testdata/mainHeadFile",
			branchName: "main",
		},
		"refs/head/master": {
			headFile:   "testdata/masterHeadFile",
			branchName: "master",
		},
		"refs/head/someBranch": {
			headFile:   "testdata/branchHeadFile",
			branchName: "someBranch",
		},
		"refs/head/detachedHead": {
			headFile:   "testdata/detachedHeadFile",
			branchName: "9fe4d9792bb5aac4d5ec60ff8a37e8160f3de631",
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			got, err := BranchName(tc.headFile)
			if err != nil {
				t.Fatalf("%s: %s", tc.headFile, err)
			}

			if got != tc.branchName {
				t.Errorf("BranchName(%q) = %q; want %q", tc.headFile, got, tc.branchName)
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
