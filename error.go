package csvutil

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
	return "csvutil: cannot unmarshal " + e.Value + " into Go value of type " + e.Type.String()
}

// An UnsupportedTypeError is returned when attempting to decode an unsupported
// value type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "csvutil: unsupported type: " + e.Type.String()
}

// An InvalidDecodeError describes an invalid argument passed to Decode.
// (The argument to Decode must be a non-nil pointer.)
type InvalidDecodeError struct {
	Type reflect.Type
}

func (e *InvalidDecodeError) Error() string {
	if e.Type == nil {
		return "csvutil: Decode(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "csvutil: Decode(non-pointer " + e.Type.String() + ")"
	}

	if indirect(reflect.New(e.Type)).Type().Kind() != reflect.Struct {
		return "csvutil: Decode(non-struct pointer)"
	}

	return "csvutil: Decode(nil " + e.Type.String() + ")"
}

// An InvalidUnmarshalError describes an invalid argument passed to Unmarshal.
// (The argument to Unmarshal must be a non-nil slice pointer.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "csvutil: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "csvutil: Unmarshal(non-pointer " + e.Type.String() + ")"
	}

	return "csvutil: Unmarshal(non-slice pointer)"
}

// InvalidEncodeError is returned by Encode when the passed argument v is not
// a struct.
type InvalidEncodeError struct {
	Type reflect.Type
}

func (e *InvalidEncodeError) Error() string {
	return "csvutil: Encode(" + e.Type.String() + ")"
}

// MarshalerError is returned by Encoder when MarshalCSV or MarshalText returned
// an error.
type MarshalerError struct {
	Type          reflect.Type
	MarshalerType string
	Err           error
}

func (e *MarshalerError) Error() string {
	return "csvutil: error calling " + e.MarshalerType + " for type " + e.Type.String() + ": " + e.Err.Error()
}
