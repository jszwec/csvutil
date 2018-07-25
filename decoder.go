package csvutil

import (
	"reflect"
)

type decField struct {
	columnIndex int
	field
	decodeFunc
	zero interface{}
}

// A Decoder reads and decodes string records into structs.
type Decoder struct {
	// Tag defines which key in the struct field's tag to scan for names and
	// options (Default: 'csv').
	Tag string

	// If not nil, Map is a function that is called for each field in the csv
	// record before decoding the data. It allows mapping certain string values
	// for specific columns or types to a known format. Decoder calls Map with
	// the current column name (taken from header) and a zero non-pointer value
	// of a type to which it is going to decode data into. Implementations
	// should use type assertions to recognize the type.
	//
	// The good example of use case for Map is if NaN values are represented by
	// eg 'n/a' string, implementing a specific Map function for all floats
	// could map 'n/a' back into 'NaN' to allow successful decoding.
	//
	// Use Map with caution. If the requirements of column or type are not met
	// Map should return 'field', since it is the original value that was
	// read from the csv input, this would indicate no change.
	//
	// If struct field is an interface v will be of type string, unless the
	// struct field contains a settable pointer value - then v will be a zero
	// value of that type.
	//
	// Map must be set before the first call to Decode and not changed after it.
	Map func(field, col string, v interface{}) string

	r       Reader
	typeKey typeKey
	hmap    map[string]int
	header  []string
	record  []string
	cache   []decField
	unused  []int
}

// NewDecoder returns a new decoder that reads from r.
//
// Decoder will match struct fields according to the given header.
//
// If header is empty NewDecoder will read one line and treat it as a header.
//
// Records coming from r must be of the same length as the header.
//
// NewDecoder may return io.EOF if there is no data in r and no header was
// provided by the caller.
func NewDecoder(r Reader, header ...string) (dec *Decoder, err error) {
	if len(header) == 0 {
		header, err = r.Read()
		if err != nil {
			return nil, err
		}
	}

	h := make([]string, len(header))
	copy(h, header)
	header = h

	m := make(map[string]int, len(header))
	for i, h := range header {
		m[h] = i
	}

	return &Decoder{
		r:      r,
		header: header,
		hmap:   m,
		unused: make([]int, 0, len(header)),
	}, nil
}

// Decode reads the next string record from its input and stores it in the value
// pointed to by v which must be a non-nil struct pointer.
//
// Decode matches all exported struct fields based on the header. Struct fields
// can be adjusted by using tags.
//
// The "omitempty" option specifies that the field should be omitted from
// the decoding if record's field is an empty string.
//
// Examples of struct field tags and their meanings:
// 	// Decode matches this field with "myName" header column.
// 	Field int `csv:"myName"`
//
// 	// Decode matches this field with "Field" header column.
// 	Field int
//
// 	// Decode matches this field with "myName" header column and decoding is not
//	// called if record's field is an empty string.
// 	Field int `csv:"myName,omitempty"`
//
// 	// Decode matches this field with "Field" header column and decoding is not
//	// called if record's field is an empty string.
// 	Field int `csv:",omitempty"`
//
// 	// Decode ignores this field.
// 	Field int `csv:"-"`
//
// By default decode looks for "csv" tag, but this can be changed by setting
// Decoder.Tag field.
//
// To Decode into a custom type v must implement csvutil.Unmarshaler or
// encoding.TextUnmarshaler.
//
// Anonymous struct fields with tags are treated like normal fields and they
// must implement csvutil.Unmarshaler or encoding.TextUnmarshaler.
//
// Anonymous struct fields without tags are populated just as if they were
// part of the main struct. However, fields in the main struct have bigger
// priority and they are populated first. If main struct and anonymous struct
// field have the same fields, the main struct's fields will be populated.
//
// Fields of type []byte expect the data to be base64 encoded strings.
//
// Float fields are decoded to NaN if a string value is 'NaN'. This check
// is case insensitive.
//
// Interface fields are decoded to strings unless they contain settable pointer
// value.
func (d *Decoder) Decode(v interface{}) (err error) {
	d.record, err = d.r.Read()
	if err != nil {
		return err
	}

	if len(d.record) != len(d.header) {
		return ErrFieldCount
	}

	return d.unmarshal(d.record, v)
}

