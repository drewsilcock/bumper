package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	configName = "config"
	configType = "toml"
)

var configPath string

func Setup() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Msgf("Error getting home directory: %v", err)
	}

	configPath = path.Join(homeDir, ".config", "bumper")

	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			// No config file - that's fine.
		} else {
			log.Fatal().Msgf("Error reading config file: %v", err)
		}
	}
}

type Config struct {
	// In the future, we could add more config options here, like branch names, etc.
	GitlabAPIKey string
	BumpType     *BumpType
	Force        bool
	LogLevel     zerolog.Level
}

func NewConfig(args Args) *Config {
	conf := Config{}

	conf.GitlabAPIKey = viper.GetString("gitlab_api_key")
	conf.Force = args.Force

	conf.LogLevel = zerolog.InfoLevel
	if args.Verbose {
		conf.LogLevel = zerolog.DebugLevel
	}

	switch strings.ToLower(args.BumpType) {
	case "major":
		conf.BumpType = Ptr(BumpTypeMajor)
	case "minor":
		conf.BumpType = Ptr(BumpTypeMinor)
	case "patch":
		conf.BumpType = Ptr(BumpTypePatch)
	case "":
		conf.BumpType = nil
	default:
		log.Fatal().Msgf("Invalid bump type: %s", args.BumpType)
	}

	return &conf
}

func (c *Config) Write() error {
	viper.Set("gitlab_api_key", c.GitlabAPIKey)
	viper.ConfigFileUsed()

	// If it already exists, that's fine.
	_ = os.Mkdir(configPath, 0700)

	confFilePath := path.Join(configPath, fmt.Sprintf("%s.%s", configName, configType))
	confFile, err := os.OpenFile(confFilePath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	if err := confFile.Close(); err != nil {
		return fmt.Errorf("error closing config file: %w", err)
	}

	return viper.WriteConfig()
}
