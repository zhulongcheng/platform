package functions

import (
	"fmt"

	"github.com/influxdata/platform/query"
	"github.com/influxdata/platform/query/execute"
	"github.com/influxdata/platform/query/plan"
	"github.com/influxdata/platform/query/values"
)

const SchemaMutationKind = "SchemaMutation"

var SchemaMutationOps = []query.OperationKind{
	RenameKind,
	DropKind,
	KeepKind,
}

func init() {
	plan.RegisterProcedureSpec(SchemaMutationKind, newSchemaMutationProcedure, SchemaMutationOps...)
	execute.RegisterTransformation(SchemaMutationKind, createSchemaMutationTransformation)
}

type BuilderContext struct {
	TableColumns []query.ColMeta
	TableKey     query.GroupKey
	ColIdxMap    []int
}

func NewBuilderContext(tbl query.Table) *BuilderContext {
	return &BuilderContext{
		TableColumns: tbl.Cols(),
		TableKey:     tbl.Key(),
		ColIdxMap:    make([]int, len(tbl.Cols())),
	}
}

func (b *BuilderContext) Cols() []query.ColMeta {
	return b.TableColumns
}

func (b *BuilderContext) Key() query.GroupKey {
	return b.TableKey
}

func (b *BuilderContext) ColMap() []int {
	return b.ColIdxMap
}

func (b *BuilderContext) UpdateCols(newCols []query.ColMeta) {
	b.TableColumns = newCols
}

func (b *BuilderContext) UpdateKey(newKey query.GroupKey) {
	b.TableKey = newKey
}

func (b *BuilderContext) UpdateColMap(newMap []int) {
	b.ColIdxMap = newMap
}

func (b *BuilderContext) Copy() *BuilderContext {
	var newColumns []query.ColMeta
	if b.TableColumns != nil {
		newColumns = make([]query.ColMeta, len(b.TableColumns))
		copy(newColumns, b.TableColumns)
	}

	var newKey query.GroupKey
	if b.TableKey != nil {
		vs := make([]values.Value, 0, len(b.TableKey.Cols()))
		for i := range b.TableKey.Cols() {
			vs = append(vs, b.TableKey.Value(i))
		}
		newKey = execute.NewGroupKey(b.TableKey.Cols(), vs)
	}

	var newIdxMap []int
	if b.ColIdxMap != nil {
		newIdxMap := make([]int, len(b.ColIdxMap))
		for i, idx := range b.ColIdxMap {
			newIdxMap[i] = idx
		}
	}

	return &BuilderContext{
		TableColumns: newColumns,
		TableKey:     newKey,
		ColIdxMap:    newIdxMap,
	}
}

type SchemaMutator interface {
	Mutate(ctx *BuilderContext) error
	Copy() SchemaMutator
}

type SchemaMutationOperationSpec interface {
	query.OperationSpec
	GetMutator() (SchemaMutator, error)
}

type SchemaMutationProcedureSpec struct {
	Mutations []SchemaMutator
}

func (s *SchemaMutationProcedureSpec) Kind() plan.ProcedureKind {
	return SchemaMutationKind
}

func (s *SchemaMutationProcedureSpec) Copy() plan.ProcedureSpec {
	newMutations := make([]SchemaMutator, len(s.Mutations))
	for _, m := range newMutations {
		newMutations = append(newMutations, m.Copy())
	}
	return &SchemaMutationProcedureSpec{
		Mutations: newMutations,
	}
}

func (s *SchemaMutationProcedureSpec) PushDownRules() []plan.PushDownProcedureSpec {
	return nil
}

func (s *SchemaMutationProcedureSpec) PushDown() (root *plan.Procedure, dup func() *plan.Procedure) {
	return nil, nil
}

func newSchemaMutationProcedure(qs query.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	s, ok := qs.(SchemaMutationOperationSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T doesn't implement SchemaMutationProcedureSpec", qs)
	}

	mutator, err := s.GetMutator()
	if err != nil {
		return nil, err
	}
	return &SchemaMutationProcedureSpec{
		Mutations: []SchemaMutator{mutator},
	}, nil
}

type schemaMutationTransformation struct {
	d         execute.Dataset
	cache     execute.TableBuilderCache
	mutations []SchemaMutator
}

func createSchemaMutationTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)

	t, err := NewSchemaMutationTransformation(d, cache, spec)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

func NewSchemaMutationTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec plan.ProcedureSpec) (*schemaMutationTransformation, error) {
	s, ok := spec.(*SchemaMutationProcedureSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", spec)
	}

	return &schemaMutationTransformation{
		d:         d,
		cache:     cache,
		mutations: s.Mutations,
	}, nil
}

func (t *schemaMutationTransformation) Process(id execute.DatasetID, tbl query.Table) error {
	ctx := NewBuilderContext(tbl)
	for _, m := range t.mutations {
		m.Mutate(ctx)
	}

	builder, created := t.cache.TableBuilder(ctx.Key())
	if created {
		for _, c := range ctx.Cols() {
			builder.AddCol(c)
		}
	}

	return tbl.Do(func(cr query.ColReader) error {
		for i := 0; i < cr.Len(); i++ {
			execute.AppendMappedRecord(i, cr, builder, ctx.ColMap())
		}
		return nil
	})
}

func (t *schemaMutationTransformation) RetractTable(id execute.DatasetID, key query.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *schemaMutationTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *schemaMutationTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *schemaMutationTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
