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

func (r *consoleReporter) finish(results syncResults) {
	if r.spinner != nil {
		r.spinner.stop()
	}

	if r.quietWanted {
		r.printErrorsOnly(results)

		return
	}

	r.printFull(results)
}

func (r *consoleReporter) printFull(results syncResults) {
	for _, res := range results {
		fmt.Println(r.formatResultMessage(res))
	}
}

//nolint:cyclop //This is a straightfoward switch.
func (r *consoleReporter) formatResultMessage(res result) string {
	var msg strings.Builder
	msg.WriteString(r.indent)
	msg.WriteString(res.plugin)
	msg.WriteString(": ")

	direction := " from start/ to opt/"
	const optToStart = " from opt/ to start/"

	switch {
	case res.opResult.has(opError):
		msg.WriteString("failed")
	case res.opResult.has(opInstalled):
		msg.WriteString("installed")
	case res.opResult.has(opRemoved):
		msg.WriteString("removed")
	case res.opResult.has(opReinstalled):
		msg.WriteString("reinstalled")
	case res.opResult.has(opUpdated) && res.opResult.has(opMoved):
		msg.WriteString("updated and moved")
		if !res.toOpt {
			direction = optToStart
		}
		msg.WriteString(direction)
	case res.opResult.has(opUpdated):
		msg.WriteString("updated")
	case res.opResult.has(opPinned) && res.opResult.has(opMoved):
		msg.WriteString("moved")
		if !res.toOpt {
			direction = optToStart
		}
		msg.WriteString(direction)
		msg.WriteString(" and pinned (no update attempted)")
	case res.opResult.has(opPinned):
		msg.WriteString("pinned (no update attempted)")
	case res.opResult.has(opMoved):
		msg.WriteString("moved")
		if !res.toOpt {
			direction = optToStart
		}
		msg.WriteString(direction)
	default: // opUpToDate
		msg.WriteString("already up-to-date")
	}

	return msg.String()
}

func (r *consoleReporter) printErrorsOnly(results syncResults) {
	for _, res := range results {
		if res.opResult.has(opError) {
			fmt.Printf("%sfailed: %s\n", r.indent, res.plugin)
		}
	}
}
