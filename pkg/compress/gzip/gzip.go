package gzip

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/ovh/utask/pkg/compress"
)

const AlgorithmName = "gzip"

type gzipCompression struct{}

// New returns a new compress.Compression with gzip as the compression algorithm.
func New() compress.Compression {
	return &gzipCompression{}
}

// Compress transforms data into a compressed form.
func (c *gzipCompression) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	if _, err := zw.Write(data); err != nil {
		_ = zw.Close()
		return nil, err
	}

	// Need to close the gzip writer before accessing the underlying bytes
	_ = zw.Close()
	return buf.Bytes(), nil
}

// Decompress transforms compressed form into an uncompressed form.
func (c *gzipCompression) Decompress(data []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = zr.Close() }()

	return io.ReadAll(zr)
}
