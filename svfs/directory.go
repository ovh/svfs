package svfs

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/xlucas/swift"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

const (
	DirContentType = "application/directory"
	ObjContentType = "application/octet-stream"
)

var (
	FolderRegex      = regexp.MustCompile("^.+/$")
	SubdirRegex      = regexp.MustCompile(".*/.*$")
	SegmentPathRegex = regexp.MustCompile("^([^/]+)/(.*)$")
	DirectoryCache   = new(Cache)
	DirectoryLister  = new(DirLister)
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
				if SegmentPathRegex.Match([]byte(h[ManifestHeader])) {
					t.o.segmented = true
				}
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
	cs   *swift.Container
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
		cs:   d.cs,
	}

	// Get object handler handler
	h, err := node.open(fuse.OpenWriteOnly)
	if err != nil {
		return nil, nil, fuse.EIO
	}

	// Cache it
	DirectoryCache.Set(d.c.Name, d.path, req.Name, node)

	return node, h, nil
}

func (d *Directory) Export() fuse.Dirent {
	return fuse.Dirent{
		Name: d.name,
		Type: fuse.DT_Dir,
	}
}

func (d *Directory) ReadDirAll(ctx context.Context) (direntries []fuse.Dirent, err error) {
	var (
		dirs         = make(map[string]bool)
		largeObjects = make(chan *Object, DirectoryLister.concurrency)
		count        = 0
	)

	defer close(largeObjects)

	// Cache check
	if _, nodes := DirectoryCache.GetAll(d.c.Name, d.path); nodes != nil {
		for _, node := range nodes {
			direntries = append(direntries, node.Export())
		}
		return direntries, nil
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
				cs:   d.cs,
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
					cs:   d.cs,
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
				cs:   d.cs,
				so:   &o,
				p:    d,
			}

			// Large object
			if o.Bytes == 0 &&
				!o.PseudoDirectory &&
				o.ContentType != DirContentType {
				DirectoryLister.AddTask(obj, largeObjects)
				child = nil
				count++
			} else {
				//Standard object
				child = obj
			}

		}

		if child != nil {
			direntries = append(direntries, child.Export())
			children[child.Name()] = child
		}
	}

	if count > 0 {
		done := 0
		for o := range largeObjects {
			done++
			direntries = append(direntries, o.Export())
			children[o.name] = o
			if done == count {
				break
			}
		}
	}

	DirectoryCache.AddAll(d.c.Name, d.path, d, children)

	return direntries, nil
}

func (d *Directory) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	if _, found := DirectoryCache.Peek(d.c.Name, d.path); !found {
		d.ReadDirAll(ctx)
	}

	// Find matching child
	if item := DirectoryCache.Get(d.c.Name, d.path, req.Name); item != nil {
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
		cs:   d.cs,
	}

	// Cache eviction
	DirectoryCache.Set(d.c.Name, d.path, req.Name, node)

	return node, nil
}

func (d *Directory) Name() string {
	return d.name
}

func (d *Directory) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	path := d.path + req.Name

	if req.Dir {
		path += "/"
		node, found := DirectoryCache.Peek(d.c.Name, d.path)
		if !found {
			return fuse.ENOTSUP
		}
		if _, ok := node.(*Directory); !ok {
			return fuse.ENOTSUP
		}
		SwiftConnection.ObjectDelete(d.c.Name, path)
		DirectoryCache.DeleteAll(d.c.Name, d.path)
	}

	if !req.Dir {
		// Get the old node from the cache
		node := DirectoryCache.Get(d.c.Name, d.path, req.Name)
		if object, ok := node.(*Object); ok {
			// Segmented object removal. We need to find all segments
			// using the manifest segment prefix then bulk delete
			// them and remove the manifest enventually.
			if object.segmented {
				_, h, err := SwiftConnection.Object(d.c.Name, path)
				if err != nil {
					return err
				}
				if !SegmentPathRegex.Match([]byte(h[ManifestHeader])) {
					return fmt.Errorf("Invalid segment path for manifest %s", req.Name)
				}
				deleteSegments(d.cs.Name, h[ManifestHeader])
			}
		}
		SwiftConnection.ObjectDelete(d.c.Name, path)
		DirectoryCache.Delete(d.c.Name, d.path, req.Name)
	}

	return nil
}

func (d *Directory) move(oldContainer, oldPath, oldName, newContainer, newPath, newName string) error {
	// Get the old node from the cache
	oldNode := DirectoryCache.Get(d.c.Name, d.path, oldName)

	if oldObject, ok := oldNode.(*Object); ok {
		// Move a manifest, not the aggregated result of its segments
		if oldObject.segmented {
			return d.moveManifest(oldContainer, oldPath, oldName, newContainer, newPath, newName, oldObject)
		}
		// Move a standard object
		if !oldObject.segmented {
			return d.moveObject(oldContainer, oldPath, oldName, newContainer, newPath, newName, oldObject)
		}
	}

	return fuse.ENOTSUP
}

func (d *Directory) moveObject(oldContainer, oldPath, oldName, newContainer, newPath, newName string, o *Object) error {
	err := SwiftConnection.ObjectMove(oldContainer, oldPath+oldName, newContainer, newPath+newName)
	if err != nil {
		return err
	}

	o.name = newName
	o.path = newPath + newName

	DirectoryCache.Delete(oldContainer, oldPath, oldName)
	DirectoryCache.Set(newContainer, newPath, newName, o)

	return nil
}

func (d *Directory) moveManifest(oldContainer, oldPath, oldName, newContainer, newPath, newName string, o *Object) error {
	_, err := SwiftConnection.ManifestCopy(oldContainer, oldPath+oldName, newContainer, newPath+newName, nil)
	if err != nil {
		return err
	}
	err = SwiftConnection.ObjectDelete(oldContainer, oldPath+oldName)
	if err != nil {
		return err
	}

	o.name = newName
	o.path = newPath + newName

	DirectoryCache.Delete(oldContainer, oldPath, oldName)
	DirectoryCache.Set(newContainer, newPath, newName, o)

	return nil
}

func (d *Directory) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	if t, ok := newDir.(*Container); ok {
		return d.move(d.c.Name, d.path, req.OldName, t.c.Name, t.path, req.NewName)
	}
	if t, ok := newDir.(*Directory); ok {
		return d.move(d.c.Name, d.path, req.OldName, t.c.Name, t.path, req.NewName)
	}
	return fuse.ENOTSUP
}

var (
	_ Node           = (*Directory)(nil)
	_ fs.Node        = (*Directory)(nil)
	_ fs.NodeCreater = (*Directory)(nil)
	_ fs.NodeRemover = (*Directory)(nil)
	_ fs.NodeMkdirer = (*Directory)(nil)
	_ fs.NodeRenamer = (*Directory)(nil)
)
