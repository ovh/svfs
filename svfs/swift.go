package svfs

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"bazil.org/fuse"

	"github.com/xlucas/swift"
)

func newReader(fh *ObjectHandle) (io.ReadSeeker, error) {
	rd, headers, err := SwiftConnection.ObjectOpen(fh.target.c.Name, fh.target.path, false, nil)
	if err != nil {
		return nil, err
	}

	if Encryption && headers[ObjectNonceHeader] != "" {
		crd := NewCryptoReadSeeker(rd, ChunkSize, int64(Cipher.Overhead()))
		nonce, err := hex.DecodeString(headers[ObjectNonceHeader])
		if err != nil {
			return nil, fmt.Errorf("Failed to decode nonce")
		}
		crd.SetCipher(Cipher, nonce)
		fh.nonce = hex.EncodeToString(crd.Nonce)
		return crd, nil
	}

	return rd, nil
}

func newWriter(container, path string, iv *string) (io.WriteCloser, error) {
	var (
		nonce []byte
		err   error
	)

	headers := map[string]string{"AutoContent": "true"}

	if Encryption {
		if *iv == "" {
			nonce, err = newNonce(Cipher)
			if err != nil {
				return nil, err
			}
		}
		if *iv != "" {
			nonce, err = hex.DecodeString(*iv)
			if err != nil {
				return nil, err
			}
		}
	}

	wd, err := SwiftConnection.ObjectCreate(container, path, false, "", "", headers)
	if err != nil {
		return nil, err
	}

	if Encryption {
		cwd := NewCryptoWriter(wd, ChunkSize, int64(Cipher.Overhead()))
		cwd.SetCipher(Cipher, nonce)
		if *iv == "" {
			*iv = hex.EncodeToString(cwd.Nonce)
		}
		return cwd, err
	}

	return wd, nil
}

func initSegment(c, prefix string, id *uint, t *swift.Object, d []byte, up *uint64, iv *string) (io.WriteCloser, error) {
	segment, err := createSegment(c, prefix, id, up, iv)
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

func createSegment(container, prefix string, id *uint, uploaded *uint64, iv *string) (fh io.WriteCloser, err error) {
	segmentName := segmentPath(prefix, id)
	fh, err = newWriter(container, segmentName, iv)
	*uploaded = 0
	return
}

func getMtime(object *swift.Object, headers swift.Headers) time.Time {
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

func isSymlink(object swift.Object, path string) bool {
	return (object.ContentType == LinkContentType)
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
