package csvutil

import (
	"encoding"
	"errors"
	"reflect"
	"strconv"
)

var (
	textUnmarshaler = reflect.TypeOf(new(encoding.TextUnmarshaler)).Elem()
	csvUnmarshaler  = reflect.TypeOf(new(Unmarshaler)).Elem()
)

type decodeFunc func(s string, v reflect.Value) error

func decodeString(s string, v reflect.Value) error {
	v.SetString(s)
	return nil
}

func decodeInt(s string, v reflect.Value) error {
	n, err := strconv.ParseInt(s, 10, v.Type().Bits())
	if err != nil {
		return &UnmarshalTypeError{Value: s, Type: v.Type()}
	}
	v.SetInt(n)
	return nil
}

func decodeUint(s string, v reflect.Value) error {
	n, err := strconv.ParseUint(s, 10, v.Type().Bits())
	if err != nil {
		return &UnmarshalTypeError{Value: s, Type: v.Type()}
	}
	v.SetUint(n)
	return nil
}

func decodeFloat(s string, v reflect.Value) error {
	n, err := strconv.ParseFloat(s, v.Type().Bits())
	if err != nil {
		return &UnmarshalTypeError{Value: s, Type: v.Type()}
	}
	v.SetFloat(n)
	return nil
}

func decodeBool(s string, v reflect.Value) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return &UnmarshalTypeError{Value: s, Type: v.Type()}
	}
	v.SetBool(b)
	return nil
}

func decodePtrTextUnmarshaler(s string, v reflect.Value) error {
	if v.CanAddr() {
		return decodeTextUnmarshaler(s, v.Addr())
	}
	return errors.New("cannot take pointer")
}

func decodeTextUnmarshaler(s string, v reflect.Value) error {
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	return v.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s))
}

func decodePtrFieldUnmarshaler(s string, v reflect.Value) error {
	if v.CanAddr() {
		return decodeFieldUnmarshaler(s, v.Addr())
	}
	return errors.New("cannot take pointer")
}

func decodeFieldUnmarshaler(s string, v reflect.Value) error {
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	return v.Interface().(Unmarshaler).UnmarshalCSV(s)
}

func decodePtr(s string, v reflect.Value) error {
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	elem := v.Elem()

	decode, err := decodeFn(elem.Type())
	if err != nil {
		return err
	}
	return decode(s, elem)
}

func decodeInterface(s string, v reflect.Value) error {
	if v.NumMethod() == 0 {
		v.Set(reflect.ValueOf(s))
		return nil
	}
	return &UnmarshalTypeError{
		Value: s,
		Type:  v.Type(),
	}
}

func decodeFn(typ reflect.Type) (decodeFunc, error) {
	if typ.Implements(csvUnmarshaler) {
		return decodeFieldUnmarshaler, nil
	}
	if reflect.PtrTo(typ).Implements(csvUnmarshaler) {
		return decodePtrFieldUnmarshaler, nil
	}
	if typ.Implements(textUnmarshaler) {
		return decodeTextUnmarshaler, nil
	}
	if reflect.PtrTo(typ).Implements(textUnmarshaler) {
		return decodePtrTextUnmarshaler, nil
	}

	switch typ.Kind() {
	case reflect.Ptr:
		return decodePtr, nil
	case reflect.Interface:
		return decodeInterface, nil
	case reflect.String:
		return decodeString, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return decodeInt, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return decodeUint, nil
	case reflect.Float32, reflect.Float64:
		return decodeFloat, nil
	case reflect.Bool:
		return decodeBool, nil
	}

	return nil, &UnsupportedTypeError{Type: typ}
}
