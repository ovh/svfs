package svfs

import (
	"fmt"
	"strings"
	"time"

	"bazil.org/fuse"

	"github.com/xlucas/swift"
)

func initSegment(c, prefix string, id *uint, t *swift.Object, d []byte, up *uint64) (*swift.ObjectCreateFile, error) {
	segment, err := createSegment(c, prefix, id, up)
	if err != nil {
		return nil, err
	}
	err = writeSegmentData(segment, t, d, up)
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

func createManifest(container, segmentsPath, path string) error {
	headers := map[string]string{
		ManifestHeader:   segmentsPath,
		"Content-Length": "0",
		AutoContent:      "true",
	}

	manifest, err := SwiftConnection.ObjectCreate(container, path, false, "", "", headers)
	if err != nil {
		return err
	}

	manifest.Write(nil)
	manifest.Close()

	return nil
}

func createSegment(container, prefix string, id *uint, uploaded *uint64) (fh *swift.ObjectCreateFile, err error) {
	segmentName := segmentPath(prefix, id)
	fh, err = SwiftConnection.ObjectCreate(container, segmentName, false, "", ObjContentType, nil)
	*uploaded = 0
	return
}

func getMtime(object *swift.Object, headers *swift.Headers) time.Time {
	if ExtraAttr && headers != nil {
		if mtime, err := headers.ObjectMetadata().GetModTime(); err == nil {
			return mtime
		}
	}
	return object.LastModified
}

func isDirectory(object swift.Object, path string) bool {
	return (object.ContentType == DirContentType) && (object.Name != path) && !object.PseudoDirectory
}

func isLargeObject(object *swift.Object) bool {
	return (object.Bytes == 0) && !object.PseudoDirectory && (object.ContentType != DirContentType)
}

func isPseudoDirectory(object swift.Object, path string) bool {
	return object.PseudoDirectory && (object.Name != path)
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

type swiftACLAuth struct {
	swift.Authenticator
	storageURL string
}

func newSwiftACLAuth(baseAuth swift.Authenticator, storageURL string) *swiftACLAuth {
	return &swiftACLAuth{
		Authenticator: baseAuth,
		storageURL:    storageURL,
	}
}

func (a *swiftACLAuth) StorageURL(Internal bool) string {
	return a.storageURL
}

var _ swift.Authenticator = (*swiftACLAuth)(nil)
