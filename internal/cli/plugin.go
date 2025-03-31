package cli

import "path/filepath"

// Plugin information about a plugin.
type Plugin struct {
	URL    string
	Name   string
	Branch string
	Opt    bool
	Pin    bool
}

func (plugin *Plugin) fullPath(dataDir string) string {
	switch plugin.Opt {
	case true:
		return filepath.Join(dataDir, "opt", plugin.Name)
	default:
		return filepath.Join(dataDir, "start", plugin.Name)
	}
}

func (plugin *Plugin) installArgs(fullPath string) []string {
	args := make([]string, 0, 6)

	args = append(args, "clone", "--filter=blob:none")
	if plugin.Branch != "" {
		args = append(args, "-b", plugin.Branch)
	}
	args = append(args, plugin.URL, fullPath)

	return args
}
