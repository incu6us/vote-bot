package config

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Type string

const (
	JSONConfigType Type = "json"
	YAMLConfigType      = "yaml"
)

type Config struct {
	*viper.Viper
}

func New(envPrefix, filePath string, configType Type) (*Config, error) {
	cfg := viper.New()
	cfg.SetEnvPrefix(envPrefix)
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	cfg.AutomaticEnv()
	cfg.SetConfigType(string(configType))

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	if err := cfg.ReadConfig(f); err != nil {
		return nil, errors.Wrap(err, "failed to read config")
	}

	return &Config{cfg}, nil
}
