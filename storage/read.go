package storage

import (
	"context"

	"github.com/influxdata/platform/query"
	"github.com/influxdata/platform/query/execute"
	"github.com/influxdata/platform/query/functions/storage"
)

// ReadService provides a mechanism for reading raw series as blocks from storage.
type ReadService interface {
	Read(ctx context.Context, trace map[string]string, rs storage.ReadSpec, start, stop execute.Time) (query.TableIterator, error)
	Close()
}
