package functions

import (
	"fmt"

	"github.com/influxdata/platform/query/compiler"

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
const KeepKind = "keep"

type RenameOpSpec struct {
	RenameCols map[string]string            `json:"columns"`
	RenameFn   *semantic.FunctionExpression `json:"fn"`
}

type DropOpSpec struct {
	DropCols      []string                     `json:"columns"`
	DropPredicate *semantic.FunctionExpression `json:"fn"`
}

type KeepOpSpec struct {
	KeepCols      []string                     `json:"columns"`
	KeepPredicate *semantic.FunctionExpression `json:"fn"`
}

var renameSignature = query.DefaultFunctionSignature()
var dropSignature = query.DefaultFunctionSignature()
var keepSignature = query.DefaultFunctionSignature()

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

	keepSignature.Params["columns"] = semantic.Object
	keepSignature.Params["fn"] = semantic.Function

	query.RegisterFunction(KeepKind, createKeepOpSpec, keepSignature)
	query.RegisterOpSpec(KeepKind, newKeepOp)
	// TODO: RegisterProcedureSpec
	plan.RegisterProcedureSpec(KeepKind, newDropProcedure, KeepKind)

	execute.RegisterTransformation(RenameKind, createRenameDropTransformation)
	execute.RegisterTransformation(DropKind, createRenameDropTransformation)
	execute.RegisterTransformation(KeepKind, createRenameDropTransformation)
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
		fn, err := interpreter.ResolveFunction(f)
		if err != nil {
			return nil, err
		}

		dropPredicate = fn
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

func createKeepOpSpec(args query.Arguments, a *query.Administration) (query.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	var cols values.Array
	if c, ok, err := args.GetArray("columns", semantic.String); err != nil {
		return nil, err
	} else if ok {
		cols = c
	}

	var keepPredicate *semantic.FunctionExpression
	if f, ok, err := args.GetFunction("fn"); err != nil {
		return nil, err
	} else if ok {
		fn, err := interpreter.ResolveFunction(f)
		if err != nil {
			return nil, err
		}

		keepPredicate = fn
	}

	if cols == nil && keepPredicate == nil {
		return nil, errors.New("keep error: neither column list nor predicate function provided")
	}

	if cols != nil && keepPredicate != nil {
		return nil, errors.New("keep error: both column list and predicate provided")
	}

	var keepCols []string
	var err error
	if cols != nil {
		keepCols, err = interpreter.ToStringArray(cols)
		if err != nil {
			return nil, err
		}
	}

	return &KeepOpSpec{
		KeepCols:      keepCols,
		KeepPredicate: keepPredicate,
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

func newKeepOp() query.OperationSpec {
	return new(KeepOpSpec)
}

func (s *KeepOpSpec) Kind() query.OperationKind {
	return KeepKind
}

type RenameDropProcedureSpec struct {
	RenameCols map[string]string
	RenameFn   *semantic.FunctionExpression
	// The same field is used for both columns to drop and columns to keep
	DropKeepCols map[string]bool
	// the same field is used for the drop predicate and the keep predicate
	DropKeepPredicate *semantic.FunctionExpression
	// Denotes whether we're going to do a drop or a keep
	KeepSpecified bool
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
	ns.DropKeepCols = s.DropKeepCols
	ns.RenameFn = s.RenameFn
	ns.DropKeepPredicate = s.DropKeepPredicate
	return ns
}

func (s *RenameDropProcedureSpec) PushDownRules() []plan.PushDownProcedureSpec {
	return nil
}

func (s *RenameDropProcedureSpec) PushDown() (root *plan.Procedure, dup func() *plan.Procedure) {
	return nil, nil
}

// Keep and Drop are only inverses, so they share the same procedure constructor
func newDropProcedure(qs query.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	pr := &RenameDropProcedureSpec{}
	switch s := qs.(type) {
	case *DropOpSpec:
		if s.DropCols != nil {
			pr.DropKeepCols = toStringSet(s.DropCols)
		}
		pr.DropKeepPredicate = s.DropPredicate
	case *KeepOpSpec:
		// flip use of dropCols field from drop to keep
		// we can't completely invert the list of columns or the predicate yet at this step,
		// so we have to rely on this flag
		pr.KeepSpecified = true
		if s.KeepCols != nil {
			pr.DropKeepCols = toStringSet(s.KeepCols)
		}
		pr.DropKeepPredicate = s.KeepPredicate

	default:
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return pr, nil
}

func toStringSet(arr []string) map[string]bool {
	if arr == nil {
		return nil
	}
	ret := make(map[string]bool, len(arr))
	for _, s := range arr {
		ret[s] = true
	}
	return ret
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
	d                 execute.Dataset
	cache             execute.TableBuilderCache
	renameCols        map[string]string
	renameFn          compiler.Func
	renameColParam    string
	dropKeepCols      map[string]bool
	keepSpecified     bool
	dropKeepPredicate compiler.Func
	dropKeepColParam  string
}

func newFunc(fn *semantic.FunctionExpression, types [2]semantic.Type) (compiler.Func, string, error) {
	scope, decls := query.BuiltIns()
	compileCache := compiler.NewCompilationCache(fn, scope, decls)
	if len(fn.Params) != 1 {
		return nil, "", fmt.Errorf("function should only have a single parameter, got %d", len(fn.Params))
	}
	paramName := fn.Params[0].Key.Name

	compiled, err := compileCache.Compile(map[string]semantic.Type{
		paramName: types[0],
	})
	if err != nil {
		return nil, "", err
	}

	if compiled.Type() != types[1] {
		return nil, "", fmt.Errorf("provided function does not evaluate to type %s", types[1].Kind())
	}

	return compiled, paramName, nil
}

func NewRenameDropTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec plan.ProcedureSpec) (*renameDropTransformation, error) {
	s, ok := spec.(*RenameDropProcedureSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", spec)
	}

	var renameMapFn compiler.Func
	var renameColParam string
	if s.RenameFn != nil {
		compiledFn, param, err := newFunc(s.RenameFn, [2]semantic.Type{semantic.String, semantic.String})
		if err != nil {
			return nil, err
		}
		renameMapFn = compiledFn
		renameColParam = param
	}

	var dropKeepPredicate compiler.Func
	var dropKeepColParam string
	if s.DropKeepPredicate != nil {
		compiledFn, param, err := newFunc(s.DropKeepPredicate, [2]semantic.Type{semantic.String, semantic.Bool})
		if err != nil {
			return nil, err
		}

		dropKeepPredicate = compiledFn
		dropKeepColParam = param
	}

	return &renameDropTransformation{
		d:                 d,
		cache:             cache,
		renameCols:        s.RenameCols,
		renameFn:          renameMapFn,
		renameColParam:    renameColParam,
		dropKeepCols:      s.DropKeepCols,
		dropKeepPredicate: dropKeepPredicate,
		dropKeepColParam:  dropKeepColParam,
		keepSpecified:     s.KeepSpecified,
	}, nil
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

	if t.dropKeepCols != nil && t.renameCols != nil {
		for k := range t.renameCols {
			if _, ok := t.dropKeepCols[k]; ok {
				return fmt.Errorf(`Cannot rename column "%s" which is marked for drop`, k)
			}
		}
	}

	// these could be merged into one, but separate for clarity.

	var renameFnScope map[string]values.Value
	if t.renameFn != nil {
		renameFnScope = make(map[string]values.Value, 1)
	}

	var dropFnScope map[string]values.Value
	if t.dropKeepPredicate != nil {
		dropFnScope = make(map[string]values.Value, 1)
	}

	builderCols := tbl.Cols()
	if t.renameCols != nil {
		for c := range t.renameCols {
			if execute.ColIdx(c, builderCols) < 0 {
				return fmt.Errorf(`rename error: column "%s" doesn't exist`, c)
			}
		}
	}

	if t.dropKeepCols != nil {
		for c := range t.dropKeepCols {
			if execute.ColIdx(c, builderCols) < 0 {
				if t.keepSpecified {
					return fmt.Errorf(`keep error: column "%s" doesn't exist`, c)
				}
				return fmt.Errorf(`drop error: column "%s" doesn't exist`, c)
			}
		}
	}

	// If `keepSpecified` is true, i.e., we want to exclusively keep the columns listed in dropCols
	// as opposed to exclusively dropping them, we update dropCols to be the list of all columns to drop
	// to simplify further logic.
	if t.keepSpecified && t.dropKeepCols != nil {
		exclusiveDropCols := make(map[string]bool, len(tbl.Cols()))
		for _, c := range tbl.Cols() {
			if _, ok := t.dropKeepCols[c.Label]; !ok {
				exclusiveDropCols[c.Label] = true
			}
		}
		t.dropKeepCols = exclusiveDropCols
	}

	for i, c := range tbl.Cols() {
		name := c.Label
		// Cannot have both column list and dropKeepPredicate; one must be nil
		if t.dropKeepCols != nil {
			if _, exists := t.dropKeepCols[name]; exists {
				continue
			}
		} else if t.dropKeepPredicate != nil {
			dropFnScope[t.dropKeepColParam] = values.NewStringValue(name)
			if pass, err := t.dropKeepPredicate.EvalBool(dropFnScope); err != nil {
				return err
			} else if pass != t.keepSpecified {
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
			renameFnScope[t.renameColParam] = values.NewStringValue(name)
			newName, err := t.renameFn.EvalString(renameFnScope)
			if err != nil {
				return err
			}
			col.Label = newName
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
