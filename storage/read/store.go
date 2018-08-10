package read

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/influxdata/influxdb/logger"
	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/influxdb/tsdb"
	"github.com/influxdata/influxdb/tsdb/engine/tsm1"
	"github.com/influxdata/influxdb/tsdb/index/tsi1"
	"github.com/influxdata/influxql"
	"go.uber.org/zap"
)

var (
	// ErrShardNotFound is returned when trying to get a non existing shard.
	ErrShardNotFound = fmt.Errorf("shard not found")
	// ErrStoreClosed is returned when trying to use a closed Store.
	ErrStoreClosed = fmt.Errorf("store is closed")
)

// Store defines the set of methods needed to interact with a local TSDB store.
type Store interface {
	WithLogger(logger *zap.Logger)
	Open() error
	Close() error

	CreateShard(id uint64, doRecover bool) error
	DeleteShard(id uint64) error
	Shard(id uint64) Shard
	ShardIDs() []uint64
	ShardPath(id uint64) string
}

type Shard interface {
	// ID returns the shard's ID.
	ID() uint64

	// Path returns the location of the shard on disk.
	Path() string

	// CreateSeriesCursor creates a cursor to itertate over series IDs.
	CreateSeriesCursor(context.Context, tsdb.SeriesCursorRequest, influxql.Expr) (tsdb.SeriesCursor, error)

	// DeleteSeriesRangeWithPredicate deletes all series data iterated over if fn returns
	// true for that series.
	DeleteSeriesRangeWithPredicate(itr tsdb.SeriesIterator, fn func([]byte, models.Tags) (int64, int64, bool)) error

	// SeriesN returns the cardinality of the shard.
	SeriesN() int64

	// WritePoints writes the provided points to the shard.
	WritePoints(pt []models.Point) error

	Close() error

	Engine() (tsdb.Engine, error)
	Index() (tsdb.Index, error)
}

type localStore struct {
	path  string
	wPath string // TODO(jgm) remove WAL support

	mu sync.RWMutex

	config     StoreConfig
	logger     *zap.Logger
	baseLogger *zap.Logger

	shards             map[uint64]*tsdb.Shard
	sfileByShard       map[uint64]*tsdb.SeriesFile
	generationsByShard map[uint64]int64

	closing chan struct{}
	opened  bool
}

// NewStore initialises a new local store, as well as some configuration options
// relating to compactions.
func NewStore(config StoreConfig) *localStore {
	return &localStore{
		path:       config.EngineOptions.Config.Dir,
		wPath:      config.EngineOptions.Config.WALDir,
		config:     config,
		logger:     zap.NewNop(),
		baseLogger: zap.NewNop(),
	}
}

// WithLogger sets the logger for the service. It should be called before Open.
func (s *localStore) WithLogger(log *zap.Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.opened {
		s.logger.Warn("cannot set logger when Store is open")
		return
	}
	s.logger = log.With(zap.String("service", "store"))
	s.baseLogger = log
}

// Open opens the local store. The local store must be opened before any shards
// can be created.
func (s *localStore) Open() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.opened {
		return nil
	}

	if err := os.MkdirAll(s.path, 0777); err != nil {
		return err
	}

	s.closing = make(chan struct{})
	s.shards = make(map[uint64]*tsdb.Shard)
	s.sfileByShard = make(map[uint64]*tsdb.SeriesFile)
	s.generationsByShard = make(map[uint64]int64)
	s.opened = true

	// TODO monitorShards()
	return nil
}

func (s *localStore) Close() error {
	for _, id := range s.ShardIDs() {
		if err := s.CloseShard(id); err != nil {
			return err
		}
	}
	return nil
}

// CreateShard creates a new shard and opens it.
func (s *localStore) CreateShard(id uint64, doRecover bool) error {
	// ensure we close anything we open, even in the case of errors.
	var cleanup []io.Closer
	var err error
	defer func() {
		if err == nil {
			return
		}
		for _, cl := range cleanup {
			cl.Close()
		}
	}()

	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.closing:
		return ErrStoreClosed
	default:
	}

	if _, found := s.shards[id]; found {
		// Shard already exists.
		return nil
	}

	log := s.baseLogger.With(logger.Shard(id))
	path := s.ShardPath(id)

	// Create or open the Series File for this shard.
	sfile := tsdb.NewSeriesFile(filepath.Join(path, tsdb.SeriesFileDirectory))
	sfile.Logger = log
	if err := sfile.Open(); err != nil {
		return err
	}
	cleanup = append(cleanup, sfile)

	opt := s.config.EngineOptions
	// Provides a function for the tsdb.Store and engine to access shards' series id sets.
	opt.SeriesIDSets = s

	// Create the tsdb.Shard.
	shard := tsdb.NewShard(id, path, s.walPath(id), sfile, opt)
	shard.WithLogger(log)
	if err := shard.Open(); err != nil {
		return err
	}
	cleanup = append(cleanup, shard)

	// Create file naming closure to inject kafka offset.
	var formatFileNameFunc tsm1.FormatFileNameFunc = func(_, sequence int) string {
		return fmt.Sprintf("%09d-%09d", s.ShardGeneration(id), sequence)
	}

	var engine *tsm1.Engine
	if e, err := shard.Engine(); err != nil {
		return err
	} else if tsm1Engine, ok := e.(*tsm1.Engine); !ok {
		return errors.New("cannot load tsm1 engine, invalid type")
	} else {
		engine = tsm1Engine
	}
	engine.WithFormatFileNameFunc(formatFileNameFunc)

	s.shards[id] = shard
	s.sfileByShard[id] = sfile

	return nil
}

