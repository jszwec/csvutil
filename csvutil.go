package csvutil

import (
	"bytes"
	"encoding/csv"
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

	c := countRecords(data)
	slice := reflect.MakeSlice(typ, c, c)

	dec, err := NewDecoder(newCSVReader(bytes.NewReader(data)))
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}

	var i int
	for ; ; i++ {
		// just in case countRecords counts it wrong.
		if i >= c && i >= slice.Len() {
			slice = reflect.Append(slice, reflect.New(typ.Elem()).Elem())
		}

		if err := dec.Decode(slice.Index(i).Addr().Interface()); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
	}

	vv.Elem().Set(slice.Slice3(0, i, i))
	return nil
}

// Marshal returns the CSV encoding of slice v. If v is not a slice, Marshal
// returns InvalidMarshalError. If slice elements are not structs, Marshal will
// return InvalidEncodeError.
//
// Marshal uses the std encoding/csv.Writer with its default settings for csv
// encoding.
//
// For the exact encoding rules look at Encoder.Encode method.
func Marshal(v interface{}) ([]byte, error) {
	val := walkPtr(reflect.ValueOf(v))
	if val.Kind() != reflect.Slice {
		return nil, &InvalidMarshalError{Type: val.Type()}
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	enc := NewEncoder(w)

	l := val.Len()
	for i := 0; i < l; i++ {
		if err := enc.Encode(val.Index(i).Interface()); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func countRecords(s []byte) (n int) {
	var prev byte
	inQuote := false
	for {
		if len(s) == 0 && prev != '"' {
			return n
		}

		i := bytes.IndexAny(s, "\n\"")
		if i == -1 {
			return n + 1
		}

		switch s[i] {
		case '\n':
			if !inQuote && (i > 0 || prev == '"') {
				n++
			}
		case '"':
			inQuote = !inQuote
		}

		prev = s[i]
		s = s[i+1:]
	}
}

func newCSVReader(r io.Reader) *csv.Reader {
	rr := csv.NewReader(r)
	rr.ReuseRecord = true
	return rr
}
