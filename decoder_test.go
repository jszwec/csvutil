package csvutil

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"
)

var Binary = []byte("binary-data")

var EncodedBinary = base64.StdEncoding.EncodeToString(Binary)

var BinaryLarge = bytes.Repeat([]byte("1"), 128*1024)

var EncodedBinaryLarge = base64.StdEncoding.EncodeToString(BinaryLarge)

type Float float64

type Enum uint8

const (
	EnumDefault = iota
	EnumFirst
	EnumSecond
)

func (e Enum) MarshalCSV() ([]byte, error) {
	switch e {
	case EnumFirst:
		return []byte("first"), nil
	case EnumSecond:
		return []byte("second"), nil
	default:
		return []byte("default"), nil
	}
}

func (e *Enum) UnmarshalCSV(data []byte) error {
	s := string(data)
	switch s {
	case "first":
		*e = EnumFirst
	case "second":
		*e = EnumSecond
	default:
		*e = EnumDefault
	}
	return nil
}

type EnumType struct {
	Enum Enum `csv:"enum"`
}

type Embedded1 struct {
	String string  `csv:"string"`
	Float  float64 `csv:"float"`
}

type Embedded2 struct {
	Float float64 `csv:"float"`
	Bool  bool    `csv:"bool"`
}

type Embedded3 map[string]string

func (e *Embedded3) UnmarshalCSV(s []byte) error {
	return json.Unmarshal(s, e)
}

func (e Embedded3) MarshalCSV() ([]byte, error) {
	return json.Marshal(e)
}

type Embedded4 interface{}

type Embedded5 struct {
	Embedded6
	Embedded7
	Embedded8
}

type Embedded6 struct {
	X int
}

type Embedded7 Embedded6

type Embedded8 struct {
	Embedded9
}

type Embedded9 struct {
	X int
	Y int
}

type Embedded10 struct {
	Embedded11
	Embedded12
	Embedded13
}

type Embedded11 struct {
	Embedded6
}

type Embedded12 struct {
	Embedded6
}

type Embedded13 struct {
	Embedded8
}

type TypeA struct {
	Embedded1
	String string `csv:"string"`
	Embedded2
	Int int `csv:"int"`
}

type TypeB struct {
	Embedded3 `csv:"json"`
	String    string `csv:"string"`
}

type TypeC struct {
	*Embedded1
	String string `csv:"string"`
}

type TypeD struct {
	*Embedded3 `csv:"json"`
	String     string `csv:"string"`
}

type TypeE struct {
	String **string `csv:"string"`
	Int    *int     `csv:"int"`
}

type TypeF struct {
	Int      int          `csv:"int"`
	Pint     *int         `csv:"pint"`
	Int8     int8         `csv:"int8"`
	Pint8    *int8        `csv:"pint8"`
	Int16    int16        `csv:"int16"`
	Pint16   *int16       `csv:"pint16"`
	Int32    int32        `csv:"int32"`
	Pint32   *int32       `csv:"pint32"`
	Int64    int64        `csv:"int64"`
	Pint64   *int64       `csv:"pint64"`
	UInt     uint         `csv:"uint"`
	Puint    *uint        `csv:"puint"`
	Uint8    uint8        `csv:"uint8"`
	Puint8   *uint8       `csv:"puint8"`
	Uint16   uint16       `csv:"uint16"`
	Puint16  *uint16      `csv:"puint16"`
	Uint32   uint32       `csv:"uint32"`
	Puint32  *uint32      `csv:"puint32"`
	Uint64   uint64       `csv:"uint64"`
	Puint64  *uint64      `csv:"puint64"`
	Float32  float32      `csv:"float32"`
	Pfloat32 *float32     `csv:"pfloat32"`
	Float64  float64      `csv:"float64"`
	Pfloat64 *float64     `csv:"pfloat64"`
	String   string       `csv:"string"`
	PString  *string      `csv:"pstring"`
	Bool     bool         `csv:"bool"`
	Pbool    *bool        `csv:"pbool"`
	V        interface{}  `csv:"interface"`
	Pv       *interface{} `csv:"pinterface"`
	Binary   []byte       `csv:"binary"`
	PBinary  *[]byte      `csv:"pbinary"`
}