func (s *localStore) indexTSMFile(index *tsi1.Index, tsmFilename string, log *zap.Logger) error {
	fi, err := os.Stat(tsmFilename)
	if err != nil {
		return err
	}
	log.Info("Indexing TSM file",
		zap.String("filename", tsmFilename),
		zap.Int64("size", fi.Size()))
	tsmFileIndexStartTime := time.Now()

	f, err := os.Open(tsmFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	r, err := tsm1.NewTSMReader(f)
	if err != nil {
		return err
	}
	defer r.Close()

	for i := 0; i < r.KeyCount(); i++ {
		key, _ := r.KeyAt(i)
		seriesKey, _ := tsm1.SeriesAndFieldFromCompositeKey(key)
		name, tags := models.ParseKey(seriesKey)

		if err := index.CreateSeriesIfNotExists(seriesKey, []byte(name), tags); err != nil {
			return fmt.Errorf("cannot create series: %s %s (%s)", name, tags.String(), err)
		}
	}

	log.Info("Indexed TSM file",
		zap.String("filename", tsmFilename),
		zap.Duration("duration", time.Since(tsmFileIndexStartTime)))

	return r.Close() // Also closes f.
}

// ShardPath returns the absolute path to a shard's directory on disk.
func (s *localStore) ShardPath(id uint64) string {
	return filepath.Join(s.path, fmt.Sprint(id))
}

// walPath returns the absolute path to the WAL directory.
func (s *localStore) walPath(id uint64) string {
	return filepath.Join(s.wPath, fmt.Sprint(id))
}

func (s *localStore) CloseShard(id uint64) error {
	s.mu.Lock()

	shard := s.shards[id]
	sfile := s.sfileByShard[id]

	delete(s.shards, id)
	delete(s.sfileByShard, id)

	s.mu.Unlock()

	var err error
	if shard != nil {
		if serr := shard.Close(); err == nil {
			err = serr
		}
	}
	if sfile != nil {
		if serr := sfile.Close(); err == nil {
			err = serr
		}
	}

	return err
}

func (s *localStore) DeleteShard(id uint64) error {
	if err := s.CloseShard(id); err != nil {
		return err
	}

	if err := os.RemoveAll(s.ShardPath(id)); err != nil {
		return err
	}

	return os.RemoveAll(s.walPath(id))
}

func (s *localStore) Shard(id uint64) Shard {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.shards[id]
}

func (s *localStore) ShardIDs() []uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]uint64, 0, len(s.shards))
	for id := range s.shards {
		ids = append(ids, id)
	}

	return ids
}

func (s *localStore) Shards() []Shard {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shards := make([]Shard, 0, len(s.shards))
	for _, shard := range s.shards {
		shards = append(shards, shard)
	}

	return shards
}

// ForEach provides access to each series ID set (a bitmap storing all series ids
// for series in each shard).
//
// ForEach implements the tsdb.SeriesIDSets interface and is used when deleting
// series.
func (s *localStore) ForEach(f func(ids *tsdb.SeriesIDSet)) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, shard := range s.shards {
		idx, err := shard.Index()
		if err != nil {
			return err
		}
		f(idx.SeriesIDSet())
	}
	return nil
}

// ShardGeneration returns the last set TSM generation number for a given shard.
func (s *localStore) ShardGeneration(id uint64) int64 {
	s.mu.RLock()
	generation := s.generationsByShard[id]
	s.mu.RUnlock()
	return generation
}

// SetShardGeneration sets the TSM generation number for a given shard.
func (s *localStore) SetShardGeneration(id uint64, generation int64) {
	s.mu.Lock()
	s.generationsByShard[id] = generation
	s.mu.Unlock()
}
