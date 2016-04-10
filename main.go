// SVFS implements a virtual file system for Openstack Swift.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"time"

	"github.com/ovh/svfs/svfs"
	"github.com/xlucas/swift"

	fuse "bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
)

func parseFlags(debug *bool, conf *svfs.Config, cconf *svfs.CacheConfig, sc *swift.Connection, profAddr, cpuProf, memProf *string) {
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
	flag.Uint64Var(&conf.SegmentSizeMB, "os-segment-size", 256, "Swift segment size in MB")

	// Hubic
	flag.StringVar(&svfs.HubicAuthorization, "hubic-authorization", "", "Hubic authorization code")
	flag.StringVar(&svfs.HubicRefreshToken, "hubic-refresh-token", "", "Hubic refresh token")

	// Permissions
	flag.Uint64Var(&svfs.DefaultUID, "default-uid", 0, "Default UID (default 0)")
	flag.Uint64Var(&svfs.DefaultGID, "default-gid", 0, "Default GID (default 0)")
	flag.Uint64Var(&svfs.DefaultMode, "default-mode", 0700, "Default permissions")
	flag.BoolVar(&conf.MountAllowRoot, "allow-root", false, "Fuse allow_root option")
	flag.BoolVar(&conf.MountAllowOther, "allow-other", true, "Fuse allow_other option")
	flag.BoolVar(&conf.MountDefaultPerm, "default-permissions", true, "Fuse default_permissions option")

	// Prefetch
	flag.Uint64Var(&conf.ListConcurrency, "readdir-concurrency", 20, "Directory listing concurrency")
	flag.BoolVar(&svfs.ExtraAttr, "readdir-extra-attributes", false, "Fetch extra attributes")
	flag.UintVar(&conf.ReadAheadSize, "readahead-size", 131072, "Per file readahead size in bytes")

	// Cache Options
	flag.DurationVar(&cconf.Timeout, "cache-ttl", 1*time.Minute, "Cache timeout")
	flag.Int64Var(&cconf.MaxEntries, "cache-max-entries", -1, "Maximum overall entries allowed in cache")
	flag.Int64Var(&cconf.MaxAccess, "cache-max-access", -1, "Maximum access count to cached entries")

	// Debug and profiling
	log.SetOutput(os.Stdout)
	flag.BoolVar(debug, "debug", false, "Enable fuse debug log")
	flag.StringVar(profAddr, "profile-bind", "", "Profiling information will be served at this address")
	flag.StringVar(cpuProf, "profile-cpu", "", "Write cpu profile to this file")
	flag.StringVar(memProf, "profile-ram", "", "Write memory profile to this file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage : %s [OPTIONS] DEVICE MOUNTPOINT\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Available options :\n")
		flag.PrintDefaults()
	}

	flag.Parse()
}

func mountOptions(device string, conf *svfs.Config) (options []fuse.MountOption) {
	if conf.MountAllowRoot && conf.MountAllowOther {
		log.Fatalf("AllowRoot and AllowOther are mutually exclusive")
	}

	if conf.MountAllowOther {
		options = append(options, fuse.AllowOther())
	}
	if conf.MountAllowRoot {
		options = append(options, fuse.AllowRoot())
	}
	if conf.MountDefaultPerm {
		options = append(options, fuse.DefaultPermissions())
	}

	options = append(options, fuse.MaxReadahead(uint32(conf.ReadAheadSize)))
	options = append(options, fuse.Subtype("svfs"))
	options = append(options, fuse.FSName(device))

	return options
}

func setDebug() {
	fuse.Debug = func(msg interface{}) {
		log.Printf("FUSE: %s\n", msg)
	}
}

func createCPUProf(cpuProf string) {
	f, err := os.Create(cpuProf)
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
}

func createMemProf(memProf string) {
	f, err := os.Create(memProf)
	if err != nil {
		log.Fatal(err)
	}
	pprof.WriteHeapProfile(f)
	f.Close()
}

func main() {
	var (
		debug    bool
		fs       = &svfs.SVFS{}
		sc       = swift.Connection{}
		srv      *fusefs.Server
		conf     = svfs.Config{}
		cconf    = svfs.CacheConfig{}
		profAddr string
		cpuProf  string
		memProf  string
	)

	parseFlags(
		&debug,
		&conf,
		&cconf,
		&sc,
		&profAddr,
		&cpuProf,
		&memProf,
	)

	// Debug
	if debug {
		setDebug()
	}

	// Live profiling
	if profAddr != "" {
		go func() {
			if err := http.ListenAndServe(profAddr, nil); err != nil {
				log.Fatal(err)
			}
		}()
	}

	// CPU profiling
	if cpuProf != "" {
		createCPUProf(cpuProf)
		defer pprof.StopCPUProfile()
	}

	// Mountpoint is mandatory
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}

	device := os.Args[len(os.Args)-2]
	mountpoint := os.Args[len(os.Args)-1]

	// Mount SVFS
	c, err := fuse.Mount(mountpoint, mountOptions(device, &conf)...)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Initialize SVFS
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

	// Memory profiling
	if memProf != "" {
		createMemProf(memProf)
	}

	if err = c.MountError; err != nil {
		goto Err
	}

	return

Err:
	fuse.Unmount(mountpoint)
	log.Fatal(err)
}
