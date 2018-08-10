package storage

import (
	"os"
	"runtime"
	"sync"

	"github.com/influxdata/influxdb/logger"
	"github.com/influxdata/influxdb/pkg/limiter"
	"github.com/influxdata/influxdb/services/storage"
	"github.com/influxdata/influxdb/tsdb"
	"github.com/influxdata/platform/storage/read"
	"github.com/influxdata/platform/storage/retention"
	"github.com/influxdata/platform/storage/tsm"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Service is the main storage service.
type Service struct {
	config Config
	logger *zap.Logger

	// Store can interact directly with the InfluxDB storage APIs.
	Store read.Store

	// TODO(edd): We need some sort of consumer service, if the storage layer
	// is going to pull writes down.

	// FIXME(edd): Need to figure out where the read service will live.
	// ReadService provides a server for queryd to access storage APIs remotely.
	ReadService *storage.Service

	// The RetentionService periodically deletes data that is outside of its
	// bucket's retention period.
	RetentionService *retention.Service

	nodeID     int   // This node's unique node ID.
	allNodeIDs []int // The set of all node IDs in the cluster.

	mu       sync.RWMutex
	_closing chan struct{}

	pprof struct {
		cpu   *os.File
		heap  *os.File
		mutex *os.File
		block *os.File
	}

	wg sync.WaitGroup
}

// NewService initialises a new Service, configuring all sub-services, but not
// opening them.
func NewService(config Config) *Service {
	// Override the compaction planner.
	config.Store.EngineOptions.CompactionPlannerCreator = func(_ tsdb.Config) interface{} {
		return tsm.NewChunkedCompactionPlanner(nil)
	}
	config.Store.EngineOptions.OpenLimiter = limiter.NewFixed(runtime.GOMAXPROCS(0))
	store := read.NewStore(config.Store)

	s := &Service{
		config: config,
		logger: zap.NewNop(),
		Store:  store,
	}
	return s
}

// WithLogger set the logger on this service and all sub-services. It must be
// called before Open.
func (s *Service) WithLogger(l *zap.Logger) {
	s.logger = l
	s.Store.WithLogger(s.logger)

	if s.ReadService != nil {
		s.ReadService.WithLogger(s.logger)
	}

	if s.RetentionService != nil {
		s.RetentionService.WithLogger(s.logger)
	}
}

func (s *Service) PrometheusCollectors() []prometheus.Collector {
	var cs []prometheus.Collector
	cs = append(cs, s.RetentionService.PrometheusCollectors()...)
	return cs
}

// Open opens the service and all sub-services. Re-opening an opened service is
// a no-op.
func (s *Service) Open() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s._closing != nil {
		return nil // Already open or opening.
	}
	s._closing = make(chan struct{})

	_, logEnd := logger.NewOperation(s.logger, "Storage service opening", "service_opening")
	defer logEnd()

	if err := s.Store.Open(); err != nil {
		return err
	}

	if err := s.ReadService.Open(); err != nil {
		return err
	}

	return s.RetentionService.Open()
}

// Closing returns a channel that can be used to signal the service is closing.
func (s *Service) Closing() chan struct{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s._closing
}

// Close closes the service and all sub-services. Re-closing a closed service is
// a no-op.
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s._closing == nil {
		return nil // Already closed.
	}
	// Signal to other goroutines that the service is closing.
	close(s._closing)

	_, logEnd := logger.NewOperation(s.logger, "Service closing", "service_closing")
	defer logEnd()

	// Close things in the reverse order to how they were opened.
	if err := s.RetentionService.Close(); err != nil {
		return err
	}

	// Close the storage service.
	if err := s.ReadService.Close(); err != nil {
		return err
	}

	// Close the TSDB store
	if err := s.Store.Close(); err != nil {
		return err
	}

	s._closing = nil
	return nil
}
