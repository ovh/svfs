package swift

import (
	"fmt"

	lib "github.com/xlucas/swift"
)

type Connection struct {
	*lib.Connection
	StoragePolicy string
}

func (con *Connection) createContainers(names []string) (err error) {
	for _, name := range names {
		err = con.ContainerCreate(name, lib.Headers{
			StoragePolicyHeader: con.StoragePolicy,
		})
		if err != nil {
			return
		}
	}
	return
}

func (con *Connection) deleteContainers(names []string) (err error) {
	for _, name := range names {
		if err = con.ContainerDelete(name); err != nil {
			return err
		}
	}
	return
}

func (con *Connection) getContainers() (list ContainerList, err error) {
	list = make(ContainerList)

	containers, err := con.ContainersAll(nil)
	if err != nil {
		return
	}

	for _, iter := range containers {
		container := iter
		list[container.Name] = &Container{
			Container: &container,
			Headers:   nil,
		}
	}

	return
}

func (con *Connection) getContainersByNames(names []string) (
	list ContainerList, err error,
) {
	list = make(ContainerList)

	for _, name := range names {
		container, headers, cerr := con.Container(name)
		if cerr == lib.ContainerNotFound {
			continue
		}
		if cerr != nil {
			return nil, cerr
		}
		list[name] = &Container{
			Container: &container,
			Headers:   headers,
		}

	}

	if con.StoragePolicy != "" {
		list = list.FilterByStoragePolicy(con.StoragePolicy)
	}

	return
}

func (con *Connection) DeleteLogicalContainer(container *LogicalContainer,
) (err error,
) {
	for _, name := range []string{
		container.MainContainer.Name,
		container.SegmentContainer.Name,
	} {
		err = con.ContainerDelete(name)
		if err != nil {
			return
		}
	}
	return
}

func (con *Connection) Account() (account *Account, err error) {
	acc, headers, err := con.Connection.Account()
	account = &Account{&acc, headers}
	return
}

func (con *Connection) LogicalContainer(name string) (
	container *LogicalContainer, err error,
) {
	segmentContainer := name + SegmentContainerSuffix
	names := []string{name, segmentContainer}

	// Fetch all containers composing the logical container
	containers, err := con.getContainersByNames(names)
	if err != nil {
		return
	}

	// Main container should exist.
	if containers[name] == nil {
		return nil, fmt.Errorf("Container %s not found", name)
	}
	// Segment container is missing.
	if containers[segmentContainer] == nil {
		// Create it with the adequate storage policy.
		err = con.ContainerCreate(segmentContainer,
			containers[name].SelectHeaders(StoragePolicyHeader))
		if err != nil {
			return
		}
		containers[segmentContainer] = &Container{
			&lib.Container{
				Name: segmentContainer,
			},
			containers[name].SelectHeaders(StoragePolicyHeader),
		}
	}

	container = &LogicalContainer{
		MainContainer:    containers[name],
		SegmentContainer: containers[segmentContainer],
	}

	return
}
