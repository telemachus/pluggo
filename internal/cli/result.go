package cli

type result struct {
	plugin string
	detail resultDetail
	kind   resultKind
}

type resultKind int

const (
	resultInstalled resultKind = iota
	resultReinstalled
	resultUpdated
	resultMoved
	resultPinned
	resultUpToDate
	resultRemoved
	resultError
)

type resultDetail struct {
	reason     string
	oldHash    string
	newHash    string
	movedToOpt bool
}
