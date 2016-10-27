package svfs

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"bazil.org/fuse"

	"github.com/xlucas/swift"
)

func canonicalHeaderKey(header string) string {
	return http.CanonicalHeaderKey(strings.Replace(header, "_", "-", -1))
}

func newReader(fh *ObjectHandle) (io.ReadSeeker, error) {
	rd, _, err := SwiftConnection.ObjectOpen(fh.target.c.Name, fh.target.path, false, nil)
	return rd, err
}

func newWriter(container, path string) (io.WriteCloser, error) {
	headers := map[string]string{"autoContent": "true"}
	return SwiftConnection.ObjectCreate(container, path, false, "", "", headers)
}

func initSegment(c, prefix string, id *uint, t *swift.Object, d []byte, up *uint64) (io.WriteCloser, error) {
	segment, err := createSegment(c, prefix, id, up)
	if err != nil {
		return nil, err
	}
	return segment, writeSegmentData(segment, t, d, up)
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

func createManifest(obj *Object, container, segmentsPath, path string) error {
	// Swift requires ampersand and question marks to be percent-encoded
	segmentsPath = strings.Replace(segmentsPath, "&", "%26", -1)
	segmentsPath = strings.Replace(segmentsPath, "?", "%3F", -1)

	obj.sh = map[string]string{
		manifestHeader:    segmentsPath,
		"Content-Length":  "0",
		autoContentHeader: "true",
	}

	manifest, err := SwiftConnection.ObjectCreate(container, path, false, "", "", obj.sh)
	if err != nil {
		return err
	}

	manifest.Write(nil)
	manifest.Close()

	return nil
}

func createSegment(container, prefix string, id *uint, uploaded *uint64) (io.WriteCloser, error) {
	segmentName := segmentPath(prefix, id)
	*uploaded = 0
	return newWriter(container, segmentName)
}

func getMtime(object *swift.Object, headers swift.Headers) time.Time {
	if Attr && len(headers) > 0 {
		if HubicTimes {
			if mtime, err := headers.ObjectMetadata().GetHubicModTime(); err == nil {
				return mtime
			}
		} else {
			if mtime, err := headers.ObjectMetadata().GetModTime(); err == nil {
				return mtime
			}
		}
	}
	if object.LastModified.IsZero() {
		return time.Now()
	}
	return object.LastModified
}

func isDirectory(object swift.Object, path string) bool {
	return (object.ContentType == dirContentType) && (object.Name != path) && !object.PseudoDirectory
}

func isLargeObject(object *swift.Object) bool {
	return (object.Bytes == 0) && !object.PseudoDirectory && (object.ContentType != dirContentType)
}

func isPseudoDirectory(object swift.Object, path string) bool {
	return object.PseudoDirectory && (object.Name != path)
}

func isSymlink(object swift.Object, path string) bool {
	return (object.ContentType == linkContentType)
}

func deleteSegments(container, manifestHeader string) error {
	prefix := strings.TrimPrefix(manifestHeader, container+"/")

	// Decode manifest header percent-encoded chars
	prefix = strings.Replace(prefix, "%26", "&", -1)
	prefix = strings.Replace(prefix, "%3F", "?", -1)

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

func formatTime(t time.Time) string {
	if HubicTimes {
		return hubicDateRegex.ReplaceAllString(t.Format(time.RFC3339), "")
	}
	return swift.TimeToFloatString(t)
}

func segmentPath(segmentPrefix string, segmentID *uint) string {
	*segmentID++
	return fmt.Sprintf("%s/%08d", segmentPrefix, *segmentID)
}

func writeSegmentData(fh io.WriteCloser, t *swift.Object, data []byte, uploaded *uint64) error {
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
