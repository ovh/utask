package noop_test

import (
	"testing"

	"github.com/ovh/utask/pkg/compress/noop"
	"github.com/ovh/utask/pkg/compress/tests"
)

func TestCompression(t *testing.T) {
	tests.CompressionTests(t, noop.New())
}