type TypeG struct {
	String      string
	Int         int
	Float       float64 `csv:"-"`
	unexported1 int
	unexported2 int `csv:"unexported2"`
}

type TypeI struct {
	String string `csv:",omitempty"`
	Int    int    `csv:"int,omitempty"`
}

type Unmarshalers struct {
	CSVUnmarshaler      CSVUnmarshaler      `csv:"csv"`
	PCSVUnmarshaler     *CSVUnmarshaler     `csv:"pcsv"`
	TextUnmarshaler     TextUnmarshaler     `csv:"text"`
	PTextUnmarshaler    *TextUnmarshaler    `csv:"ptext"`
	CSVTextUnmarshaler  CSVTextUnmarshaler  `csv:"csv-text"`
	PCSVTextUnmarshaler *CSVTextUnmarshaler `csv:"pcsv-text"`
}

type EmbeddedUnmarshalers struct {
	CSVUnmarshaler     `csv:"csv"`
	TextUnmarshaler    `csv:"text"`
	CSVTextUnmarshaler `csv:"csv-text"`
}

type EmbeddedPtrUnmarshalers struct {
	*CSVUnmarshaler     `csv:"csv"`
	*TextUnmarshaler    `csv:"text"`
	*CSVTextUnmarshaler `csv:"csv-text"`
}

type CSVUnmarshaler struct {
	String string `csv:"string"`
}

func (t *CSVUnmarshaler) UnmarshalCSV(s []byte) error {
	t.String = "unmarshalCSV:" + string(s)
	return nil
}

type TextUnmarshaler struct {
	String string `csv:"string"`
}

func (t *TextUnmarshaler) UnmarshalText(text []byte) error {
	t.String = "unmarshalText:" + string(text)
	return nil
}

type CSVTextUnmarshaler struct {
	String string `csv:"string"`
}

func (t *CSVTextUnmarshaler) UnmarshalCSV(s []byte) error {
	t.String = "unmarshalCSV:" + string(s)
	return nil
}

func (t *CSVTextUnmarshaler) UnmarshalText(text []byte) error {
	t.String = "unmarshalText:" + string(text)
	return nil
}

type TypeWithInvalidField struct {
	String TypeI `csv:"string"`
}

type InvalidType struct {
	String struct{}
}

type TagPriority struct {
	Foo int
	Bar int `csv:"Foo"`
}

type embedded struct {
	Foo int `csv:"foo"`
	bar int `csv:"bar"`
}

type UnexportedEmbedded struct {
	embedded
}

type A struct {
	B
	X int
}

type B struct {
	*A
	Y int
}

var Int = 10
var String = "string"
var PString = &String

func pint(n int) *int                       { return &n }
func pint8(n int8) *int8                    { return &n }
func pint16(n int16) *int16                 { return &n }
func pint32(n int32) *int32                 { return &n }
func pint64(n int64) *int64                 { return &n }
func puint(n uint) *uint                    { return &n }
func puint8(n uint8) *uint8                 { return &n }
func puint16(n uint16) *uint16              { return &n }
func puint32(n uint32) *uint32              { return &n }
func puint64(n uint64) *uint64              { return &n }
func pfloat32(f float32) *float32           { return &f }
func pfloat64(f float64) *float64           { return &f }
func pstring(s string) *string              { return &s }
func pbool(b bool) *bool                    { return &b }
func pinterface(v interface{}) *interface{} { return &v }

