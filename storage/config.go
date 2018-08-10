package storage // Config has configuration options for a Service and its sub-services.
import (
	"time"

	"github.com/influxdata/platform/storage/read"
)

// Config defines the configuration for the service.
type Config struct {
	DataDir string // TSM data directory.

	ProfileStartup       bool          // If true, the service will generate CPU and memory profiles during startup.
	BlockProfileRate     time.Duration // If non-zero and ProfileStartup, generates a block profile during startup.
	MutexProfileFraction int           // If non-zero and ProfileStartup, generates a mutex profile during startup.

	EtcdAddrs     []string // etcd cluster addresses.
	EtcdTimeout   int64    // etcd cluster dial timeout in seconds.
	EtcdNamespace string   // namespace prefix for this service's keys in etcd.

	Storage read.Config      // ReadService configuration
	Store   read.StoreConfig // Store configuration

	RetentionInterval int64 // Frequency that RetentionService runs. 0 disables.
}

// NewConfig initialises a new Config with sane defaults.
func NewConfig(dataDir string) Config {
	return Config{
		DataDir: dataDir,

		// etcd configuration.
		EtcdAddrs:     []string{"127.0.0.1:2379"},
		EtcdTimeout:   5,
		EtcdNamespace: "storage",

		Storage:           read.NewConfig(),
		Store:             read.NewStoreConfig(dataDir),
		RetentionInterval: 3600, // 1 hour.
	}
}
