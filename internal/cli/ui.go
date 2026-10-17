package cli

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

// spinner provides animated feedback during long operations.
type spinner struct {
	stopCh chan struct{}
	frames []string
	done   sync.WaitGroup
}

func newSpinner() *spinner {
	return &spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		stopCh: make(chan struct{}),
	}
}

func (s *spinner) start(message string) {
	s.done.Go(func() {
		tick := time.Tick(100 * time.Millisecond)
		frame := 0
		for {
			select {
			case <-s.stopCh:
				// Clear the line.
				fmt.Fprint(os.Stderr, "\r\033[K")
				return
			case <-tick:
				fmt.Fprintf(os.Stderr, "\r%s %s", s.frames[frame], message)
				frame = (frame + 1) % len(s.frames)
			}
		}
	})
}

func (s *spinner) stop() {
	close(s.stopCh)
	s.done.Wait()
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}

	fi, err := f.Stat()

	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// reporter handles progress display and result formatting.
type reporter struct {
	spinner       *spinner
	indent        string
	quietWanted   bool
	spinnerWanted bool
}

// newReporter returns a reporter. The spinner appears only on a terminal, and
// not in quiet mode or in debug mode, where warnings would garble it.
func newReporter(indent string, quietWanted, debugWanted bool) *reporter {
	return &reporter{
		indent:        indent,
		quietWanted:   quietWanted,
		spinnerWanted: !quietWanted && !debugWanted && isTerminal(os.Stderr),
	}
}

func (r *reporter) start(banner string) {
	if r.spinnerWanted {
		r.spinner = newSpinner()
		r.spinner.start(banner)
	}
}

func (r *reporter) finish(results []result) {
	if r.spinner != nil {
		r.spinner.stop()
	}

	slices.SortFunc(results, func(a, b result) int {
		return strings.Compare(a.plugin, b.plugin)
	})

	if r.quietWanted {
		r.printErrorsOnly(results)
		return
	}

	r.printFull(results)
}

func (r *reporter) printFull(results []result) {
	for _, res := range results {
		fmt.Println(r.formatResult(res))
	}
}

func (r *reporter) printErrorsOnly(results []result) {
	for _, res := range results {
		if res.err != nil {
			fmt.Printf("%sfailed: %s\n", r.indent, res.plugin)
		}
	}
}

func (r *reporter) formatResult(res result) string {
	return r.indent + res.plugin + ": " + formatStatus(res)
}

func formatStatus(res result) string {
	if res.err != nil {
		if res.movedTo != "" {
			return "failed (after moving to " + res.movedTo + "/)"
		}

		return "failed"
	}

	switch res.status {
	case installed:
		return "installed"
	case removed:
		return "removed"
	case reinstalled:
		return formatReinstalled(res)
	case updated:
		return formatUpdated(res)
	case unchanged:
		return formatUnchanged(res)
	default:
		panic(fmt.Sprintf("unreachable: invalid status %d", res.status))
	}
}

func formatReinstalled(res result) string {
	if res.reason != "" {
		return "reinstalled (" + res.reason + ")"
	}

	return "reinstalled"
}

func formatUpdated(res result) string {
	if res.movedTo != "" {
		return "updated and moved to " + res.movedTo + "/"
	}

	return "updated"
}

func formatUnchanged(res result) string {
	// Case 1: the plugin was moved.
	if res.movedTo != "" {
		msg := "moved to " + res.movedTo + "/"
		if res.pinned {
			msg += " and pinned (no update attempted)"
		}

		return msg
	}

	// Case 2: the plugin is pinned and was not moved.
	if res.pinned {
		return "pinned (no update attempted)"
	}

	// Case 3: the plugin wasn't moved and there were no updates.
	return "already up-to-date"
}
