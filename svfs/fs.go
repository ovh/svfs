package svfs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/xlucas/swift"
	"golang.org/x/net/context"
)

var (
	// SwiftConnection represents a connection to a swift provider.
	// It should be ready for authentication before initializing svfs.
	SwiftConnection = new(swift.Connection)
	// TargetContainer is an existing container ready to be served.
	TargetContainer string
	// ExtraAttr represents extra attributes fetching mode activation.
	ExtraAttr bool
	// HubicTimes represents the usage of hubiC synchronization clients
	// meta headers to read and store file times.
	HubicTimes bool
	// SegmentSize is the size of a segment in bytes.
	SegmentSize uint64

	// AllowRoot represents FUSE allow_root option.
	AllowRoot bool
	// AllowOther represents FUSE allow_other option.
	AllowOther bool
	// DefaultGID is the gid mapped to svfs files.
	DefaultGID uint64
	// DefaultUID is the uid mapped to svfs files.
	DefaultUID uint64
	// DefaultMode is the mode mapped to svfs files.
	DefaultMode uint64
	// DefaultPermissions are permissions mapped to svfs files.
	DefaultPermissions bool
	// BlockSize is the filesystem block size in bytes.
	BlockSize uint
	// ReadAheadSize is the filesystem readahead size in bytes.
	ReadAheadSize uint
	// ReadOnly represents the filesystem readonly access mode activation.
	ReadOnly bool
	// TransferMode represents a mode of operation enabled by the user to indicate
	// that svfs will interact only with processes transferring files. This is seen
	// as an opportunity to optimize network access by reducing requests to Swift.
	TransferMode bool
)

// SVFS implements the Swift Virtual File System.
type SVFS struct{}

// Init sets up the filesystem. It sets configuration settings, starts mandatory
// services and make sure authentication in Swift has succeeded.
func (s *SVFS) Init() (err error) {
	// Copy storage URL option
	overloadStorageURL := SwiftConnection.StorageUrl

	// Use file times set by hubic synchronization clients
	if HubicTimes {
		objectMtimeHeader = hubicMtimeHeader
	}

	// Hubic special authentication
	if HubicAuthorization != "" && HubicRefreshToken != "" {
		SwiftConnection.Auth = new(HubicAuth)
	}

	// Start directory lister
	directoryLister.Start()

	// Authenticate if we don't have a token and storage URL
	if !SwiftConnection.Authenticated() {
		err = SwiftConnection.Authenticate()
	}

	// Swift ACL special authentication
	if overloadStorageURL != "" {
		SwiftConnection.StorageUrl = overloadStorageURL
		SwiftConnection.Auth = newSwiftACLAuth(SwiftConnection.Auth, overloadStorageURL)
	}

	return err
}

// Root gets the root node of the filesystem. It can either be a fake root node
// filled with all the containers found for the given Openstack tenant or a container
// node if a container name have been specified in mount options.
func (s *SVFS) Root() (fs.Node, error) {
	// Mount a specific container
	if TargetContainer != "" {
		return s.rootContainer(TargetContainer)
	}
	// Mount all containers within an account
	return &Root{
		Directory: &Directory{
			apex: true,
		},
	}, nil
}

// Statfs gets the filesystem meta information. It's notably used to report filesystem metrics
// to the host. If the target account is using quota it will be reported as the device size.
// If no quota was found the device size will be equal to the underlying type maximum value.
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
	// Mounting a specific container, then get container usage.
	if TargetContainer != "" {
		c, _, err := SwiftConnection.Container(TargetContainer)
		if err != nil {
			return err
		}
		cs, _, err := SwiftConnection.Container(TargetContainer + segmentContainerSuffix)
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

func (s *SVFS) rootContainer(container string) (fs.Node, error) {
	baseContainer, _, err := SwiftConnection.Container(container)
	if err != nil {
		return nil, err
	}

	// Find segment container too
	segmentContainerName := container + segmentContainerSuffix
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

	return &Directory{
		apex: true,
		c:    &baseContainer,
		cs:   &segmentContainer,
	}, nil
}

var (
	_ fs.FS         = (*SVFS)(nil)
	_ fs.FSStatfser = (*SVFS)(nil)
)
