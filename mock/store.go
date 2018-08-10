package mock

import (
	"fmt"

	"github.com/influxdata/platform/storage/read"
	"go.uber.org/zap"
)

// Store is a mockable implementation of a storage.Store
type Store struct {
	WithLoggerFn  func(*zap.Logger)
	OpenFn        func() error
	CloseFn       func() error
	CreateShardFn func(id uint64, doRecover bool) error
	DeleteShardFn func(id uint64) error
	ShardFn       func(id uint64) read.Shard
	ShardIDsFn    func() []uint64
	ShardPathFn   func(id uint64) string
}

// NewStore returns a mock Store where all methods return the zero values.
func NewStore() *Store {
	return &Store{
		WithLoggerFn:  func(*zap.Logger) {},
		OpenFn:        func() error { return nil },
		CloseFn:       func() error { return nil },
		CreateShardFn: func(id uint64, doRecover bool) error { return nil },
		DeleteShardFn: func(id uint64) error { return nil },
		ShardFn:       func(id uint64) read.Shard { return nil },
		ShardIDsFn:    func() []uint64 { return nil },
		ShardPathFn:   func(id uint64) string { return fmt.Sprintf("/path/to/%d", id) },
	}
}

// WithLogger sets the logger on the Store.
func (s *Store) WithLogger(l *zap.Logger) { s.WithLoggerFn(l) }

// Open opens the Store.
func (s *Store) Open() error { return s.OpenFn() }

// Close closes the Store.
func (s *Store) Close() error { return s.CloseFn() }

// CreateShard creates a Shard.
func (s *Store) CreateShard(id uint64, doRecover bool) error {
	return s.CreateShardFn(id, doRecover)
}

// DeleteShard deletes the shard with the provided ID.
func (s *Store) DeleteShard(id uint64) error { return s.DeleteShardFn(id) }

// Shard returns a shard with id.
func (s *Store) Shard(id uint64) read.Shard { return s.ShardFn(id) }

// ShardIDs returns all shard IDs within the store.
func (s *Store) ShardIDs() []uint64 { return s.ShardIDsFn() }

// ShardPath returns the path of the shard.
func (s *Store) ShardPath(id uint64) string { return s.ShardPathFn(id) }
