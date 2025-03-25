package cli

import "path/filepath"

// Repo stores information about a git repository.
type Repo struct {
	URL    string
	Name   string
	Branch string
	Opt    bool
	Pin    bool
}

func (r *Repo) fullPath(dataDir string) string {
	switch r.Opt {
	case true:
		return filepath.Join(dataDir, "opt", r.Name)
	default:
		return filepath.Join(dataDir, "start", r.Name)
	}
}

func (r *Repo) installArgs(fullPath string) []string {
	args := make([]string, 0, 6)

	args = append(args, "clone", "--filter=blob:none")
	if r.Branch != "" {
		args = append(args, "-b", r.Branch)
	}
	args = append(args, r.URL, fullPath)

	return args
}
