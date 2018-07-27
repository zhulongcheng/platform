package functions

import (
	"fmt"

	"github.com/influxdata/platform/query/interpreter"
	"github.com/pkg/errors"

	"github.com/influxdata/platform/query"
	"github.com/influxdata/platform/query/execute"
	"github.com/influxdata/platform/query/plan"
	"github.com/influxdata/platform/query/semantic"
	"github.com/influxdata/platform/query/values"
)

const RenameKind = "rename"
const DropKind = "drop"

// TODO: `keep` operation?

type RenameOpSpec struct {
	RenameCols map[string]string `json:"columns"`
	RenameFn   *semantic.FunctionExpression
}

type DropOpSpec struct {
	DropCols      []string `json:"columns"`
	DropPredicate *semantic.FunctionExpression
}

var renameSignature = query.DefaultFunctionSignature()
var dropSignature = query.DefaultFunctionSignature()

func init() {
	renameSignature.Params["columns"] = semantic.Object
	renameSignature.Params["fn"] = semantic.Function

	query.RegisterFunction(RenameKind, createRenameOpSpec, renameSignature)
	query.RegisterOpSpec(RenameKind, newRenameOp)
	plan.RegisterProcedureSpec(RenameKind, newRenameProcedure, RenameKind)

	dropSignature.Params["columns"] = semantic.NewArrayType(semantic.String)
	dropSignature.Params["fn"] = semantic.Function

	query.RegisterFunction(DropKind, createDropOpSpec, dropSignature)
	query.RegisterOpSpec(DropKind, newDropOp)
	plan.RegisterProcedureSpec(DropKind, newDropProcedure, DropKind)

	execute.RegisterTransformation(RenameKind, createRenameDropTransformation)
	execute.RegisterTransformation(DropKind, createRenameDropTransformation)
}

func createRenameOpSpec(args query.Arguments, a *query.Administration) (query.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	var cols values.Object
	if c, ok, err := args.GetObject("columns"); err != nil {
		return nil, err
	} else if ok {
		cols = c
	}

	var renameFn *semantic.FunctionExpression
	if f, ok, err := args.GetFunction("fn"); err != nil {
		return nil, err
	} else if ok {
		if fn, err := interpreter.ResolveFunction(f); err != nil {
			return nil, err
		} else {
			renameFn = fn
		}
	}

	if cols == nil && renameFn == nil {
		return nil, errors.New("rename error: neither column list nor map function provided")
	}

	if cols != nil && renameFn != nil {
		return nil, errors.New("rename error: both column list and map function provided")
	}

	spec := &RenameOpSpec{
		RenameFn: renameFn,
	}

	if cols != nil {
		var err error
		renameCols := make(map[string]string, cols.Len())
		// Check types of object values manually
		cols.Range(func(name string, v values.Value) {
			if err != nil {
				return
			}
			if v.Type() != semantic.String {
				err = fmt.Errorf("rename error: columns object contains non-string value of type %s", v.Type())
				return
			}
			renameCols[name] = v.Str()
		})
		if err != nil {
			return nil, err
		}
		spec.RenameCols = renameCols
	}

	return spec, nil
}

func createDropOpSpec(args query.Arguments, a *query.Administration) (query.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	var cols values.Array
	if c, ok, err := args.GetArray("columns", semantic.String); err != nil {
		return nil, err
	} else if ok {
		cols = c
	}

	var dropPredicate *semantic.FunctionExpression
	if f, ok, err := args.GetFunction("fn"); err != nil {
		return nil, err
	} else if ok {
		if fn, err := interpreter.ResolveFunction(f); err != nil {
			return nil, err
		} else {
			dropPredicate = fn
		}
	}

	if cols == nil && dropPredicate == nil {
		return nil, errors.New("drop error: neither column list nor predicate function provided")
	}

	if cols != nil && dropPredicate != nil {
		return nil, errors.New("drop error: both column list and predicate provided")
	}

	var dropCols []string
	var err error
	if cols != nil {
		dropCols, err = interpreter.ToStringArray(cols)
		if err != nil {
			return nil, err
		}
	}

	return &DropOpSpec{
		DropCols:      dropCols,
		DropPredicate: dropPredicate,
	}, nil
}

func newRenameOp() query.OperationSpec {
	return new(RenameOpSpec)
}

func (s *RenameOpSpec) Kind() query.OperationKind {
	return RenameKind
}

func newDropOp() query.OperationSpec {
	return new(DropOpSpec)
}

func (s *DropOpSpec) Kind() query.OperationKind {
	return DropKind
}

type RenameDropProcedureSpec struct {
	RenameCols    map[string]string
	RenameFn      *semantic.FunctionExpression
	DropCols      map[string]bool
	DropPredicate *semantic.FunctionExpression
}

func newRenameProcedure(qs query.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	s, ok := qs.(*RenameOpSpec)

	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	var renameCols map[string]string
	if s.RenameCols != nil {
		renameCols = s.RenameCols
	}

	return &RenameDropProcedureSpec{
		RenameCols: renameCols,
		RenameFn:   s.RenameFn,
	}, nil
}

func (s *RenameDropProcedureSpec) Kind() plan.ProcedureKind {
	return RenameKind
}

func (s *RenameDropProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(RenameDropProcedureSpec)
	ns.RenameCols = s.RenameCols
	ns.DropCols = s.DropCols
	ns.RenameFn = s.RenameFn
	ns.DropPredicate = s.DropPredicate
	return ns
}

