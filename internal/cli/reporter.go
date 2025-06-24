package cli

import (
	"fmt"
	"strings"
)

type consoleReporter struct {
	spinner     *spinner
	indent      string
	quietWanted bool
}

func newConsoleReporter(indent string, quietWanted bool) *consoleReporter {
	return &consoleReporter{
		indent:      indent,
		quietWanted: quietWanted,
	}
}

func (r *consoleReporter) start(banner string) {
	if !r.quietWanted {
		r.spinner = newSpinner()
		r.spinner.start(banner)
	}
}

func (r *consoleReporter) finish(results *syncResults) {
	if r.spinner != nil {
		r.spinner.stop()
	}

	if r.quietWanted {
		r.printErrorsOnly(results)

		return
	}

	r.printFullSummary(results)
}

func (r *consoleReporter) printFullSummary(results *syncResults) {
	r.printSuccessResults(results)
	r.printErrorResults(results)
}

func (r *consoleReporter) printSuccessResults(results *syncResults) {
	r.printInstalled(results.installed)
	r.printReinstalled(results.reinstalled)
	r.printUpdated(results.updated)
	r.printMoved(results.moved)
	r.printPinned(results.pinned)
	r.printUpToDate(results.upToDate)
	r.printRemoved(results.removed)
}

func (r *consoleReporter) printInstalled(installed []string) {
	for _, plugin := range installed {
		fmt.Printf("%s%s: installed\n", r.indent, plugin)
	}
}

func (r *consoleReporter) printReinstalled(reinstalled []pluginReinstall) {
	for _, reinstall := range reinstalled {
		fmt.Printf("%s%s: reinstalled (%s)\n", r.indent, reinstall.name, reinstall.reason)
	}
}

func (r *consoleReporter) printUpdated(updated []pluginUpdate) {
	for _, update := range updated {
		fmt.Printf("%s%s: updated from %.*s to %.*s\n", r.indent, update.name, hashDisplayLen, update.oldHash, hashDisplayLen, update.newHash)
	}
}

func (r *consoleReporter) printMoved(moved []pluginMove) {
	for _, move := range moved {
		direction := "start/ to opt/"
		if !move.toOpt {
			direction = "opt/ to start/"
		}
		fmt.Printf("%s%s: moved from %s\n", r.indent, move.name, direction)
	}
}

func (r *consoleReporter) printPinned(pinned []string) {
	for _, plugin := range pinned {
		fmt.Printf("%s%s: pinned (no update attempted)\n", r.indent, plugin)
	}
}

func (r *consoleReporter) printUpToDate(upToDate []string) {
	for _, plugin := range upToDate {
		fmt.Printf("%s%s: already up to date\n", r.indent, plugin)
	}
}

func (r *consoleReporter) printRemoved(removed []string) {
	for _, plugin := range removed {
		fmt.Printf("%s%s: removed (not in configuration)\n", r.indent, plugin)
	}
}

func (r *consoleReporter) printErrorResults(results *syncResults) {
	if len(results.errors) > 0 {
		fmt.Printf("%serrors: %s\n", r.indent, strings.Join(results.errors, ", "))
	}
}

func (r *consoleReporter) printErrorsOnly(results *syncResults) {
	for _, plugin := range results.errors {
		fmt.Printf("%sfailed: %s\n", r.indent, plugin)
	}
}
