package fs

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type File struct {
	Type       fuse.DirentType
	Content    []byte
	Attributes fuse.Attr
}

var _ = (fs.Node)((*File)(nil))
var _ = (fs.HandleWriter)((*File)(nil))
var _ = (fs.HandleReadAller)((*File)(nil))
var _ = (fs.NodeSetattrer)((*File)(nil))
var _ = (EntryGetter)((*File)(nil))

func NewFile(content []byte) *File {
	log.Println("NewFile called")
	atomic.AddUint64(&inodeCount, 1)
	return &File{
		Type:    fuse.DT_File,
		Content: content,
		Attributes: fuse.Attr{
			Inode: inodeCount,
			Atime: time.Now(),
			Mtime: time.Now(),
			Ctime: time.Now(),
			Mode:  0o777,
		},
	}
}

func (f *File) GetDirentType() fuse.DirentType {
	return f.Type
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	*a = f.Attributes
	log.Println("Attr: Modified At", a.Mtime.String())
	return nil
}

func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	log.Println("ReadAll called")
	return f.Content, nil
}

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	log.Println("Write called: Size ", f.Attributes.Size)
	log.Println("Data to write: ", string(req.Data))
	f.Content = req.Data
	resp.Size = len(req.Data)
	f.Attributes.Size = uint64(resp.Size)
	return nil
}

func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	if req.Valid.Atime() {
		f.Attributes.Atime = req.Atime
	}
	if req.Valid.Mtime() {
		f.Attributes.Mtime = req.Mtime
	}
	if req.Valid.Size() {
		f.Attributes.Size = req.Size
	}
	log.Println("Setattr called: Attributes ", f.Attributes)
	return nil
}
