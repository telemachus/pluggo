package git

import "os"

// FileReader implements the fs.ReadFileFS interface.
type FileReader interface {
	ReadFile(name string) ([]byte, error)
}

type osFileReader struct{}

func (osFileReader) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// Provide os.ReadFile as a default for production use.
var defaultFileReader FileReader = osFileReader{}
