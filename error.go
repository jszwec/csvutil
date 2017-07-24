package recenc

import (
	"errors"
	"reflect"
)

// ErrFieldCount is returned when header's length doesn't match the length of
// the read record.
var ErrFieldCount = errors.New("wrong number of fields in record")

// An UnmarshalTypeError describes a string value that was not appropriate for
// a value of a specific Go type.
type UnmarshalTypeError struct {
	Value string       // string value
	Type  reflect.Type // type of Go value it could not be assigned to
}

func (e *UnmarshalTypeError) Error() string {
	return "recenc: cannot unmarshal " + e.Value + " into Go value of type " + e.Type.String()
}

// An UnsupportedTypeError is returned when attempting to decode an unsupported
// value type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "recenc: unsupported type: " + e.Type.String()
}

// An InvalidUnmarshalError describes an invalid argument passed to Decode.
// (The argument to Decode must be a non-nil pointer.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "recenc: Decode(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "recenc: Decode(non-pointer " + e.Type.String() + ")"
	}

	if indirect(reflect.New(e.Type)).Type().Kind() != reflect.Struct {
		return "recenc: Decode(non-struct pointer)"
	}

	return "recenc: Decode(nil " + e.Type.String() + ")"
}
