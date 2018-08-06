package functions_test

import (
	"testing"
	"time"

	"github.com/influxdata/platform/query"
	"github.com/influxdata/platform/query/execute"
	"github.com/influxdata/platform/query/functions"
	"github.com/influxdata/platform/query/querytest"
)

func TestFromProm_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name:    "fromProm no args",
			Raw:     `fromProm()`,
			WantErr: true,
		},
		{
			Name:    "fromProm conflicting args",
			Raw:     `fromProm(host: "foo", hosts:["foo"])`,
			WantErr: true,
		},
		{
			Name:    "fromProm repeat arg",
			Raw:     `fromProm(host:"foo", host:"oops")`,
			WantErr: true,
		},
		{
			Name:    "fromProm invalid arg",
			Raw:     `fromProm(host:"telegraf", chicken:"what is this?")`,
			WantErr: true,
		},
		{
			Name: "fromProm with host",
			Raw:  `fromProm(host:"localhost:9090") |> range(start:-4h, stop:-2h) |> sum()`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "fromProm0",
						Spec: &functions.FromPromOpSpec{
							Hosts: []string{"localhost:9090"},
						},
					},
					{
						ID: "range1",
						Spec: &functions.RangeOpSpec{
							Start: query.Time{
								Relative:   -4 * time.Hour,
								IsRelative: true,
							},
							Stop: query.Time{
								Relative:   -2 * time.Hour,
								IsRelative: true,
							},
							TimeCol:  "_time",
							StartCol: "_start",
							StopCol:  "_stop",
						},
					},
					{
						ID: "sum2",
						Spec: &functions.SumOpSpec{
							AggregateConfig: execute.DefaultAggregateConfig,
						},
					},
				},
				Edges: []query.Edge{
					{Parent: "fromProm0", Child: "range1"},
					{Parent: "range1", Child: "sum2"},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}