func TestDecoder(t *testing.T) {
	fixtures := []struct {
		desc           string
		in             string
		out            interface{}
		expected       interface{}
		expectedRecord []string
		inheader       []string
		header         []string
		unused         []int
		err            error
	}{
		{
			desc: "embedded type - no tag - conflicting float tag",
			in:   "string,int,float,bool\nstring,5,2.5,t",
			out:  &TypeA{},
			expected: &TypeA{
				Embedded1: Embedded1{},
				Embedded2: Embedded2{Bool: true},
				String:    "string",
				Int:       5,
			},
			unused:         []int{2},
			expectedRecord: []string{"string", "5", "2.5", "t"},
			header:         []string{"string", "int", "float", "bool"},
		},
		{
			desc: "embedded type - with tag",
			in: `string,json
string,"{""key"":""value""}"
`,
			out: &TypeB{},
			expected: &TypeB{
				Embedded3: Embedded3{"key": "value"},
				String:    "string",
			},
			expectedRecord: []string{"string", `{"key":"value"}`},
			header:         []string{"string", "json"},
		},
		{
			desc: "embedded pointer type - no tag - type with conflicting tag",
			in:   "string,float\nstring,2.5",
			out:  &TypeC{},
			expected: &TypeC{
				Embedded1: &Embedded1{Float: 2.5},
				String:    "string",
			},
			expectedRecord: []string{"string", "2.5"},
			header:         []string{"string", "float"},
		},
		{
			desc: "embedded pointer type - with tag ",
			in: `string,json
string,"{""key"":""value""}"
`,
			out: &TypeD{},
			expected: &TypeD{
				Embedded3: &Embedded3{"key": "value"},
				String:    "string",
			},
			expectedRecord: []string{"string", `{"key":"value"}`},
			header:         []string{"string", "json"},
		},
		{
			desc: "pointer types",
			in:   "string,int\nstring,10",
			out:  &TypeE{},
			expected: &TypeE{
				String: &PString,
				Int:    &Int,
			},
			expectedRecord: []string{"string", "10"},
			header:         []string{"string", "int"},
		},
		{
			desc: "basic types",
			in: "int,pint,int8,pint8,int16,pint16,int32,pint32,int64,pint64,uint," +
				"puint,uint8,puint8,uint16,puint16,uint32,puint32,uint64,puint64,float32," +
				"pfloat32,float64,pfloat64,string,pstring,bool,pbool,interface,pinterface,binary,pbinary\n" +
				"1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,true,true,true,1," +
				EncodedBinary + "," + EncodedBinaryLarge,
			out: &TypeF{},
			expected: &TypeF{
				Int:      1,
				Pint:     pint(2),
				Int8:     3,
				Pint8:    pint8(4),
				Int16:    5,
				Pint16:   pint16(6),
				Int32:    7,
				Pint32:   pint32(8),
				Int64:    9,
				Pint64:   pint64(10),
				UInt:     11,
				Puint:    puint(12),
				Uint8:    13,
				Puint8:   puint8(14),
				Uint16:   15,
				Puint16:  puint16(16),
				Uint32:   17,
				Puint32:  puint32(18),
				Uint64:   19,
				Puint64:  puint64(20),
				Float32:  21,
				Pfloat32: pfloat32(22),
				Float64:  23,
				Pfloat64: pfloat64(24),
				String:   "25",
				PString:  pstring("26"),
				Bool:     true,
				Pbool:    pbool(true),
				V:        "true",
				Pv:       pinterface("1"),
				Binary:   Binary,
				PBinary:  &BinaryLarge,
			},
			expectedRecord: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12",
				"13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26",
				"true", "true", "true", "1", EncodedBinary, EncodedBinaryLarge},
			header: []string{"int",
				"pint",
				"int8",
				"pint8",
				"int16",
				"pint16",
				"int32",
				"pint32",
				"int64",
				"pint64",
				"uint",
				"puint",
				"uint8",
				"puint8",
				"uint16",
				"puint16",
				"uint32",
				"puint32",
				"uint64",
				"puint64",
				"float32",
				"pfloat32",
				"float64",
				"pfloat64",
				"string",
				"pstring",
				"bool",
				"pbool",
				"interface",
				"pinterface",
				"binary",
				"pbinary",
			},
		},
		{
			desc: "tags and unexported fields",
			in:   "String,int,Float64,unexported1,unexported2\nstring,10,2.5,1,1",
			out:  &TypeG{},
			expected: &TypeG{
				String: "string",
			},
			expectedRecord: []string{"string", "10", "2.5", "1", "1"},
			unused:         []int{1, 2, 3, 4},
			header:         []string{"String", "int", "Float64", "unexported1", "unexported2"},
		},
		{
			desc:           "omitempty tag",
			in:             "String,int\n,",
			out:            &TypeI{},
			expected:       &TypeI{},
			expectedRecord: []string{"", ""},
			header:         []string{"String", "int"},
		},
		{
			desc: "decode unmarshalers",
			in:   "csv,pcsv,text,ptext,csv-text,pcsv-text\nfield,field,field,field,field,field",
			out:  &Unmarshalers{},
			expected: &Unmarshalers{
				CSVUnmarshaler:      CSVUnmarshaler{"unmarshalCSV:field"},
				PCSVUnmarshaler:     &CSVUnmarshaler{"unmarshalCSV:field"},
				TextUnmarshaler:     TextUnmarshaler{"unmarshalText:field"},
				PTextUnmarshaler:    &TextUnmarshaler{"unmarshalText:field"},
				CSVTextUnmarshaler:  CSVTextUnmarshaler{"unmarshalCSV:field"},
				PCSVTextUnmarshaler: &CSVTextUnmarshaler{"unmarshalCSV:field"},
			},
			expectedRecord: []string{"field", "field", "field", "field", "field", "field"},
			header:         []string{"csv", "pcsv", "text", "ptext", "csv-text", "pcsv-text"},
		},
		{
			desc: "decode embedded tagged unmarshalers",
			in:   "csv,text,csv-text\nfield,field,field",
			out:  &EmbeddedUnmarshalers{},
			expected: &EmbeddedUnmarshalers{
				CSVUnmarshaler:     CSVUnmarshaler{"unmarshalCSV:field"},
				TextUnmarshaler:    TextUnmarshaler{"unmarshalText:field"},
				CSVTextUnmarshaler: CSVTextUnmarshaler{"unmarshalCSV:field"},
			},
			expectedRecord: []string{"field", "field", "field"},
			header:         []string{"csv", "text", "csv-text"},
		},
		{
			desc: "decode pointer embedded tagged unmarshalers",
			in:   "csv,text,csv-text\nfield,field,field",
			out:  &EmbeddedPtrUnmarshalers{},
			expected: &EmbeddedPtrUnmarshalers{
				CSVUnmarshaler:     &CSVUnmarshaler{"unmarshalCSV:field"},
				TextUnmarshaler:    &TextUnmarshaler{"unmarshalText:field"},
				CSVTextUnmarshaler: &CSVTextUnmarshaler{"unmarshalCSV:field"},
			},
			expectedRecord: []string{"field", "field", "field"},
			header:         []string{"csv", "text", "csv-text"},
		},
		{
			desc: "custom header",
			in:   "string,10",
			out:  &TypeI{},
			expected: &TypeI{
				String: "string",
				Int:    10,
			},
			expectedRecord: []string{"string", "10"},
			inheader:       []string{"String", "int"},
			header:         []string{"String", "int"},
		},
		{
			desc: "tag priority over field",
			in:   "Foo\n1",
			out:  &TagPriority{},
			expected: &TagPriority{
				Foo: 0,
				Bar: 1,
			},
			expectedRecord: []string{"1"},
			header:         []string{"Foo"},
		},
		{
			desc: "decode into unexported embedded field",
			in:   "foo,bar\n1,1",
			out:  &UnexportedEmbedded{},
			expected: &UnexportedEmbedded{
				embedded{
					Foo: 1,
					bar: 0,
				},
			},
			expectedRecord: []string{"1", "1"},
			header:         []string{"foo", "bar"},
			unused:         []int{1},
		},
		{
			desc: "embedded field conflict #1",
			in:   "X,Y\n1,2",
			out:  &Embedded5{},
			expected: &Embedded5{
				Embedded8: Embedded8{
					Embedded9: Embedded9{Y: 2},
				},
			},
			expectedRecord: []string{"1", "2"},
			header:         []string{"X", "Y"},
			unused:         []int{0},
		},
		{
			desc: "embedded field conflict #2",
			in:   "X,Y\n1,2",
			out:  &Embedded10{},
			expected: &Embedded10{
				Embedded13: Embedded13{
					Embedded8: Embedded8{
						Embedded9: Embedded9{Y: 2},
					},
				},
			},
			expectedRecord: []string{"1", "2"},
			header:         []string{"X", "Y"},
			unused:         []int{0},
		},
		{
			desc:           "circular reference",
			in:             "X,Y\n1,2",
			out:            &A{},
			expected:       &A{X: 1, B: B{Y: 2}},
			expectedRecord: []string{"1", "2"},
			header:         []string{"X", "Y"},
		},
		{
			desc:           "primitive type alias with Unmarshaler",
			in:             "enum\nfirst",
			out:            &EnumType{},
			expected:       &EnumType{Enum: EnumFirst},
			expectedRecord: []string{"first"},
			header:         []string{"enum"},
		},
		{
			desc:           "alias type",
			in:             "Float\n3.14",
			out:            &struct{ Float float64 }{},
			expected:       &struct{ Float float64 }{3.14},
			expectedRecord: []string{"3.14"},
			header:         []string{"Float"},
		},
		{
			desc: "unsupported type",
			in:   "string,int\ns,1",
			out:  &TypeWithInvalidField{},
			err: &UnsupportedTypeError{
				Type: reflect.TypeOf(TypeI{}),
			},
		},
		{
			desc: "invalid int",
			in:   "Int,Foo\n,",
			out:  &struct{ Int int }{},
			err:  &UnmarshalTypeError{Value: "", Type: reflect.TypeOf(int(0))},
		},
		{
			desc: "invalid int pointer",
			in:   "Int,Foo\n,",
			out:  &struct{ Int *int }{},
			err:  &UnmarshalTypeError{Value: "", Type: reflect.TypeOf(int(0))},
		},
		{
			desc: "invalid type pointer",
			in:   "Int,Foo\n,",
			out:  &struct{ Int *struct{} }{},
			err:  &UnsupportedTypeError{Type: reflect.TypeOf(struct{}{})},
		},
		{
			desc: "invalid uint",
			in:   "Uint,Foo\n,",
			out:  &struct{ Uint uint }{},
			err:  &UnmarshalTypeError{Value: "", Type: reflect.TypeOf(uint(0))},
		},
		{
			desc: "invalid float",
			in:   "Float,Foo\n,",
			out:  &struct{ Float float64 }{},
			err:  &UnmarshalTypeError{Value: "", Type: reflect.TypeOf(float64(0))},
		},
		{
			desc: "invalid bool",
			in:   "Bool,Foo\n,",
			out:  &struct{ Bool bool }{},
			err:  &UnmarshalTypeError{Value: "", Type: reflect.TypeOf(bool(false))},
		},
		{
			desc: "invalid interface",
			in:   "Interface,Foo\n,",
			out:  &struct{ Interface Unmarshaler }{},
			err:  &UnmarshalTypeError{Value: "", Type: csvUnmarshaler},
		},
		{
			desc: "invalid interface pointer",
			in:   "Interface,Foo\n,",
			out:  &struct{ Interface *Unmarshaler }{},
			err:  &UnmarshalTypeError{Value: "", Type: csvUnmarshaler},
		},
		{
			desc: "invalid field in embedded type",
			in:   "String,int\n1,1",
			out:  &struct{ InvalidType }{},
			err:  &UnsupportedTypeError{Type: reflect.TypeOf(struct{}{})},
		},
		{
			desc: "not a struct in decode",
			in:   "string,int\n1,1",
			out:  &Int,
			err:  &InvalidDecodeError{Type: reflect.TypeOf(&Int)},
		},
	}

	for _, f := range fixtures {
		t.Run(f.desc, func(t *testing.T) {
			r, err := NewDecoder(csv.NewReader(strings.NewReader(f.in)), f.inheader...)
			if err != nil {
				t.Fatal(err)
			}

			err = r.Decode(&f.out)
			if f.err != nil {
				if !reflect.DeepEqual(f.err, err) {
					t.Errorf("want err=%v; got %v", f.err, err)
				}
				return
			}

			if err != nil {
				t.Errorf("want err=nil; got %v", err)
			}

			if !reflect.DeepEqual(r.Record(), f.expectedRecord) {
				t.Errorf("want rec=%q; got %q", f.expectedRecord, r.Record())
			}

			if !reflect.DeepEqual(f.out, f.expected) {
				t.Errorf("want %#v; got %#v", f.expected, f.out)
			}

			if !reflect.DeepEqual(r.Unused(), f.unused) {
				t.Errorf("want unused=%v; got %v", f.unused, r.Unused())
			}

			if !reflect.DeepEqual(r.Header(), f.header) {
				t.Errorf("want header=%v; got %v", f.header, r.Header())
			}
		})
	}

	t.Run("decode with custom tag", func(t *testing.T) {
		type Type struct {
			String string `customtag:"string"`
			Int    int    `customtag:"int"`
		}

		dec, err := NewDecoder(NewReader([]string{"string", "10"}), "string", "int")
		if err != nil {
			t.Fatal(err)
		}
		dec.Tag = "customtag"

		var tt Type
		if err := dec.Decode(&tt); err != nil {
			t.Errorf("want err=nil; got %v", err)
		}

		expected := Type{"string", 10}
		if !reflect.DeepEqual(tt, expected) {
			t.Errorf("want tt=%v; got %v", expected, tt)
		}
	})

	t.Run("invalid unmarshal tests", func(t *testing.T) {
		var fixtures = []struct {
			v        interface{}
			expected string
		}{
			{nil, "csvutil: Decode(nil)"},
			{struct{}{}, "csvutil: Decode(non-pointer struct {})"},
			{(*int)(nil), "csvutil: Decode(non-struct pointer)"},
			{(*TypeA)(nil), "csvutil: Decode(nil *csvutil.TypeA)"},
		}

		for _, f := range fixtures {
			r, err := NewDecoder(csv.NewReader(strings.NewReader("string\ns")))
			if err != nil {
				t.Fatal(err)
			}
			err = r.Decode(f.v)
			if err == nil {
				t.Errorf("Decode expecting error, got nil")
				continue
			}
			if got := err.Error(); got != f.expected {
				t.Errorf("want Decode=%q; got %q", got, f.expected)
			}
		}
	})

	t.Run("header and field length mismatch", func(t *testing.T) {
		type Foo struct {
			Col1 string `csv:"col1"`
			Col2 string `csv:"col2"`
		}
		data := []byte("1,1,1")
		r, err := NewDecoder(csv.NewReader(bytes.NewReader(data)), "col1", "col2")
		if err != nil {
			t.Fatal(err)
		}

		var foo Foo
		if err := r.Decode(&foo); err != ErrFieldCount {
			t.Errorf("want err=%v; got %v", ErrFieldCount, err)
		}
	})
}

