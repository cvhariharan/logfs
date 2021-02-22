package fs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/cvhariharan/logfs/index"
)

var inodeCount uint64
var Index index.IndexStore

type EntryGetter interface {
	GetDirentType() fuse.DirentType
}

type FS struct {
	Index index.IndexStore
}

func NewFS() FS {
	Index = index.NewMemIndexStore()
	return FS{}
}

func (f FS) Root() (fs.Node, error) {
	return NewDir(), nil
}
