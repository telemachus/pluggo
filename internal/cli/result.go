package cli

import (
	"fmt"
	"os"
)

type result struct {
	msg   string
	isErr bool
}

func (r result) publish() {
	if r.isErr {
		fmt.Fprintln(os.Stderr, r.msg)
		return
	}
	fmt.Fprintln(os.Stdout, r.msg)
}

func (r result) publishError() {
	if r.isErr {
		fmt.Fprintln(os.Stderr, r.msg)
	}
}
