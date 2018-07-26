package functions_test

import (
	"regexp"
	"testing"

	"github.com/influxdata/platform/query/ast"
	"github.com/influxdata/platform/query/semantic"

	"github.com/influxdata/platform/query/plan"

	"github.com/influxdata/platform/query/execute/executetest"

	"github.com/influxdata/platform/query"
	"github.com/influxdata/platform/query/execute"
	"github.com/influxdata/platform/query/functions"
	"github.com/influxdata/platform/query/querytest"
)

func TestRenameDrop_NewQuery(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "test rename query",
			Raw:  `from(db:"mydb") |> rename(columns:{old:"new"}) |> sum()`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "rename1",
						Spec: &functions.RenameOpSpec{
							RenameCols: map[string]string{
								"old": "new",
							},
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
					{Parent: "from0", Child: "rename1"},
					{Parent: "rename1", Child: "sum2"},
				},
			},
		},
		{
			Name: "test drop query",
			Raw:  `from(db:"mydb") |> drop(columns:["col1", "col2", "col3"]) |> sum()`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "drop1",
						Spec: &functions.DropOpSpec{
							DropCols: []string{"col1", "col2", "col3"},
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
					{Parent: "from0", Child: "drop1"},
					{Parent: "drop1", Child: "sum2"},
				},
			},
		},
		{
			Name: "test drop query fn param",
			Raw:  `from(db:"mydb") |> drop(fn: (col) => col =~ /reg*/) |> sum()`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "drop1",
						Spec: &functions.DropOpSpec{
							DropPredicate: &semantic.FunctionExpression{
								Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "col"}}},
								Body: &semantic.BinaryExpression{
									Operator: ast.RegexpMatchOperator,
									Left: &semantic.IdentifierExpression{
										Name: "col",
									},
									Right: &semantic.RegexpLiteral{
										Value: regexp.MustCompile(`reg*`),
									},
								},
							},
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
					{Parent: "from0", Child: "drop1"},
					{Parent: "drop1", Child: "sum2"},
				},
			},
		},
		{
			Name: "test rename query fn param",
			Raw:  `from(db:"mydb") |> rename(fn: (col) => "new_name") |> sum()`,
			Want: &query.Spec{
				Operations: []*query.Operation{
					{
						ID: "from0",
						Spec: &functions.FromOpSpec{
							Database: "mydb",
						},
					},
					{
						ID: "rename1",
						Spec: &functions.RenameOpSpec{
							RenameFn: &semantic.FunctionExpression{
								Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "col"}}},
								Body: &semantic.StringLiteral{
									Value: "new_name",
								},
							},
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
					{Parent: "from0", Child: "rename1"},
					{Parent: "rename1", Child: "sum2"},
				},
			},
		},
		{
			Name:    "test rename query invalid",
			Raw:     `from(db:"mydb") |> rename(fn: (col) => "new_name", columns: {a:"b", c:"d"}) |> sum()`,
			Want:    nil,
			WantErr: true,
		},
		{
			Name:    "test drop query invalid",
			Raw:     `from(db:"mydb") |> drop(fn: (col) => col == target, columns: ["a", "b"]) |> sum()`,
			Want:    nil,
			WantErr: true,
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

func TestRenameDrop_Process(t *testing.T) {
	testCases := []struct {
		name    string
		spec    plan.ProcedureSpec
		data    []query.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "rename multiple cols",
			spec: &functions.RenameDropProcedureSpec{
				RenameCols: map[string]string{
					"1a": "1b",
					"2a": "2b",
					"3a": "3b",
				},
			},
			data: []query.Table{&executetest.Table{
				ColMeta: []query.ColMeta{
					{Label: "1a", Type: query.TFloat},
					{Label: "2a", Type: query.TFloat},
					{Label: "3a", Type: query.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []query.ColMeta{
					{Label: "1b", Type: query.TFloat},
					{Label: "2b", Type: query.TFloat},
					{Label: "3b", Type: query.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
		},
		{
			name: "drop multiple cols",
			spec: &functions.RenameDropProcedureSpec{
				DropCols: map[string]bool{
					"a": true,
					"b": true,
				},
			},
			data: []query.Table{&executetest.Table{
				ColMeta: []query.ColMeta{
					{Label: "a", Type: query.TFloat},
					{Label: "b", Type: query.TFloat},
					{Label: "c", Type: query.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []query.ColMeta{
					{Label: "c", Type: query.TFloat},
				},
				Data: [][]interface{}{
					{3.0},
					{13.0},
					{23.0},
				},
			}},
		},
		{
			name: "rename map fn (col) => name",
			spec: &functions.RenameDropProcedureSpec{
				RenameFn: &semantic.FunctionExpression{
					Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "col"}}},
					Body: &semantic.StringLiteral{
						Value: "new_name",
					},
				},
			},
			data: []query.Table{&executetest.Table{
				ColMeta: []query.ColMeta{
					{Label: "1a", Type: query.TFloat},
					{Label: "2a", Type: query.TFloat},
					{Label: "3a", Type: query.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []query.ColMeta{
					{Label: "new_name", Type: query.TFloat},
					{Label: "new_name", Type: query.TFloat},
					{Label: "new_name", Type: query.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
		},
		{
			name: "drop predicate (col) => col ~= /reg/",
			spec: &functions.RenameDropProcedureSpec{
				DropPredicate: &semantic.FunctionExpression{
					Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "col"}}},
					Body: &semantic.BinaryExpression{
						Operator: ast.RegexpMatchOperator,
						Left: &semantic.IdentifierExpression{
							Name: "col",
						},
						Right: &semantic.RegexpLiteral{
							Value: regexp.MustCompile(`server*`),
						},
					},
				},
			},
			data: []query.Table{&executetest.Table{
				ColMeta: []query.ColMeta{
					{Label: "server1", Type: query.TFloat},
					{Label: "local", Type: query.TFloat},
					{Label: "server2", Type: query.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []query.ColMeta{
					{Label: "local", Type: query.TFloat},
				},
				Data: [][]interface{}{
					{2.0},
					{12.0},
					{22.0},
				},
			}},
		},
		{
			name: "drop and rename",
			spec: &functions.RenameDropProcedureSpec{
				DropCols: map[string]bool{
					"server1": true,
					"server2": true,
				},
				RenameCols: map[string]string{
					"local": "localhost",
				},
			},
			data: []query.Table{&executetest.Table{
				ColMeta: []query.ColMeta{
					{Label: "server1", Type: query.TFloat},
					{Label: "local", Type: query.TFloat},
					{Label: "server2", Type: query.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []query.ColMeta{
					{Label: "localhost", Type: query.TFloat},
				},
				Data: [][]interface{}{
					{2.0},
					{12.0},
					{22.0},
				},
			}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				tc.wantErr,
				func(d execute.Dataset, c execute.TableBuilderCache) execute.Transformation {
					tr, err := functions.NewRenameDropTransformation(d, c, tc.spec)
					if err != nil {
						t.Fatal(err)
					}
					return tr
				},
			)
		})
	}
}

/*

	{
		name: "drop predicate wrong signature",
		spec: &functions.RenameDropProcedureSpec{
			DropPredicate: &semantic.FunctionExpression{
				Params: []*semantic.FunctionParam{{Key: &semantic.Identifier{Name: "col"}}},
				Body: &semantic.FloatLiteral{
					Value: 3.14159,
				},
			},
		},
		data: []query.Table{&executetest.Table{
			ColMeta: []query.ColMeta{
				{Label: "server1", Type: query.TFloat},
				{Label: "local", Type: query.TFloat},
				{Label: "server2", Type: query.TFloat},
			},
			Data: [][]interface{}{
				{1.0, 2.0, 3.0},
				{11.0, 12.0, 13.0},
				{21.0, 22.0, 23.0},
			},
		}},
		want: []*executetest.Table{{
			ColMeta: []query.ColMeta{
				{Label: "local", Type: query.TFloat},
			},
			Data: [][]interface{}{
				{2.0},
				{12.0},
				{22.0},
			},
		}},
	},
*/
