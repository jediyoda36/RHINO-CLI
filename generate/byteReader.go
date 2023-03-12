package generate

import "io"

// in version 0.1.0, we create the template by downloading a zip file from github and unzip it,
// but to manage the templates better, we decide to bind the zip file with our cli,
// which means we store the zip file in a very big byte array in ./zz_filesystem_generated.go.
// here we use zip.NewReader to read the byte array, which requires us to implement io.ReaderAt interface

type byteReader []byte

func (b byteReader) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(b)) {
		return 0, io.EOF
	}

	n := copy(p, b[off:])
	if n < len(p) {
		return n, io.ErrUnexpectedEOF
	}
	return n, nil
}
