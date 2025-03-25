package opts

import "errors"

// ErrAlreadyParsed signals an attempt to parse a Group more than once.
var ErrAlreadyParsed = errors.New("opts: option group already parsed")
