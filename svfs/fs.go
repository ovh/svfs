package svfs

import (
	"crypto/cipher"

	"bazil.org/fuse/fs"
	"github.com/xlucas/swift"
)

var (
	// Swift
	SwiftConnection = new(swift.Connection)
	TargetContainer string
	ExtraAttr       bool
	SegmentSize     uint64

	// FS
	AllowRoot          bool
	AllowOther         bool
	DefaultGID         uint64
	DefaultUID         uint64
	DefaultMode        uint64
	DefaultPermissions bool
	BlockSize          uint
	ReadAheadSize      uint

	// Encryption
	Cipher     cipher.AEAD
	Encryption bool
	KeyFile    string
	Key        []byte
	ChunkSize  int64
)

// SVFS implements the Swift Virtual File System.
type SVFS struct{}

// Init sets up the filesystem. It sets configuration settings, starts mandatory
// services and make sure authentication in Swift has succeeded.
func (s *SVFS) Init() (err error) {
	// Copy storage URL option
	overloadStorageURL := SwiftConnection.StorageUrl

	// Hubic special authentication
	if HubicAuthorization != "" && HubicRefreshToken != "" {
		SwiftConnection.Auth = new(HubicAuth)
	}

	// Start directory lister
	DirectoryLister.Start()

	// Authenticate if we don't have a token and storage URL
	if !SwiftConnection.Authenticated() {
		err = SwiftConnection.Authenticate()
	}

	// Swift ACL special authentication
	if overloadStorageURL != "" {
		SwiftConnection.StorageUrl = overloadStorageURL
		SwiftConnection.Auth = newSwiftACLAuth(SwiftConnection.Auth, overloadStorageURL)
	}

	// Data encryption
	if Encryption {
		Cipher, err = newCipher(Key)
	}

	return err
}

// Root gets the root node of the filesystem. It can either be a fake root node
// filled with all the containers found for the given Openstack tenant or a container
// node if a container name have been specified in mount options.
func (s *SVFS) Root() (fs.Node, error) {
	// Mount a specific container
	if TargetContainer != "" {
		baseContainer, _, err := SwiftConnection.Container(TargetContainer)
		if err != nil {
			return nil, err
		}

		// Find segment container too
		segmentContainerName := TargetContainer + SegmentContainerSuffix
		segmentContainer, _, err := SwiftConnection.Container(segmentContainerName)

		// Create it if missing
		if err == swift.ContainerNotFound {
			var container *swift.Container
			container, err = createContainer(segmentContainerName)
			segmentContainer = *container
		}
		if err != nil && err != swift.ContainerNotFound {
			return nil, err
		}

		return &Container{
			Directory: &Directory{
				apex: true,
				c:    &baseContainer,
				cs:   &segmentContainer,
			},
		}, nil
	}

	// Mount all containers within an account
	return &Root{
		Directory: &Directory{
			apex: true,
		},
	}, nil
}

var _ fs.FS = (*SVFS)(nil)
