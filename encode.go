package csvutil

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"reflect"
	"strconv"
)

var (
	textMarshaler = reflect.TypeOf(new(encoding.TextMarshaler)).Elem()
	csvMarshaler  = reflect.TypeOf(new(Marshaler)).Elem()
)

type encodeFunc func(v reflect.Value, buf *bytes.Buffer, omitempty bool) (int, error)

func encodeString(v reflect.Value, buf *bytes.Buffer, omitempty bool) (int, error) {
	return buf.WriteString(v.String())
}

func encodeInt() encodeFunc {
	var b [64]byte
	return func(v reflect.Value, buf *bytes.Buffer, omitempty bool) (int, error) {
		n := v.Int()
		if n == 0 && omitempty {
			return 0, nil
		}
		return buf.Write(strconv.AppendInt(b[:0], n, 10))
	}
}

func encodeUint() encodeFunc {
	var b [64]byte
	return func(v reflect.Value, buf *bytes.Buffer, omitempty bool) (int, error) {
		n := v.Uint()
		if n == 0 && omitempty {
			return 0, nil
		}
		return buf.Write(strconv.AppendUint(b[:0], n, 10))
	}
}

func encodeFloat(typ reflect.Type) encodeFunc {
	bits := typ.Bits()
	var b [64]byte
	return func(v reflect.Value, buf *bytes.Buffer, omitempty bool) (int, error) {
		f := v.Float()
		if f == 0 && omitempty {
			return 0, nil
		}
		return buf.Write(strconv.AppendFloat(b[:0], f, 'G', -1, bits))
	}
}

func encodeBool() encodeFunc {
	var b [5]byte // 'true' or 'false'
	return func(v reflect.Value, buf *bytes.Buffer, omitempty bool) (int, error) {
		t := v.Bool()
		if !t && omitempty {
			return 0, nil
		}
		return buf.Write(strconv.AppendBool(b[:0], t))
	}
}

func encodeInterface(v reflect.Value, buf *bytes.Buffer, omitempty bool) (int, error) {
	if !v.IsValid() || v.IsNil() || !v.Elem().IsValid() {
		return 0, nil
	}

	v = v.Elem()

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return 0, nil
		}
	default:
	}

	enc, err := encodeFn(v.Type())
	if err != nil {
		return 0, err
	}
	return enc(walkPtr(v), buf, omitempty)
}

func encodePtrMarshaler(v reflect.Value, buf *bytes.Buffer, omitempty bool) (int, error) {
	if v.CanAddr() {
		return encodeMarshaler(v.Addr(), buf, omitempty)
	}
	return 0, nil
}

func encodeTextMarshaler(v reflect.Value, buf *bytes.Buffer, _ bool) (int, error) {
	b, err := v.Interface().(encoding.TextMarshaler).MarshalText()
	if err != nil {
		return 0, &MarshalerError{Type: v.Type(), MarshalerType: "MarshalText", Err: err}
	}
	return buf.Write(b)
}

func encodePtrTextMarshaler(v reflect.Value, buf *bytes.Buffer, omitempty bool) (int, error) {
	if v.CanAddr() {
		return encodeTextMarshaler(v.Addr(), buf, omitempty)
	}
	return 0, nil
}

func encodeMarshaler(v reflect.Value, buf *bytes.Buffer, _ bool) (int, error) {
	b, err := v.Interface().(Marshaler).MarshalCSV()
	if err != nil {
		return 0, &MarshalerError{Type: v.Type(), MarshalerType: "MarshalCSV", Err: err}
	}
	return buf.Write(b)
}

func encodePtr(typ reflect.Type) (encodeFunc, error) {
	next, err := encodeFn(typ.Elem())
	if err != nil {
		return nil, err
	}
	return func(v reflect.Value, buf *bytes.Buffer, omitempty bool) (int, error) {
		return next(v, buf, omitempty)
	}, nil
}

func encodeBytes(v reflect.Value, buf *bytes.Buffer, _ bool) (int, error) {
	b := v.Bytes()
	w := base64.NewEncoder(base64.StdEncoding, buf)
	w.Write(b)
	w.Close()
	return base64.StdEncoding.EncodedLen(len(b)), nil
}

func encodeFn(typ reflect.Type) (encodeFunc, error) {
	if typ.Implements(csvMarshaler) {
		return encodeMarshaler, nil
	}
	if reflect.PtrTo(typ).Implements(csvMarshaler) {
		return encodePtrMarshaler, nil
	}

	if typ.Implements(textMarshaler) {
		return encodeTextMarshaler, nil
	}
	if reflect.PtrTo(typ).Implements(textMarshaler) {
		return encodePtrTextMarshaler, nil
	}

	switch typ.Kind() {
	case reflect.String:
		return encodeString, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return encodeInt(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return encodeUint(), nil
	case reflect.Float32, reflect.Float64:
		return encodeFloat(typ), nil
	case reflect.Bool:
		return encodeBool(), nil
	case reflect.Interface:
		return encodeInterface, nil
	case reflect.Ptr:
		return encodePtr(typ)
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			return encodeBytes, nil
		}
	}

	return nil, &UnsupportedTypeError{Type: typ}
}