/*
Push Down Rule cases:

drop into drop, i.e.:

|> drop(columns: ["a", "b", "c"])
|> drop(columns: ["d", "e", "f"])
==> drop(columns: ["a","b","c", "d", "e","f"]) //lists merged
|> drop(fn: (col) => col =~ /a/)
|> drop(fn: (col) => col =~ /b/)
==> drop(fn: (col) => col =~ /a/ or col =~ /b/) // predicate logic merged

cannot merge procedure with drop column argument and procedure with drop fn argument

rename into rename, i.e.:

|> rename(columns: {a: "b", c:"d"})
|> rename(columns: {b:"c", e:"f"})
==> rename(columns:{a:"c", c:"d", e:"f"})
|> rename(fn: (col) => "{col}_new")
|> rename(fn: (col) => "{col}_1")
==> rename(fn: (col) => {
	res = ((col) => "{col}_new")(col)
	return "{res}_1"
}) // nest function results

drop into rename
rename(columns: {a:"b"})
|> drop(columns: ["b"])
==> drop(columns:["a"])
rename(fn: (col) => "{col}_new")
|> drop(columns: ["a_new"])
==> no simplification, all internal
rename(fn: (col) => "{col}_new")
|> drop(fn: (col) => col == "c_new")
==> no simplification, all internal
rename(columns:{a:"b"})
|>   drop(fn: (col) => col == "b")
==> drop(columns: ["a"])


rename into drop:
drop(columns:["c"])
|> rename(columns:{"a","b"})
==> no simplification, all internal
drop(columns:["c"])
|> rename(columns:{c:"d"})
==> invalid!
*/

func (s *RenameDropProcedureSpec) PushDownRules() []plan.PushDownProcedureSpec {
	return nil
}

func (s *RenameDropProcedureSpec) PushDown() (root *plan.Procedure, dup func() *plan.Procedure) {
	return nil, nil
}

func newDropProcedure(qs query.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	s, ok := qs.(*DropOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	var dropCols map[string]bool
	if s.DropCols != nil {
		dropCols = make(map[string]bool)
		for _, c := range s.DropCols {
			dropCols[c] = true
		}
	}

	return &RenameDropProcedureSpec{
		DropCols:      dropCols,
		DropPredicate: s.DropPredicate,
	}, nil
}

func createRenameDropTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)

	t, err := NewRenameDropTransformation(d, cache, spec)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

type renameDropTransformation struct {
	d             execute.Dataset
	cache         execute.TableBuilderCache
	renameCols    map[string]string
	renameFn      *execute.ColumnMapFn
	dropCols      map[string]bool
	dropPredicate *execute.ColumnPredicateFn
}

func NewRenameDropTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec plan.ProcedureSpec) (*renameDropTransformation, error) {

	s, ok := spec.(*RenameDropProcedureSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", spec)
	}

	var renameMapFn *execute.ColumnMapFn
	var err error
	if s.RenameFn != nil {
		renameMapFn, err = execute.NewColumnMapFn(s.RenameFn)
		if err != nil {
			return nil, err
		}
	}

	var dropPredicate *execute.ColumnPredicateFn
	if s.DropPredicate != nil {
		dropPredicate, err = execute.NewColumnPredicateFn(s.DropPredicate)
		if err != nil {
			return nil, err
		}
	}

	return &renameDropTransformation{
		d:             d,
		cache:         cache,
		renameCols:    s.RenameCols,
		renameFn:      renameMapFn,
		dropCols:      s.DropCols,
		dropPredicate: dropPredicate,
	}, nil
	// May never trigger
}

func (t *renameDropTransformation) Process(id execute.DatasetID, tbl query.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return fmt.Errorf("rename found duplicate table with key: %v", tbl.Key())
	}

	// If we remove columns, column indices will be different between the
	// builder and the table - we need to keep track
	colMap := make([]int, builder.NCols())
	// TODO: error if overlap between dropCols and renameCols?
	if t.dropPredicate != nil {
		if err := t.dropPredicate.Prepare(); err != nil {
			return err
		}
	} else if t.renameFn != nil {
		if err := t.renameFn.Prepare(); err != nil {
			return err
		}
	}

	for i, c := range tbl.Cols() {
		name := c.Label

		// Cannot have both column list and dropPredicate; one must be nil
		if t.dropCols != nil {
			if _, exists := t.dropCols[name]; exists {
				continue
			}
		} else if t.dropPredicate != nil {
			if pass, err := t.dropPredicate.Eval(name); err != nil {
				return err
			} else if pass {
				continue
			}
		}

		col := c

		// Cannot have both column list and renameFn; one must be nil
		if t.renameCols != nil {
			if newName, ok := t.renameCols[name]; ok {
				col.Label = newName
			}
		} else if t.renameFn != nil {
			if newName, err := t.renameFn.Eval(name); err != nil {
				return err
			} else {
				col.Label = newName
			}
		}
		colMap = append(colMap, i)
		builder.AddCol(col)
	}

	err := tbl.Do(func(cr query.ColReader) error {
		for i := 0; i < cr.Len(); i++ {
			execute.AppendMappedRecord(i, cr, builder, colMap)
		}
		return nil
	})

	return err
}

func (t *renameDropTransformation) RetractTable(id execute.DatasetID, key query.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *renameDropTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *renameDropTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *renameDropTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
