package fs

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
	"github.com/xlucas/svfs/fs/file"
	"golang.org/x/net/context"
)

var (
	SegmentRegex = regexp.MustCompile("^.+_segments$")
	RootRegex    = regexp.MustCompile("^[^/]+/?$")
	FolderRegex  = regexp.MustCompile("^.+/$")
)

type Dir struct {
	s *swift.Connection
	f file.Node
}

func (d *Dir) container() (*file.Container, error) {
	dir, ok := d.f.(*file.Directory)
	if ok {
		return dir.Container, nil
	}
	cont, ok := d.f.(*file.Container)
	if ok {
		return cont, nil
	}
	return nil, fmt.Errorf("Unable to find relevant swift container")
}

func (d *Dir) readRoot() ([]fuse.Dirent, error) {
	// Retrieve all the containers for this account
	containers, err := d.s.ContainersAll(nil)
	if err != nil {
		return nil, err
	}

	// Convert swift containers to fuse directories
	// but hide segment containers
	var dirs []fuse.Dirent
	for _, container := range containers {
		if !SegmentRegex.Match([]byte(container.Name)) {
			dirs = append(dirs, fuse.Dirent{
				Name: container.Name,
				Type: fuse.DT_Dir,
			})
		}
	}
	return dirs, nil
}

func (d *Dir) read() ([]fuse.Dirent, error) {
	// Build filter prefix
	prefix := d.f.Path()
	if prefix != "" {
		prefix = fmt.Sprintf("%s/", d.f.Path())
	}

	// Find relevant container
	c, err := d.container()
	if err != nil {
		return nil, err
	}

	// Fetch objects
	objects, err := d.s.ObjectsAll(c.C.Name, &swift.ObjectsOpts{
		Delimiter: '/',
		Prefix:    prefix,
	})
	if err != nil {
		return nil, err
	}

	// Convert them to fuse directories
	var dirs []fuse.Dirent
	for _, object := range objects {
		// Simple file
		fileName := strings.TrimPrefix(object.Name, prefix)
		fileType := fuse.DT_File

		// Directory
		if FolderRegex.Match([]byte(fileName)) {
			fileName = fileName[:len(fileName)-1]
			fileType = fuse.DT_Dir
		}

		//fmt.Fprintf(os.Stderr, "Got : %s\n", fileName)
		dirs = append(dirs, fuse.Dirent{
			Name: fileName,
			Type: fileType,
		})

	}
	return dirs, nil
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	if d.f != nil {
		a.Size = d.f.Size()
		a.Mode = d.f.Mode()
	} else {
		a.Mode = os.ModeDir
	}
	return nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	if d.f == nil {
		return d.readRoot()
	}
	return d.read()
}

func (d *Dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	var (
		fn   file.Node
		path = ""
	)

	if d.f != nil {
		if d.f.Path() == "" {
			path = req.Name
		} else {
			path = fmt.Sprintf("%s/%s", d.f.Path(), req.Name)
		}
	}

	// Root lookup
	if d.f == nil {
		// Main container
		container, _, err := d.s.Container(req.Name)
		if err != nil {
			return nil, err
		}

		// Segment container
		seg, _, err := d.s.Container(req.Name + "_segments")

		// Build fuse entry
		if err == swift.ContainerNotFound {
			fn = &file.Container{File: file.NewFile(""), C: &container}
		} else if err == nil {
			fn = &file.Container{File: file.NewFile(""), C: &container, CS: &seg}
		}
		return &Dir{s: d.s, f: fn}, nil
	}

	// Container or directory lookup
	c, err := d.container()
	if err != nil {
		return nil, err
	}

	obj, _, err := d.s.Object(c.C.Name, path)
	if err == swift.ObjectNotFound {
		fn = &file.Directory{File: file.NewFile(path), Container: c, Label: req.Name}
	} else if err == nil {
		fn = &file.Object{File: file.NewFile(path), Container: c, SO: &obj}
	}

	return &Dir{s: d.s, f: fn}, nil
}

// Check that we satisify the Node interface.
var _ fs.Node = (*Dir)(nil)
