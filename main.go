package main

import (
	"fmt"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	swiftfs "github.com/ovh/svfs/fs/swift"
	swiftfuse "github.com/ovh/svfs/fuse"
	"github.com/ovh/svfs/signal"
	ctx "golang.org/x/net/context"
)

func init() {
	fuse.Debug = func(msg interface{}) {
		fmt.Println(msg)
	}
}

func cleanup(mountpoint interface{}) {
	fmt.Println("Cleaning up")
	fuse.Unmount(mountpoint.(string))
}

func main() {
	mountpoint := "/tmp/svfs"

	fs := &swiftfuse.SVFS{Fs: &swiftfs.Fs{}}

	context := ctx.Background()

	opts := &swiftfs.FsConfiguration{
		BlockSize:     4096,
		Connections:   30,
		Perms:         0700,
		Gid:           3882,
		Uid:           3882,
		OsAuthURL:     "https://auth.cloud.ovh.net/v2.0",
		OsUserName:    "user",
		OsPassword:    "passwd",
		OsTenantName:  "tenant",
		OsRegionName:  "region",
		StoragePolicy: "PCS",
	}

	fs.Setup(context, opts)

	// Mount the FUSE fs
	c, err := fuse.Mount(mountpoint, fuse.Subtype("svfs"))
	srv := fusefs.New(c, nil)

	// Catch signals
	signal.Trap(cleanup, mountpoint)

	// Serve FUSE requests
	srv.Serve(fs)

	// Teardown
	<-c.Ready

	if err != nil {
		fmt.Println(err)
	}
}
