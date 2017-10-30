package csvutil

import (
	"fmt"
	"reflect"
)

type typeCache struct {
	fieldIndex  int
	columnIndex int
	embedded    bool
	tag         tag
	decodeFunc
}

// A Decoder reads and decodes string records into structs.
type Decoder struct {
	Tag string

	r      Reader
	header []string
	record []string
	hmap   map[string]int
	cache  map[reflect.Type][]typeCache
	used   []bool
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

	m := make(map[string]int, len(header))
	for i, h := range header {
		m[h] = i
	}

	return &Decoder{
		r:      r,
		header: header,
		hmap:   m,
		cache:  make(map[reflect.Type][]typeCache),
		used:   make([]bool, len(header)),
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

// Header returns the first line that came from the reader.
func (d *Decoder) Header() []string {
	header := make([]string, len(d.hmap))
	copy(header, d.header)
	return header
}

// Unused returns a list of column indexes that were not used during decoding
// due to lack of matching struct field.
func (d *Decoder) Unused() (indexes []int) {
	for i, b := range d.used {
		if !b {
			indexes = append(indexes, i)
		}
	}
	return
}

func (d *Decoder) unmarshal(record []string, v interface{}) error {
	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Ptr || vv.IsNil() {
		return &InvalidDecodeError{Type: reflect.TypeOf(v)}
	}

	elem := indirect(vv.Elem())
	typ := elem.Type()

	if typ.Kind() != reflect.Struct {
		return &InvalidDecodeError{Type: reflect.PtrTo(typ)}
	}

	for i := range d.used {
		d.used[i] = false
	}

	return d.unmarshalStruct(record, elem, typ, d.used)
}

func (d *Decoder) unmarshalStruct(record []string, v reflect.Value, typ reflect.Type, used []bool) error {
	if _, ok := d.cache[typ]; !ok {
		hmap := make(map[string]int, len(d.hmap))
		for k, v := range d.hmap {
			hmap[k] = v
		}
		if err := d.scanStruct(v, typ, hmap); err != nil {
			return err
		}
	}

	for _, c := range d.cache[typ] {
		s := record[c.columnIndex]
		if c.tag.omitEmpty && s == "" {
			used[c.columnIndex] = true
			continue
		}

		field := v.Field(c.fieldIndex)

		if c.embedded {
			field = indirect(field)
			if err := d.unmarshalStruct(record, field, field.Type(), used); err != nil {
				return err
			}
			continue
		}

		used[c.columnIndex] = true

		if err := c.decodeFunc(s, field); err != nil {
			return err
		}

	}
	return nil
}

func (d *Decoder) scanStruct(v reflect.Value, typ reflect.Type, hmap map[string]int) error {
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("unsupported type %v", typ)
	}

	var anonymous []typeCache
	numField := v.NumField()
	cs := make([]typeCache, 0, numField)

	for i := 0; i < numField; i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		typeField := typ.Field(i)
		tag := parseTag(d.tag(), typeField)
		if tag.ignore {
			continue
		}

		if typeField.Anonymous && tag.empty {
			anonymous = append(anonymous, typeCache{
				fieldIndex: i,
				embedded:   true,
			})
			continue
		}

		index, ok := hmap[tag.name]
		if !ok {
			continue
		}
		delete(hmap, tag.name)

		decode, err := decodeFn(field.Type())
		if err != nil {
			return err
		}

		cs = append(cs, typeCache{
			fieldIndex:  i,
			columnIndex: index,
			tag:         tag,
			decodeFunc:  decode,
		})
	}

	for _, a := range anonymous {
		field := indirect(v.Field(a.fieldIndex))
		if err := d.scanStruct(field, field.Type(), hmap); err != nil {
			return err
		}
	}
	d.cache[typ] = append(cs, anonymous...)
	return nil
}

func (d *Decoder) tag() string {
	if d.Tag == "" {
		return "csv"
	}
	return d.Tag
}

func indirect(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() {
				v = e
				continue
			}
		}
		if v.Kind() != reflect.Ptr {
			break
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}
