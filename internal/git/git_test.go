package git

import (
	"testing"
	"testing/fstest"
)

func TestBranchRef(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		// TODO: add an error field.
		content   string
		branchRef string
	}{
		"refs/head/main": {
			content:   "ref: refs/heads/main\n",
			branchRef: "refs/heads/main",
		},
		"refs/head/master": {
			content:   "ref: refs/heads/master\n",
			branchRef: "refs/heads/master",
		},
		"refs/head/someBranch": {
			content:   "ref: refs/heads/someBranch\n",
			branchRef: "refs/heads/someBranch",
		},
		"detached HEAD": {
			// TODO: branchRef should be "" and err should be non-nil.
			content:   "9fe4d9792bb5aac4d5ec60ff8a37e8160f3de631\n",
			branchRef: "9fe4d9792bb5aac4d5ec60ff8a37e8160f3de631",
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			// TODO: MapFS should include ".git" already.
			mockFS := fstest.MapFS{
				".git/HEAD": &fstest.MapFile{
					Data: []byte(tc.content),
				},
			}

			repo := NewRepoWithFS(mockFS)
			got, err := repo.BranchRef()
			if err != nil {
				t.Fatalf("BranchRef() error: %s", err)
			}

			if got != tc.branchRef {
				t.Errorf("BranchRef() = %q; want %q", got, tc.branchRef)
			}
		})
	}
}

func TestBranchName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		// TODO: add an error field.
		content    string
		branchName string
	}{
		"refs/head/main": {
			content:    "ref: refs/heads/main\n",
			branchName: "main",
		},
		"refs/head/master": {
			content:    "ref: refs/heads/master\n",
			branchName: "master",
		},
		"refs/head/someBranch": {
			content:    "ref: refs/heads/someBranch\n",
			branchName: "someBranch",
		},
		"detached HEAD": {
			// TODO: error should be non-nil here.
			content:    "9fe4d9792bb5aac4d5ec60ff8a37e8160f3de631\n",
			branchName: "",
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			mockFS := fstest.MapFS{
				// TODO: MapFS should include .git already.
				".git/HEAD": &fstest.MapFile{
					Data: []byte(tc.content),
				},
			}

			repo := NewRepoWithFS(mockFS)
			got, err := repo.BranchName()
			if err != nil {
				t.Fatalf("BranchName() error: %s", err)
			}

			if got != tc.branchName {
				t.Errorf("BranchName() = %q; want %q", got, tc.branchName)
			}
		})
	}
}

func TestHeadDigest(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		headContent    string
		branchRef      string
		branchContent  string
		expectedDigest string
	}{
		"main branch": {
			headContent:    "ref: refs/heads/main\n",
			branchRef:      "refs/heads/main",
			branchContent:  "089293721eb4f586907a17a18783fee1eae2f445\n",
			expectedDigest: "089293721eb4f586907a17a18783fee1eae2f445",
		},
		"master branch": {
			headContent:    "ref: refs/heads/master\n",
			branchRef:      "refs/heads/master",
			branchContent:  "fc558a102bc00e11580aef6033692f92d964a638\n",
			expectedDigest: "fc558a102bc00e11580aef6033692f92d964a638",
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			// TODO: the MapFS should already contain ".git".
			mockFS := fstest.MapFS{
				".git/HEAD": &fstest.MapFile{
					Data: []byte(tc.headContent),
				},
			}

			// TODO: the MapFS should already contain ".git".
			branchPath := ".git/" + tc.branchRef
			mockFS[branchPath] = &fstest.MapFile{
				Data: []byte(tc.branchContent),
			}

			repo := NewRepoWithFS(mockFS)
			digest, err := repo.HeadDigest()
			if err != nil {
				t.Fatalf("HeadDigest() error: %s", err)
			}

			got := string(digest)
			if got != tc.expectedDigest {
				t.Errorf("HeadDigest() = %q; want %q", got, tc.expectedDigest)
			}
		})
	}
}

func TestDigestEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		digest1  []byte
		digest2  []byte
		expected bool
	}{
		"identical digests": {
			digest1:  []byte("089293721eb4f586907a17a18783fee1eae2f445"),
			digest2:  []byte("089293721eb4f586907a17a18783fee1eae2f445"),
			expected: true,
		},
		"different digests": {
			digest1:  []byte("089293721eb4f586907a17a18783fee1eae2f445"),
			digest2:  []byte("fc558a102bc00e11580aef6033692f92d964a638"),
			expected: false,
		},
		"empty digest": {
			digest1:  []byte("089293721eb4f586907a17a18783fee1eae2f445"),
			digest2:  []byte{},
			expected: false,
		},
	}

	for msg, tc := range testCases {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			d1 := Digest(tc.digest1)
			d2 := Digest(tc.digest2)

			got := d1.Equals(d2)
			if got != tc.expected {
				t.Errorf("Digest(%q).Equals(Digest(%q)) = %v; want %v",
					string(tc.digest1), string(tc.digest2), got, tc.expected)
			}
		})
	}
}
