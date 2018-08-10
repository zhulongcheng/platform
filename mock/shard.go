package mock

import (
	"context"

	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/influxdb/tsdb"
	"github.com/influxdata/influxdb/tsdb/engine/tsm1"
	"github.com/influxdata/influxql"
)

// Shard is a mockable implementation of a storage.
type Shard struct {
	IDFn                             func() uint64
	PathFn                           func() string
	CreateSeriesCursorFn             func(context.Context, tsdb.SeriesCursorRequest, influxql.Expr) (tsdb.SeriesCursor, error)
	DeleteSeriesRangeWithPredicateFn func(tsdb.SeriesIterator, func([]byte, models.Tags) (int64, int64, bool)) error
	SeriesNFn                        func() int64
	WritePointsFn                    func([]models.Point) error
	CloseFn                          func() error
	EngineFn                         func() (tsdb.Engine, error)
	IndexFn                          func() (tsdb.Index, error)

	// The SeriesCursor is returned by CreateSeriesCursor.
	SeriesCursor *SeriesCursor

	TSM1Engine *tsm1.Engine
}

// NewShard returns a mock Shard where each method will by default return the
// zero values.
func NewShard() *Shard {
	s := &Shard{
		IDFn:   func() uint64 { return 0 },
		PathFn: func() string { return "" },
		DeleteSeriesRangeWithPredicateFn: func(tsdb.SeriesIterator, func([]byte, models.Tags) (int64, int64, bool)) error {
			return nil
		},
		SeriesNFn:     func() int64 { return 0 },
		WritePointsFn: func([]models.Point) error { return nil },
		CloseFn:       func() error { return nil },
		IndexFn:       func() (tsdb.Index, error) { return nil, nil },
	}

	s.CreateSeriesCursorFn = func(context.Context, tsdb.SeriesCursorRequest, influxql.Expr) (tsdb.SeriesCursor, error) {
		return s.SeriesCursor, nil
	}

	s.SeriesCursor = &SeriesCursor{
		CloseFn: func() error { return nil },
		NextFn:  func() (*tsdb.SeriesCursorRow, error) { return nil, nil },
	}

	s.TSM1Engine = tsm1.NewEngine(uint64(0), nil, "", "", nil, tsdb.EngineOptions{}).(*tsm1.Engine)
	s.EngineFn = func() (tsdb.Engine, error) { return s.TSM1Engine, nil }
	return s
}

// ID returns the shard's ID.
func (s *Shard) ID() uint64 { return s.IDFn() }

// Path returns the location of the shard on disk.
func (s *Shard) Path() string { return s.PathFn() }

// CreateSeriesCursor creates a cursor to itertate over series IDs.
func (s *Shard) CreateSeriesCursor(ctx context.Context, cr tsdb.SeriesCursorRequest, expr influxql.Expr) (tsdb.SeriesCursor, error) {
	return s.CreateSeriesCursorFn(ctx, cr, expr)
}

// DeleteSeriesRangeWithPredicate deletes all series data iterated over if fn returns
// true for that series.
func (s *Shard) DeleteSeriesRangeWithPredicate(itr tsdb.SeriesIterator, fn func([]byte, models.Tags) (int64, int64, bool)) error {
	return s.DeleteSeriesRangeWithPredicateFn(itr, fn)
}

// SeriesN returns the cardinality of the shard.
func (s *Shard) SeriesN() int64 { return s.SeriesNFn() }

// WritePoints writes the provided points to the shard.
func (s *Shard) WritePoints(pt []models.Point) error { return s.WritePointsFn(pt) }

func (s *Shard) Close() error { return s.CloseFn() }

func (s *Shard) Engine() (tsdb.Engine, error) { return s.EngineFn() }

func (s *Shard) Index() (tsdb.Index, error) { return s.IndexFn() }

// SeriesCursor is a mockable implementation of a tsdb.SeriesCursor.
type SeriesCursor struct {
	CloseFn func() error
	NextFn  func() (*tsdb.SeriesCursorRow, error)
}

// Close closes the cursor.
func (s *SeriesCursor) Close() error { return s.CloseFn() }

// Next returns the next row in the cursor.
func (s *SeriesCursor) Next() (*tsdb.SeriesCursorRow, error) { return s.NextFn() }