func BenchmarkDecode(b *testing.B) {
	type A struct {
		A int     `csv:"a"`
		B float64 `csv:"b"`
		C string  `csv:"c"`
		D int64   `csv:"d"`
		E int8    `csv:"e"`
		F float32 `csv:"f"`
		G float32 `csv:"g"`
		H float32 `csv:"h"`
		I string  `csv:"i"`
		J int     `csv:"j"`
	}

	header := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	record := []string{"1", "2.5", "xD", "6", "7", "8", "9", "10", "lol", "10"}

	fixtures := []struct {
		desc string
		len  int
	}{
		{"10 field struct 1 record", 1},
		{"10 field struct 10 records", 10},
		{"10 field struct 100 records", 100},
		{"10 field struct 1000 records", 1000},
		{"10 field struct 10000 records", 10000},
	}

	for _, f := range fixtures {
		var records [][]string
		for i := 0; i < f.len; i++ {
			records = append(records, record)
		}

		b.Run(f.desc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				dec, err := NewDecoder(NewReader(records...), header...)
				if err != nil {
					b.Fatal(err)
				}
				var a A
				b.StartTimer()

				for {
					if err := dec.Decode(&a); err == io.EOF {
						break
					} else if err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}

	b.Run("10 field struct first decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			dec, err := NewDecoder(NewReader(record), header...)
			if err != nil {
				b.Fatal(err)
			}

			var a A
			b.StartTimer()

			if err := dec.Decode(&a); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("10 field struct second decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			r, err := NewDecoder(NewReader(record, record), header...)
			if err != nil {
				b.Fatal(err)
			}

			var a A
			if err := r.Decode(&a); err != nil {
				b.Fatal(err)
			}
			a = A{}
			b.StartTimer()

			if err := r.Decode(&a); err != nil {
				b.Fatal(err)
			}
		}
	})
}

type reader struct {
	records [][]string
	i       int
}

func NewReader(records ...[]string) Reader {
	return &reader{records, 0}
}

func (r *reader) Read() ([]string, error) {
	if r.i >= len(r.records) {
		return nil, io.EOF
	}
	r.i++
	return r.records[r.i-1], nil
}
