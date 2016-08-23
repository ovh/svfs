package main

import "github.com/Sirupsen/logrus"
import "github.com/ovh/svfs/cmd"

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
