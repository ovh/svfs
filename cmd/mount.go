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
	//Swift options
	mountCmd.PersistentFlags().StringVar(&svfs.SwiftConnection.AuthUrl, "os-auth-url", "https://auth.cloud.ovh.net/v2.0", "Authentification URL")
	mountCmd.PersistentFlags().StringVar(&svfs.TargetContainer, "os-container-name", "", "Container name")
	mountCmd.PersistentFlags().StringVar(&svfs.SwiftConnection.AuthToken, "os-auth-token", "", "Authentification token")
	mountCmd.PersistentFlags().StringVar(&svfs.SwiftConnection.UserName, "os-username", "", "Username")
	mountCmd.PersistentFlags().StringVar(&svfs.SwiftConnection.ApiKey, "os-password", "", "User password")
	mountCmd.PersistentFlags().StringVar(&svfs.SwiftConnection.Region, "os-region-name", "", "Region name")
	mountCmd.PersistentFlags().StringVar(&svfs.SwiftConnection.StorageUrl, "os-storage-url", "", "Storage URL")
	mountCmd.PersistentFlags().StringVar(&svfs.SwiftConnection.Tenant, "os-tenant-name", "", "Tenant name")
	mountCmd.PersistentFlags().IntVar(&svfs.SwiftConnection.AuthVersion, "os-auth-version", 0, "Authentification version, 0 = auto")
	mountCmd.PersistentFlags().DurationVar(&svfs.SwiftConnection.ConnectTimeout, "os-connect-timeout", 15*time.Second, "Swift connection timeout")
	mountCmd.PersistentFlags().DurationVar(&svfs.SwiftConnection.Timeout, "os-request-timeout", 5*time.Minute, "Swift operation timeout")
	mountCmd.PersistentFlags().Uint64Var(&svfs.SegmentSize, "os-segment-size", 256, "Swift segment size in MiB")
	mountCmd.PersistentFlags().StringVar(&swift.DefaultUserAgent, "user-agent", "svfs/"+svfs.Version, "Default User-Agent")

	//HubiC options
	mountCmd.PersistentFlags().StringVar(&svfs.HubicAuthorization, "hubic-autorization", "", "hubiC authorization code")
	mountCmd.PersistentFlags().StringVar(&svfs.HubicRefreshToken, "hubic-refresh-token", "", "hubiC refresh token")
	mountCmd.PersistentFlags().BoolVar(&svfs.HubicTimes, "hubic-times", false, "Use file times set by hubiC synchronization clients")

	// Permissions
	mountCmd.PersistentFlags().Uint64Var(&svfs.DefaultUID, "default-uid", 0, "Default UID (default 0)")
	mountCmd.PersistentFlags().Uint64Var(&svfs.DefaultGID, "default-gid", 0, "Default GID (default 0)")
	mountCmd.PersistentFlags().Uint64Var(&svfs.DefaultMode, "default-mode", 0700, "Default permissions")
	mountCmd.PersistentFlags().BoolVar(&svfs.AllowRoot, "allow-root", false, "Fuse allow-root option")
	mountCmd.PersistentFlags().BoolVar(&svfs.AllowOther, "allow-other", true, "Fuse allow_other option")
	mountCmd.PersistentFlags().BoolVar(&svfs.DefaultPermissions, "default-permissions", true, "Fuse default_permissions option")
	mountCmd.PersistentFlags().BoolVar(&svfs.ReadOnly, "read-only", false, "Read only access")

	// Prefetch
	mountCmd.PersistentFlags().Uint64Var(&svfs.ListerConcurrency, "readdir-concurrency", 20, "Directory listing concurrency")
	mountCmd.PersistentFlags().BoolVar(&svfs.ExtraAttr, "readdir-extra-attributes", false, "Fetch extra attributes")
	mountCmd.PersistentFlags().UintVar(&svfs.BlockSize, "block-size", 4096, "Block size in bytes")
	mountCmd.PersistentFlags().UintVar(&svfs.ReadAheadSize, "readahead-size", 128, "Per file readhead size in KiB")
	mountCmd.PersistentFlags().BoolVar(&svfs.TransferMode, "transfer-mode", false, "Enable transfer mode")

	// Cache Options
	mountCmd.PersistentFlags().DurationVar(&svfs.CacheTimeout, "cache-ttl", 1*time.Minute, "Cache timeout")
	mountCmd.PersistentFlags().Int64Var(&svfs.CacheMaxEntries, "cache-max-entires", -1, "Maximum overall entires allowed in cache")
	mountCmd.PersistentFlags().Int64Var(&svfs.CacheMaxAccess, "cache-max-access", -1, "Maximum access count to cached entries")

	// Debug and profiling
	mountCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable fuse debug log")
	mountCmd.PersistentFlags().StringVar(&profAddr, "profile-bind", "", "Profiling information will be served at this address")
	mountCmd.PersistentFlags().StringVar(&cpuProf, "profile-cpu", "", "Write cpu profile to this file")
	mountCmd.PersistentFlags().StringVar(&memProf, "profile-mem", "", "Write memory profile to this file")

	// Mandatory flags
	mountCmd.PersistentFlags().StringVar(&device, "device", "", "Device name")
	mountCmd.PersistentFlags().StringVar(&mountpoint, "mountpoint", "", "Mountpoint")

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
