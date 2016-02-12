package fs

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
	"golang.org/x/net/context"
)

var (
	SegmentRegex = regexp.MustCompile("^.+_segments$")
	RootRegex    = regexp.MustCompile("^[^/]+/?$")
	FolderRegex  = regexp.MustCompile("^.+/$")
)

type Directory struct {
	s *swift.Connection
	f *File
}

func (d *Directory) readRoot() ([]fuse.Dirent, error) {
	// Retrieve all the containers for this account
	containers, err := d.s.ContainersAll(nil)
	if err != nil {
		return nil, err
	}

	// Convert swift containers to fuse directories
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

func (d *Directory) readDirectory() ([]fuse.Dirent, error) {
	// Retrieve all directories in our path
	prefix := d.f.path
	if prefix != "" {
		prefix = fmt.Sprintf("%s/", d.f.path)
	}
	objects, err := d.s.ObjectsAll(d.f.c.Name, &swift.ObjectsOpts{
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

		fmt.Fprintf(os.Stderr, "Got : %s\n", fileName)
		dirs = append(dirs, fuse.Dirent{
			Name: fileName,
			Type: fileType,
		})

	}
	return dirs, nil
}

func (d *Directory) Attr(ctx context.Context, a *fuse.Attr) error {
	if d.f != nil {
		a.Size = d.f.Size()
		if d.f.directory {
			a.Mode = os.ModeDir
		}
	}
	if d.f == nil {
		a.Mode = os.ModeDir
	}
	return nil
}

func (d *Directory) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	// Root directory
	if d.f == nil {
		fmt.Fprintf(os.Stderr, "Readdir request for root\n")
		return d.readRoot()
	}
	// Inside a container
	fmt.Fprintf(os.Stderr, "Readdir request for %s\n", d.f.Name())
	return d.readDirectory()
}

func (d *Directory) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	var (
		file *File
		path = ""
	)

	if d.f != nil {
		if d.f.path == "" {
			path = req.Name
		} else {
			path = fmt.Sprintf("%s/%s", d.f.path, req.Name)
		}
	}

	// Container root
	if d.f == nil {
		fmt.Fprintf(os.Stderr, "Lookup request for root %s with path %s\n", req.Name, path)
		container, headers, err := d.s.Container(req.Name)
		if err != nil {
			return nil, err
		}
		seg, segH, err := d.s.Container(req.Name + "_segments")
		if err != nil {
			file = NewFromContainer(path, container, headers)
		} else {
			file = NewFromContainerWithSegments(path, container, seg, headers, segH)
		}
		return &Directory{s: d.s, f: file}, nil
	}

	// Inside a regular container
	fmt.Fprintf(os.Stderr, "Lookup request for directory %s with path %s\n", req.Name, path)
	obj, headers, err := d.s.Object(d.f.c.Name, path)
	if err != nil {
		file = NewFromObjectWithSegments(path, true, &obj, headers, d.f)
	} else {
		file = NewFromObjectWithSegments(path, false, &obj, headers, d.f)
	}

	return &Directory{s: d.s, f: file}, nil
}

// Check that we satisify the Dir interface.
var _ fs.Node = (*Directory)(nil)
