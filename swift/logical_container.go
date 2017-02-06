package swift

import (
	"time"

	lib "github.com/xlucas/swift"
)

const (
	SegmentContainerSuffix = "_segments"
)

// LogicalContainer is a collection of containers accessed as one. It is used
// to represent a container couple formed by a main container holding non
// segmented objects and manifests and a segment container which will hold
// segments referenced by manifests.
type LogicalContainer struct {
	MainContainer    *Container
	SegmentContainer *Container
}

func NewLogicalContainer(con *Connection, mainContainerName string) (
	container *LogicalContainer, err error,
) {
	segmentContainerName := mainContainerName + SegmentContainerSuffix
	containers := []string{mainContainerName, segmentContainerName}

	for _, name := range containers {
		err = con.ContainerCreate(name,
			lib.Headers{
				StoragePolicyHeader: con.StorageUrl,
			})
		if err != nil {
			return
		}
	}

	list, err := con.getContainersByNames(containers)

	container = &LogicalContainer{
		MainContainer:    list[mainContainerName],
		SegmentContainer: list[segmentContainerName],
	}

	return
}

func (c *LogicalContainer) Bytes() int64 {
	return c.MainContainer.Bytes + c.SegmentContainer.Bytes
}

func (c *LogicalContainer) CreationTime() time.Time {
	return c.MainContainer.CreationTime()
}
