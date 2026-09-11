package opelog

import (
	"bytes"
	"slices"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/opeloggrpc"
)

type OpeLogManager interface {
	Create(source, target string) error
	Open(readOnly bool) error
	Sync() error
	Close() error
	PutLogicalEntry(relPath string, ole *LogicalEntry) error
	GetLogicalEntry(relPath string) (*LogicalEntry, error)
}

type EventCode opeloggrpc.EventCode

const (
	EVT_UNSPECIFIED = EventCode(opeloggrpc.EventCode_EVT_UNSPECIFIED)
	EVT_ABS         = EventCode(opeloggrpc.EventCode_EVT_ABS)
	EVT_EXIST       = EventCode(opeloggrpc.EventCode_EVT_EXIST)
	EVT_CR_MOD      = EventCode(opeloggrpc.EventCode_EVT_CR_MOD)
	EVT_ATTS_CHG    = EventCode(opeloggrpc.EventCode_EVT_ATTS_CHG)
	EVT_START_DIRUP = EventCode(opeloggrpc.EventCode_EVT_START_DIRUP)
	EVT_END_DIRUP   = EventCode(opeloggrpc.EventCode_EVT_END_DIRUP)
)

type OriginCode opeloggrpc.OriginCode

const (
	ORI_UNSPECIFIED = OriginCode(opeloggrpc.OriginCode_ORI_UNSPECIFIED)
	ORI_LIST        = OriginCode(opeloggrpc.OriginCode_ORI_LIST)
	ORI_STAT        = OriginCode(opeloggrpc.OriginCode_ORI_STAT)
	ORI_READ        = OriginCode(opeloggrpc.OriginCode_ORI_READ)
	ORI_MKDIR       = OriginCode(opeloggrpc.OriginCode_ORI_MKDIR)
	ORI_WRITE       = OriginCode(opeloggrpc.OriginCode_ORI_WRITE)
	ORI_SET_STAT    = OriginCode(opeloggrpc.OriginCode_ORI_SET_STAT)
	ORI_RM          = OriginCode(opeloggrpc.OriginCode_ORI_RM)
)

type Rights struct {
	Read    bool
	Write   bool
	Execute bool
}

func (nr *Rights) Equal(or *Rights) (result bool) {
	if nr.Read != or.Read {
		return
	}
	if nr.Write != or.Write {
		return
	}
	if nr.Execute != or.Execute {
		return
	}
	result = true
	return
}

type StoredEntry struct {
	IsPresent     bool
	IsDir         bool
	Size          int64
	Mtime         int64
	User          int32
	UserRights    *Rights
	Group         int32
	GroupRights   *Rights
	OtherRights   *Rights
	IsSymLink     bool
	SymLinkTarget string
	Children      []string
	AddMeta       []byte
}

func (se *StoredEntry) HasChild(cChild string) bool {
	return slices.Contains(se.Children, cChild)
}

func (nse *StoredEntry) Equal(ose *StoredEntry) (result bool) {
	if nse.IsPresent != ose.IsPresent {
		return
	}
	if nse.IsDir != ose.IsDir {
		return
	}
	if nse.Size != ose.Size {
		return
	}
	if nse.Mtime != ose.Mtime {
		return
	}
	if nse.User != ose.User {
		return
	}
	if !nse.UserRights.Equal(ose.UserRights) {
		return
	}
	if nse.Group != ose.Group {
		return
	}
	if !nse.GroupRights.Equal(ose.GroupRights) {
		return
	}
	if !nse.OtherRights.Equal(ose.OtherRights) {
		return
	}
	if nse.IsSymLink != ose.IsSymLink {
		return
	}
	if nse.SymLinkTarget != ose.SymLinkTarget {
		return
	}
	if len(nse.Children) != len(ose.Children) {
		return
	}
	for _, nChild := range nse.Children {
		if !ose.HasChild(nChild) {
			return
		}
	}
	if !bytes.Equal(nse.AddMeta, ose.AddMeta) {
		return
	}
	result = true
	return
}

func dr2r(dr *dssa.Rights) *Rights {
	return &Rights{Read: dr.Read, Write: dr.Write, Execute: dr.Execute}
}

func FromDataEntry(dse *dssa.DataEntry, children []string) *StoredEntry {
	return &StoredEntry{
		IsDir:         dse.IsDir,
		Size:          dse.Size,
		Mtime:         dse.Mtime,
		User:          int32(dse.User),
		UserRights:    dr2r(&dse.UserRights),
		Group:         int32(dse.Group),
		GroupRights:   dr2r(&dse.GroupRights),
		OtherRights:   dr2r(&dse.OtherRights),
		IsSymLink:     dse.IsSymLink,
		SymLinkTarget: dse.SymLinkTarget,
		Children:      slices.Clone(children),
		AddMeta:       bytes.Clone(dse.AddMeta),
	}
}

type Event struct {
	Kind       EventCode
	Origin     OriginCode
	TimeStamp  int64
	StateIndex int32
	// comma-separated list algo:hexa-of-checksum
	Checksums string
	Error     string
}

type Verification struct {
	TimeStamp    int64
	WithChecksum bool
	NewStatus    *StoredEntry
	NewChecksums string
}

type AggInfo struct {
	Number int64
	Size   int64
}

type ComputedStats struct {
	TimeStamp        int64
	SourceListOrStat *AggInfo
	TargetListOrStat *AggInfo
	Read             *AggInfo
	Create           *AggInfo
	Update           *AggInfo
	Remove           *AggInfo
	ModChange        *AggInfo
	Error            *AggInfo
}

type LogicalEntry struct {
	// keeping source and target states out of event saves storage when unchanged
	SourceStates  []*StoredEntry
	SourceEvents  []*Event
	SourceVerif   *Verification
	TargetStates  []*StoredEntry
	DepCount      int32
	DirupState    *StoredEntry
	DirupChildren []string
	TargetEvents  []*Event
	TargetVerif   *Verification
	StatsList     []*ComputedStats
}

// For a memory to file simple implementation, limited to 2GiB, cf https://protobuf.dev/programming-guides/proto-limits/#total
type OpeLogAllInOne struct {
	SourceRoot string
	TargetRoot string
	// the key is the relative path
	LogicalEntries map[string]*LogicalEntry
}
