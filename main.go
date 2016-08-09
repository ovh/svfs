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

func parseFlags(debug *bool, profAddr, cpuProf, memProf *string) {
	// Swift options
	flag.StringVar(&svfs.SwiftConnection.AuthUrl, "os-auth-url", "https://auth.cloud.ovh.net/v2.0", "Authentication URL")
	flag.StringVar(&svfs.TargetContainer, "os-container-name", "", "Container name")
	flag.StringVar(&svfs.SwiftConnection.AuthToken, "os-auth-token", "", "Authentication token")
	flag.StringVar(&svfs.SwiftConnection.ApiKey, "os-password", "", "User password")
	flag.StringVar(&svfs.SwiftConnection.UserName, "os-username", "", "User name")
	flag.StringVar(&svfs.SwiftConnection.Region, "os-region-name", "", "Region")
	flag.StringVar(&svfs.SwiftConnection.StorageUrl, "os-storage-url", "", "Storage URL")
	flag.StringVar(&svfs.SwiftConnection.Tenant, "os-tenant-name", "", "Tenant name")
	flag.IntVar(&svfs.SwiftConnection.AuthVersion, "os-auth-version", 0, "Authentication version, 0 = auto")
	flag.DurationVar(&svfs.SwiftConnection.ConnectTimeout, "os-connect-timeout", 15*time.Second, "Swift connection timeout")
	flag.DurationVar(&svfs.SwiftConnection.Timeout, "os-request-timeout", 5*time.Minute, "Swift operation timeout")
	flag.Uint64Var(&svfs.SegmentSize, "os-segment-size", 256, "Swift segment size in MiB")
	flag.StringVar(&swift.DefaultUserAgent, "user-agent", "svfs/"+svfs.Version, "Default User-Agent")

	// Hubic
	flag.StringVar(&svfs.HubicAuthorization, "hubic-authorization", "", "Hubic authorization code")
	flag.StringVar(&svfs.HubicRefreshToken, "hubic-refresh-token", "", "Hubic refresh token")
	flag.BoolVar(&svfs.HubicTimes, "hubic-times", false, "Use file times set by hubiC synchronization clients")

	// Permissions
	flag.Uint64Var(&svfs.DefaultUID, "default-uid", 0, "Default UID (default 0)")
	flag.Uint64Var(&svfs.DefaultGID, "default-gid", 0, "Default GID (default 0)")
	flag.Uint64Var(&svfs.DefaultMode, "default-mode", 0700, "Default permissions")
	flag.BoolVar(&svfs.AllowRoot, "allow-root", false, "Fuse allow_root option")
	flag.BoolVar(&svfs.AllowOther, "allow-other", true, "Fuse allow_other option")
	flag.BoolVar(&svfs.DefaultPermissions, "default-permissions", true, "Fuse default_permissions option")
	flag.BoolVar(&svfs.ReadOnly, "read-only", false, "Read only access")

	// Prefetch
	flag.Uint64Var(&svfs.ListerConcurrency, "readdir-concurrency", 20, "Directory listing concurrency")
	flag.BoolVar(&svfs.ExtraAttr, "readdir-extra-attributes", false, "Fetch extra attributes")
	flag.UintVar(&svfs.BlockSize, "block-size", 4096, "Block size in bytes")
	flag.UintVar(&svfs.ReadAheadSize, "readahead-size", 128, "Per file readahead size in KiB")

	// Cache Options
	flag.DurationVar(&svfs.CacheTimeout, "cache-ttl", 1*time.Minute, "Cache timeout")
	flag.Int64Var(&svfs.CacheMaxEntries, "cache-max-entries", -1, "Maximum overall entries allowed in cache")
	flag.Int64Var(&svfs.CacheMaxAccess, "cache-max-access", -1, "Maximum access count to cached entries")

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

func mountOptions(device string) (options []fuse.MountOption) {
	if svfs.AllowOther {
		options = append(options, fuse.AllowOther())
	}
	if svfs.AllowRoot {
		options = append(options, fuse.AllowRoot())
	}
	if svfs.DefaultPermissions {
		options = append(options, fuse.DefaultPermissions())
	}
	if svfs.ReadOnly {
		options = append(options, fuse.ReadOnly())
	}

	options = append(options, fuse.MaxReadahead(uint32(svfs.ReadAheadSize)))
	options = append(options, fuse.Subtype("svfs"))
	options = append(options, fuse.FSName(device))

	return options
}

func checkOptions() error {
	// Convert to MB
	svfs.SegmentSize *= (1 << 20)
	svfs.ReadAheadSize *= (1 << 10)

	// Should not exceed swift maximum object size.
	if svfs.SegmentSize > 5*(1<<30) {
		return fmt.Errorf("Segment size can't exceed 5 GiB")
	}
	return nil
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
		fs       svfs.SVFS
		srv      *fusefs.Server
		profAddr string
		cpuProf  string
		memProf  string
	)

	parseFlags(&debug, &profAddr, &cpuProf, &memProf)

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

	if err := checkOptions(); err != nil {
		log.Fatal(err)
	}

	// Mount SVFS
	c, err := fuse.Mount(mountpoint, mountOptions(device)...)
	if err != nil {
		goto Err
	}
	defer c.Close()

	// Initialize SVFS
	if err = fs.Init(); err != nil {
		goto Err
	}

	// Serve SVFS
	srv = fusefs.New(c, nil)
	if err = srv.Serve(&fs); err != nil {
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
