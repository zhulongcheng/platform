package errors

import (
	"fmt"
	"net/http"
	"strings"
)

// Type of an error.
type Type int

const (
	// InternalError indicates an unexpected error condition.
	InternalError Type = iota
	// MalformedData indicates malformed input, such as unparsable JSON.
	MalformedData
	// InvalidData indicates that data is well-formed, but invalid.
	InvalidData
	// Forbidden indicates a forbidden operation.
	Forbidden
	// NotFound indicates a correct operation returns no results.
	NotFound
)

var typCode = []int{
	http.StatusInternalServerError,
	http.StatusBadRequest,
	http.StatusUnprocessableEntity,
	http.StatusForbidden,
	http.StatusUnprocessableEntity,
}

var typStr = []string{
	"Internal Error: ",
	"Malformed Data: ",
	"Invalid Data: ",
	"Forbidden: ",
	"Not Found: ",
}

// String returns the string value of the type
func (t Type) String() string {
	return typStr[t]
}

// TypedError wraps error with a immutable type
type TypedError interface {
	Reference() Type
	// WrapLocation will append the location to trace the error
	WrapLocation(loc string)
	error
}

type typedErr struct {
	typ Type
	err error
	loc []string
}

func (e typedErr) Error() string {
	msg := fmt.Sprintf("%s: %v", typStr[e.typ], e.err)
	var loc string
	if len(e.loc) > 0 {
		loc = "; location: " + strings.Join(e.loc, ">")
	}
	return msg + loc
}

func (e typedErr) Reference() Type {
	return e.typ
}

func (e typedErr) WrapLocation(loc string) {
	e.loc = append(e.loc, loc)
}

func newTypedError(typ Type) func(err error) TypedError {
	return func(err error) TypedError {
		if err == nil {
			return nil
		}
		return typedErr{
			typ: typ,
			err: err,
			loc: make([]string, 0),
		}
	}
}

// funcs
var (
	InternalErrorf = newTypedError(InternalError)
	MalformedDataf = newTypedError(MalformedData)
	InvalidDataf   = newTypedError(InvalidData)
	Forbiddenf     = newTypedError(Forbidden)
	NotFoundf      = newTypedError(NotFound)
)
