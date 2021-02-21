package fs

import (
	"context"
	"log"
	"os"
	"sync/atomic"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type Dir struct {
	Type       fuse.DirentType
	Attributes fuse.Attr
	Entries    map[string]interface{}
}

var _ = (fs.Node)((*Dir)(nil))
var _ = (fs.NodeMkdirer)((*Dir)(nil))
var _ = (fs.NodeCreater)((*Dir)(nil))
var _ = (fs.HandleReadDirAller)((*Dir)(nil))
var _ = (fs.NodeSetattrer)((*Dir)(nil))
var _ = (EntryGetter)((*Dir)(nil))

func NewDir() *Dir {
	log.Println("NewDir called")
	atomic.AddUint64(&inodeCount, 1)
	return &Dir{
		Type: fuse.DT_Dir,
		Attributes: fuse.Attr{
			Inode: inodeCount,
			Atime: time.Now(),
			Mtime: time.Now(),
			Ctime: time.Now(),
			Mode:  os.ModeDir | 0o777,
		},
		Entries: map[string]interface{}{},
	}
}

func (d *Dir) GetDirentType() fuse.DirentType {
	return d.Type
}

func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	log.Println("Mkdir called with name: ", req.Name)
	dir := NewDir()
	d.Entries[req.Name] = dir
	return dir, nil
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	*a = d.Attributes
	log.Println("Attr permissions: ", a.Mode)
	log.Println("Attr: Modified At", a.Mtime.String())
	return nil
}

func (d *Dir) LookUp(ctx context.Context, name string) (fs.Node, error) {
	node, ok := d.Entries[name]
	if ok {
		return node.(fs.Node), nil
	}
	return nil, syscall.ENOENT
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("ReadDirAll called")
	var entries []fuse.Dirent

	for k, v := range d.Entries {
		var a fuse.Attr
		v.(fs.Node).Attr(ctx, &a)
		entries = append(entries, fuse.Dirent{
			Inode: a.Inode,
			Type:  v.(EntryGetter).GetDirentType(),
			Name:  k,
		})
	}
	return entries, nil
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	log.Println("Create called with filename: ", req.Name)
	f := NewFile(nil)
	log.Println("Create: Modified at", f.Attributes.Mtime.String())
	d.Entries[req.Name] = f
	return f, f, nil
}

func (d *Dir) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	if req.Valid.Atime() {
		d.Attributes.Atime = req.Atime
	}
	if req.Valid.Mtime() {
		d.Attributes.Mtime = req.Mtime
	}
	if req.Valid.Size() {
		d.Attributes.Size = req.Size
	}
	log.Println("Setattr called: Attributes ", d.Attributes)
	return nil
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	delete(d.Entries, req.Name)
	return nil
}
