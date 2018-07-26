package execute

import (
	"fmt"

	"github.com/influxdata/platform/query"
	"github.com/influxdata/platform/query/compiler"
	"github.com/influxdata/platform/query/semantic"
	"github.com/influxdata/platform/query/values"
	"github.com/pkg/errors"
)

type columnFn struct {
	fn               *semantic.FunctionExpression
	compilationCache *compiler.CompilationCache
	scope            compiler.Scope
	preparedFn       compiler.Func
	colParamName     string
}

func newColumnFn(fn *semantic.FunctionExpression) (columnFn, error) {
	if len(fn.Params) != 1 {
		return columnFn{}, fmt.Errorf("function should only have a single parameter, got %d", len(fn.Params))
	}
	scope, decls := query.BuiltIns()
	return columnFn{
		compilationCache: compiler.NewCompilationCache(fn, scope, decls),
		scope:            make(compiler.Scope, 1),
		colParamName:     fn.Params[0].Key.Name,
	}, nil
}

func (f *columnFn) prepare() error {
	// Compile fn for given types
	fn, err := f.compilationCache.Compile(map[string]semantic.Type{
		f.colParamName: semantic.String,
	})
	if err != nil {
		return err
	}
	f.preparedFn = fn
	return nil
}

func (f *columnFn) eval(column string) (values.Value, error) {
	f.scope[f.colParamName] = values.NewStringValue(column)
	return f.preparedFn.Eval(f.scope)
}

type ColumnPredicateFn struct {
	columnFn
}

func NewColumnPredicateFn(fn *semantic.FunctionExpression) (*ColumnPredicateFn, error) {
	c, err := newColumnFn(fn)
	if err != nil {
		return nil, err
	}
	return &ColumnPredicateFn{
		columnFn: c,
	}, nil
}

func (f *ColumnPredicateFn) Prepare() error {
	err := f.columnFn.prepare()
	if err != nil {
		return err
	}
	if f.preparedFn.Type() != semantic.Bool {
		return errors.New("column predicate function does not evaluate to a boolean")
	}
	return nil
}

func (f *ColumnPredicateFn) Eval(column string) (bool, error) {
	v, err := f.columnFn.eval(column)
	if err != nil {
		return false, err
	}
	return v.Bool(), nil
}

type ColumnMapFn struct {
	columnFn
}

func NewColumnMapFn(fn *semantic.FunctionExpression) (*ColumnMapFn, error) {
	c, err := newColumnFn(fn)
	if err != nil {
		return nil, err
	}
	return &ColumnMapFn{
		columnFn: c,
	}, nil
}

func (f *ColumnMapFn) Prepare() error {
	err := f.columnFn.prepare()
	if err != nil {
		return err
	}
	if f.preparedFn.Type() != semantic.String {
		return errors.New("column map function does not evaluate to a string")
	}
	return nil
}

func (f *ColumnMapFn) Type() semantic.Type {
	return f.preparedFn.Type()
}

func (f *ColumnMapFn) Eval(column string) (string, error) {
	v, err := f.columnFn.eval(column)
	if err != nil {
		return "", err
	}
	return v.Str(), nil
}
