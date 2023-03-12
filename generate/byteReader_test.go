package generate

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testBytes byteReader = []byte{
	'H', 'e', 'l', 'l', 'o', ',',
	'W', 'o', 'r', 'l', 'd', '\n',
}

func TestByteReader(t *testing.T) {
	var p []byte = make([]byte, 5)
	testBytes.ReadAt(p, 0)
	assert.Equal(t, "Hello", string(p))

	testBytes.ReadAt(p, 6)
	assert.Equal(t, "World", string(p))
}

func TestByteReaderEOF(t *testing.T) {
	var p []byte = make([]byte, 5)
	n, err := testBytes.ReadAt(p, 32)
	assert.Equal(t, 0, n)
	assert.Equal(t, io.EOF, err)
}

func TestByteReaderUnexpectedEOF(t *testing.T) {
	var p []byte = make([]byte, 5)
	n, err := testBytes.ReadAt(p, 10)

	assert.Equal(t, 2, n)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
	assert.Equal(t, "d\n\x00\x00\x00", string(p))
}
