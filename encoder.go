package csvutil

import (
	"bytes"
	"reflect"
)

type encField struct {
	field
	encodeFunc
}

type encCache struct {
	types  map[typeKey][]encField
	buf    bytes.Buffer
	index  []int
	record []string
}

func (c *encCache) fields(k typeKey) ([]encField, error) {
	encFields, ok := c.types[k]
	if !ok {
		fields := cachedFields(k)
		encFields = make([]encField, len(fields))

		for i, f := range fields {
			fn, err := encodeFn(f.typ)
			if err != nil {
				return nil, err
			}

			encFields[i] = encField{
				field:      f,
				encodeFunc: fn,
			}
		}
		c.types[k] = encFields
	}
	return encFields, nil
}

func (c *encCache) reset(fieldsLen int) {
	c.buf.Reset()

	if fieldsLen != len(c.index) {
		c.index = make([]int, fieldsLen)
		c.record = make([]string, fieldsLen)
		return
	}

	for i := range c.index {
		c.index[i] = 0
		c.record[i] = ""
	}
}

// Encoder writes structs CSV representations to the output stream.
type Encoder struct {
	// Tag defines which key in the struct field's tag to scan for names and
	// options (Default: 'csv').
	Tag string

	w        Writer
	cache    encCache
	noHeader bool
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w Writer) *Encoder {
	return &Encoder{
		w:        w,
		noHeader: true,
		cache:    encCache{types: make(map[typeKey][]encField)},
	}
}

// Encode writes the CSV encoding of v to the output stream. The provided
// argument v must be a struct.
//
// Only the exported fields will be encoded.
//
// First call to Encode will write a header. Header names can be customized by
// using tags ('csv' by default), otherwise original Field names are used.
//
// Header and fields are written in the same order as struct fields are defined.
// Embedded struct's fields are treated as if they were part of the outer struct.
// Fields that are embedded types and that are tagged are treated like any
// other field, but they have to implement Marshaler or encoding.TextMarshaler
// interfaces.
//
// Marshaler interface has the priority over encoding.TextMarshaler.
//
// Tagged fields have the priority over non tagged fields with the same name.
//
// Following the Go vibility rules if there are multiple fields with the same
// name (tagged or not tagged) on the same level and choice between them is
// ambiguous, then all these fields will be ignored.
//
// Nil values will be encoded as empty strings. Same will happen if 'omitempty'
// tag is set, and the value is a default value like 0, false or nil interface.
//
// Bool types are encoded as 'true' or 'false'.
//
// Float types are encoded using strconv.FormatFloat with precision -1 and 'G'
// format.
//
// Fields of type []byte are being encoded as base64-encoded strings.
//
// Fields can be excluded from encoding by using '-' tag option.
//
// Examples of struct tags:
//
// 	// Field appears as 'myName' header in CSV encoding.
// 	Field int `csv:"myName"`
//
// 	// Field appears as 'Field' header in CSV encoding.
// 	Field int
//
// 	// Field appears as 'myName' header in CSV encoding and is an empty string
//	// if Field is 0.
// 	Field int `csv:"myName,omitempty"`
//
// 	// Field appears as 'Field' header in CSV encoding and is an empty string
//	// if Field is 0.
// 	Field int `csv:",omitempty"`
//
// 	// Encode ignores this field.
// 	Field int `csv:"-"`
//
// Encode doesn't flush data. The caller is responsible for calling Flush() if
// the used Writer supports it.
func (e *Encoder) Encode(v interface{}) error {
	return e.encode(reflect.ValueOf(v))
}

func (e *Encoder) encode(v reflect.Value) error {
	v = indirect(v)
	if v.Kind() != reflect.Struct {
		return &InvalidEncodeError{v.Type()}
	}

	if e.noHeader {
		k := typeKey{e.tag(), v.Type()}
		fields, err := e.cache.fields(k)
		if err != nil {
			return err
		}

		if err := e.encodeHeader(fields); err != nil {
			return err
		}
		e.noHeader = false
	}

	return e.marshal(v)
}

func (e *Encoder) encodeHeader(fields []encField) error {
	e.cache.reset(len(fields))
	for i, f := range fields {
		e.cache.record[i] = f.tag.name
	}
	return e.w.Write(e.cache.record)
}

func (e *Encoder) marshal(v reflect.Value) error {
	k := typeKey{e.tag(), v.Type()}

	fields, err := e.cache.fields(k)
	if err != nil {
		return err
	}

	e.cache.reset(len(fields))
	buf, index, record := &e.cache.buf, e.cache.index, e.cache.record

	for i, f := range fields {
		v, ok := walkIndex(v, f.index)
		if !ok {
			continue
		}

		n, err := f.encodeFunc(v, buf, f.tag.omitEmpty)
		if err != nil {
			return err
		}
		index[i] = n
	}

	out := buf.String()
	for i, n := range index {
		record[i], out = out[:n], out[n:]
	}

	return e.w.Write(record)
}

func (e *Encoder) tag() string {
	if e.Tag == "" {
		return defaultTag
	}
	return e.Tag
}

func walkIndex(v reflect.Value, index []int) (reflect.Value, bool) {
	for _, i := range index {
		v = walkPtr(v)
		if !v.IsValid() {
			return reflect.Value{}, false
		}
		v = v.Field(i)
	}

	v = walkPtr(v)
	return v, v.IsValid()
}

func walkPtr(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}
