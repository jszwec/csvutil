package recenc

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"
)

type Embedded1 struct {
	String string  `recenc:"string"`
	Float  float64 `recenc:"float"`
}

type Embedded2 struct {
	Float float64 `recenc:"float"`
	Bool  bool    `recenc:"bool"`
}

type Embedded3 map[string]string

func (e *Embedded3) UnmarshalCSV(s string) error {
	return json.Unmarshal([]byte(s), e)
}

type Embedded4 interface{}

type TypeA struct {
	Embedded1
	String string `recenc:"string"`
	Embedded2
	Int int `recenc:"int"`
}

type TypeB struct {
	Embedded3 `recenc:"json"`
	String    string `recenc:"string"`
}

type TypeC struct {
	*Embedded1
	String string `recenc:"string"`
}

type TypeD struct {
	*Embedded3 `recenc:"json"`
	String     string `recenc:"string"`
}

type TypeE struct {
	String **string `recenc:"string"`
	Int    *int     `recenc:"int"`
}

type TypeF struct {
	Int      int          `recenc:"int"`
	Pint     *int         `recenc:"pint"`
	Int8     int8         `recenc:"int8"`
	Pint8    *int8        `recenc:"pint8"`
	Int16    int16        `recenc:"int16"`
	Pint16   *int16       `recenc:"pint16"`
	Int32    int32        `recenc:"int32"`
	Pint32   *int32       `recenc:"pint32"`
	Int64    int64        `recenc:"int64"`
	Pint64   *int64       `recenc:"pint64"`
	UInt     uint         `recenc:"uint"`
	Puint    *uint        `recenc:"puint"`
	Uint8    uint8        `recenc:"uint8"`
	Puint8   *uint8       `recenc:"puint8"`
	Uint16   uint16       `recenc:"uint16"`
	Puint16  *uint16      `recenc:"puint16"`
	Uint32   uint32       `recenc:"uint32"`
	Puint32  *uint32      `recenc:"puint32"`
	Uint64   uint64       `recenc:"uint64"`
	Puint64  *uint64      `recenc:"puint64"`
	Float32  float32      `recenc:"float32"`
	Pfloat32 *float32     `recenc:"pfloat32"`
	Float64  float64      `recenc:"float64"`
	Pfloat64 *float64     `recenc:"pfloat64"`
	String   string       `recenc:"string"`
	PString  *string      `recenc:"pstring"`
	Bool     bool         `recenc:"bool"`
	Pbool    *bool        `recenc:"pbool"`
	V        interface{}  `recenc:"interface"`
	Pv       *interface{} `recenc:"pinterface"`
}

type TypeG struct {
	String      string
	Int         int
	Float       float64 `recenc:"-"`
	unexported1 int
	unexported2 int `recenc:"unexported2"`
}

type TypeI struct {
	String string `recenc:",omitempty"`
	Int    int    `recenc:"int,omitempty"`
}

type TypeWithInvalidField struct {
	String TypeI `recenc:"string"`
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
			desc: "embedded type - no tag",
			in:   "string,int,float,bool\nstring,5,2.5,t",
			out:  &TypeA{},
			expected: &TypeA{
				Embedded1: Embedded1{Float: 2.5},
				Embedded2: Embedded2{Bool: true},
				String:    "string",
				Int:       5,
			},
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
			desc: "embedded pointer type - no tag ",
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
				"pfloat32,float64,pfloat64,string,pstring,bool,pbool,interface,pinterface\n" +
				"1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,true,true,true,1",
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
			},
			expectedRecord: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12",
				"13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26",
				"true", "true", "true", "1"},
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
			desc: "unsupported type",
			in:   "string,int\ns,1",
			out:  &TypeWithInvalidField{},
			err: &UnsupportedTypeError{
				Type: reflect.TypeOf(TypeI{}),
			},
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
			{nil, "recenc: Decode(nil)"},
			{struct{}{}, "recenc: Decode(non-pointer struct {})"},
			{(*int)(nil), "recenc: Decode(non-struct pointer)"},
			{(*TypeA)(nil), "recenc: Decode(nil *recenc.TypeA)"},
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
}

func BenchmarkDecode(b *testing.B) {
	type A struct {
		A int     `recenc:"a"`
		B float64 `recenc:"b"`
		C string  `recenc:"c"`
		D int64   `recenc:"d"`
		E int8    `recenc:"e"`
		F float32 `recenc:"f"`
		G float32 `recenc:"g"`
		H float32 `recenc:"h"`
		I string  `recenc:"i"`
		J int     `recenc:"j"`
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
