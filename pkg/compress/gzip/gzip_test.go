package gzip_test

import (
	"testing"

	"github.com/ovh/utask/pkg/compress/gzip"
	"github.com/ovh/utask/pkg/compress/tests"
)

func TestCompression(t *testing.T) {
	tests.CompressionTests(t, gzip.New())
}
