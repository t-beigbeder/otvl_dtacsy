package csavsftp

import (
	"io"

	"github.com/pkg/sftp"
)

type sftpReader struct {
	reader *sftp.File
	cb     func()
	closed bool
}

// Close implements [io.ReadCloser].
func (s *sftpReader) Close() error {
	if s.closed {
		return s.reader.Close()
	}
	s.cb()
	s.closed = true
	return s.reader.Close()
}

// Read implements [io.ReadCloser].
func (s *sftpReader) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}

var _ io.ReadCloser = &sftpReader{}
