package cmd

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"time"

	fuse "bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	"github.com/Sirupsen/logrus"
	"github.com/ovh/svfs/svfs"
	"github.com/spf13/cobra"
	"github.com/xlucas/swift"
)

var (
	debug      bool
	fs         svfs.SVFS
	srv        *fusefs.Server
	profAddr   string
	cpuProf    string
	memProf    string
	cfgFile    string
	device     string
	mountpoint string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "svfs --device name --mountpoint path",
	Short: "The Swift Virtual File System",
	Long: "SVFS is a Virtual File System over Openstack Swift built upon fuse.\n\n" +
		"It is compatible with hubiC, OVH Public Cloud Storage and\n" +
		"basically every endpoint using a standard Openstack Swift setup.\n\n" +
		"It brings a layer of abstraction over object storage,\n" +
		"making it as accessible and convenient as a filesystem,\n" +
		"without being intrusive on the way your data is stored.\n",
	Run: func(cmd *cobra.Command, args []string) {
		//Mandatory flags
		cmd.MarkPersistentFlagRequired("device")
		cmd.MarkPersistentFlagRequired("mountpoint")

		// Debug
		if debug {
			setDebug()
		}

		// Live profiling
		if profAddr != "" {
			go func() {
				if err := http.ListenAndServe(profAddr, nil); err != nil {
					logrus.Fatal(err)
				}
			}()
		}

		// CPU profiling
		if cpuProf != "" {
			createCPUProf(cpuProf)
			defer pprof.StopCPUProfile()
		}

		// Check segment size
		if err := checkOptions(); err != nil {
			logrus.Fatal(err)
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
		logrus.Fatal(err)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	//Swift options
	RootCmd.PersistentFlags().StringVar(&svfs.SwiftConnection.AuthUrl, "os-auth-url", "https://auth.cloud.ovh.net/v2.0", "Authentification URL")
	RootCmd.PersistentFlags().StringVar(&svfs.TargetContainer, "os-container-name", "", "Container name")
	RootCmd.PersistentFlags().StringVar(&svfs.SwiftConnection.AuthToken, "os-auth-token", "", "Authentification token")
	RootCmd.PersistentFlags().StringVar(&svfs.SwiftConnection.UserName, "os-username", "", "Username")
	RootCmd.PersistentFlags().StringVar(&svfs.SwiftConnection.ApiKey, "os-password", "", "User password")
	RootCmd.PersistentFlags().StringVar(&svfs.SwiftConnection.Region, "os-region-name", "", "Region name")
	RootCmd.PersistentFlags().StringVar(&svfs.SwiftConnection.StorageUrl, "os-storage-url", "", "Storage URL")
	RootCmd.PersistentFlags().StringVar(&svfs.SwiftConnection.Tenant, "os-tenant-name", "", "Tenant name")
	RootCmd.PersistentFlags().IntVar(&svfs.SwiftConnection.AuthVersion, "os-auth-version", 0, "Authentification version, 0 = auto")
	RootCmd.PersistentFlags().DurationVar(&svfs.SwiftConnection.ConnectTimeout, "os-connect-timeout", 5*time.Minute, "Swift connection timeout")
	RootCmd.PersistentFlags().Uint64Var(&svfs.SegmentSize, "os-segment-size", 256, "Swift segment size in MiB")
	RootCmd.PersistentFlags().StringVar(&swift.DefaultUserAgent, "user-agent", "svfs/"+svfs.Version, "Default User-Agent")

	//hubiC options
	RootCmd.PersistentFlags().StringVar(&svfs.HubicAuthorization, "hubic-autorization", "", "hubiC authorization code")
	RootCmd.PersistentFlags().StringVar(&svfs.HubicRefreshToken, "hubic-refresh-token", "", "hubiC refresh token")
	RootCmd.PersistentFlags().BoolVar(&svfs.HubicTimes, "hubic-times", false, "Use file times set by hubiC synchronization clients")

	// Permissions
	RootCmd.PersistentFlags().Uint64Var(&svfs.DefaultUID, "default-uid", 0, "Default UID (default 0)")
	RootCmd.PersistentFlags().Uint64Var(&svfs.DefaultGID, "default-gid", 0, "Default GID (default 0)")
	RootCmd.PersistentFlags().Uint64Var(&svfs.DefaultMode, "default-mode", 0700, "Default permissions")
	RootCmd.PersistentFlags().BoolVar(&svfs.AllowRoot, "allow-root", false, "Fuse allow-root option")
	RootCmd.PersistentFlags().BoolVar(&svfs.AllowOther, "allow-other", true, "Fuse allow_other option")
	RootCmd.PersistentFlags().BoolVar(&svfs.DefaultPermissions, "default-permissions", true, "Fuse default_permissions option")
	RootCmd.PersistentFlags().BoolVar(&svfs.ReadOnly, "read-only", false, "Read only access")

	// Prefetch
	RootCmd.PersistentFlags().Uint64Var(&svfs.ListerConcurrency, "readdir-concurrency", 20, "Directory listing concurrency")
	RootCmd.PersistentFlags().BoolVar(&svfs.ExtraAttr, "readdir-extra-attributes", false, "Fetch extra attributes")
	RootCmd.PersistentFlags().UintVar(&svfs.BlockSize, "block-size", 4096, "Block size in bytes")
	RootCmd.PersistentFlags().UintVar(&svfs.ReadAheadSize, "readahead-size", 128, "Per file readhead size in KiB")

	// Cache Options
	RootCmd.PersistentFlags().DurationVar(&svfs.CacheTimeout, "cache-ttl", 1*time.Minute, "Cache timeout")
	RootCmd.PersistentFlags().Int64Var(&svfs.CacheMaxEntries, "cache-max-entires", -1, "Maximum overall entires allowed in cache")
	RootCmd.PersistentFlags().Int64Var(&svfs.CacheMaxAccess, "cache-max-access", -1, "Maximum access count to cached entries")

	// Debug and profiling
	logrus.SetOutput(os.Stdout)
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable fuse debug log")
	RootCmd.PersistentFlags().StringVar(&profAddr, "profile-bind", "", "Profiling information will be served at this address")
	RootCmd.PersistentFlags().StringVar(&cpuProf, "profile-cpu", "", "Write cpu profile to this file")
	RootCmd.PersistentFlags().StringVar(&memProf, "profile-mem", "", "Write memory profile to this file")

	// Mandatory flags
	RootCmd.PersistentFlags().StringVar(&device, "device", "", "Device name")
	RootCmd.PersistentFlags().StringVar(&mountpoint, "mountpoint", "", "Mountpoint")

	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
		logrus.Printf("FUSE: %s\n", msg)
	}
}

func createCPUProf(cpuProf string) {
	f, err := os.Create(cpuProf)
	if err != nil {
		logrus.Fatal(err)
	}
	pprof.StartCPUProfile(f)
}

func createMemProf(memProf string) {
	f, err := os.Create(memProf)
	if err != nil {
		logrus.Fatal(err)
	}
	pprof.WriteHeapProfile(f)

	f.Close()
}
