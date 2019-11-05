package openapi

import (
	"fmt"
	"reflect"
)

// FieldError is the error returned when an
// error related to a field occurs.
type FieldError struct {
	Name              string
	TypeName          string
	Type              reflect.Type
	Message           string
	ParameterLocation string
	Parent            reflect.Type
}

// Error implements the builtin error interface for FieldError.
func (fe *FieldError) Error() string {
	return fmt.Sprintf("%s: field=%s, type=%s", fe.Message, fe.Name, fe.TypeName)
}

// TypeError is the error returned when the generator
// encounters an unknow or unsupported type.
type TypeError struct {
	Message string
	Type    reflect.Type
}

// Error implements the builtin error interface for TypeError.
func (te *TypeError) Error() string {
	return fmt.Sprintf("%s: type=%s, kind=%s", te.Message, te.Type, te.Type.Kind())
}
