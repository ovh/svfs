package svfs

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/ncw/swift"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

var (
	FolderRegex     = regexp.MustCompile("^.+/$")
	SubdirRegex     = regexp.MustCompile(".*/.*$")
	DirContentType  = "application/directory"
	ObjContentType  = "application/octet-stream"
	EntryCache      = new(Cache)
	DirectoryLister = new(DirLister)
)

type DirLister struct {
	concurrency uint64
	taskChan    chan DirListerTask
}

type DirListerTask struct {
	o  *Object
	rc chan<- *Object
}

func (dl *DirLister) Start() {
	dl.taskChan = make(chan DirListerTask, dl.concurrency)
	for i := 0; uint64(i) < dl.concurrency; i++ {
		go func() {
			for t := range dl.taskChan {
				_, h, _ := SwiftConnection.Object(t.o.c.Name, t.o.so.Name)
				t.o.so.Bytes, _ = strconv.ParseInt(h["Content-Length"], 10, 64)
				t.rc <- t.o
			}
		}()
	}
}

func (dl *DirLister) AddTask(o *Object, c chan<- *Object) {
	go func() {
		dl.taskChan <- DirListerTask{
			o:  o,
			rc: c,
		}
	}()
}

type Directory struct {
	apex bool
	name string
	path string
	c    *swift.Container
}

func (d *Directory) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0600
	a.Size = uint64(4096)
	return nil
}

func (d *Directory) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	// Create an empty object in swift
	path := d.path + req.Name
	w, err := SwiftConnection.ObjectCreate(d.c.Name, path, false, "", ObjContentType, nil)
	if err != nil {
		return nil, nil, fuse.EIO
	}
	if _, err := w.Write([]byte(nil)); err != nil {
		return nil, nil, fuse.EIO
	}
	w.Close()

	// Retrieve it
	obj, _, err := SwiftConnection.Object(d.c.Name, path)
	if err != nil {
		return nil, nil, fuse.EIO
	}

	// New node
	node := &Object{
		name: req.Name,
		path: path,
		so:   &obj,
		c:    d.c,
	}

	// Get object handler handler
	h, err := node.open(fuse.OpenWriteOnly)
	if err != nil {
		return nil, nil, fuse.EIO
	}

	// Cache it
	EntryCache.Set(d.c.Name, d.path, req.Name, node)

	return node, h, nil
}

func (d *Directory) Export() fuse.Dirent {
	return fuse.Dirent{
		Name: d.name,
		Type: fuse.DT_Dir,
	}
}

func (d *Directory) ReadDirAll(ctx context.Context) (entries []fuse.Dirent, err error) {
	var (
		dirs  = make(map[string]bool)
		loC   = make(chan *Object, DirectoryLister.concurrency)
		count = 0
	)

	defer close(loC)

	// Cache check
	if nodes := EntryCache.GetAll(d.c.Name, d.path); nodes != nil {
		for _, node := range nodes {
			entries = append(entries, node.Export())
		}
		return entries, nil
	}

	// Fetch objects
	objects, err := SwiftConnection.ObjectsAll(d.c.Name, &swift.ObjectsOpts{
		Delimiter: '/',
		Prefix:    d.path,
	})
	if err != nil {
		return nil, err
	}

	var children = make(map[string]Node)

	// Fill cache
	for _, object := range objects {
		var (
			child    Node
			o        = object
			fileName = strings.TrimPrefix(o.Name, d.path)
		)
		// This is a directory
		if o.ContentType == DirContentType && !FolderRegex.Match([]byte(o.Name)) {
			dirs[fileName] = true
			child = &Directory{
				c:    d.c,
				path: o.Name + "/",
				name: fileName,
			}
		} else if o.PseudoDirectory &&
			FolderRegex.Match([]byte(o.Name)) && fileName != "" {
			// This is a pseudo directory. Add it only if the real directory is missing
			realName := fileName[:len(fileName)-1]
			if !dirs[realName] {
				dirs[realName] = true
				child = &Directory{
					c:    d.c,
					path: o.Name,
					name: realName,
				}
			}
		} else if !FolderRegex.Match([]byte(o.Name)) {
			// This is a swift object
			obj := &Object{
				path: o.Name,
				name: fileName,
				c:    d.c,
				so:   &o,
				p:    d,
			}

			if o.Bytes == 0 &&
				!o.PseudoDirectory &&
				o.ContentType != DirContentType {
				DirectoryLister.AddTask(obj, loC)
				child = nil
				count++
			} else {
				child = obj
			}

		}

		if child != nil {
			entries = append(entries, child.Export())
			children[child.Name()] = child
		}
	}

	if count > 0 {
		done := 0
		for o := range loC {
			done++
			entries = append(entries, o.Export())
			children[o.name] = o
			if done == count {
				break
			}
		}
	}

	EntryCache.AddAll(d.c.Name, d.path, children)

	return entries, nil
}

func (d *Directory) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	if !EntryCache.CheckGetAll(d.c.Name, d.path) {
		d.ReadDirAll(ctx)
	}

	// Find matching child
	if item := EntryCache.Get(d.c.Name, d.path, req.Name); item != nil {
		if n, ok := item.(*Container); ok {
			return n, nil
		}
		if n, ok := item.(*Directory); ok {
			return n, nil
		}
		if n, ok := item.(*Object); ok {
			return n, nil
		}
	}

	return nil, fuse.ENOENT
}

func (d *Directory) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	var (
		objName = req.Name + "/"
		absPath = d.path + objName
	)

	// Create the file in swift
	if err := SwiftConnection.ObjectPutBytes(d.c.Name, absPath, nil, DirContentType); err != nil {
		return nil, fuse.EIO
	}

	// Directory object
	node := &Directory{
		name: req.Name,
		path: absPath,
		c:    d.c,
	}

	// Cache eviction
	EntryCache.Set(d.c.Name, d.path, req.Name, node)

	return node, nil
}

func (d *Directory) Name() string {
	return d.name
}

func (d *Directory) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	path := d.path + req.Name

	if req.Dir {
		path += "/"
	}

	// Delete from swift
	err := SwiftConnection.ObjectDelete(d.c.Name, path)
	if err != nil && err != swift.ObjectNotFound {
		return err
	}

	// Cache eviction
	EntryCache.Delete(d.c.Name, d.path, req.Name)

	return nil
}

func (d *Directory) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	// Not supported
	if _, ok := newDir.(*Root); ok {
		return fuse.ENOTSUP
	}
	// Swift move = copy + delete
	if t, ok := newDir.(*Container); ok {
		SwiftConnection.ObjectMove(d.c.Name, d.path+req.OldName, t.c.Name, t.path+req.NewName)
		EntryCache.Delete(d.c.Name, d.path, req.OldName)
		EntryCache.Set(t.c.Name, t.path, req.NewName, t)
		return nil
	}
	if t, ok := newDir.(*Directory); ok {
		SwiftConnection.ObjectMove(d.c.Name, d.path+req.OldName, t.c.Name, t.path+req.NewName)
		EntryCache.Delete(d.c.Name, d.path, req.OldName)
		EntryCache.Set(t.c.Name, t.path, req.NewName, t)
		return nil
	}
	return nil
}

var (
	_ Node           = (*Directory)(nil)
	_ fs.Node        = (*Directory)(nil)
	_ fs.NodeCreater = (*Directory)(nil)
	_ fs.NodeRemover = (*Directory)(nil)
	_ fs.NodeMkdirer = (*Directory)(nil)
	_ fs.NodeRenamer = (*Directory)(nil)
)
