package config

import (
	"fmt"
	v "github.com/spf13/viper"
)

// LoadConfig reads in config file and ENV variables if set.
func LoadConfig() error {
	v.SetConfigName("svfs")
	v.SetConfigType("yaml")
	v.AddConfigPath("/etc/")
	v.AddConfigPath("$HOME/")

	v.BindEnv("os_auth_url")
	v.BindEnv("os_container_name")
	v.BindEnv("os_tenant_name")
	v.BindEnv("os_username")
	v.BindEnv("os_password")
	v.BindEnv("os_region_name")

	// Read config file
	err := v.ReadInConfig()

	if err == nil {
		fmt.Println("Using config file:", v.ConfigFileUsed())
	}

	return err
}
