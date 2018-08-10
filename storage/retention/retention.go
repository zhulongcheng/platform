package retention

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/influxdb/logger"
	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/influxdb/tsdb"
	"github.com/influxdata/influxql"
	"github.com/influxdata/platform"
	"github.com/influxdata/platform/storage/read"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	serviceName      = "retention"
	bucketAPITimeout = 10 * time.Second
	shardAPITimeout  = time.Minute
)

// ErrServiceClosed is returned when the service is unavailable.
var ErrServiceClosed = errors.New("service is currently closed")

// The Service periodically removes data that is outside of the retention
// period of the bucket associated with the data.
type Service struct {
	// Store provides access to data stored locally on the storage node.
	Store read.Store

	// BucketService provides an API for retrieving buckets associated with
	// organisations.
	BucketService BucketService

	logger   *zap.Logger
	interval time.Duration // Interval that retention service deletes data.
	nodeID   int           // Used for instrumentation.

	retentionMetrics *retentionMetrics

	mu       sync.RWMutex
	_closing chan struct{}

	wg sync.WaitGroup
}

// NewService returns a new Service that performs deletes
// every interval period. Setting interval to 0 is equivalent to disabling the
// service.
func NewService(store read.Store, bucketService BucketService, interval int64, nodeID int) *Service {
	s := &Service{
		Store:         store,
		BucketService: bucketService,
		logger:        zap.NewNop(),
		interval:      time.Duration(interval) * time.Second,
		nodeID:        nodeID,

		retentionMetrics: newRetentionMetrics(),
	}

	return s
}

// WithLogger sets the logger l on the service. It must be called before Open.
func (s *Service) WithLogger(l *zap.Logger) {
	s.logger = l.With(zap.String("service", serviceName))
	s.BucketService.WithLogger(s.logger)
}

// Open opens the service, which begins the process of removing expired data.
// Re-opening the service once it's open is a no-op.
func (s *Service) Open() error {
	if s.closing() != nil {
		return nil // Already open.
	}

	s.logger.Info("Service opening", zap.Duration("check_interval", s.interval))
	if s.interval < 0 {
		return fmt.Errorf("invalid interval %v", s.interval)
	}

	// Open BucketService implementation.
	if err := s.BucketService.Open(); err != nil {
		return err
	}

	s.mu.Lock()
	s._closing = make(chan struct{})
	s.mu.Unlock()

	s.wg.Add(1)
	go func() { defer s.wg.Done(); s.run() }()
	s.logger.Info("Service finished opening")

	return nil
}

// run periodically iterates over all owned shards, and expires (deletes) all
// data that's fallen outside of the retention period for the associated bucket.
func (s *Service) run() {
	if s.interval == 0 {
		s.logger.Info("Service disabled")
		return
	}

	deleteShardData := func(shardIDs []uint64) {
		log, logEnd := logger.NewOperation(s.logger, "Data retention check", "data_retention_check")
		defer logEnd()

		rpByBucketID, err := s.getRetentionPeriodPerBucket()
		if err != nil {
			log.Error("Unable to determine bucket:RP mapping", zap.Error(err))
			return
		}

		for _, shardID := range shardIDs {
			now := time.Now().UTC()
			status := "ok"

			shard := s.Store.Shard(shardID)
			if shard == nil { // We don't have this shard.
				log.Info("Shard not local")
				return
			}

			if err := s.expireData(shard, rpByBucketID, now); err != nil {
				log.Error("Deletion not successful", zap.Error(err))
				status = "error"
			}
			s.retentionMetrics.CheckDuration.With(prometheus.Labels{"node_id": fmt.Sprint(s.nodeID), "partition": fmt.Sprint(shardID), "status": status}).Observe(time.Since(now).Seconds())
			s.retentionMetrics.Checks.With(prometheus.Labels{"node_id": fmt.Sprint(s.nodeID), "partition": fmt.Sprint(shardID), "status": status}).Inc()
		}
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	closingCh := s.closing()
	for {
		// Refresh shard IDs.
		select {
		case <-ticker.C:
			deleteShardData(s.Store.ShardIDs())
		case <-closingCh:
			return
		}
	}
}

// expireData runs a delete operation against the provided shard.
//
// Any series data that (1) belongs to a bucket in the provided map and (2) falls outside the bucket's
// indicated retention period will be deleted.
func (s *Service) expireData(sh read.Shard, rpByBucketID map[string]time.Duration, now time.Time) error {
	_, logEnd := logger.NewOperation(s.logger, "Data deletion", "data_deletion",
		logger.Shard(sh.ID()), zap.String("path", sh.Path()))
	defer logEnd()

	ctx, cancel := context.WithTimeout(context.Background(), shardAPITimeout)
	defer cancel()
	cur, err := sh.CreateSeriesCursor(ctx, tsdb.SeriesCursorRequest{}, nil)
	if err != nil {
		return err
	}
	defer cur.Close()

	var mu sync.Mutex
	badMSketch := make(map[string]struct{})     // Badly formatted measurements.
	missingBSketch := make(map[string]struct{}) // Missing buckets.

	var seriesDeleted uint64 // Number of series where a delete is attempted.
	var seriesSkipped uint64 // Number of series that were skipped from delete.

	fn := func(name []byte, tags models.Tags) (int64, int64, bool) {
		_, bucketID, err := platform.ReadMeasurement(name)
		if err != nil {
			mu.Lock()
			badMSketch[string(bucketID)] = struct{}{}
			mu.Unlock()
			atomic.AddUint64(&seriesSkipped, 1)
			return 0, 0, false
		}

		retentionPeriod, ok := rpByBucketID[string(bucketID)]
		if !ok {
			mu.Lock()
			missingBSketch[string(bucketID)] = struct{}{}
			mu.Unlock()
			atomic.AddUint64(&seriesSkipped, 1)
			return 0, 0, false
		}

		atomic.AddUint64(&seriesDeleted, 1)
		to := now.Add(-retentionPeriod).UnixNano()
		return math.MinInt64, to, true
	}

	defer func() {
		// Track metrics on deletion.
		s.retentionMetrics.Unprocessable.With(prometheus.Labels{"node_id": fmt.Sprint(s.nodeID), "partition": fmt.Sprint(sh.ID()), "status": "bad_measurement"}).Add(float64(len(badMSketch)))
		s.retentionMetrics.Unprocessable.With(prometheus.Labels{"node_id": fmt.Sprint(s.nodeID), "partition": fmt.Sprint(sh.ID()), "status": "missing_bucket"}).Add(float64(len(missingBSketch)))

		s.retentionMetrics.Series.With(prometheus.Labels{"node_id": fmt.Sprint(s.nodeID), "partition": fmt.Sprint(sh.ID()), "status": "ok"}).Add(float64(atomic.LoadUint64(&seriesDeleted)))
		s.retentionMetrics.Series.With(prometheus.Labels{"node_id": fmt.Sprint(s.nodeID), "partition": fmt.Sprint(sh.ID()), "status": "skipped"}).Add(float64(atomic.LoadUint64(&seriesSkipped)))
	}()

	return sh.DeleteSeriesRangeWithPredicate(newSeriesIteratorAdapter(cur), fn)
}

// getRetentionPeriodPerBucket returns a map of (bucket ID -> retention period)
// for all buckets.
func (s *Service) getRetentionPeriodPerBucket() (map[string]time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), bucketAPITimeout)
	defer cancel()
	buckets, _, err := s.BucketService.FindBuckets(ctx, platform.BucketFilter{})
	if err != nil {
		return nil, err
	}
	rpByBucketID := make(map[string]time.Duration, len(buckets))
	for _, bucket := range buckets {
		rpByBucketID[string(bucket.ID)] = bucket.RetentionPeriod
	}
	return rpByBucketID, nil
}

