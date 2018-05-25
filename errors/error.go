package errors

import (
	"fmt"
)

// Common errors.
const (
	AuthorizationNotFound       = Error("authorization not found")
	AuthorizationNotFoundConext = Error("authorization not found on context")

	OrganizationNotFound = Error("organization not found")
	UserNotFound         = Error("user not found")

	TokenNoutFoundContext = Error("token not found on context")

	URLMissingID = Error("url missing id")
	EmptyValue   = Error("empty value")
)

// Error is a constant string type error.
type Error string

// Error returns the string value
func (e Error) Error() string {
	return string(e)
}

func errWithValue(f string) func(args ...interface{}) error {
	return func(args ...interface{}) error {
		return fmt.Errorf(f, args...)
	}
}

// errors with value
var (
	OrganizationNameAlreadyExist = errWithValue("organization with name %s already exists")
	UserNameAlreadyExist         = errWithValue("user with name %s already exists")
)

func errWrap(f string) func(err error) error {
	return func(err error) error {
		return fmt.Errorf(f+": %s", err)
	}
}

// errors in wrap
var (
	FailedToGetStorageHost = errWrap("failed to get storage hosts")
	FailedToGetBucketName  = errWrap("failed to get bucket name")
)
