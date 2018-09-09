package csvutil

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"testing"
)

var Error = errors.New("error")

var nilIface interface{}

var nilPtr *TypeF

var nilIfacePtr interface{} = nilPtr

type embeddedMap map[string]string

type Embedded14 Embedded3

func (e *Embedded14) MarshalCSV() ([]byte, error) {
	return json.Marshal(e)
}

type Embedded15 Embedded3

func (e *Embedded15) MarshalText() ([]byte, error) {
	return json.Marshal(Embedded3(*e))
}

type CSVMarshaler struct {
	Err error
}

func (m CSVMarshaler) MarshalCSV() ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return []byte("csvmarshaler"), nil
}

type TextMarshaler struct {
	Err error
}

func (m TextMarshaler) MarshalText() ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return []byte("textmarshaler"), nil
}

type CSVTextMarshaler struct {
	CSVMarshaler
	TextMarshaler
}

type TypeH struct {
	Int     int         `csv:"int,omitempty"`
	Int8    int8        `csv:"int8,omitempty"`
	Int16   int16       `csv:"int16,omitempty"`
	Int32   int32       `csv:"int32,omitempty"`
	Int64   int64       `csv:"int64,omitempty"`
	UInt    uint        `csv:"uint,omitempty"`
	Uint8   uint8       `csv:"uint8,omitempty"`
	Uint16  uint16      `csv:"uint16,omitempty"`
	Uint32  uint32      `csv:"uint32,omitempty"`
	Uint64  uint64      `csv:"uint64,omitempty"`
	Float32 float32     `csv:"float32,omitempty"`
	Float64 float64     `csv:"float64,omitempty"`
	String  string      `csv:"string,omitempty"`
	Bool    bool        `csv:"bool,omitempty"`
	V       interface{} `csv:"interface,omitempty"`
}

