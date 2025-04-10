package main

import (
	"strings"

	"github.com/spf13/viper"
)

// setupConfig initializes the viper configuration with defaults and environment variables
// It sets the config file name and paths to search for the config file.
// It also sets up the environment variable prefix and key replacer for environment variables.
func setupConfig() *viper.Viper {
	config := viper.New()
	config.SetEnvPrefix(appname)
	config.AutomaticEnv()
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", ""))
	config.SetDefault("debug", false)

	config.Set("app.name", appname)
	config.Set("app.version", version)
	config.Set("app.commit", commit)

	return config
}
