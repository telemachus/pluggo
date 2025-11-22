package cli

// pluginSpec represents a plugin specified in the user's configuration file.
type pluginSpec struct {
	URL    string `json:"url"`
	Name   string `json:"name"`
	Branch string `json:"branch"`
	Opt    bool   `json:"opt,omitempty"`
	Pinned bool   `json:"pin,omitempty"`
}

// pluginState represents a plugin installed locally.
type pluginState struct {
	name      string
	directory string
	url       string
	branch    string
	hash      digest
}

// status describes the final state of a plugin after processing.
type status uint8

const (
	unknown status = iota
	installed
	reinstalled
	updated
	removed
	unchanged
)

// result contains the result of a plugin operation.
type result struct {
	err     error
	plugin  string
	movedTo string // "start" or "opt"; "" if not moved
	reason  string // Additional context (e.g., "switching branches")
	status  status
	pinned  bool
}