// closing returns a channel to signal that the service is closing.
func (s *Service) closing() chan struct{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s._closing
}

// Close closes the service.
//
// If a delete of data is in-progress, then it will be allowed to complete before
// Close returns. Re-closing the service once it's closed is a no-op.
func (s *Service) Close() error {
	if s.closing() == nil {
		return nil // Already closed.
	}

	now := time.Now()
	s.logger.Info("Service closing:")
	close(s.closing())
	s.wg.Wait()

	if err := s.BucketService.Close(); err != nil {
		return err
	}
	s.logger.Info("Service closed:", zap.Duration("took", time.Since(now)))

	s.mu.Lock()
	s._closing = nil
	s.mu.Unlock()
	return nil
}

// PrometheusCollectors satisfies the prom.PrometheusCollector interface.
func (s *Service) PrometheusCollectors() []prometheus.Collector {
	return s.retentionMetrics.PrometheusCollectors()
}

// A BucketService is an platform.BucketService that the RetentionService can open,
// close and log.
type BucketService interface {
	platform.BucketService
	Open() error
	Close() error
	WithLogger(l *zap.Logger)
}

type seriesIteratorAdapter struct {
	itr  tsdb.SeriesCursor
	ea   seriesElemAdapter
	elem tsdb.SeriesElem
}

func newSeriesIteratorAdapter(itr tsdb.SeriesCursor) *seriesIteratorAdapter {
	si := &seriesIteratorAdapter{itr: itr}
	si.elem = &si.ea
	return si
}

// Next returns the next tsdb.SeriesElem.
//
// The returned tsdb.SeriesElem is valid for use until Next is called again.
func (s *seriesIteratorAdapter) Next() (tsdb.SeriesElem, error) {
	if s.itr == nil {
		return nil, nil
	}

	row, err := s.itr.Next()
	if err != nil {
		return nil, err
	}

	if row == nil {
		return nil, nil
	}

	s.ea.name = row.Name
	s.ea.tags = row.Tags
	return s.elem, nil
}

func (s *seriesIteratorAdapter) Close() error {
	if s.itr != nil {
		err := s.itr.Close()
		s.itr = nil
		return err
	}
	return nil
}

type seriesElemAdapter struct {
	name []byte
	tags models.Tags
}

func (e *seriesElemAdapter) Name() []byte        { return e.name }
func (e *seriesElemAdapter) Tags() models.Tags   { return e.tags }
func (e *seriesElemAdapter) Deleted() bool       { return false }
func (e *seriesElemAdapter) Expr() influxql.Expr { return nil }
