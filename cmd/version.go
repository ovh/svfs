package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/ovh/svfs/svfs"
	"github.com/spf13/cobra"
)

const (
	currentVersion = "v" + svfs.Version
	gitHubOwner    = "OVH"
	gitHubRepo     = "svfs"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

// Get svfs version information.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the svfs version and any available update",
	Long: "Display information about current SVFS version and\n" +
		"check if a new version is available on GitHub.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = false
		client := github.NewClient(nil)

		// Print current version information
		color.White("Version information:\n\n")
		current, _, err := client.Repositories.GetReleaseByTag(gitHubOwner, gitHubRepo, currentVersion)
		printRelease(current, err)
		if err != nil {
			return err
		}

		// Get the latest release
		releases, _, err := client.Repositories.ListReleases(
			gitHubOwner,
			gitHubRepo,
			&github.ListOptions{PerPage: 1})
		latest := releases[len(releases)-1]

		// Print if an update is available
		if *current.TagName == *latest.TagName {
			color.Green("\nYour version is up to date.")
		} else {
			color.Yellow("\nA new version is available:\n\n")
			printRelease(latest, err)
			color.Yellow("\nVisit %s for more information.\n", *latest.HTMLURL)
		}

		return err
	},
}

// Print release information.
func printRelease(release *github.RepositoryRelease, err error) {
	if err != nil {
		color.Red("\nYour version is invalid! Check the error below.\n%v", err)
		return
	}

	fmt.Printf("* Release name: %s\n", *release.Name)
	fmt.Printf("* Related git tag: %s\n", *release.TagName)
	fmt.Printf("* Stable release: %t\n", !*release.Prerelease)

	if release.PublishedAt != nil && !release.PublishedAt.IsZero() {
		fmt.Printf("* Release date: %s\n", *release.PublishedAt)
	}
}
