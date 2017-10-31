package csvutil

import (
	"bytes"
	"io"
	"reflect"
)

// Unmarshal parses the CSV-encoded data and stores the result in the slice
// pointed to by v. If v is nil or not a pointer to a slice, Unmarshal returns
// an InvalidUnmarshalError.
//
// Unmarshal uses the std encoding/csv.Reader for parsing and csvutil.Decoder
// for populating the struct elements in the provided slice. For exact decoding
// rules look at the Decoder's documentation.
//
// The first line in data is treated as a header. Decoder will use it to map
// csv columns to struct's fields.
//
// In case of success the provided slice will be reinitialized and its content
// fully replaced with decoded data.
func Unmarshal(data []byte, v interface{}) error {
	vv := reflect.ValueOf(v)

	if vv.Kind() != reflect.Ptr || vv.IsNil() {
		return &InvalidUnmarshalError{Type: reflect.TypeOf(v)}
	}

	if vv.Type().Elem().Kind() != reflect.Slice {
		return &InvalidUnmarshalError{Type: vv.Type()}
	}

	typ := vv.Type().Elem()
	slice := reflect.MakeSlice(typ, 0, bytes.Count(data, []byte("\n")))

	dec, err := NewDecoder(newCSVReader(bytes.NewReader(data)))
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}

	for {
		elem := reflect.New(typ.Elem())
		if err := dec.Decode(elem.Interface()); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		slice = reflect.Append(slice, elem.Elem())
	}

	vv.Elem().Set(slice)
	return nil
}
