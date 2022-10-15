package compress

import (
	"fmt"
	"sync"
)

var (
	compressions    = map[string]Compression{}
	compressionsMut sync.Mutex
)

type Compression interface {
	Compress([]byte) ([]byte, error)
	Decompress([]byte) ([]byte, error)
}

// RegisterAlgorithm registers a custom compression algorithm.
func RegisterAlgorithm(name string, c Compression) error {
	if c == nil {
		return nil
	}
	compressionsMut.Lock()
	defer compressionsMut.Unlock()
	_, found := compressions[name]
	if found {
		return fmt.Errorf("conflicting compression key compressions: %s", name)
	}
	compressions[name] = c
	return nil
}

func Get(name string) (Compression, error) {
	compressionsMut.Lock()
	defer compressionsMut.Unlock()

	c, ok := compressions[name]
	if !ok {
		return nil, fmt.Errorf("%s compression algorithm not found", name)
	}
	return c, nil
}
