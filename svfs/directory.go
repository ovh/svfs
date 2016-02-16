package svfs

import (
	"os"
	"regexp"
	"strings"

	"github.com/ncw/swift"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

var (
	FolderRegex = regexp.MustCompile("^.+/$")
)

type Directory struct {
	name     string
	path     string
	s        *swift.Connection
	c        *swift.Container
	children []Node
}

func (d *Directory) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0600
	a.Size = uint64(4096)
	return nil
}

func (d *Directory) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	// Create an empty object in swift
	path := d.path + req.Name
	w, err := d.s.ObjectCreate(d.c.Name, path, false, "", "application/octet-stream", nil)
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
	d.children = []Node{}

	return node, h, nil
}

func (d *Directory) Export() fuse.Dirent {
	return fuse.Dirent{
		Name: d.name,
		Type: fuse.DT_Dir,
	}
}

func (d *Directory) ReadDirAll(ctx context.Context) (entries []fuse.Dirent, err error) {
	dirs := make(map[string]bool)

	// Cache hit
	if len(d.children) > 0 {
		for _, child := range d.children {
			entries = append(entries, child.Export())
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

	// Fill cache
	for _, object := range objects {
		var (
			child    Node
			o        = object
			fileName = strings.TrimPrefix(o.Name, d.path)
		)
		// This is a directory
		if o.ContentType == "application/directory" && !FolderRegex.Match([]byte(o.Name)) {
			child = &Directory{
				s:    d.s,
				c:    d.c,
				path: o.Name + "/",
				name: fileName,
			}
			dirs[fileName] = true
		} else if o.PseudoDirectory &&
			FolderRegex.Match([]byte(o.Name)) && fileName != "" {
			// This is a pseudo directory. Add it only if the real directory is missing
			realName := fileName[:len(fileName)-1]
			if !dirs[realName] {
				child = &Directory{
					s:    d.s,
					c:    d.c,
					path: o.Name,
					name: realName,
				}
				dirs[realName] = true
			}
		} else if !FolderRegex.Match([]byte(o.Name)) {
			// This is a swift object
			child = &Object{
				path: o.Name,
				name: fileName,
				s:    d.s,
				c:    d.c,
				so:   &o,
				p:    d,
			}
		}

		if child != nil {
			d.children = append(d.children, child)
			entries = append(entries, child.Export())
		}
	}

	return entries, nil
}

func (d *Directory) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	// Get children if this node was evicted from the kernel cache
	if len(d.children) == 0 && d.c != nil {
		d.ReadDirAll(ctx)
	}

	// Find matching child
	for _, item := range d.children {
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
	if err := d.s.ObjectPutBytes(d.c.Name, absPath, nil, "application/directory"); err != nil {
		return nil, fuse.EIO
	}

	// Cache eviction
	d.children = []Node{}

	// Directory object
	return &Directory{
		name: req.Name,
		path: absPath,
		s:    d.s,
		c:    d.c,
	}, nil
}

func (d *Directory) Name() string {
	return d.name
}

func (d *Directory) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	// Delete from swift
	err := d.s.ObjectDelete(d.c.Name, d.path+req.Name)
	if err != nil && err != swift.ObjectNotFound {
		return err
	}

	// Cache eviction
	d.children = []Node{}

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
		t.children = []Node{}
		d.children = []Node{}
		return nil
	}
	if t, ok := newDir.(*Directory); ok {
		d.s.ObjectMove(d.c.Name, d.path+req.OldName, t.c.Name, t.path+req.NewName)
		t.children = []Node{}
		d.children = []Node{}
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
