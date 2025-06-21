package git_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/telemachus/pluggo/internal/git"
)

// testFileReader implements git.FileReader for testing
type testFileReader struct {
	files map[string][]byte
}

func (t testFileReader) ReadFile(name string) ([]byte, error) {
	if data, exists := t.files[name]; exists {
		return data, nil
	}
	return nil, fmt.Errorf("file not found: %s", name)
}

func TestGetBranchInfo(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		files          map[string][]byte
		repoDir        string
		expectedBranch string
		expectedHash   string
		expectError    bool
	}{
		"main branch": {
			files: map[string][]byte{
				filepath.Join("repo", ".git", "HEAD"):                  []byte("ref: refs/heads/main\n"),
				filepath.Join("repo", ".git", "refs", "heads", "main"): []byte("abc123def456\n"),
			},
			repoDir:        "repo",
			expectedBranch: "main",
			expectedHash:   "abc123def456",
			expectError:    false,
		},
		"master branch": {
			files: map[string][]byte{
				filepath.Join("repo", ".git", "HEAD"):                    []byte("ref: refs/heads/master\n"),
				filepath.Join("repo", ".git", "refs", "heads", "master"): []byte("def456abc123\n"),
			},
			repoDir:        "repo",
			expectedBranch: "master",
			expectedHash:   "def456abc123",
			expectError:    false,
		},
		"feature branch": {
			files: map[string][]byte{
				filepath.Join("repo", ".git", "HEAD"):                         []byte("ref: refs/heads/feature-xyz\n"),
				filepath.Join("repo", ".git", "refs", "heads", "feature-xyz"): []byte("789xyz456\n"),
			},
			repoDir:        "repo",
			expectedBranch: "feature-xyz",
			expectedHash:   "789xyz456",
			expectError:    false,
		},
		"detached HEAD": {
			files: map[string][]byte{
				filepath.Join("repo", ".git", "HEAD"): []byte("9fe4d9792bb5aac4d5ec60ff8a37e8160f3de631\n"),
			},
			repoDir:     "repo",
			expectError: true,
		},
		"missing HEAD file": {
			files:       map[string][]byte{},
			repoDir:     "repo",
			expectError: true,
		},
		"missing branch ref file": {
			files: map[string][]byte{
				filepath.Join("repo", ".git", "HEAD"): []byte("ref: refs/heads/main\n"),
				// Missing refs/heads/main file
			},
			repoDir:     "repo",
			expectError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testFS := testFileReader{files: tc.files}
			info, err := git.GetBranchInfoWithReader(tc.repoDir, testFS)

			if tc.expectError {
				if err == nil {
					t.Fatalf("GetBranchInfoWithReader(%q, testFS) expected error; got nil", tc.repoDir)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetBranchInfoWithReader(%q, testFS) failed: %v", tc.repoDir, err)
			}

			if info.Branch != tc.expectedBranch {
				t.Errorf("GetBranchInfoWithReader(%q, testFS).Branch = %q; want %q", tc.repoDir, info.Branch, tc.expectedBranch)
			}

			if info.Hash.String() != tc.expectedHash {
				t.Errorf("GetBranchInfo(%q, testFS).Hash = %q; want %q", tc.repoDir, info.Hash.String(), tc.expectedHash)
			}
		})
	}
}

