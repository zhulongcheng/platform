package functions

import (
	"fmt"
	"log"

	"github.com/influxdata/platform/query/interpreter"

	"github.com/influxdata/platform/query"
	"github.com/influxdata/platform/query/execute"
	"github.com/influxdata/platform/query/plan"
	"github.com/influxdata/platform/query/semantic"
	"github.com/influxdata/platform/query/values"
)

const RenameKind = "rename"
const DropKind = "drop"

// TODO: keep transformation?

type RenameOpSpec struct {
	RenameCols map[string]string `json:"columns"`
}

type DropOpSpec struct {
	DropCols []string `json:"columns"`
}

var renameSignature = query.DefaultFunctionSignature()
var dropSignature = query.DefaultFunctionSignature()

func init() {
	rangeSignature.Params["columns"] = semantic.Object

	query.RegisterFunction(RenameKind, createRenameOpSpec, renameSignature)
	query.RegisterOpSpec(RenameKind, newRangeOp)
	plan.RegisterProcedureSpec(RenameKind, newRenameProcedure, RenameKind)

	dropSignature.Params["columns"] = semantic.NewArrayType(semantic.String)

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
	cols, err := args.GetRequiredObject("columns")
	if err != nil {
		return nil, err
	}

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
	spec := &RenameOpSpec{
		RenameCols: renameCols,
	}

	return spec, nil
}

func createDropOpSpec(args query.Arguments, a *query.Administration) (query.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	cols, err := args.GetRequiredArray("columns", semantic.String)
	if err != nil {
		return nil, err
	}

	dropCols, err := interpreter.ToStringArray(cols)
	if err != nil {
		return nil, err
	}
	return &DropOpSpec{
		DropCols: dropCols,
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

type RenameProcedureSpec struct {
	RenameCols map[string]string
	DropCols   map[string]bool
	KeepCols   map[string]bool
}

func newRenameProcedure(qs query.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	s, ok := qs.(*RenameOpSpec)

	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	var renameCols map[string]string
	if s.RenameCols != nil {
		renameCols = s.RenameCols
	} else {
		renameCols = make(map[string]string)
	}

	return &RenameProcedureSpec{
		RenameCols: renameCols,
	}, nil
}

func (s *RenameProcedureSpec) Kind() plan.ProcedureKind {
	return RenameKind
}

func (s *RenameProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(RenameProcedureSpec)
	ns.RenameCols = s.RenameCols
	return ns
}

func (s *RenameProcedureSpec) PushDownRules() []plan.PushDownProcedureSpec {
	return nil
}

func (s *RenameProcedureSpec) PushDown() (root *plan.Procedure, dup func() *plan.Procedure) {
	return nil, nil
}

type DropProcedureSpec struct {
	DropCols map[string]bool
}

func newDropProcedure(qs query.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	s, ok := qs.(*DropOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	dropCols := make(map[string]bool)
	if s.DropCols != nil {
		for _, c := range s.DropCols {
			dropCols[c] = true
		}
	}
	return &DropProcedureSpec{
		DropCols: dropCols,
	}, nil
}

func (s *DropProcedureSpec) Kind() plan.ProcedureKind {
	return DropKind
}

func (s *DropProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(DropProcedureSpec)
	ns.DropCols = s.DropCols
	return ns
}

func (s *DropProcedureSpec) PushDownRules() []plan.PushDownProcedureSpec {
	return nil
}

func (s *DerivativeProcedureSpec) PushDown() (root *plan.Procedure, dup func() *plan.Procedure) {
	return nil, nil
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
	d          execute.Dataset
	cache      execute.TableBuilderCache
	renameCols map[string]string
	dropCols   map[string]bool
	keepCols   map[string]bool // TODO: Remove if not needed
}

func NewRenameDropTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec plan.ProcedureSpec) (*renameDropTransformation, error) {
	switch s := spec.(type) {
	case *RenameProcedureSpec:
		return &renameDropTransformation{
			d:          d,
			cache:      cache,
			renameCols: s.RenameCols,
		}, nil
	case *DropProcedureSpec:
		return &renameDropTransformation{
			d:        d,
			cache:    cache,
			dropCols: s.DropCols,
		}, nil
	default:
		// May never trigger
		return nil, fmt.Errorf("invalid spec type %T", spec)
	}
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
	for i, c := range tbl.Cols() {
		name := c.Label

		if t.dropCols != nil {
			if _, exists := t.dropCols[name]; exists {
				continue
			}
		}

		col := c
		if t.renameCols != nil {
			if newName, ok := t.renameCols[name]; ok {
				col = query.ColMeta{
					Label: newName,
					Type:  c.Type,
				}
			}
		}
		colMap = append(colMap, i)
		builder.AddCol(col)
	}

	log.Println("cols: ", builder.Cols())
	log.Println(colMap)
	err := tbl.Do(func(cr query.ColReader) error {
		l := cr.Len()
		if l == 0 {
			return nil
		}
		for i := 0; i < l; i++ {
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
