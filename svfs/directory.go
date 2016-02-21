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
	FolderRegex    = regexp.MustCompile("^.+/$")
	DirContentType = "application/directory"
	ObjContentType = "application/octet-stream"
)

type DirLister struct {
	c           *swift.Connection
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
				_, h, _ := dl.c.Object(t.o.c.Name, t.o.so.Name)
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
	apex  bool
	name  string
	path  string
	cache *Cache
	s     *swift.Connection
	c     *swift.Container
	l     *DirLister
}

func (d *Directory) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0600
	a.Size = uint64(4096)
	return nil
}

func (d *Directory) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	// Create an empty object in swift
	path := d.path + req.Name
	w, err := d.s.ObjectCreate(d.c.Name, path, false, "", ObjContentType, nil)
	if err != nil {
		return nil, nil, fuse.EIO
	}
	if _, err := w.Write([]byte(nil)); err != nil {
		return nil, nil, fuse.EIO
	}
	w.Close()

	// Retrieve it
	obj, _, err := d.s.Object(d.c.Name, path)
	if err != nil {
		return nil, nil, fuse.EIO
	}

	// New node
	node := &Object{
		name: req.Name,
		path: path,
		s:    d.s,
		so:   &obj,
		c:    d.c,
	}

	// Get object handler handler
	h, err := node.open(fuse.OpenWriteOnly)
	if err != nil {
		return nil, nil, fuse.EIO
	}

	// Force cache eviction
	d.cache.Delete(d.path)

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
		loC   = make(chan *Object, d.l.concurrency)
		count = 0
	)

	// Cache check
	if nodes := d.cache.Get(d.path); nodes != nil {
		for _, node := range nodes {
			entries = append(entries, node.Export())
		}
		return entries, nil
	}

	// Fetch objects
	objects, err := d.s.ObjectsAll(d.c.Name, &swift.ObjectsOpts{
		Delimiter: '/',
		Prefix:    d.path,
	})
	if err != nil {
		return nil, err
	}

	var children = make([]Node, 0)

	// Fill cache
	for _, object := range objects {
		var (
			child    Node
			o        = object
			fileName = strings.TrimPrefix(o.Name, d.path)
		)
		// This is a directory
		if o.ContentType == DirContentType && !FolderRegex.Match([]byte(o.Name)) {
			count++
			dirs[fileName] = true
			child = &Directory{
				s:     d.s,
				c:     d.c,
				l:     d.l,
				cache: d.cache,
				path:  o.Name + "/",
				name:  fileName,
			}
		} else if o.PseudoDirectory &&
			FolderRegex.Match([]byte(o.Name)) && fileName != "" {
			// This is a pseudo directory. Add it only if the real directory is missing
			count++
			realName := fileName[:len(fileName)-1]
			if !dirs[realName] {
				dirs[realName] = true
				child = &Directory{
					s:     d.s,
					c:     d.c,
					l:     d.l,
					cache: d.cache,
					path:  o.Name,
					name:  realName,
				}
			}
		} else if !FolderRegex.Match([]byte(o.Name)) {
			// This is a swift object
			obj := &Object{
				path: o.Name,
				name: fileName,
				s:    d.s,
				c:    d.c,
				so:   &o,
				p:    d,
			}

			if o.Bytes == 0 &&
				!o.PseudoDirectory &&
				o.ContentType != DirContentType {
				d.l.AddTask(obj, loC)
				child = nil
			} else {
				count++
				child = obj
			}

		}

		if child != nil {
			entries = append(entries, child.Export())
			children = append(children, child)
		}
	}

	if count != len(objects) {
		for o := range loC {
			count++
			entries = append(entries, o.Export())
			children = append(children, o)
			if count == len(objects) {
				close(loC)
				break
			}
		}
	}

	d.cache.Set(d.path, children)

	return entries, nil
}

func (d *Directory) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	var nodes []Node

	if nodes = d.cache.Get(d.path); nodes == nil {
		d.ReadDirAll(ctx)
		nodes = d.cache.Get(d.path)
	}

	// Find matching child
	for _, item := range nodes {
		if item.Name() == req.Name {
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
	}

	return nil, fuse.ENOENT
}

func (d *Directory) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	var (
		objName = req.Name + "/"
		absPath = d.path + objName
	)

	// Create the file in swift
	if err := d.s.ObjectPutBytes(d.c.Name, absPath, nil, DirContentType); err != nil {
		return nil, fuse.EIO
	}

	// Cache eviction
	d.cache.Delete(d.path)

	// Directory object
	return &Directory{
		name:  req.Name,
		path:  absPath,
		cache: d.cache,
		s:     d.s,
		c:     d.c,
	}, nil
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
	err := d.s.ObjectDelete(d.c.Name, path)
	if err != nil && err != swift.ObjectNotFound {
		return err
	}

	// Cache eviction
	d.cache.Delete(d.path)

	return nil
}

func (d *Directory) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	// Not supported
	if _, ok := newDir.(*Root); ok {
		return fuse.ENOTSUP
	}
	// Swift move = copy + delete
	if t, ok := newDir.(*Container); ok {
		d.s.ObjectMove(d.c.Name, d.path+req.OldName, t.c.Name, t.path+req.NewName)
		d.cache.Delete(d.path)
		t.cache.Delete(t.path)
		return nil
	}
	if t, ok := newDir.(*Directory); ok {
		d.s.ObjectMove(d.c.Name, d.path+req.OldName, t.c.Name, t.path+req.NewName)
		d.cache.Delete(d.path)
		t.cache.Delete(t.path)
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