func TestEncoder(t *testing.T) {
	fixtures := []struct {
		desc string
		in   []interface{}
		out  [][]string
		err  error
	}{
		{
			desc: "test all types",
			in: []interface{}{
				TypeF{
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
				TypeF{},
			},
			out: [][]string{
				{
					"int", "pint", "int8", "pint8", "int16", "pint16", "int32",
					"pint32", "int64", "pint64", "uint", "puint", "uint8", "puint8",
					"uint16", "puint16", "uint32", "puint32", "uint64", "puint64",
					"float32", "pfloat32", "float64", "pfloat64", "string", "pstring",
					"bool", "pbool", "interface", "pinterface", "binary", "pbinary",
				},
				{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11",
					"12", "13", "14", "15", "16", "17", "18", "19", "20", "21",
					"22", "23", "24", "25", "26", "true", "true", "true", "1",
					EncodedBinary, EncodedBinaryLarge,
				},
				{"0", "", "0", "", "0", "", "0", "", "0", "", "0", "",
					"0", "", "0", "", "0", "", "0", "", "0", "", "0", "", "", "",
					"false", "", "", "", "", "",
				},
			},
		},
		{
			desc: "tags and unexported fields",
			in: []interface{}{
				TypeG{
					String:      "string",
					Int:         1,
					Float:       3.14,
					unexported1: 100,
					unexported2: 200,
				},
			},
			out: [][]string{
				{"String", "Int"},
				{"string", "1"},
			},
		},
		{
			desc: "omitempty tags",
			in: []interface{}{
				TypeH{},
			},
			out: [][]string{
				{"int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16",
					"uint32", "uint64", "float32", "float64", "string", "bool", "interface",
				},
				{"", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},
			},
		},
		{
			desc: "embedded types #1",
			in: []interface{}{
				TypeA{
					Embedded1: Embedded1{
						String: "string1",
						Float:  1,
					},
					String: "string",
					Embedded2: Embedded2{
						Float: 2,
						Bool:  true,
					},
					Int: 10,
				},
			},
			out: [][]string{
				{"string", "bool", "int"},
				{"string", "true", "10"},
			},
		},
		{
			desc: "embedded non struct tagged types",
			in: []interface{}{
				TypeB{
					Embedded3: Embedded3{"key": "val"},
					String:    "string1",
				},
			},
			out: [][]string{
				{"json", "string"},
				{`{"key":"val"}`, "string1"},
			},
		},
		{
			desc: "embedded non struct tagged types with pointer receiver MarshalCSV",
			in: []interface{}{
				&struct {
					Embedded14 `csv:"json"`
					A          Embedded14 `csv:"json2"`
				}{
					Embedded14: Embedded14{"key": "val"},
					A:          Embedded14{"key1": "val1"},
				},
				struct {
					Embedded14 `csv:"json"`
					A          Embedded14 `csv:"json2"`
				}{
					Embedded14: Embedded14{"key": "val"},
					A:          Embedded14{"key1": "val1"},
				},
				struct {
					*Embedded14 `csv:"json"`
					A           *Embedded14 `csv:"json2"`
				}{
					Embedded14: &Embedded14{"key": "val"},
					A:          &Embedded14{"key1": "val1"},
				},
			},
			out: [][]string{
				{"json", "json2"},
				{`{"key":"val"}`, `{"key1":"val1"}`},
				{``, ``},
				{`{"key":"val"}`, `{"key1":"val1"}`},
			},
		},
		{
			desc: "embedded non struct tagged types with pointer receiver MarshalText",
			in: []interface{}{
				&struct {
					Embedded15 `csv:"json"`
					A          Embedded15 `csv:"json2"`
				}{
					Embedded15: Embedded15{"key": "val"},
					A:          Embedded15{"key1": "val1"},
				},
				struct {
					Embedded15 `csv:"json"`
					A          Embedded15 `csv:"json2"`
				}{
					Embedded15: Embedded15{"key": "val"},
					A:          Embedded15{"key1": "val1"},
				},
				struct {
					*Embedded15 `csv:"json"`
					A           *Embedded15 `csv:"json2"`
				}{
					Embedded15: &Embedded15{"key": "val"},
					A:          &Embedded15{"key1": "val1"},
				},
			},
			out: [][]string{
				{"json", "json2"},
				{`{"key":"val"}`, `{"key1":"val1"}`},
				{``, ``},
				{`{"key":"val"}`, `{"key1":"val1"}`},
			},
		},
		{
			desc: "embedded pointer types",
			in: []interface{}{
				TypeC{
					Embedded1: &Embedded1{
						String: "string2",
						Float:  1,
					},
					String: "string1",
				},
			},
			out: [][]string{
				{"float", "string"},
				{`1`, "string1"},
			},
		},
		{
			desc: "embedded pointer types with nil values",
			in: []interface{}{
				TypeC{
					Embedded1: nil,
					String:    "string1",
				},
			},
			out: [][]string{
				{"float", "string"},
				{``, "string1"},
			},
		},
		{
			desc: "embedded non struct tagged pointer types",
			in: []interface{}{
				TypeD{
					Embedded3: &Embedded3{"key": "val"},
					String:    "string1",
				},
			},
			out: [][]string{
				{"json", "string"},
				{`{"key":"val"}`, "string1"},
			},
		},
		{
			desc: "embedded non struct tagged pointer types with nil value",
			in: []interface{}{
				TypeD{
					Embedded3: nil,
					String:    "string1",
				},
			},
			out: [][]string{
				{"json", "string"},
				{"", "string1"},
			},
		},
		{
			desc: "tagged fields priority",
			in: []interface{}{
				TagPriority{Foo: 1, Bar: 2},
			},
			out: [][]string{
				{"Foo"},
				{"2"},
			},
		},
		{
			desc: "conflicting embedded fields #1",
			in: []interface{}{
				Embedded5{
					Embedded6: Embedded6{X: 60},
					Embedded7: Embedded7{X: 70},
					Embedded8: Embedded8{
						Embedded9: Embedded9{
							X: 90,
							Y: 91,
						},
					},
				},
			},
			out: [][]string{
				{"Y"},
				{"91"},
			},
		},
		{
			desc: "conflicting embedded fields #2",
			in: []interface{}{
				Embedded10{
					Embedded11: Embedded11{
						Embedded6: Embedded6{X: 60},
					},
					Embedded12: Embedded12{
						Embedded6: Embedded6{X: 60},
					},
					Embedded13: Embedded13{
						Embedded8: Embedded8{
							Embedded9: Embedded9{
								X: 90,
								Y: 91,
							},
						},
					},
				},
			},
			out: [][]string{
				{"Y"},
				{"91"},
			},
		},
		{
			desc: "double pointer",
			in: []interface{}{
				TypeE{
					String: &PString,
					Int:    &Int,
				},
			},
			out: [][]string{
				{"string", "int"},
				{"string", "10"},
			},
		},
		{
			desc: "nil double pointer",
			in: []interface{}{
				TypeE{},
			},
			out: [][]string{
				{"string", "int"},
				{"", ""},
			},
		},
		{
			desc: "unexported non-struct embedded",
			in: []interface{}{
				struct {
					A int
					embeddedMap
				}{1, make(embeddedMap)},
			},
			out: [][]string{
				{"A"},
				{"1"},
			},
		},
		{
			desc: "cyclic reference",
			in: []interface{}{
				A{
					B: B{Y: 2, A: &A{}},
					X: 1,
				},
			},
			out: [][]string{
				{"Y", "X"},
				{"2", "1"},
			},
		},
		{
			desc: "text marshaler",
			in: []interface{}{
				struct {
					A CSVMarshaler
				}{},
				struct {
					A TextMarshaler
				}{},
				struct {
					A struct {
						TextMarshaler
						CSVMarshaler
					}
				}{},
			},
			out: [][]string{
				{"A"},
				{"csvmarshaler"},
				{"textmarshaler"},
				{"csvmarshaler"},
			},
		},
		{
			desc: "primitive type alias implementing Marshaler",
			in: []interface{}{
				EnumType{Enum: EnumFirst},
				EnumType{Enum: EnumSecond},
			},
			out: [][]string{
				{"enum"},
				{"first"},
				{"second"},
			},
		},
		{
			desc: "aliased type",
			in: []interface{}{
				struct{ Float float64 }{3.14},
			},
			out: [][]string{
				{"Float"},
				{"3.14"},
			},
		},
		{
			desc: "embedded tagged marshalers",
			in: []interface{}{
				struct {
					CSVMarshaler  `csv:"csv"`
					TextMarshaler `csv:"text"`
				}{},
			},
			out: [][]string{
				{"csv", "text"},
				{"csvmarshaler", "textmarshaler"},
			},
		},
		{
			desc: "embedded pointer tagged marshalers",
			in: []interface{}{
				struct {
					*CSVMarshaler  `csv:"csv"`
					*TextMarshaler `csv:"text"`
				}{&CSVMarshaler{}, &TextMarshaler{}},
			},
			out: [][]string{
				{"csv", "text"},
				{"csvmarshaler", "textmarshaler"},
			},
		},
		{
			desc: "encode different types",
			// This doesnt mean the output csv is valid. Generally this is an invalid
			// use. However, we need to make sure that the encoder is doing what it is
			// asked to... correctly.
			in: []interface{}{
				struct {
					A int
				}{},
				struct {
					A int
					B string
				}{},
				struct {
					A int
				}{},
				struct{}{},
			},
			out: [][]string{
				{"A"},
				{"0"},
				{"0", ""},
				{"0"},
				{},
			},
		},
		{
			desc: "encode interface values",
			in: []interface{}{
				struct {
					V interface{}
				}{1},
				struct {
					V interface{}
				}{pint(10)},
				struct {
					V interface{}
				}{ppint(100)},
				struct {
					V interface{}
				}{pppint(1000)},
				struct {
					V *interface{}
				}{pinterface(ppint(10000))},
				struct {
					V *interface{}
				}{func() *interface{} {
					var v interface{} = pppint(100000)
					var vv interface{} = v
					return &vv
				}()},
				struct {
					V interface{}
				}{func() interface{} {
					var v interface{} = &CSVMarshaler{}
					var vv interface{} = v
					return &vv
				}()},
				struct {
					V interface{}
				}{func() interface{} {
					var v interface{} = CSVMarshaler{}
					var vv interface{} = v
					return &vv
				}()},
				struct {
					V interface{}
				}{func() interface{} {
					var v interface{} = &CSVMarshaler{}
					var vv interface{} = v
					return vv
				}()},
				struct {
					V interface{}
				}{func() interface{} {
					var v interface{}
					var vv interface{} = v
					return &vv
				}()},
			},
			out: [][]string{
				{"V"},
				{"1"},
				{"10"},
				{"100"},
				{"1000"},
				{"10000"},
				{"100000"},
				{"csvmarshaler"},
				{"csvmarshaler"},
				{"csvmarshaler"},
				{""},
			},
		},
		{
			desc: "encode NaN",
			in: []interface{}{
				struct {
					Float float64
				}{math.NaN()},
			},
			out: [][]string{
				{"Float"},
				{"NaN"},
			},
		},
		{
			desc: "encode NaN with aliased type",
			in: []interface{}{
				struct {
					Float Float
				}{Float(math.NaN())},
			},
			out: [][]string{
				{"Float"},
				{"NaN"},
			},
		},
		{
			desc: "empty struct",
			in: []interface{}{
				struct{}{},
			},
			out: [][]string{{}, {}},
		},
		{
			desc: "value wrapped in interfaces and pointers",
			in: []interface{}{
				func() (v interface{}) { v = &struct{ A int }{5}; return v }(),
			},
			out: [][]string{{"A"}, {"5"}},
		},
		{
			desc: "csv marshaler error",
			in: []interface{}{
				struct {
					A CSVMarshaler
				}{
					A: CSVMarshaler{Err: Error},
				},
			},
			err: &MarshalerError{Type: reflect.TypeOf(CSVMarshaler{}), MarshalerType: "MarshalCSV", Err: Error},
		},
		{
			desc: "text marshaler error",
			in: []interface{}{
				struct {
					A TextMarshaler
				}{
					A: TextMarshaler{Err: Error},
				},
			},
			err: &MarshalerError{Type: reflect.TypeOf(TextMarshaler{}), MarshalerType: "MarshalText", Err: Error},
		},
		{
			desc: "unsupported type",
			in: []interface{}{
				InvalidType{},
			},
			err: &UnsupportedTypeError{
				Type: reflect.TypeOf(struct{}{}),
			},
		},
		{
			desc: "unsupported double pointer type",
			in: []interface{}{
				struct {
					A **struct{}
				}{},
			},
			err: &UnsupportedTypeError{
				Type: reflect.TypeOf(struct{}{}),
			},
		},
		{
			desc: "unsupported interface type",
			in: []interface{}{
				TypeF{V: TypeA{}},
			},
			err: &UnsupportedTypeError{
				Type: reflect.TypeOf(TypeA{}),
			},
		},
		{
			desc: "encode not a struct",
			in:   []interface{}{int(1)},
			err: &InvalidEncodeError{
				Type: reflect.TypeOf(int(1)),
			},
		},
		{
			desc: "encode nil interface",
			in:   []interface{}{nilIface},
			err: &InvalidEncodeError{
				Type: reflect.TypeOf(nilIface),
			},
		},
		{
			desc: "encode nil ptr",
			in:   []interface{}{nilPtr},
			err:  &InvalidEncodeError{},
		},
		{
			desc: "encode nil interface pointer",
			in:   []interface{}{nilIfacePtr},
			err:  &InvalidEncodeError{},
		},
	}

	for _, f := range fixtures {
		t.Run(f.desc, func(t *testing.T) {
			var buf bytes.Buffer
			w := csv.NewWriter(&buf)
			enc := NewEncoder(w)
			for _, v := range f.in {
				err := enc.Encode(v)
				if f.err != nil {
					if !reflect.DeepEqual(f.err, err) {
						t.Errorf("want err=%v; got %v", f.err, err)
					}
					return
				} else if err != nil {
					t.Errorf("want err=nil; got %v", err)
				}
			}
			w.Flush()
			if err := w.Error(); err != nil {
				t.Errorf("want err=nil; got %v", err)
			}

			var out bytes.Buffer
			if err := csv.NewWriter(&out).WriteAll(f.out); err != nil {
				t.Errorf("want err=nil; got %v", err)
			}

			if buf.String() != out.String() {
				t.Errorf("want=%s; got %s", out.String(), buf.String())
			}
		})
	}

	t.Run("test decoder tags", func(t *testing.T) {
		type Test struct {
			A int     `custom:"1"`
			B string  `custom:"2"`
			C float64 `custom:"-"`
		}

		test := &Test{
			A: 1,
			B: "b",
			C: 2.5,
		}

		var bufs [4]bytes.Buffer
		for i := 0; i < 4; i += 2 {
			encode(t, &bufs[i], test, "")
			encode(t, &bufs[i+1], test, "custom")
		}

		if b1, b2 := bufs[0].String(), bufs[2].String(); b1 != b2 {
			t.Errorf("buffers are not equal: %s vs %s", b1, b2)
		}
		if b1, b2 := bufs[1].String(), bufs[3].String(); b1 != b2 {
			t.Errorf("buffers are not equal: %s vs %s", b1, b2)
		}

		expected1 := [][]string{
			{"A", "B", "C"},
			{"1", "b", "2.5"},
		}
		expected2 := [][]string{
			{"1", "2"},
			{"1", "b"},
		}

		if b1, b2 := bufs[0].String(), encodeCSV(t, expected1); b1 != b2 {
			t.Errorf("want buf=%s; got %s", b2, b1)
		}
		if b1, b2 := bufs[1].String(), encodeCSV(t, expected2); b1 != b2 {
			t.Errorf("want buf=%s; got %s", b2, b1)
		}
	})

	t.Run("error messages", func(t *testing.T) {
		fixtures := []struct {
			desc     string
			expected string
			v        interface{}
		}{
			{
				desc:     "invalid encode error message",
				expected: "csvutil: Encode(int64)",
				v:        int64(1),
			},
			{
				desc:     "invalid encode error message with nil interface",
				expected: "csvutil: Encode(nil)",
				v:        nilIface,
			},
			{
				desc:     "invalid encode error message with nil value",
				expected: "csvutil: Encode(nil)",
				v:        nilPtr,
			},
			{
				desc:     "unsupported type error message",
				expected: "csvutil: unsupported type: struct {}",
				v:        struct{ InvalidType }{},
			},
			{
				desc:     "marshaler error message",
				expected: "csvutil: error calling MarshalText for type csvutil.TextMarshaler: " + Error.Error(),
				v:        struct{ M TextMarshaler }{TextMarshaler{Error}},
			},
		}

		for _, f := range fixtures {
			t.Run(f.desc, func(t *testing.T) {
				err := NewEncoder(csv.NewWriter(bytes.NewBuffer(nil))).Encode(f.v)
				if err == nil {
					t.Fatal("want err not to be nil")
				}
				if err.Error() != f.expected {
					t.Errorf("want=%s; got %s", f.expected, err.Error())
				}
			})
		}
	})

	t.Run("EncodeHeader", func(t *testing.T) {
		t.Run("no double header with encode", func(t *testing.T) {
			var buf bytes.Buffer
			w := csv.NewWriter(&buf)
			enc := NewEncoder(w)
			if err := enc.EncodeHeader(TypeI{}); err != nil {
				t.Errorf("want err=nil; got %v", err)
			}
			if err := enc.Encode(TypeI{}); err != nil {
				t.Errorf("want err=nil; got %v", err)
			}
			w.Flush()

			expected := encodeCSV(t, [][]string{
				{"String", "int"},
				{"", ""},
			})

			if buf.String() != expected {
				t.Errorf("want out=%s; got %s", expected, buf.String())
			}
		})

		t.Run("encode writes header if EncodeHeader fails", func(t *testing.T) {
			var buf bytes.Buffer
			w := csv.NewWriter(&buf)
			enc := NewEncoder(w)

			if err := enc.EncodeHeader(InvalidType{}); err == nil {
				t.Errorf("expected not nil error")
			}

			if err := enc.Encode(TypeI{}); err != nil {
				t.Errorf("want err=nil; got %v", err)
			}

			w.Flush()

			expected := encodeCSV(t, [][]string{
				{"String", "int"},
				{"", ""},
			})

			if buf.String() != expected {
				t.Errorf("want out=%s; got %s", expected, buf.String())
			}
		})

		fixtures := []struct {
			desc string
			in   interface{}
			tag  string
			out  [][]string
			err  error
		}{
			{
				desc: "conflicting fields",
				in:   &Embedded10{},
				out: [][]string{
					{"Y"},
				},
			},
			{
				desc: "custom tag",
				in:   TypeJ{},
				tag:  "json",
				out: [][]string{
					{"string", "bool", "Uint", "Float"},
				},
			},
			{
				desc: "nil interface ptr value",
				in:   nilIfacePtr,
				out: [][]string{
					{
						"int",
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
			},
			{
				desc: "ptr to nil interface ptr value",
				in:   &nilIfacePtr,
				out: [][]string{
					{
						"int",
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
			},
			{
				desc: "nil ptr value",
				in:   nilPtr,
				out: [][]string{
					{
						"int",
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
			},
			{
				desc: "ptr to nil ptr value",
				in:   &nilPtr,
				out: [][]string{
					{
						"int",
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
			},
			{
				desc: "ptr to nil interface",
				in:   &nilIface,
				err:  &UnsupportedTypeError{Type: reflect.ValueOf(&nilIface).Type().Elem()},
			},
			{
				desc: "nil value",
				err:  &UnsupportedTypeError{},
			},
			{
				desc: "ptr - not a struct",
				in:   &[]int{},
				err:  &UnsupportedTypeError{Type: reflect.TypeOf([]int{})},
			},
			{
				desc: "not a struct",
				in:   int(1),
				err:  &UnsupportedTypeError{Type: reflect.TypeOf(int(0))},
			},
		}

		for _, f := range fixtures {
			t.Run(f.desc, func(t *testing.T) {
				var buf bytes.Buffer
				w := csv.NewWriter(&buf)

				enc := NewEncoder(w)
				enc.Tag = f.tag

				err := enc.EncodeHeader(f.in)
				w.Flush()

				if !reflect.DeepEqual(err, f.err) {
					t.Errorf("want err=%v; got %v", f.err, err)
				}

				if f.err != nil {
					return
				}

				if expected := encodeCSV(t, f.out); buf.String() != expected {
					t.Errorf("want out=%s; got %s", expected, buf.String())
				}
			})
		}
	})

	t.Run("AutoHeader false", func(t *testing.T) {
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		enc := NewEncoder(w)
		enc.AutoHeader = false

		if err := enc.Encode(TypeG{
			String: "s",
			Int:    10,
		}); err != nil {
			t.Fatalf("want err=nil; got %v", err)
		}
		w.Flush()

		expected := encodeCSV(t, [][]string{{"s", "10"}})
		if expected != buf.String() {
			t.Errorf("want %s; got %s", expected, buf.String())
		}
	})

	t.Run("fail on type encoding without header", func(t *testing.T) {
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		enc := NewEncoder(w)
		enc.AutoHeader = false

		err := enc.Encode(struct {
			Invalid InvalidType
		}{})

		expected := &UnsupportedTypeError{Type: reflect.TypeOf(InvalidType{})}
		if !reflect.DeepEqual(err, expected) {
			t.Errorf("want %v; got %v", expected, err)
		}
	})

	t.Run("fail while writing header", func(t *testing.T) {
		Error := errors.New("error")
		enc := NewEncoder(failingWriter{Err: Error})
		if err := enc.EncodeHeader(TypeA{}); err != Error {
			t.Errorf("want %v; got %v", Error, err)
		}
	})
}

func encode(t *testing.T, buf *bytes.Buffer, v interface{}, tag string) {
	w := csv.NewWriter(buf)
	enc := NewEncoder(w)
	enc.Tag = tag
	if err := enc.Encode(v); err != nil {
		t.Fatalf("want err=nil; got %v", err)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatalf("want err=nil; got %v", err)
	}
}

func encodeCSV(t *testing.T, recs [][]string) string {
	var buf bytes.Buffer
	if err := csv.NewWriter(&buf).WriteAll(recs); err != nil {
		t.Fatalf("want err=nil; got %v", err)
	}
	return buf.String()
}

type failingWriter struct {
	Err error
}

func (w failingWriter) Write([]string) error {
	return w.Err
}
