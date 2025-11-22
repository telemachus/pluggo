package cli

import (
	"fmt"
	"os"
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
	s.done.Add(1)
	go func() {
		defer s.done.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		frame := 0
		for {
			select {
			case <-s.stopCh:
				// Clear the line.
				fmt.Fprint(os.Stderr, "\r\033[K")
				return
			case <-ticker.C:
				fmt.Fprintf(os.Stderr, "\r%s %s", s.frames[frame], message)
				frame = (frame + 1) % len(s.frames)
			}
		}
	}()
}

func (s *spinner) stop() {
	close(s.stopCh)
	s.done.Wait()
}

// reporter handles progress display and result formatting.
type reporter struct {
	spinner     *spinner
	indent      string
	quietWanted bool
}

func newReporter(indent string, quietWanted bool) *reporter {
	return &reporter{
		indent:      indent,
		quietWanted: quietWanted,
	}
}

func (r *reporter) start(banner string) {
	if !r.quietWanted {
		r.spinner = newSpinner()
		r.spinner.start(banner)
	}
}

func (r *reporter) finish(results []result) {
	if r.spinner != nil {
		r.spinner.stop()
	}

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
	var msg strings.Builder
	msg.WriteString(r.indent)
	msg.WriteString(res.plugin)
	msg.WriteString(": ")
	msg.WriteString(r.formatStatus(res))

	return msg.String()
}

func (r *reporter) formatStatus(res result) string {
	if res.err != nil {
		return "failed"
	}

	switch res.status {
	case installed:
		return "installed"
	case removed:
		return "removed"
	case reinstalled:
		return r.formatReinstalled(res)
	case updated:
		return r.formatUpdated(res)
	case unchanged:
		return r.formatUnchanged(res)
	default:
		panic(fmt.Sprintf("unreachable: invalid status %d", res.status))
	}
}

func (r *reporter) formatReinstalled(res result) string {
	if res.reason != "" {
		return "reinstalled (" + res.reason + ")"
	}

	return "reinstalled"
}

func (r *reporter) formatUpdated(res result) string {
	if res.movedTo != "" {
		return "updated and moved to " + res.movedTo + "/"
	}

	return "updated"
}

func (r *reporter) formatUnchanged(res result) string {
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
