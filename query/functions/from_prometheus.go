package functions

import (
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/platform/query"
	"github.com/influxdata/platform/query/ast"
	"github.com/influxdata/platform/query/execute"
	"github.com/influxdata/platform/query/functions/prometheus"
	"github.com/influxdata/platform/query/plan"
	"github.com/influxdata/platform/query/semantic"
	"github.com/pkg/errors"
)

const FromPromKind = "fromProm"

type FromPromOpSpec struct {
	Host    string
	Matcher string
}

var fromPromSignature = semantic.FunctionSignature{
	Params: map[string]semantic.Type{
		"host":    semantic.String,
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
		spec.Host = host
	} else {
		return nil, errors.New("must specify a host")
	}

	if matcher, ok, err := args.GetString("matcher"); err != nil {
		return nil, err
	} else if ok {
		spec.Matcher = matcher
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
	Host    string
	Matcher string

	Bounds plan.BoundsSpec
}

func newFromPromProcedure(qs query.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromPromOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &FromPromProcedureSpec{
		Host:    spec.Host,
		Matcher: spec.Matcher,
	}, nil
}

func (s *FromPromProcedureSpec) Kind() plan.ProcedureKind {
	return FromPromKind
}
func (s *FromPromProcedureSpec) TimeBounds() plan.BoundsSpec {
	return s.Bounds
}
func (s *FromPromProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromPromProcedureSpec)

	ns.Matcher = s.Matcher
	ns.Bounds = s.Bounds
	ns.Host = s.Host

	return ns
}
func (s *FromPromProcedureSpec) SetMatcherFromFilter(fn *semantic.FunctionExpression) error {
	fmt.Println("pushing down")
	if s.Matcher != "" {
		return nil
	}
	m, err := s.toMatcher(fn.Body.(semantic.Expression))
	if err != nil {
		fmt.Println("matcher error: ", err)
		return err
	}
	s.Matcher = fmt.Sprintf("{%s}", m)
	fmt.Println("matcher ", s.Matcher)
	return nil
}

func (s *FromPromProcedureSpec) toMatcher(n semantic.Expression) (string, error) {
	switch n := n.(type) {
	case *semantic.LogicalExpression:
		left, err := s.toMatcher(n.Left)
		if err != nil {
			return "", errors.Wrap(err, "left hand side")
		}
		right, err := s.toMatcher(n.Right)
		if err != nil {
			return "", errors.Wrap(err, "right hand side")
		}
		switch n.Operator {
		case ast.AndOperator:
			return fmt.Sprintf("%s,%s", left, right), nil
		case ast.OrOperator:
			return "", errors.New("or operator not supported in fromProm")
		default:
			return "", fmt.Errorf("unknown logical operator %v", n.Operator)
		}
	case *semantic.BinaryExpression:
		left, err := s.toMatcher(n.Left)
		if err != nil {
			return "", errors.Wrap(err, "left hand side")
		}
		right, err := s.toMatcher(n.Right)
		if err != nil {
			return "", errors.Wrap(err, "right hand side")
		}

		switch n.Operator {
		case ast.EqualOperator:
			return fmt.Sprintf("%s=\"%s\"", left, right), nil
		case ast.NotEqualOperator:
			return fmt.Sprintf("%s!=\"%s\"", left, right), nil
		case ast.RegexpMatchOperator:
			return fmt.Sprintf("%s=~\"%s\"", left, right), nil
		case ast.NotRegexpMatchOperator:
			return fmt.Sprintf("%s!~\"%s\"", left, right), nil
		case ast.StartsWithOperator:
			return fmt.Sprintf("%s=~\"^%s.*\"", left, right), nil
		case ast.LessThanOperator:
			return "", errors.New("< not supported")
		case ast.LessThanEqualOperator:
			return "", errors.New("<= not supported")
		case ast.GreaterThanOperator:
			return "", errors.New("> not supported")
		case ast.GreaterThanEqualOperator:
			return "", errors.New(">= not supported")
		case ast.InOperator:
			return fmt.Sprintf("%s=~\"%s\"", left, right), nil
		default:
			return "", fmt.Errorf("unknown operator %v", n.Operator)
		}
	case *semantic.StringLiteral:
		return n.Value, nil
	case *semantic.IntegerLiteral:
		return fmt.Sprintf("%d", n.Value), nil
	case *semantic.BooleanLiteral:
		if n.Value {
			return "true", nil
		}
		return "false", nil
	case *semantic.FloatLiteral:
		return fmt.Sprintf("%f", n.Value), nil
	case *semantic.RegexpLiteral:
		return n.Value.String(), nil
	case *semantic.MemberExpression:
		if n.Property == "_value" {
			return "", errors.New("unable to push value filtering down to Prometheus")
		}
		return n.Property, nil
	// 	// Sanity check that the object is the objectName identifier
	// 	if ident, ok := n.Object.(*semantic.IdentifierExpression); !ok || ident.Name != objectName {
	// 		return nil, fmt.Errorf("unknown object %q", n.Object)
	// 	}
	// 	if n.Property == "_value" {
	// 		return &Node{
	// 			NodeType: NodeTypeFieldRef,
	// 			Value: &Node_FieldRefValue{
	// 				FieldRefValue: "_value",
	// 			},
	// 		}, nil
	// 	}
	// 	return &Node{
	// 		NodeType: NodeTypeTagRef,
	// 		Value: &Node_TagRefValue{
	// 			TagRefValue: n.Property,
	// 		},
	// 	}, nil
	case *semantic.ArrayExpression:
		vals := make([]string, 0, len(n.Elements))
		for _, e := range n.Elements {
			vals = append(vals, e.(*semantic.StringLiteral).Value)
		}
		return strings.Join(vals, "|"), nil
	case *semantic.DurationLiteral:
		return "", errors.New("duration literals not supported in storage predicates")
	case *semantic.DateTimeLiteral:
		return "", errors.New("time literals not supported in storage predicates")
	default:
		return "", fmt.Errorf("unsupported semantic expression type %T", n)
	}
}

func createFromPromSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec := prSpec.(*FromPromProcedureSpec)

	t := time.Now()
	return prometheus.NewSource(dsid, spec.Host, spec.Matcher, spec.Bounds.Start.Time(t), spec.Bounds.Stop.Time(t)), nil
}
