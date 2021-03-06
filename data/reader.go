package data

import (
	"io"
)

type Reader interface {
	io.ByteReader
	io.Reader
	Len() int
}

type LimitByteReader struct {
	R Reader // underlying reader
	N int64  // max bytes remaining
}

func LimitedByteReader(r Reader, n int64) *LimitByteReader {
	return &LimitByteReader{r, n}
}

func (l *LimitByteReader) Len() int {
	return int(l.N)
}

func (l *LimitByteReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.Read(p)
	l.N -= int64(n)
	return
}

func (l *LimitByteReader) ReadByte() (c byte, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}
	l.N--
	return l.R.ReadByte()
}
