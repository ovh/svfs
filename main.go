package main

import (
	"fmt"
	"log"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	"github.com/ovh/svfs/driver"
	swiftfs "github.com/ovh/svfs/fs/swift"
	swiftfuse "github.com/ovh/svfs/fuse"
	"github.com/ovh/svfs/signal"
	"github.com/ovh/svfs/store"
	ctx "golang.org/x/net/context"

	_ "github.com/ovh/svfs/store/drivers"
	_ "github.com/ovh/svfs/store/drivers/register"
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

	// Initialize store
	store := driver.GetGroup("store").Get("Bolt").(store.Store)
	store.Init("/dev/shm/svfs.store")

	fs, err := swiftfuse.NewSVFS(store)
	if err != nil {
		log.Fatal(err)
	}

	context := ctx.Background()

	// Filesystem configuration
	// Hardcoded version here, for a first draft
	opts := &swiftfs.FsConfiguration{
		BlockSize:     4096,
		MaxConn:       30,
		Perms:         0700,
		Gid:           1000,
		Uid:           1000,
		OsAuthURL:     "https://auth.provider/version",
		OsUserName:    "user",
		OsPassword:    "pass",
		OsTenantName:  "tenant",
		OsRegionName:  "region",
		StoragePolicy: "Policy",
		StorePath:     "/dev/shm/svfs.db",
		StoreDriver:   "bolt",
	}

	err = fs.Setup(context, opts)
	if err != nil {
		log.Fatal(err)
	}

	// Mount the FUSE fs
	c, err := fuse.Mount(mountpoint, fuse.Subtype("svfs"))
	srv := fusefs.New(c, nil)

	// Catch signals
	signal.Trap(cleanup, mountpoint)

	// Serve FUSE requests
	err = srv.Serve(fs)

	// Teardown
	<-c.Ready

	if err != nil {
		log.Fatal(err)
	}
}
