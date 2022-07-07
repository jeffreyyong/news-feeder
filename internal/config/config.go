package config

import (
	"os"

	yaml "gopkg.in/yaml.v3"

	"github.com/pkg/errors"
)

const (
	defaultConfigFilePath = "config.yaml"
)

// Config variables for the application
type Config struct {
	PostgresDSN      string            `yaml:"postgres_dsn"`
	PrivilegedTokens map[string]string `yaml:"privileged_tokens"`
	MigrationPath    string            `yaml:"migration_path"`
	Worker           struct {
		URLSources []string `yaml:"url_sources"`
		Interval   int      `yaml:"interval"`
	} `yaml:"worker"`
	Social struct {
		Twitter struct {
			ConsumerKey    string `yaml:"consumer_key"`
			ConsumerSecret string `yaml:"consumer_secret"`
			AccessToken    string `yaml:"access_token"`
			AccessSecret   string `yaml:"access_secret"`
		} `yaml:"twitter"`
	} `yaml:"social"`
}

// Load loads the configuration for the application.
func Load() (Config, error) {
	var config Config

	file, err := os.Open(defaultConfigFilePath)
	if err != nil {
		return Config{}, errors.Wrap(err, "can't open file config file")
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return Config{}, errors.Wrap(err, "failed to decode config")
	}

	return config, nil
}
