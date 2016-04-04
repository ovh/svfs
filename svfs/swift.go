package svfs

import (
	"fmt"
	"strings"

	"bazil.org/fuse"

	"github.com/xlucas/swift"
)

func createAndWriteSegment(container, segmentPrefix string, segmentID *uint, t *swift.Object, data []byte, uploaded *uint64) (*swift.ObjectCreateFile, error) {
	segment, err := createSegment(container, segmentPrefix, segmentID, uploaded)
	if err != nil {
		return nil, err
	}
	err = writeSegmentData(segment, t, data, uploaded)
	if err != nil {
		return nil, err
	}
	return segment, nil
}

func createContainer(name string) (*swift.Container, error) {
	err := SwiftConnection.ContainerCreate(name, nil)
	if err != nil {
		return nil, err
	}
	c, _, err := SwiftConnection.Container(name)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func createSegment(container, segmentPrefix string, segmentID *uint, uploaded *uint64) (fh *swift.ObjectCreateFile, err error) {
	segmentName := segmentPath(segmentPrefix, segmentID)
	fh, err = SwiftConnection.ObjectCreate(container, segmentName, false, "", ObjContentType, nil)
	*uploaded = 0
	return
}

func deleteSegments(container, manifestHeader string) error {
	prefix := strings.TrimPrefix(manifestHeader, container+"/")

	// Custom segment container name is not supported
	if prefix == manifestHeader {
		return fuse.ENOTSUP
	}

	// Find segments
	segments, err := SwiftConnection.ObjectNamesAll(container, &swift.ObjectsOpts{
		Prefix: prefix,
	})
	if err != nil {
		return err
	}

	// Delete segments
	for _, segment := range segments {
		if err := SwiftConnection.ObjectDelete(container, segment); err != nil {
			return err
		}
	}

	return nil
}

func segmentPath(segmentPrefix string, segmentID *uint) string {
	*segmentID++
	return fmt.Sprintf("%s/%08d", segmentPrefix, *segmentID)
}

func writeSegmentData(fh *swift.ObjectCreateFile, t *swift.Object, data []byte, uploaded *uint64) error {
	_, err := fh.Write(data)
	t.Bytes += int64(len(data))
	*uploaded += uint64(len(data))
	return err
}
