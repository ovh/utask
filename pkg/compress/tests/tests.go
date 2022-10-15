package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ovh/utask/pkg/compress"
)

// CompressionTests executes a range of tests to test a `Compression` module.
func CompressionTests(t *testing.T, c compress.Compression) {
	tests := []struct {
		name string
		want string
	}{
		{name: "Hello world", want: "Hello world!"},
		{name: "Empty string", want: " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.Compress([]byte(tt.want))
			require.Nilf(t, err, "Compress(): %s", err)

			got, err = c.Decompress(got)

			require.Nilf(t, err, "Decompress(): %s", err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}
