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

// DefaultFileReader provides os.ReadFile for production use.
var DefaultFileReader FileReader = osFileReader{}