// Record returns the most recently read record. The slice is valid until the
// next call to Decode.
func (d *Decoder) Record() []string {
	return d.record
}

// Header returns the first line that came from the reader, or returns the
// defined header by the caller.
func (d *Decoder) Header() []string {
	header := make([]string, len(d.header))
	copy(header, d.header)
	return header
}

// Unused returns a list of column indexes that were not used during decoding
// due to lack of matching struct field.
func (d *Decoder) Unused() []int {
	if len(d.unused) == 0 {
		return nil
	}

	indices := make([]int, len(d.unused))
	copy(indices, d.unused)
	return indices
}

func (d *Decoder) unmarshal(record []string, v interface{}) error {
	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Ptr || vv.IsNil() {
		return &InvalidDecodeError{Type: reflect.TypeOf(v)}
	}

	elem := indirect(vv.Elem())
	if typ := elem.Type(); typ.Kind() != reflect.Struct {
		return &InvalidDecodeError{Type: reflect.PtrTo(typ)}
	}

	return d.unmarshalStruct(record, elem)
}

func (d *Decoder) unmarshalStruct(record []string, v reflect.Value) error {
	fields, err := d.fields(typeKey{d.tag(), v.Type()})
	if err != nil {
		return err
	}

	for _, f := range fields {
		if f.tag.omitEmpty && record[f.columnIndex] == "" {
			continue
		}

		fv := v
		for _, i := range f.index {
			fv = fv.Field(i)
			if fv.Kind() == reflect.Ptr {
				if fv.IsNil() {
					// this can happen if a field is an unexported embedded
					// pointer type. In Go prior to 1.10 it was possible to
					// set such value because of a bug in the reflect package
					// https://github.com/golang/go/issues/21353
					if !fv.CanSet() {
						return errPtrUnexportedStruct(fv.Type())
					}
					fv.Set(reflect.New(fv.Type().Elem()))
				}
				fv = fv.Elem()
			}
		}

		s := record[f.columnIndex]
		if d.Map != nil && f.zero != nil {
			zero := f.zero
			if fv.Kind() == reflect.Interface && !fv.IsNil() {
				if v := walkValue(fv); v.CanSet() {
					zero = reflect.Zero(v.Type()).Interface()
				}
			}
			s = d.Map(s, d.header[f.columnIndex], zero)
		}

		if err := f.decodeFunc(s, fv); err != nil {
			return err
		}
	}
	return nil
}

func (d *Decoder) fields(k typeKey) ([]decField, error) {
	if k == d.typeKey {
		return d.cache, nil
	}

	fields := cachedFields(k)
	decFields := make([]decField, 0, len(fields))
	used := make([]bool, len(d.header))

	for _, f := range fields {
		i, ok := d.hmap[f.tag.name]
		if !ok {
			continue
		}

		fn, err := decodeFn(f.typ)
		if err != nil {
			return nil, err
		}

		df := decField{
			columnIndex: i,
			field:       f,
			decodeFunc:  fn,
		}

		if d.Map != nil {
			switch f.typ.Kind() {
			case reflect.Interface:
				df.zero = "" // interface values are decoded to strings
			default:
				df.zero = reflect.Zero(walkType(f.typ)).Interface()
			}
		}

		decFields = append(decFields, df)
		used[i] = true
	}

	d.unused = d.unused[:0]
	for i, b := range used {
		if !b {
			d.unused = append(d.unused, i)
		}
	}

	d.cache, d.typeKey = decFields, k
	return d.cache, nil
}

func (d *Decoder) tag() string {
	if d.Tag == "" {
		return defaultTag
	}
	return d.Tag
}

func indirect(v reflect.Value) reflect.Value {
	for {
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				return v
			}
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() {
				v = e
				continue
			}
			return v
		case reflect.Ptr:
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		default:
			return v
		}
	}
}
