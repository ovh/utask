package noop

import "github.com/ovh/utask/pkg/compress"

const AlgorithmName = "noop"

type noneCompression struct{}

// New returns a new compress.Compression with no compression algorithm.
func New() compress.Compression {
	return &noneCompression{}
}

func (c *noneCompression) Compress(data []byte) ([]byte, error) {
	return data, nil
}

func (c *noneCompression) Decompress(data []byte) ([]byte, error) {
	return data, nil
}
