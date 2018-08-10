package read

import (
	"github.com/influxdata/influxdb/monitor/diagnostics"
)

// Config represents a configuration for a HTTP service.
type Config struct {
	LogEnabled bool `toml:"log-enabled"` // verbose logging
}

// NewConfig returns a new Config with default settings.
func NewConfig() Config {
	return Config{
		LogEnabled: true,
	}
}

// Diagnostics returns a diagnostics representation of a subset of the Config.
func (c Config) Diagnostics() (*diagnostics.Diagnostics, error) {
	return diagnostics.RowFromMap(map[string]interface{}{
		"log-enabled": c.LogEnabled,
	}), nil
}
