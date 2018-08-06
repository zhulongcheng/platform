package prometheus

import (
	"context"
	"time"

	"github.com/influxdata/platform/query"
	"github.com/influxdata/platform/query/execute"
	"github.com/influxdata/platform/query/values"
	opentracing "github.com/opentracing/opentracing-go"
)

type Reader interface {
	Read(ctx context.Context, trace map[string]string, host, matcher string, start, end time.Time) (query.TableIterator, error)
}

type source struct {
	id        execute.DatasetID
	host      string
	matcher   string
	startTime time.Time
	endTime   time.Time

	transformations []execute.Transformation
}

func NewSource(id execute.DatasetID, host string, matcher string, startTime, endTime time.Time) execute.Source {
	return &source{
		id:        id,
		host:      host,
		matcher:   matcher,
		startTime: startTime,
		endTime:   endTime,
	}
}

func (s *source) AddTransformation(t execute.Transformation) {
	s.transformations = append(s.transformations, t)
}

func (s *source) Run(ctx context.Context) {
	err := s.run(ctx)
	for _, t := range s.transformations {
		t.Finish(s.id, err)
	}
}

func (s *source) run(ctx context.Context) error {
	var trace map[string]string
	if span := opentracing.SpanFromContext(ctx); span != nil {
		trace = make(map[string]string)
		span = opentracing.StartSpan("prom_source.run", opentracing.ChildOf(span.Context()))
		_ = opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, opentracing.TextMapCarrier(trace))
	}

	r := &reader{}

	ti, err := r.Read(ctx, trace, s.host, s.matcher, s.startTime, s.endTime)
	if err != nil {
		return err
	}

	err = ti.Do(func(tbl query.Table) error {
		for _, t := range s.transformations {
			if err := t.Process(s.id, tbl); err != nil {
				return err
			}
		}

		return nil
	})

	for _, t := range s.transformations {
		t.UpdateWatermark(s.id, values.Time(s.endTime.UnixNano()))
	}

	return err
}
