package functions

import (
	"fmt"

	"github.com/influxdata/platform/query"
	"github.com/influxdata/platform/query/execute"
	"github.com/influxdata/platform/query/interpreter"
	"github.com/influxdata/platform/query/plan"
	"github.com/influxdata/platform/query/semantic"
	"github.com/pkg/errors"
)

const FromPromKind = "fromProm"

type FromPromOpSpec struct {
	Hosts   []string
	Matcher string
}

var fromPromSignature = semantic.FunctionSignature{
	Params: map[string]semantic.Type{
		"host":    semantic.String,
		"hosts":   semantic.Array,
		"matcher": semantic.String,
	},
	ReturnType: query.TableObjectType,
}

func init() {
	query.RegisterFunction(FromPromKind, createFromPromOpSpec, fromPromSignature)
	query.RegisterOpSpec(FromPromKind, newFromPromOp)
	plan.RegisterProcedureSpec(FromPromKind, newFromPromProcedure, FromPromKind)
	execute.RegisterSource(FromPromKind, createFromPromSource)
}

func createFromPromOpSpec(args query.Arguments, a *query.Administration) (query.OperationSpec, error) {
	spec := new(FromPromOpSpec)

	if host, ok, err := args.GetString("host"); err != nil {
		return nil, err
	} else if ok {
		spec.Hosts = append(spec.Hosts, host)
	}

	if matcher, ok, err := args.GetString("matcher"); err != nil {
		return nil, err
	} else if ok {
		spec.Matcher = matcher
	}

	if array, ok, err := args.GetArray("hosts", semantic.String); err != nil {
		return nil, err
	} else if ok {
		if len(spec.Hosts) > 0 {
			return nil, errors.New("must specify either host or an array of hosts")
		}
		spec.Hosts, err = interpreter.ToStringArray(array)
		if err != nil {
			return nil, err
		}
	}

	if len(spec.Hosts) == 0 {
		return nil, errors.New("must specify either a host or hosts to query from")
	}

	return spec, nil
}

func newFromPromOp() query.OperationSpec {
	return new(FromPromOpSpec)
}

func (s *FromPromOpSpec) Kind() query.OperationKind {
	return FromPromKind
}

type FromPromProcedureSpec struct {
	Hosts   []string
	Matcher string

	Bounds plan.BoundsSpec
}

func newFromPromProcedure(qs query.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromPromOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &FromPromProcedureSpec{
		Hosts:   spec.Hosts,
		Matcher: spec.Matcher,
	}, nil
}

func (s *FromPromProcedureSpec) Kind() plan.ProcedureKind {
	return FromKind
}
func (s *FromPromProcedureSpec) TimeBounds() plan.BoundsSpec {
	return s.Bounds
}
func (s *FromPromProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromPromProcedureSpec)

	if len(s.Hosts) > 0 {
		ns.Hosts = make([]string, len(s.Hosts))
		copy(ns.Hosts, s.Hosts)
	}

	ns.Matcher = s.Matcher
	ns.Bounds = s.Bounds

	return ns
}

func createFromPromSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	return nil, errors.New("not implemented")
}
