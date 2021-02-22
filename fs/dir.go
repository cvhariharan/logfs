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
	Type  fuse.DirentType
	Inode uint64
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
	Index.SetAttributes(inodeCount, fuse.Attr{
		Inode: inodeCount,
		Atime: time.Now(),
		Mtime: time.Now(),
		Ctime: time.Now(),
		Mode:  os.ModeDir | 0o777,
	})
	return &Dir{
		Type:  fuse.DT_Dir,
		Inode: inodeCount,
	}
}

func (d *Dir) GetDirentType() fuse.DirentType {
	return d.Type
}

func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	log.Println("Mkdir called with name: ", req.Name)
	dir := NewDir()
	Index.SetInDir(d.Inode, req.Name, dir)
	return dir, nil
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	*a, _ = Index.GetAttributes(d.Inode)
	log.Println("Attr permissions: ", a.Mode)
	log.Println("Attr: Modified At", a.Mtime.String())
	return nil
}

func (d *Dir) LookUp(ctx context.Context, name string) (fs.Node, error) {
	node, ok := Index.GetInDir(d.Inode, name)
	if ok {
		return node.(fs.Node), nil
	}
	return nil, syscall.ENOENT
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("ReadDirAll called")
	var entries []fuse.Dirent
	dirEntries, _ := Index.GetAllFromDir(d.Inode)
	for k, v := range dirEntries {
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
	attr, _ := Index.GetAttributes(f.Inode)
	log.Println("Create: Modified at", attr.Mtime)
	Index.SetInDir(d.Inode, req.Name, f)
	return f, f, nil
}

func (d *Dir) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	if req.Valid.Atime() {
		attr, _ := Index.GetAttributes(d.Inode)
		attr.Atime = req.Atime
		Index.SetAttributes(d.Inode, attr)
	}
	if req.Valid.Mtime() {
		attr, _ := Index.GetAttributes(d.Inode)
		attr.Mtime = req.Mtime
		Index.SetAttributes(d.Inode, attr)
	}
	if req.Valid.Size() {
		attr, _ := Index.GetAttributes(d.Inode)
		attr.Size = req.Size
		Index.SetAttributes(d.Inode, attr)
	}
	attr, _ := Index.GetAttributes(d.Inode)
	log.Println("Setattr called: Attributes ", attr)
	return nil
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	entries, _ := Index.GetAllFromDir(d.Inode)
	delete(entries, req.Name)
	Index.SetDirEntries(d.Inode, entries)
	return nil
}
