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
	"github.com/fatih/color"
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

func init() {
	// Logger
	formatter := new(logrus.TextFormatter)
	formatter.TimestampFormat = time.RFC3339
	formatter.FullTimestamp = true
	logrus.SetFormatter(formatter)
	logrus.SetOutput(os.Stdout)

	setFlags()
	RootCmd.AddCommand(mountCmd)
}

// mountCmd represents the base command when called without any subcommands
var mountCmd = &cobra.Command{
	Use:   "mount --device name --mountpoint path",
	Short: "Mount object storage as a device",
	Long: "Mount object storage either from HubiC or a vanilla Swift access\n" +
		"as a device at the given mountpoint.",
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

func setFlags() {
	flags := mountCmd.PersistentFlags()

	//Swift options
	flags.StringVar(&svfs.SwiftConnection.AuthUrl, "os-auth-url", "https://auth.cloud.ovh.net/v2.0", "Authentification URL")
	flags.StringVar(&svfs.TargetContainer, "os-container-name", "", "Container name")
	flags.StringVar(&svfs.SwiftConnection.AuthToken, "os-auth-token", "", "Authentification token")
	flags.StringVar(&svfs.SwiftConnection.UserName, "os-username", "", "Username")
	flags.StringVar(&svfs.SwiftConnection.ApiKey, "os-password", "", "User password")
	flags.StringVar(&svfs.SwiftConnection.Region, "os-region-name", "", "Region name")
	flags.StringVar(&svfs.SwiftConnection.StorageUrl, "os-storage-url", "", "Storage URL")
	flags.StringVar(&svfs.SwiftConnection.Tenant, "os-tenant-name", "", "Tenant name")
	flags.IntVar(&svfs.SwiftConnection.AuthVersion, "os-auth-version", 0, "Authentification version, 0 = auto")
	flags.DurationVar(&svfs.SwiftConnection.ConnectTimeout, "os-connect-timeout", 15*time.Second, "Swift connection timeout")
	flags.DurationVar(&svfs.SwiftConnection.Timeout, "os-request-timeout", 5*time.Minute, "Swift operation timeout")
	flags.Uint64Var(&svfs.SegmentSize, "os-segment-size", 256, "Swift segment size in MiB")
	flags.StringVar(&swift.DefaultUserAgent, "user-agent", "svfs/"+svfs.Version, "Default User-Agent")

	//HubiC options
	flags.StringVar(&svfs.HubicAuthorization, "hubic-autorization", "", "hubiC authorization code")
	flags.StringVar(&svfs.HubicRefreshToken, "hubic-refresh-token", "", "hubiC refresh token")
	flags.BoolVar(&svfs.HubicTimes, "hubic-times", false, "Use file times set by hubiC synchronization clients")

	// Permissions
	flags.Uint64Var(&svfs.DefaultUID, "default-uid", 0, "Default UID (default 0)")
	flags.Uint64Var(&svfs.DefaultGID, "default-gid", 0, "Default GID (default 0)")
	flags.Uint64Var(&svfs.DefaultMode, "default-mode", 0700, "Default permissions")
	flags.BoolVar(&svfs.AllowRoot, "allow-root", false, "Fuse allow-root option")
	flags.BoolVar(&svfs.AllowOther, "allow-other", true, "Fuse allow_other option")
	flags.BoolVar(&svfs.DefaultPermissions, "default-permissions", true, "Fuse default_permissions option")
	flags.BoolVar(&svfs.ReadOnly, "read-only", false, "Read only access")

	// Prefetch
	flags.Uint64Var(&svfs.ListerConcurrency, "readdir-concurrency", 20, "Directory listing concurrency")
	flags.BoolVar(&svfs.ExtraAttr, "readdir-extra-attributes", false, "Fetch extra attributes")
	flags.UintVar(&svfs.BlockSize, "block-size", 4096, "Block size in bytes")
	flags.UintVar(&svfs.ReadAheadSize, "readahead-size", 128, "Per file readhead size in KiB")
	flags.BoolVar(&svfs.TransferMode, "transfer-mode", false, "Enable transfer mode")

	// Cache Options
	flags.DurationVar(&svfs.CacheTimeout, "cache-ttl", 1*time.Minute, "Cache timeout")
	flags.Int64Var(&svfs.CacheMaxEntries, "cache-max-entires", -1, "Maximum overall entires allowed in cache")
	flags.Int64Var(&svfs.CacheMaxAccess, "cache-max-access", -1, "Maximum access count to cached entries")

	// Debug and profiling
	flags.BoolVar(&debug, "debug", false, "Enable fuse debug log")
	flags.StringVar(&profAddr, "profile-bind", "", "Profiling information will be served at this address")
	flags.StringVar(&cpuProf, "profile-cpu", "", "Write cpu profile to this file")
	flags.StringVar(&memProf, "profile-ram", "", "Write memory profile to this file")

	// Mandatory flags
	flags.StringVar(&device, "device", "", "Device name")
	flags.StringVar(&mountpoint, "mountpoint", "", "Mountpoint")

	mountCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
	logrus.SetLevel(logrus.DebugLevel)
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	fuse.Debug = func(msg interface{}) {
		logrus.WithField("source", yellow("fuse")).Debugln(blue(msg))
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
