package cli

type opResult uint8

const (
	opNone        opResult = 0
	opMoved       opResult = 1 << iota // 2
	opUpdated                          // 4
	opPinned                           // 8
	opError                            // 16
	opInstalled                        // 32
	opReinstalled                      // 64
	opRemoved                          // 128
)

type result struct {
	plugin   string
	reason   string
	toOpt    bool
	opResult opResult
}

func (or opResult) has(op opResult) bool {
	return or&op != 0
}

func (or *opResult) set(op opResult) {
	*or |= op
}
