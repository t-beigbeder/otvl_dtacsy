package opelogimpl

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
	"github.com/t-beigbeder/vdasync/opeloggrpc"
	"google.golang.org/protobuf/proto"
)

type m2fMng struct {
	path       string
	mx         sync.Mutex
	source     string
	target     string
	les        map[string]*opelog.LogicalEntry
	readOnly   bool
	isOpen     bool
	hasUpdates bool
}

func (m *m2fMng) save() error {
	aio := opeloggrpc.OpeLogAllInOne{
		SourceRoot:     m.source,
		TargetRoot:     m.target,
		LogicalEntries: make(map[string]*opeloggrpc.LogicalEntry, len(m.les)),
	}
	for rp, le := range m.les {
		aio.LogicalEntries[rp] = opelog.LogicalEntry2GrpcLogicalEntry(le)
	}
	bs, err := proto.Marshal(&aio)
	if err != nil {
		return err
	}
	if err = common.WriteFile(m.path, bs); err != nil {
		return err
	}
	return nil
}

// GetLogicalEntry implements [opelog.OpeLogManager].
func (m *m2fMng) GetLogicalEntry(relPath string) (*opelog.LogicalEntry, error) {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return nil, errors.New("m2fMng.GetLogicalEntry: not opened")
	}
	le, _ := m.les[relPath]
	return le, nil
}

// Sync implements [opelog.OpeLogManager].
func (m *m2fMng) Sync() error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("m2fMng.Sync: not opened")
	}
	if !m.hasUpdates {
		return nil
	}
	if err := m.save(); err != nil {
		return err
	}
	m.hasUpdates = false
	return nil
}

// Close implements [opelog.OpeLogManager].
func (m *m2fMng) Close() error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("m2fMng.Close: not opened")
	}
	if !m.hasUpdates {
		return nil
	}
	if err := m.save(); err != nil {
		return err
	}
	m.hasUpdates = false
	m.isOpen = false
	if !m.readOnly {
		if err := os.Remove(fmt.Sprintf("%s.lock", m.path)); err != nil {
			return err
		}
	}
	return nil
}

// Create implements [opelog.OpeLogManager].
func (m *m2fMng) Create(source string, target string) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if common.FileExists(m.path) {
		return fmt.Errorf("m2fMng.Create: %s already exists", m.path)
	}
	if m.les != nil {
		return fmt.Errorf("m2fMng.Create: %s should be created without entries", m.path)
	}
	aio := opeloggrpc.OpeLogAllInOne{
		SourceRoot: source,
		TargetRoot: target,
	}
	bs, err := proto.Marshal(&aio)
	if err != nil {
		return err
	}
	if err = common.WriteFile(m.path, bs); err != nil {
		return err
	}
	return nil
}

// NewSession implements [opelog.OpeLogManager].
func (m *m2fMng) Open(readOnly bool) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	lock := fmt.Sprintf("%s.lock", m.path)
	if m.isOpen {
		return errors.New("m2fMng.Open: already opened")
	}
	if !readOnly && common.FileExists(lock) {
		return fmt.Errorf("m2fMng.Open: locked (%s)", lock)
	}
	bs, err := common.UnsafeLoadFile(m.path)
	if err != nil {
		return err
	}
	var aio opeloggrpc.OpeLogAllInOne
	if err = proto.Unmarshal(bs, &aio); err != nil {
		return err
	}
	m.source = aio.SourceRoot
	m.target = aio.TargetRoot
	m.les = make(map[string]*opelog.LogicalEntry, len(aio.LogicalEntries))
	for rp, gle := range aio.LogicalEntries {
		m.les[rp] = opelog.GrpcLogicalEntry2LogicalEntry(gle)
	}
	if !readOnly {
		if err := common.WriteFile(lock, []byte{}); err != nil {
			return err
		}
	}
	m.isOpen = true
	m.readOnly = readOnly
	return nil
}

// PutLogicalEntry implements [opelog.OpeLogManager].
func (m *m2fMng) PutLogicalEntry(relPath string, ole *opelog.LogicalEntry) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("m2fMng.PutEntryLog: not opened")
	}
	if m.readOnly {
		return errors.New("m2fMng.PutEntryLog: opened in read-only")
	}
	m.les[relPath] = ole
	m.hasUpdates = true
	return nil
}

var _ opelog.OpeLogManager = &m2fMng{}

func MakeM2fManager(path string) (opelog.OpeLogManager, error) {
	return &m2fMng{path: path}, nil
}
