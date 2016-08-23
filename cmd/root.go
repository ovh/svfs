package cmd

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Short: "The Swift Virtual File System",
	Long: "SVFS is a Virtual File System over Openstack Swift built upon fuse.\n\n" +
		"It is compatible with hubiC, OVH Public Cloud Storage and\n" +
		"basically every endpoint using a standard Openstack Swift setup.\n\n" +
		"It brings a layer of abstraction over object storage,\n" +
		"making it as accessible and convenient as a file system,\n" +
		"without being intrusive on the way your data is stored.\n",
}
