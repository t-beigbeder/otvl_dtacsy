package csavsftp

import (
	"io"

	"github.com/pkg/sftp"
)

type sftpWriter struct {
	writer *sftp.File
	cb     func()
	closed bool
}

// Close implements [io.WriteCloser].
func (s *sftpWriter) Close() error {
	if s.closed {
		return s.writer.Close()
	}
	s.cb()
	s.closed = true
	return s.writer.Close()
}

// Write implements [io.WriteCloser].
func (s *sftpWriter) Write(p []byte) (n int, err error) {
	return s.writer.Write(p)
}

var _ io.WriteCloser = &sftpWriter{}
