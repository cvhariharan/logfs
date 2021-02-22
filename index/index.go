package index

import (
	"sync"
	"time"

	"bazil.org/fuse"
)

type AttrHandler interface {
	SetAttrAtime(inode uint64, atime time.Time)
	SetAttrMtime(inode uint64, mtime time.Time)
	SetAttrSize(inode uint64, size uint64)
	GetAttrAtime(inode uint64) (time.Time, bool)
	GetAttrMtime(inode uint64) (time.Time, bool)
	GetAttrSize(inode uint64) (uint64, bool)
}

type IndexStore interface {
	AttrHandler
	GetAllFromDir(inode uint64) (map[string]interface{}, bool)
	SetDirEntries(inode uint64, val map[string]interface{})
	SetInDir(inode uint64, name string, val interface{})
	GetInDir(inode uint64, name string) (interface{}, bool)
	GetAttributes(inode uint64) (fuse.Attr, bool)
	SetAttributes(inode uint64, attr fuse.Attr)
}

type MemIndexStore struct {
	mutex   *sync.Mutex
	tree    map[uint64]map[string]interface{}
	attrMap map[uint64]fuse.Attr
}

func NewMemIndexStore() IndexStore {
	return &MemIndexStore{
		mutex:   &sync.Mutex{},
		tree:    make(map[uint64]map[string]interface{}),
		attrMap: make(map[uint64]fuse.Attr),
	}
}

func (m *MemIndexStore) GetAllFromDir(inode uint64) (map[string]interface{}, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	val, ok := m.tree[inode]
	return val, ok
}

func (m *MemIndexStore) SetDirEntries(inode uint64, val map[string]interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.tree[inode] = val
}

func (m *MemIndexStore) SetInDir(inode uint64, name string, val interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	entries := m.tree[inode]
	if entries == nil {
		entries = make(map[string]interface{})
	}
	entries[name] = val
	m.tree[inode] = entries
}

func (m *MemIndexStore) GetInDir(inode uint64, name string) (interface{}, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	entries := m.tree[inode]
	val, ok := entries[name]
	return val, ok
}

func (m *MemIndexStore) GetAttributes(inode uint64) (fuse.Attr, bool) {
	val, ok := m.attrMap[inode]
	return val, ok
}

func (m *MemIndexStore) SetAttributes(inode uint64, attr fuse.Attr) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.attrMap[inode] = attr
}

func (m *MemIndexStore) SetAttrAtime(inode uint64, atime time.Time) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	attr, _ := m.GetAttributes(inode)
	attr.Atime = atime
	m.SetAttributes(inode, attr)
}

func (m *MemIndexStore) SetAttrMtime(inode uint64, mtime time.Time) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	attr, _ := m.GetAttributes(inode)
	attr.Mtime = mtime
	m.SetAttributes(inode, attr)
}

func (m *MemIndexStore) SetAttrSize(inode uint64, size uint64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	attr, _ := m.GetAttributes(inode)
	attr.Size = size
	m.SetAttributes(inode, attr)
}

func (m *MemIndexStore) GetAttrAtime(inode uint64) (time.Time, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	val, ok := m.GetAttributes(inode)
	return val.Atime, ok
}

func (m *MemIndexStore) GetAttrMtime(inode uint64) (time.Time, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	val, ok := m.GetAttributes(inode)
	return val.Mtime, ok
}

func (m *MemIndexStore) GetAttrSize(inode uint64) (uint64, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	val, ok := m.GetAttributes(inode)
	return val.Size, ok
}