func TestBranchName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		files          map[string][]byte
		repoDir        string
		expectedBranch string
		expectError    bool
		expectDetached bool
	}{
		"main branch": {
			files: map[string][]byte{
				filepath.Join("repo", ".git", "HEAD"): []byte("ref: refs/heads/main\n"),
			},
			repoDir:        "repo",
			expectedBranch: "main",
			expectError:    false,
		},
		"master branch": {
			files: map[string][]byte{
				filepath.Join("repo", ".git", "HEAD"): []byte("ref: refs/heads/master\n"),
			},
			repoDir:        "repo",
			expectedBranch: "master",
			expectError:    false,
		},
		"feature branch": {
			files: map[string][]byte{
				filepath.Join("repo", ".git", "HEAD"): []byte("ref: refs/heads/feature-xyz\n"),
			},
			repoDir:        "repo",
			expectedBranch: "feature-xyz",
			expectError:    false,
		},
		"detached HEAD": {
			files: map[string][]byte{
				filepath.Join("repo", ".git", "HEAD"): []byte("9fe4d9792bb5aac4d5ec60ff8a37e8160f3de631\n"),
			},
			repoDir:        "repo",
			expectError:    true,
			expectDetached: true,
		},
		"missing HEAD file": {
			files:       map[string][]byte{},
			repoDir:     "repo",
			expectError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testFS := testFileReader{files: tc.files}
			branch, err := git.BranchNameWithReader(tc.repoDir, testFS)

			if tc.expectError {
				if err == nil {
					t.Fatalf("BranchNameWithReader(%q, testFS) expected error but got none", tc.repoDir)
				}
				if tc.expectDetached && !errors.Is(err, git.ErrDetachedHead) {
					t.Errorf("BranchNameWithReader(%q, testFS) expected ErrDetachedHead but got %v", tc.repoDir, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("BranchNameWithReader(%q, testFS) failed: %v", tc.repoDir, err)
			}

			if branch != tc.expectedBranch {
				t.Errorf("BranchNameWithReader(%q, testFs) = %q; want %q", tc.repoDir, branch, tc.expectedBranch)
			}
		})
	}
}

func TestHeadDigest(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		files        map[string][]byte
		repoDir      string
		expectedHash string
		expectError  bool
	}{
		"main branch": {
			files: map[string][]byte{
				filepath.Join("repo", ".git", "HEAD"):                  []byte("ref: refs/heads/main\n"),
				filepath.Join("repo", ".git", "refs", "heads", "main"): []byte("abc123def456\n"),
			},
			repoDir:      "repo",
			expectedHash: "abc123def456",
			expectError:  false,
		},
		"master branch": {
			files: map[string][]byte{
				filepath.Join("repo", ".git", "HEAD"):                    []byte("ref: refs/heads/master\n"),
				filepath.Join("repo", ".git", "refs", "heads", "master"): []byte("def456abc123\n"),
			},
			repoDir:      "repo",
			expectedHash: "def456abc123",
			expectError:  false,
		},
		"detached HEAD": {
			files: map[string][]byte{
				filepath.Join("repo", ".git", "HEAD"): []byte("9fe4d9792bb5aac4d5ec60ff8a37e8160f3de631\n"),
			},
			repoDir:     "repo",
			expectError: true,
		},
		"missing HEAD file": {
			files:       map[string][]byte{},
			repoDir:     "repo",
			expectError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testFS := testFileReader{files: tc.files}
			digest, err := git.HeadDigestWithReader(tc.repoDir, testFS)

			if tc.expectError {
				if err == nil {
					t.Fatalf("HeadDigestWithReader(%q, testFS) expected error but got none", tc.repoDir)
				}
				return
			}

			if err != nil {
				t.Fatalf("HeadDigestWithReader(%q, testFS) failed: %v", tc.repoDir, err)
			}

			if digest.String() != tc.expectedHash {
				t.Errorf("HeadDigestWithReader(%q, testFS) = %q; want %q", tc.repoDir, digest.String(), tc.expectedHash)
			}
		})
	}
}

func TestDigestEquals(t *testing.T) {
	t.Parallel()

	// Create digests using the constructor from the git package
	digest1 := git.Digest([]byte("abc123"))
	digest2 := git.Digest([]byte("abc123"))
	digest3 := git.Digest([]byte("def456"))
	var nilDigest git.Digest

	if !digest1.Equals(digest2) {
		t.Error("identical digests should be equal")
	}

	if digest1.Equals(digest3) {
		t.Error("different digests should not be equal")
	}

	if digest1.Equals(nilDigest) {
		t.Error("digest should not equal nil digest")
	}

	if digest1.String() != "abc123" {
		t.Errorf("digest.String() = %q; want %q", digest1.String(), "abc123")
	}
}
