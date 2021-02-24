package fs

import (
	"context"
	"log"
	"path/filepath"
	"sync/atomic"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type File struct {
	Path    string
	Type    fuse.DirentType
	Content []byte
	Inode   uint64
}

var _ = (fs.Node)((*File)(nil))
var _ = (fs.HandleWriter)((*File)(nil))
var _ = (fs.HandleReadAller)((*File)(nil))
var _ = (fs.NodeSetattrer)((*File)(nil))
var _ = (EntryGetter)((*File)(nil))

func NewFile(name, parentPath string, content []byte) *File {
	log.Println("NewFile called")
	f := &File{
		Path:    filepath.Join(parentPath, name),
		Type:    fuse.DT_File,
		Content: content,
	}
	f.Inode = atomic.AddUint64(&inodeCount, 1)
	log.Println("NewFile inodeCount: ", inodeCount, " inode: ", f.Inode)
	Index.SetAttributes(inodeCount, fuse.Attr{
		Inode: f.Inode,
		Atime: time.Now(),
		Mtime: time.Now(),
		Ctime: time.Now(),
		Mode:  0o777,
	})
	return f
}

func (f *File) GetDirentType() fuse.DirentType {
	return f.Type
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	*a, _ = Index.GetAttributes(f.Inode)
	log.Println("Attr: Modified At", a.Mtime.String())
	return nil
}

func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	log.Println("ReadAll called")
	return f.Content, nil
}

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	log.Println("Write called, data to write: ", string(req.Data))
	f.Content = req.Data
	resp.Size = len(req.Data)
	attr, _ := Index.GetAttributes(f.Inode)
	attr.Size = uint64(resp.Size)
	Index.SetAttributes(f.Inode, attr)
	return nil
}

func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	if req.Valid.Atime() {
		attr, _ := Index.GetAttributes(f.Inode)
		attr.Atime = req.Atime
		Index.SetAttributes(f.Inode, attr)
	}
	if req.Valid.Mtime() {
		attr, _ := Index.GetAttributes(f.Inode)
		attr.Mtime = req.Mtime
		Index.SetAttributes(f.Inode, attr)
	}
	if req.Valid.Size() {
		attr, _ := Index.GetAttributes(f.Inode)
		attr.Size = req.Size
		Index.SetAttributes(f.Inode, attr)
	}
	attr, _ := Index.GetAttributes(f.Inode)
	log.Println("Setattr called: Attributes ", attr)
	return nil
}
