package config

import v "github.com/spf13/viper"

// LoadConfig reads configuration from a configuration file or the environment.
func LoadConfig() error {
	v.SetConfigName("svfs")
	v.SetConfigType("yaml")
	v.AddConfigPath("/etc/")

	v.BindEnv("os_auth_url")
	v.BindEnv("os_auth_token")
	v.BindEnv("os_tenant_name")
	v.BindEnv("os_username")
	v.BindEnv("os_password")
	v.BindEnv("os_region_name")
	v.BindEnv("os_domain")
	v.BindEnv("os_storage_url")
	v.BindEnv("hubic_auth")
	v.BindEnv("hubic_token")

	return v.ReadInConfig()
}
