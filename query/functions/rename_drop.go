package functions

import (
	"fmt"

	"github.com/influxdata/platform/query/compiler"

	"github.com/influxdata/platform/query/interpreter"
	"github.com/pkg/errors"

	"github.com/influxdata/platform/query"
	"github.com/influxdata/platform/query/execute"
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

	dropSignature.Params["columns"] = semantic.NewArrayType(semantic.String)
	dropSignature.Params["fn"] = semantic.Function

	query.RegisterFunction(DropKind, createDropOpSpec, dropSignature)
	query.RegisterOpSpec(DropKind, newDropOp)

	keepSignature.Params["columns"] = semantic.Object
	keepSignature.Params["fn"] = semantic.Function

	query.RegisterFunction(KeepKind, createKeepOpSpec, keepSignature)
	query.RegisterOpSpec(KeepKind, newKeepOp)
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

type RenameMutator struct {
	RenameCols map[string]string
	RenameFn   compiler.Func
	scope      map[string]values.Value
	paramName  string
}

func newRenameMutator(qs query.OperationSpec) (*RenameMutator, error) {
	s, ok := qs.(*RenameOpSpec)

	m := &RenameMutator{}
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	if s.RenameCols != nil {
		m.RenameCols = s.RenameCols
	}

	if s.RenameFn != nil {
		compiledFn, param, err := newFunc(s.RenameFn, [2]semantic.Type{semantic.String, semantic.String})
		if err != nil {
			return nil, err
		}

		m.RenameFn = compiledFn
		m.paramName = param
		m.scope = make(map[string]values.Value, 1)
	}
	return m, nil
}

func (m *RenameMutator) checkColumnReferences(cols []query.ColMeta) error {
	if m.RenameCols != nil {
		for c := range m.RenameCols {
			if execute.ColIdx(c, cols) < 0 {
				return fmt.Errorf(`rename error: column "%s" doesn't exist`, c)
			}
		}
	}
	return nil
}

func (m *RenameMutator) renameCol(col *query.ColMeta) error {
	if col == nil {
		return errors.New("rename error: cannot rename nil column")
	}
	if m.RenameCols != nil {
		if newName, ok := m.RenameCols[col.Label]; ok {
			col.Label = newName
		}
	} else if m.RenameFn != nil {
		m.scope[m.paramName] = values.NewStringValue(col.Label)
		newName, err := m.RenameFn.EvalString(m.scope)
		if err != nil {
			return err
		}
		col.Label = newName
	}
	return nil
}

func (m *RenameMutator) Mutate(ctx *BuilderContext) error {

	if err := m.checkColumnReferences(ctx.Cols()); err != nil {
		return err
	}

	keyCols := make([]query.ColMeta, 0, len(ctx.Cols()))
	keyValues := make([]values.Value, 0, len(ctx.Cols()))
	newCols := make([]query.ColMeta, 0, len(ctx.Cols()))
	newColMap := make([]int, 0, len(ctx.Cols()))

	for i, c := range ctx.Cols() {
		keyIdx := execute.ColIdx(c.Label, ctx.Key().Cols())
		keyed := keyIdx >= 0

		if err := m.renameCol(&c); err != nil {
			return err
		}

		if keyed {
			keyCols = append(keyCols, c)
			keyValues = append(keyValues, ctx.Key().Value(keyIdx))
		}
		newCols = append(newCols, c)
		newColMap = append(newColMap, i)
	}

	ctx.UpdateCols(newCols)
	ctx.UpdateKey(execute.NewGroupKey(keyCols, keyValues))
	ctx.UpdateColMap(newColMap)

	return nil
}

func (m *RenameMutator) Copy() SchemaMutator {
	nm := new(RenameMutator)
	nm.RenameCols = m.RenameCols
	nm.RenameFn = m.RenameFn
	return nm
}

type DropKeepMutator struct {
	DropKeepCols      map[string]bool
	dropExclusiveCols map[string]bool
	DropKeepPredicate compiler.Func
	Keep              bool
	paramName         string
	scope             map[string]values.Value
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

func newDropKeepMutator(qs query.OperationSpec) (*DropKeepMutator, error) {
	m := &DropKeepMutator{}
	switch s := qs.(type) {
	case *DropOpSpec:
		if s.DropCols != nil {
			m.DropKeepCols = toStringSet(s.DropCols)
		}
		if s.DropPredicate != nil {
			compiledFn, param, err := newFunc(s.DropPredicate, [2]semantic.Type{semantic.String, semantic.Bool})
			if err != nil {
				return nil, err
			}
			m.DropKeepPredicate = compiledFn
			m.paramName = param
			m.scope = make(map[string]values.Value, 1)
		}
	case *KeepOpSpec:
		if s.KeepCols != nil {
			m.DropKeepCols = toStringSet(s.KeepCols)
		}
		if s.KeepPredicate != nil {
			compiledFn, param, err := newFunc(s.KeepPredicate, [2]semantic.Type{semantic.String, semantic.Bool})
			if err != nil {
				return nil, err
			}
			m.DropKeepPredicate = compiledFn
			m.paramName = param
			m.scope = make(map[string]values.Value, 1)
		}
		m.Keep = true
	}

	return m, nil
}

func (m *DropKeepMutator) checkColumnReferences(cols []query.ColMeta) error {
	if m.DropKeepCols != nil {
		for c := range m.DropKeepCols {
			if execute.ColIdx(c, cols) < 0 {
				if m.Keep {
					return fmt.Errorf(`keep error: column "%s" doesn't exist`, c)
				}
				return fmt.Errorf(`drop error: column "%s" doesn't exist`, c)
			}
		}
	}
	return nil
}

func (m *DropKeepMutator) shouldDrop(col string) (bool, error) {
	m.scope[m.paramName] = values.NewStringValue(col)
	if shouldDrop, err := m.DropKeepPredicate.EvalBool(m.scope); err != nil {
		return false, err
	} else if m.Keep {
		return !shouldDrop, nil
	} else {
		return shouldDrop, nil
	}
}

func (m *DropKeepMutator) shouldDropCol(col string) (bool, error) {
	if m.dropExclusiveCols != nil {
		if _, exists := m.dropExclusiveCols[col]; exists {
			return true, nil
		}
	} else if m.DropKeepPredicate != nil {
		return m.shouldDrop(col)
	}
	return false, nil
}

func (m *DropKeepMutator) keepToDropCols(cols []query.ColMeta) {

	// If `m.Keep` is true, i.e., we want to exclusively keep the columns listed in DropKeepCols
	// as opposed to exclusively dropping them. So in that case, we invert the DropKeepCols map and store it
	// in exclusiveDropCols; exclusiveDropCols may be changed with each call to `Mutate`, but
	// `DropKeepCols` will not be.
	if m.Keep && m.DropKeepCols != nil {
		exclusiveDropCols := make(map[string]bool, len(cols))
		for _, c := range cols {
			if _, ok := m.DropKeepCols[c.Label]; !ok {
				exclusiveDropCols[c.Label] = true
			}
		}
		m.dropExclusiveCols = exclusiveDropCols
	} else if m.DropKeepCols != nil {
		m.dropExclusiveCols = m.DropKeepCols
	}
}

func (m *DropKeepMutator) Mutate(ctx *BuilderContext) error {
	if err := m.checkColumnReferences(ctx.Cols()); err != nil {
		return err
	}

	m.keepToDropCols(ctx.Cols())

	keyCols := make([]query.ColMeta, 0, len(ctx.Cols()))
	keyValues := make([]values.Value, 0, len(ctx.Cols()))
	newCols := make([]query.ColMeta, 0, len(ctx.Cols()))
	newColMap := make([]int, 0, len(ctx.Cols()))

	for i, c := range ctx.Cols() {
		if shouldDrop, err := m.shouldDropCol(c.Label); err != nil {
			return err
		} else if shouldDrop {
			continue
		}

		keyIdx := execute.ColIdx(c.Label, ctx.Key().Cols())
		if keyIdx >= 0 {
			keyCols = append(keyCols, c)
			keyValues = append(keyValues, ctx.Key().Value(keyIdx))
		}
		newCols = append(newCols, c)
		newColMap = append(newColMap, i)
	}

	ctx.UpdateCols(newCols)
	ctx.UpdateKey(execute.NewGroupKey(keyCols, keyValues))
	ctx.UpdateColMap(newColMap)

	return nil
}

func (m *DropKeepMutator) Copy() SchemaMutator {
	nm := new(DropKeepMutator)
	nm.DropKeepCols = m.DropKeepCols
	nm.DropKeepPredicate = m.DropKeepPredicate
	return nm
}

func (s *RenameOpSpec) GetMutator() (SchemaMutator, error) {
	m, err := newRenameMutator(s)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *DropOpSpec) GetMutator() (SchemaMutator, error) {
	m, err := newDropKeepMutator(s)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *KeepOpSpec) GetMutator() (SchemaMutator, error) {
	m, err := newDropKeepMutator(s)
	if err != nil {
		return nil, err
	}
	return m, nil
}
