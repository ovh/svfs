package main

import "github.com/Sirupsen/logrus"
import "github.com/ovh/svfs/cmd"

// Just init cobra
func main() {
	// Execute adds all child commands to the root command sets flags appropriately.
	if err := cmd.RootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
