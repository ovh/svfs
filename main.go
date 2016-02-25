// SVFS implements a virtual file system for Openstack Swift.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ncw/swift"
	"github.com/xlucas/svfs/svfs"

	fuse "bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
)

func main() {
	var (
		debug bool
		fs    *svfs.SVFS
		sc    = swift.Connection{}
		srv   *fusefs.Server
		conf  = svfs.Config{}
		cconf = svfs.CacheConfig{}
	)

	// Logger
	log.SetOutput(os.Stdout)
	flag.BoolVar(&debug, "debug", false, "Enable fuse debug log")

	// Swift options
	flag.StringVar(&sc.AuthUrl, "os-auth-url", "https://auth.cloud.ovh.net/v2.0", "Authentication URL")
	flag.StringVar(&conf.Container, "os-container-name", "", "Container name")
	flag.StringVar(&sc.AuthToken, "os-auth-token", "", "Authentication token")
	flag.StringVar(&sc.ApiKey, "os-password", "", "User password")
	flag.StringVar(&sc.UserName, "os-username", "", "User name")
	flag.StringVar(&sc.Region, "os-region-name", "", "Region")
	flag.StringVar(&sc.StorageUrl, "os-storage-url", "", "Storage URL")
	flag.StringVar(&sc.Tenant, "os-tenant-name", "", "Tenant name")
	flag.IntVar(&sc.AuthVersion, "os-auth-version", 0, "Authentication version, 0 = auto")
	flag.DurationVar(&conf.ConnectTimeout, "os-connect-timeout", 5*time.Minute, "Swift connection timeout")

	// Concurrency
	flag.Uint64Var(&conf.MaxReaddirConcurrency, "max-readdir-concurrency", 20, "Overall concurrency factor when listing directories")
	flag.UintVar(&conf.ReadAheadSize, "readahead-size", 131072, "Per file readahead size in bytes")

	// Cache Options
	flag.DurationVar(&cconf.Timeout, "cache-ttl", 1*time.Minute, "Cache timeout")
	flag.Int64Var(&cconf.MaxEntries, "cache-max-entries", -1, "Maximum overall entries allowed in cache")
	flag.Int64Var(&cconf.MaxAccess, "cache-max-access", -1, "Maximum access count to cached entries")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage : %s [OPTIONS] MOUNTPOINT\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Available options :\n")
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
		fuse.MaxReadahead(uint32(conf.ReadAheadSize)),
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
	if err = fs.Init(&sc, &conf, &cconf); err != nil {
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
