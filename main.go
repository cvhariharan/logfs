package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
)

var inodeCount uint64

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}
	mountpoint := flag.Arg(0)

	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("logfs"),
		fuse.Subtype("logfs"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	err = fs.Serve(c, FS{})
	if err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT\n", os.Args[0])
	flag.PrintDefaults()
}

type EntryGetter interface {
	GetDirentType() fuse.DirentType
}

type FS struct{}

func (f FS) Root() (fs.Node, error) {
	return NewDir(), nil
}

type Dir struct {
	Type    fuse.DirentType
	Inode   uint64
	Entries map[string]interface{}
}

var _ = (fs.Node)((*Dir)(nil))
var _ = (fs.NodeMkdirer)((*Dir)(nil))
var _ = (fs.NodeCreater)((*Dir)(nil))
var _ = (fs.HandleReadDirAller)((*Dir)(nil))
var _ = (EntryGetter)((*Dir)(nil))

func NewDir() *Dir {
	atomic.AddUint64(&inodeCount, 1)
	return &Dir{
		Type:    fuse.DT_Dir,
		Inode:   inodeCount,
		Entries: map[string]interface{}{},
	}
}

func (d *Dir) GetDirentType() fuse.DirentType {
	return d.Type
}

func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	dir := NewDir()
	d.Entries[req.Name] = dir
	return dir, nil
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = d.Inode
	a.Mode = os.ModeDir | 0o777
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
	f := NewFile(nil)
	log.Println("Create: Modified at", f.Attributes.Mtime.String())
	d.Entries[req.Name] = f
	return f, f, nil
}

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
	f.Content = append(f.Content, req.Data...)
	resp.Size = len(req.Data)
	f.Attributes.Size = uint64(resp.Size)
	log.Println("Write called: Size ", f.Attributes.Size)
	return nil
}

func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	f.Attributes.Mtime = req.Mtime
	f.Attributes.Atime = req.Atime
	f.Attributes.Size += req.Size
	resp.Attr.Size = req.Size
	resp.Attr.Valid = time.Minute
	log.Println("Setattr called: Modified at ", f.Attributes.Mtime)
	return nil
}
