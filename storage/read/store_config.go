package read

import (
	"path/filepath"
	"runtime"
	"time"

	"github.com/influxdata/influxdb/pkg/limiter"
	"github.com/influxdata/influxdb/toml"
	"github.com/influxdata/influxdb/tsdb"
)

// StoreConfig contains the configuration options pertaining to the tsdb.Store,
// tsm1.Engine, and the ObjectStore.
type StoreConfig struct {
	EngineOptions tsdb.EngineOptions
}

// NewStoreConfig initialises a new StoreConfig with some sane defaults.
func NewStoreConfig(dataDir string) StoreConfig {
	// TSDB specific options.
	tsdbConfig := tsdb.NewConfig()
	tsdbConfig.Dir = filepath.Join(dataDir, "data")
	tsdbConfig.WALDir = filepath.Join(dataDir, "wal")
	tsdbConfig.Index = "tsi1"
	tsdbConfig.CacheSnapshotWriteColdDuration = toml.Duration(10 * time.Second)
	tsdbConfig.CacheSnapshotMemorySize = 256 << 20
	tsdbConfig.MaxConcurrentCompactions = 8

	// TSDB Engine options
	engineOptions := tsdb.NewEngineOptions()
	engineOptions.Config = tsdbConfig
	engineOptions.EngineVersion = tsdbConfig.Engine
	engineOptions.IndexVersion = tsdbConfig.Index
	engineOptions.WALEnabled = false

	// Setup compactions options.
	lim := engineOptions.Config.MaxConcurrentCompactions
	if lim == 0 {
		lim = runtime.GOMAXPROCS(0) / 2 // Default to 50% of cores for compactions

		// On systems with more cores, cap at 4 to reduce disk utilization.
		if lim > 4 {
			lim = 4
		}

		if lim < 1 {
			lim = 1
		}
	}

	// Don't allow more compactions to run than cores.
	if lim > runtime.GOMAXPROCS(0) {
		lim = runtime.GOMAXPROCS(0)
	}

	engineOptions.CompactionLimiter = limiter.NewFixed(lim)

	// This option determines how much IO compactions can use. The value was taken
	// from InfluxDB (tsdb/store.go). It was probably determined empirically by
	// Jason Wilder.
	engineOptions.CompactionThroughputLimiter = limiter.NewRate(48*1024*1024, 48*1024*1024)

	return StoreConfig{
		EngineOptions: engineOptions,
	}
}
