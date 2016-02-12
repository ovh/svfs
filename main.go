// SVFS implements a virtual file system for Openstack Swift.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ncw/swift"
	"github.com/xlucas/svfs/fs"

	fuse "bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
)

func main() {
	sc := swift.Connection{}
	log.SetOutput(os.Stderr)

	// FS options
	flag.StringVar(&sc.UserName, "u", "", "User name")
	flag.StringVar(&sc.ApiKey, "p", "", "User password")
	flag.StringVar(&sc.AuthUrl, "a", "https://auth.cloud.ovh.net/v2.0", "Authentication URL")
	flag.StringVar(&sc.Region, "r", "", "Region")
	flag.StringVar(&sc.Tenant, "t", "", "Tenant name")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s :\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

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
	if err = sc.Authenticate(); err != nil {
		log.Fatal(err)
	}

	// Init SVFS
	svfs := &fs.SVFS{}
	if err = svfs.Init(&sc); err != nil {
		log.Fatal(err)
	}

	// Serve SVFS
	srv := fusefs.New(c, nil)
	if err = srv.Serve(svfs); err != nil {
		log.Fatal(err)
	}

	// Check for mount errors
	<-c.Ready
	if err = c.MountError; err != nil {
		log.Fatal(err)
	}
}
