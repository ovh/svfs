package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ovh/svfs/cmd"
	"github.com/ovh/svfs/config"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Warnf("Could not load configuration file: %s", err)
	}

	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
