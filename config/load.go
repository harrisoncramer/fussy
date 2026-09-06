package config

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

// EnvVar names the environment variable holding the path to the configuration file.
const EnvVar = "FUSSY_CONFIG"

// LoadFromEnv reads the configuration the environment points at, and returns Default when it
// points at nothing.
func LoadFromEnv() (Config, error) {
	return Load(os.Getenv(EnvVar))
}

// Default is the configuration a run with no config file uses, which holds to the three rules
// that say something about any Go package and leaves off the two that assume a repository's
// own shape.
func Default() Config {
	return Config{
		ContextTimeout: ContextTimeoutConfig{Skip: true},
		ForbidGetenv:   ForbidGetenvConfig{Skip: true},
	}
}

// Load reads a YAML file holding the same settings block golangci-lint passes the plugin, and
// returns Default when the path is empty.
func Load(path string) (Config, error) {
	if path == "" {
		return Default(), nil
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("reading %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing %s: %w", path, err)
	}

	return cfg, nil
}
