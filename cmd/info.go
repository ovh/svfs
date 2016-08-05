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
	gitOwner       = "OVH"
	gitRepo        = "svfs"
)

// Print some SVFS information
func printReleaseInfo(name, tag string, time *github.Timestamp, prerelease bool, err error) {
	p := fmt.Printf

	p("* Version name : %s\n", name)
	p("* Tag name : %s\n", tag)
	p("* Is pre-release : %t\n", prerelease)
	if time != nil {
		p("* Published at : %s\n", *time)
	}
	if err != nil {
		color.Red("\nYour SVFS version is invalid ! Check error just bellow.")
	}
}

// Get more information on GitHub API about the current release you have
func getCurrentReleaseInfo(client *github.Client) (cName, cTag string, cTime *github.Timestamp, cPrerelease bool, err error) {
	orgs, _, err := client.Repositories.GetReleaseByTag(gitOwner, gitRepo, currentVersion)
	if err != nil {
		return "?", "?", nil, false, err
	}
	c := orgs

	return *(c.Name), *(c.TagName), c.PublishedAt, *(c.Prerelease), nil
}

// Get more information on GitHub API about the latest release
func getLastReleaseInfo(client *github.Client) (lName, lTag string, lTime *github.Timestamp, lPrerelease bool) {
	opt := &github.ListOptions{PerPage: 1}
	orgs, _, err := client.Repositories.ListReleases(gitOwner, gitRepo, opt)
	if err != nil {
		return "?", "?", nil, false
	}

	l := orgs[len(orgs)-1]

	return *(l.Name), *(l.TagName), l.PublishedAt, *(l.Prerelease)
}

// Represents the info command, lVar for lastest release, cVar for current release
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display fs-wide information",
	Long: "Display some information about SVFS current version and\n" +
		"check if a new version of SVFS exist using GitHub API.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = false

		client := github.NewClient(nil)

		color.White("Current SVFS version:\n\n")

		cName, cTag, cTime, cPrerelease, err := getCurrentReleaseInfo(client)
		printReleaseInfo(cName, cTag, cTime, cPrerelease, err)
		if err != nil {
			return err
		}

		lName, lTag, lTime, lPrerelease := getLastReleaseInfo(client)

		if currentVersion == lTag {
			color.Green("\n\nYou have the latest version of svfs.")
		} else if lTag != "" {
			color.Red("\n\nA new svfs release exist.\nVisit https://github.com/ovh/svfs/ for more information.\n\n")
			printReleaseInfo(lName, lTag, lTime, lPrerelease, err)
			if err != nil {
				return err
			}
		} else {
			color.Yellow("\n\nCan't join GitHub API.\n")
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(infoCmd)
}
