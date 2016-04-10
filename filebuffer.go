// filebuffer is a package implementing a few file like interfaces
// backed by a byte buffer.
// Reader ReaderAt Writer Seeker
// TODO: Closer
package filebuffer

import (
	"bytes"
	"errors"
	"io"
)

// Buffer implements interfaces implemented by files.
// The main purpose of this type is to have an in memory replacement for a
// file.
type Buffer struct {
	// Buff is the backing buffer
	Buff *bytes.Buffer
	// Index indicates where in the buffer we are at
	Index int64
}

// New returns a new populated Buffer
func New(b []byte) *Buffer {
	return &Buffer{Buff: bytes.NewBuffer(b)}
}

// Bytes returns the bytes
func (fbuff *Buffer) Bytes() []byte {
	if fbuff.Index >= int64(fbuff.Buff.Len()) {
		return []byte{}
	}
	return fbuff.Buff.Bytes()[fbuff.Index:]
}

// Read reads up to len(p) bytes into p. It returns the number of bytes read (0 <= n <= len(p))
// and any error encountered. Even if Read returns n < len(p), it may use all of p as scratch
// space during the call. If some data is available but not len(p) bytes, Read conventionally
// returns what is available instead of waiting for more.

// When Read encounters an error or end-of-file condition after successfully reading n > 0 bytes,
// it returns the number of bytes read. It may return the (non-nil) error from the same call or
// return the error (and n == 0) from a subsequent call. An instance of this general case is
// that a Reader returning a non-zero number of bytes at the end of the input stream may return
// either err == EOF or err == nil. The next Read should return 0, EOF.
func (fbuff *Buffer) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	if fbuff.Index >= int64(fbuff.Buff.Len()) {
		return 0, io.EOF
	}
	n, err = bytes.NewBuffer(fbuff.Buff.Bytes()[fbuff.Index:]).Read(b)
	fbuff.Index += int64(n)

	return n, err
}

// ReadAt reads len(p) bytes into p starting at offset off in the underlying input source.
// It returns the number of bytes read (0 <= n <= len(p)) and any error encountered.
//
// When ReadAt returns n < len(p), it returns a non-nil error explaining why more bytes were not returned.
// In this respect, ReadAt is stricter than Read.
//
// Even if ReadAt returns n < len(p), it may use all of p as scratch space during the call.
// If some data is available but not len(p) bytes, ReadAt blocks until either all the data is available or an error occurs.
// In this respect ReadAt is different from Read.
//
// If the n = len(p) bytes returned by ReadAt are at the end of the input source,
// ReadAt may return either err == EOF or err == nil.
//
// If ReadAt is reading from an input source with a seek offset,
// ReadAt should not affect nor be affected by the underlying seek offset.
// Clients of ReadAt can execute parallel ReadAt calls on the same input source.
func (fbuff *Buffer) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errors.New("filebuffer.ReadAt: negative offset")
	}
	reqLen := len(p)
	buffLen := int64(fbuff.Buff.Len())
	if off >= buffLen {
		return 0, io.EOF
	}

	n = copy(p, fbuff.Buff.Bytes()[off:])
	if n < reqLen {
		err = io.EOF
	}
	return n, err
}

func (f *Buffer) Write(p []byte) (int, error) {
	if f.Index < 0 {
		return 0, io.EOF
	}
	buffLen := int64(f.Buff.Len())
	if f.Index >= buffLen {
		return 0, io.EOF
	}
	n, err := bytes.NewBuffer(f.Buff.Bytes()[f.Index:]).Write(p)
	if err == nil {
		f.Index = int64(f.Buff.Len())
	}

	return n, err
}

func (fbuff *Buffer) Seek(offset int64, whence int) (idx int64, err error) {

	var abs int64
	switch whence {
	case 0:
		abs = offset
	case 1:
		abs = int64(fbuff.Index) + offset
	case 2:
		abs = int64(fbuff.Buff.Len()) + offset
	default:
		return 0, errors.New("filebuffer.Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("filebuffer.Seek: negative position")
	}
	fbuff.Index = abs
	return abs, nil
}
