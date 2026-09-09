package dssa

import "io"

type Rights struct {
	Read    bool
	Write   bool
	Execute bool
}

type DataEntry struct {
	IsDir         bool
	Path          string
	Size          int64
	Mtime         int64
	User          int
	UserRights    Rights
	Group         int
	GroupRights   Rights
	OtherRights   Rights
	NoLStat       bool
	IsSymLink     bool
	SymLinkTarget string
	Error         error
	ErrNotExist   bool
	Id            string
	AddMeta       []byte
}

type Dssa interface {
	NewSession() error
	EndSession() error
	List(string) ([]*DataEntry, error)
	Mkdir(*DataEntry) error
	Stat(string) (*DataEntry, error)
	SetStat(_ *DataEntry, noPerm, noMtime bool) error
	Checksum(algos, path_ string) (string, error)
	GetReadCloser(string) (io.ReadCloser, error)
	GetWriteCloser(string) (io.WriteCloser, error)
	Rm(string) error
	Symlink(old, new_ string) error
}

type DssaMaker interface {
	MakeDssa(args ...any) (Dssa, error)
}
