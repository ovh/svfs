package svfs

import (
	"crypto/cipher"

	"golang.org/x/net/context"

	"bazil.org/fuse"
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

func (s *SVFS) Statfs(ctx context.Context, req *fuse.StatfsRequest, resp *fuse.StatfsResponse) error {
	account, _, err := SwiftConnection.Account()
	if err != nil {
		return err
	}

	resp.Bsize = uint32(BlockSize)

	// Not mounting a specific container, then get account
	// information.
	if TargetContainer == "" {
		resp.Files = uint64(account.Objects)
		resp.Blocks = uint64(account.BytesUsed) / uint64(resp.Bsize)
	}
	// Mount a specific container, then get container usage.
	if TargetContainer != "" {
		c, _, err := SwiftConnection.Container(TargetContainer)
		if err != nil {
			return err
		}
		cs, _, err := SwiftConnection.Container(TargetContainer + SegmentContainerSuffix)
		if err != nil {
			return err
		}
		resp.Files = uint64(c.Count)
		resp.Blocks = uint64(c.Bytes+cs.Bytes) / uint64(resp.Bsize)
	}
	// An account quota has been set, compute relative free space.
	if account.Quota > 0 {
		resp.Bavail = uint64(account.Quota-account.BytesUsed) / uint64(resp.Bsize)
		resp.Bfree = resp.Bavail
		if TargetContainer == "" {
			resp.Blocks = uint64(account.Quota) / uint64(resp.Bsize)
		} else {
			resp.Blocks = uint64(account.Quota-account.BytesUsed)/uint64(resp.Bsize) + resp.Blocks
		}
	} else {
		// Else there's theorically no limit to available storage space.
		used := resp.Blocks
		resp.Blocks = uint64(1<<63-1) / uint64(resp.Bsize)
		resp.Bavail = resp.Blocks - used
		resp.Bfree = resp.Bavail
	}

	return nil
}

var (
	_ fs.FS         = (*SVFS)(nil)
	_ fs.FSStatfser = (*SVFS)(nil)
)
