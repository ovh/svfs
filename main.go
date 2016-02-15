// SVFS implements a virtual file system for Openstack Swift.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ncw/swift"
	"github.com/xlucas/svfs/svfs"

	fuse "bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
)

func main() {
	var (
		debug bool
		fs    *svfs.SVFS
		srv   *fusefs.Server
		sc    = swift.Connection{}
	)

	// Logger
	log.SetOutput(os.Stdout)

	// FS options
	flag.StringVar(&sc.UserName, "u", "", "User name")
	flag.StringVar(&sc.ApiKey, "p", "", "User password")
	flag.StringVar(&sc.AuthUrl, "a", "https://auth.cloud.ovh.net/v2.0", "Authentication URL")
	flag.StringVar(&sc.Region, "r", "", "Region")
	flag.StringVar(&sc.Tenant, "t", "", "Tenant name")
	flag.StringVar(&sc.StorageUrl, "s", "", "Storage URL")
	flag.StringVar(&sc.AuthToken, "k", "", "Authentication token")
	flag.IntVar(&sc.AuthVersion, "v", 0, "Authentication version")
	flag.BoolVar(&debug, "debug", false, "Enable fuse debug log")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s :\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	// Debug
	if debug {
		fuse.Debug = func(msg interface{}) {
			log.Printf("FUSE: %s\n", msg)
		}
	}

	// Mountpoint is mandatory
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	mountpoint := os.Args[len(os.Args)-1]

	// Mount SVFS
	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("svfs"),
		fuse.Subtype("svfs"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Pre-Serve: authenticate to identity endpoint
	// if no token is specified
	if !sc.Authenticated() {
		if err = sc.Authenticate(); err != nil {
			goto Err
		}
	}

	// Init SVFS
	fs = &svfs.SVFS{}
	if err = fs.Init(&sc); err != nil {
		goto Err
	}

	// Serve SVFS
	srv = fusefs.New(c, nil)
	if err = srv.Serve(fs); err != nil {
		goto Err
	}

	// Check for mount errors
	<-c.Ready
	if err = c.MountError; err != nil {
		goto Err
	}

	return

Err:
	fuse.Unmount(mountpoint)
	log.Fatal(err)
}
