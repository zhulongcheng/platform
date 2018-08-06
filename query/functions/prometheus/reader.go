package prometheus

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/influxdata/platform/query"
	"github.com/influxdata/platform/query/execute"
	"github.com/influxdata/platform/query/values"
	promclient "github.com/ryotarai/prometheus-query/client"
)

type reader struct {
}

const defaultStep = 10 * time.Second

func (r *reader) Read(ctx context.Context, trace map[string]string, host, matcher string, start, stop time.Time) (query.TableIterator, error) {
	prom, err := promclient.NewClient(host)
	if err != nil {
		return nil, err
	}

	fmt.Println("requesting to prometheus ", matcher, start, stop)

	promRes, err := prom.QueryRange(matcher, start, stop, defaultStep)
	if err != nil {
		return nil, err
	}

	// convert the results to one table per series returned from Prometheus
	pti := &promTableIterator{}

	for _, series := range promRes.Data.Result {
		keyCols := make([]query.ColMeta, 0, len(series.Metric)+2)
		keyValues := make([]values.Value, 0, len(series.Metric)+2)

		names := make([]string, 0, len(series.Metric))
		for n := range series.Metric {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, name := range names {
			value := series.Metric[name]
			keyCols = append(keyCols, query.ColMeta{Label: name, Type: query.TString})
			keyValues = append(keyValues, values.NewStringValue(value))
		}
		keyCols = append(keyCols, query.ColMeta{Label: "_start", Type: query.TTime})
		keyCols = append(keyCols, query.ColMeta{Label: "_stop", Type: query.TTime})
		keyValues = append(keyValues, values.NewTimeValue(values.ConvertTime(start)))
		keyValues = append(keyValues, values.NewTimeValue(values.ConvertTime(stop)))

		key := execute.NewGroupKey(keyCols, keyValues)
		builder := execute.NewColListTableBuilder(key, &execute.Allocator{Limit: math.MaxInt64})

		for _, c := range keyCols {
			builder.AddCol(c)
		}
		valueIdx := len(keyCols)
		timeIdx := valueIdx + 1

		builder.AddCol(query.ColMeta{Label: "_value", Type: query.TFloat})
		builder.AddCol(query.ColMeta{Label: "_time", Type: query.TTime})

		// builder.AddCol(query.ColMeta{Label: "_start", Type: query.TTime})
		// builder.AddCol(query.ColMeta{Label: "_stop", Type: query.TTime})
		// startIdx := timeIdx + 1
		// stopIdx := startIdx + 1

		for _, v := range series.Values {
			val, err := v.Value()
			if err != nil {
				continue
			}
			l := len(keyValues) - 2
			for i, v := range keyValues[:l] {
				builder.AppendString(i, v.Str())
			}

			// maybe this
			builder.AppendTime(l, values.ConvertTime(start))
			builder.AppendTime(l+1, values.ConvertTime(stop))

			builder.AppendFloat(valueIdx, val)
			builder.AppendTime(timeIdx, values.Time(v.Time().UnixNano()))
			// builder.AppendTime(startIdx, values.ConvertTime(start))
			// builder.AppendTime(stopIdx, values.Time(stop.UnixNano()))
		}

		pti.tables = append(pti.tables, builder.RawTable())
	}

	return pti, nil
}

// promTableIterator is the collection of all results from Prometheus that
// implements the query.TableIterator interface since
// all results from a Prometheus source are brought into memory.
type promTableIterator struct {
	tables []*execute.ColListTable
}

func (p *promTableIterator) Do(f func(query.Table) error) error {
	for _, t := range p.tables {
		if err := f(t); err != nil {
			return err
		}
	}

	return nil
}
