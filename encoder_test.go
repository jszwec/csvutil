package csvutil

import (
	"bytes"
	"encoding"
	"encoding/csv"
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strconv"
	"testing"
)

var Error = errors.New("error")

var nilIface any

var nilPtr *TypeF

var nilIfacePtr any = nilPtr

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

type PtrRecCSVMarshaler int

func (m *PtrRecCSVMarshaler) MarshalCSV() ([]byte, error) {
	return []byte("ptrreccsvmarshaler"), nil
}

func (m *PtrRecCSVMarshaler) CSV() ([]byte, error) {
	return []byte("ptrreccsvmarshaler.CSV"), nil
}

type PtrRecTextMarshaler int

func (m *PtrRecTextMarshaler) MarshalText() ([]byte, error) {
	return []byte("ptrrectextmarshaler"), nil
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

type Inline struct {
	J1      TypeJ  `csv:",inline"`
	J2      TypeJ  `csv:"prefix-,inline"`
	String  string `csv:"top-string"`
	String2 string `csv:"STR"`
}

type Inline2 struct {
	S string
	A Inline3 `csv:"A,inline"`
	B Inline3 `csv:",inline"`
}

type Inline3 struct {
	Inline4 `csv:",inline"`
}

type Inline4 struct {
	A string
}

type Inline5 struct {
	A Inline2 `csv:"A,inline"`
	B Inline2 `csv:",inline"`
}

type Inline6 struct {
	A Inline7 `csv:",inline"`
}

type Inline7 struct {
	A *Inline6 `csv:",inline"`
	X int
}

type Inline8 struct {
	F  *Inline4 `csv:"A,inline"`
	AA int
}

type TypeH struct {
	Int     int     `csv:"int,omitempty"`
	Int8    int8    `csv:"int8,omitempty"`
	Int16   int16   `csv:"int16,omitempty"`
	Int32   int32   `csv:"int32,omitempty"`
	Int64   int64   `csv:"int64,omitempty"`
	UInt    uint    `csv:"uint,omitempty"`
	Uint8   uint8   `csv:"uint8,omitempty"`
	Uint16  uint16  `csv:"uint16,omitempty"`
	Uint32  uint32  `csv:"uint32,omitempty"`
	Uint64  uint64  `csv:"uint64,omitempty"`
	Float32 float32 `csv:"float32,omitempty"`
	Float64 float64 `csv:"float64,omitempty"`
	String  string  `csv:"string,omitempty"`
	Bool    bool    `csv:"bool,omitempty"`
	V       any     `csv:"interface,omitempty"`
}

type TypeM struct {
	*TextMarshaler `csv:"text"`
}

func TestEncoder(t *testing.T) {
	fixtures := []struct {
		desc    string
		in      []any
		regFunc marshalersSlice
		out     [][]string
		err     error
	}{
		{
			desc: "test all types",
			in: []any{
				TypeF{
					Int:      1,
					Pint:     ptr(2),
					Int8:     3,
					Pint8:    ptr[int8](4),
					Int16:    5,
					Pint16:   ptr[int16](6),
					Int32:    7,
					Pint32:   ptr[int32](8),
					Int64:    9,
					Pint64:   ptr[int64](10),
					UInt:     11,
					Puint:    ptr[uint](12),
					Uint8:    13,
					Puint8:   ptr[uint8](14),
					Uint16:   15,
					Puint16:  ptr[uint16](16),
					Uint32:   17,
					Puint32:  ptr[uint32](18),
					Uint64:   19,
					Puint64:  ptr[uint64](20),
					Float32:  21,
					Pfloat32: ptr[float32](22),
					Float64:  23,
					Pfloat64: ptr[float64](24),
					String:   "25",
					PString:  ptr("26"),
					Bool:     true,
					Pbool:    ptr(true),
					V:        "true",
					Pv:       ptr[any]("1"),
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
			in: []any{
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
			in: []any{
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
			desc: "omitempty tags on pointers - non nil default values",
			in: []any{
				struct {
					Pint    *int    `csv:",omitempty"`
					PPint   **int   `csv:",omitempty"`
					PPint2  **int   `csv:",omitempty"`
					PString *string `csv:",omitempty"`
					PBool   *bool   `csv:",omitempty"`
					Iint    *any    `csv:",omitempty"`
				}{
					ptr(0),
					pptr(0),
					new(*int),
					ptr(""),
					ptr(false),
					ptr[any](0),
				},
			},
			out: [][]string{
				{"Pint", "PPint", "PPint2", "PString", "PBool", "Iint"},
				{"0", "0", "", "", "false", "0"},
			},
		},
		{
			desc: "omitempty tags on pointers - nil ptrs",
			in: []any{
				struct {
					Pint    *int    `csv:",omitempty"`
					PPint   **int   `csv:",omitempty"`
					PString *string `csv:",omitempty"`
					PBool   *bool   `csv:",omitempty"`
					Iint    *any    `csv:",omitempty"`
				}{},
			},
			out: [][]string{
				{"Pint", "PPint", "PString", "PBool", "Iint"},
				{"", "", "", "", ""},
			},
		},
		{
			desc: "omitempty tags on interfaces - non nil default values",
			in: []any{
				struct {
					Iint  any `csv:",omitempty"`
					IPint any `csv:",omitempty"`
				}{
					0,
					ptr(0),
				},
				struct {
					Iint  any `csv:",omitempty"`
					IPint any `csv:",omitempty"`
				}{
					1,
					ptr(1),
				},
			},
			out: [][]string{
				{"Iint", "IPint"},
				{"0", "0"},
				{"1", "1"},
			},
		},
		{
			desc: "omitempty tags on interfaces - nil",
			in: []any{
				struct {
					Iint  any `csv:",omitempty"`
					IPint any `csv:",omitempty"`
				}{
					nil,
					nil,
				},
				struct {
					Iint  any `csv:",omitempty"`
					IPint any `csv:",omitempty"`
				}{
					(*int)(nil),
					ptr[any]((*int)(nil)),
				},
			},
			out: [][]string{
				{"Iint", "IPint"},
				{"", ""},
				{"", ""},
			},
		},
		{
			desc: "embedded types #1",
			in: []any{
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
			in: []any{
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
			in: []any{
				&struct {
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
				{`{"key":"val"}`, `{"key1":"val1"}`},
			},
		},
		{
			desc: "embedded non struct tagged types with pointer receiver MarshalText",
			in: []any{
				&struct {
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
				{`{"key":"val"}`, `{"key1":"val1"}`},
			},
		},
		{
			desc: "embedded pointer types",
			in: []any{
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
			in: []any{
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
			in: []any{
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
			desc: "embedded non struct tagged pointer types with nil value - textmarshaler",
			in: []any{
				TypeM{
					TextMarshaler: nil,
				},
			},
			out: [][]string{
				{"text"},
				{""},
			},
		},
		{
			desc: "embedded non struct tagged pointer types with nil value - csvmarshaler",
			in: []any{
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
			in: []any{
				TagPriority{Foo: 1, Bar: 2},
			},
			out: [][]string{
				{"Foo"},
				{"2"},
			},
		},
		{
			desc: "conflicting embedded fields #1",
			in: []any{
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
			in: []any{
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
			in: []any{
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
			in: []any{
				TypeE{},
			},
			out: [][]string{
				{"string", "int"},
				{"", ""},
			},
		},
		{
			desc: "unexported non-struct embedded",
			in: []any{
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
			in: []any{
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
			desc: "ptr receiver csv marshaler",
			in: []any{
				&struct {
					A PtrRecCSVMarshaler
				}{},
				struct {
					A PtrRecCSVMarshaler
				}{},
				struct {
					A *PtrRecCSVMarshaler
				}{new(PtrRecCSVMarshaler)},
				&struct {
					A *PtrRecCSVMarshaler
				}{new(PtrRecCSVMarshaler)},
				&struct {
					A *PtrRecCSVMarshaler
				}{},
			},
			out: [][]string{
				{"A"},
				{"ptrreccsvmarshaler"},
				{"0"},
				{"ptrreccsvmarshaler"},
				{"ptrreccsvmarshaler"},
				{""},
			},
		},
		{
			desc: "ptr receiver text marshaler",
			in: []any{
				&struct {
					A PtrRecTextMarshaler
				}{},
				struct {
					A PtrRecTextMarshaler
				}{},
				struct {
					A *PtrRecTextMarshaler
				}{new(PtrRecTextMarshaler)},
				&struct {
					A *PtrRecTextMarshaler
				}{new(PtrRecTextMarshaler)},
				&struct {
					A *PtrRecTextMarshaler
				}{},
			},
			out: [][]string{
				{"A"},
				{"ptrrectextmarshaler"},
				{"0"},
				{"ptrrectextmarshaler"},
				{"ptrrectextmarshaler"},
				{""},
			},
		},
		{
			desc: "text marshaler",
			in: []any{
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
			in: []any{
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
			in: []any{
				struct{ Float float64 }{3.14},
			},
			out: [][]string{
				{"Float"},
				{"3.14"},
			},
		},
		{
			desc: "embedded tagged marshalers",
			in: []any{
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
			in: []any{
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
			desc: "inline fields",
			in: []any{
				Inline{
					J1: TypeJ{
						String:     "j1",
						Int:        "1",
						Float:      "1",
						Embedded16: Embedded16{Bool: true, Uint8: 1},
					},
					J2: TypeJ{
						String:     "j2",
						Int:        "2",
						Float:      "2",
						Embedded16: Embedded16{Bool: true, Uint8: 2},
					},
					String:  "top-level-str",
					String2: "STR",
				},
			},
			out: [][]string{
				{"int", "Bool", "Uint8", "float", "prefix-STR", "prefix-int", "prefix-Bool", "prefix-Uint8", "prefix-float", "top-string", "STR"},
				{"1", "true", "1", "1", "j2", "2", "true", "2", "2", "top-level-str", "STR"},
			},
		},
		{
			desc: "inline chain",
			in: []any{
				Inline5{
					A: Inline2{
						S: "1",
						A: Inline3{
							Inline4: Inline4{A: "11"},
						},
						B: Inline3{
							Inline4: Inline4{A: "12"},
						},
					},
					B: Inline2{
						S: "2",
						A: Inline3{
							Inline4: Inline4{A: "21"},
						},
						B: Inline3{
							Inline4: Inline4{A: "22"},
						},
					},
				},
			},
			out: [][]string{
				{"AS", "AAA", "S", "A"},
				{"1", "11", "2", "22"},
			},
		},
		{
			desc: "cyclic inline - no prefix",
			in: []any{
				Inline6{
					A: Inline7{
						A: &Inline6{A: Inline7{
							A: &Inline6{},
							X: 10,
						}},
						X: 1,
					},
				},
			},
			out: [][]string{
				{"X"},
				{"1"},
			},
		},
		{
			desc: "embedded with inline tag",
			in: []any{
				struct {
					Inline7 `csv:"A,inline"`
				}{
					Inline7: Inline7{
						A: &Inline6{A: Inline7{
							A: &Inline6{},
							X: 10,
						}},
						X: 1,
					},
				},
			},
			out: [][]string{
				{"AX"},
				{"1"},
			},
		},
		{
			desc: "embedded with empty inline tag",
			in: []any{
				struct {
					Inline7 `csv:",inline"`
				}{
					Inline7: Inline7{
						A: &Inline6{A: Inline7{
							A: &Inline6{},
							X: 10,
						}},
						X: 1,
					},
				},
			},
			out: [][]string{
				{"X"},
				{"1"},
			},
		},
		{
			desc: "embedded with ptr inline tag",
			in: []any{
				struct {
					*Inline7 `csv:"A,inline"`
				}{
					Inline7: &Inline7{
						A: &Inline6{A: Inline7{
							A: &Inline6{},
							X: 10,
						}},
						X: 1,
					},
				},
			},
			out: [][]string{
				{"AX"},
				{"1"},
			},
		},
		{
			desc: "inline visibility rules - top field first",
			in: []any{
				struct {
					AA string
					F  Inline4 `csv:"A,inline"`
				}{
					AA: "1",
					F:  Inline4{A: "10"},
				},
			},
			out: [][]string{
				{"AA"},
				{"1"},
			},
		},
		{
			desc: "inline visibility rules - top field last",
			in: []any{
				Inline8{
					F:  &Inline4{A: "10"},
					AA: 1,
				},
			},
			out: [][]string{
				{"AA"},
				{"1"},
			},
		},
		{
			desc: "ignore inline tag on non struct",
			in: []any{
				struct {
					X int `csv:",inline"`
					Y int `csv:"y,inline"`
				}{
					X: 1,
					Y: 2,
				},
			},
			out: [][]string{
				{"X", "y"},
				{"1", "2"},
			},
		},
		{
			desc: "registered func - non ptr elem",
			in: []any{
				struct {
					Int    int
					Pint   *int
					Iface  any
					Piface *any
				}{
					Pint:   ptr(0),
					Iface:  34,
					Piface: ptr[any](34),
				},
			},
			regFunc: marshalersSlice{
				marshalerFunc(func(int) ([]byte, error) { return []byte("int"), nil }),
			},
			out: [][]string{
				{"Int", "Pint", "Iface", "Piface"},
				{"int", "int", "int", "int"},
			},
		},
		{
			desc: "registered func - ptr elem",
			in: []any{
				&struct {
					Int    int
					Pint   *int
					Iface  any
					Piface *any
				}{
					Pint:   ptr(0),
					Iface:  34,
					Piface: ptr[any](34),
				},
			},
			regFunc: marshalersSlice{
				marshalerFunc(func(int) ([]byte, error) { return []byte("int"), nil }),
			},
			out: [][]string{
				{"Int", "Pint", "Iface", "Piface"},
				{"int", "int", "int", "int"},
			},
		},
		{
			desc: "registered func - ptr type - non ptr elem",
			in: []any{
				struct {
					Int    int
					Pint   *int
					Iface  any
					Piface *any
				}{
					Pint:   ptr(0),
					Iface:  34,
					Piface: ptr[any](ptr(34)),
				},
			},
			regFunc: marshalersSlice{
				marshalerFunc(func(*int) ([]byte, error) { return []byte("int"), nil }),
			},
			out: [][]string{
				{"Int", "Pint", "Iface", "Piface"},
				{"0", "int", "34", "int"},
			},
		},
		{
			desc: "registered func - ptr type - ptr elem",
			in: []any{
				&struct {
					Int    int
					Pint   *int
					Iface  any
					Piface *any
				}{
					Pint:   ptr(0),
					Iface:  34,
					Piface: ptr[any](ptr(34)),
				},
			},
			regFunc: marshalersSlice{
				marshalerFunc(func(*int) ([]byte, error) { return []byte("int"), nil }),
			},
			out: [][]string{
				{"Int", "Pint", "Iface", "Piface"},
				{"int", "int", "34", "int"},
			},
		},
		{
			desc: "registered func - mixed types - non ptr elem",
			in: []any{
				struct {
					Int    int
					Pint   *int
					Iface  any
					Piface *any
				}{
					Pint:   ptr(0),
					Iface:  34,
					Piface: ptr[any](ptr(34)),
				},
			},
			regFunc: marshalersSlice{
				marshalerFunc(func(int) ([]byte, error) { return []byte("int"), nil }),
				marshalerFunc(func(*int) ([]byte, error) { return []byte("*int"), nil }),
			},
			out: [][]string{
				{"Int", "Pint", "Iface", "Piface"},
				{"int", "*int", "int", "*int"},
			},
		},
		{
			desc: "registered func - mixed types - ptr elem",
			in: []any{
				&struct {
					Int    int
					Pint   *int
					Iface  any
					Piface *any
				}{
					Pint:   ptr(0),
					Iface:  34,
					Piface: ptr[any](ptr(34)),
				},
			},
			regFunc: marshalersSlice{
				marshalerFunc(func(int) ([]byte, error) { return []byte("int"), nil }),
				marshalerFunc(func(*int) ([]byte, error) { return []byte("*int"), nil }),
			},
			out: [][]string{
				{"Int", "Pint", "Iface", "Piface"},
				{"int", "*int", "int", "*int"},
			},
		},
		{
			desc: "registered func - interfaces",
			in: []any{
				&struct {
					CSVMarshaler        Marshaler
					Marshaler           CSVMarshaler
					PMarshaler          *CSVMarshaler
					CSVTextMarshaler    CSVTextMarshaler
					PCSVTextMarshaler   *CSVTextMarshaler
					PtrRecCSVMarshaler  PtrRecCSVMarshaler
					PtrRecTextMarshaler PtrRecTextMarshaler
				}{
					PMarshaler:        &CSVMarshaler{},
					PCSVTextMarshaler: &CSVTextMarshaler{},
				},
			},
			regFunc: marshalersSlice{
				marshalerFunc(func(Marshaler) ([]byte, error) { return []byte("registered.marshaler"), nil }),
				marshalerFunc(func(encoding.TextMarshaler) ([]byte, error) { return []byte("registered.textmarshaler"), nil }),
			},
			out: [][]string{
				{"CSVMarshaler", "Marshaler", "PMarshaler", "CSVTextMarshaler", "PCSVTextMarshaler", "PtrRecCSVMarshaler", "PtrRecTextMarshaler"},
				{"registered.marshaler", "registered.marshaler", "registered.marshaler", "registered.marshaler", "registered.marshaler", "registered.marshaler", "registered.textmarshaler"},
			},
		},
		{
			desc: "registered func - interface order",
			in: []any{
				&struct {
					CSVTextMarshaler  CSVTextMarshaler
					PCSVTextMarshaler *CSVTextMarshaler
				}{
					PCSVTextMarshaler: &CSVTextMarshaler{},
				},
			},
			regFunc: marshalersSlice{
				marshalerFunc(func(encoding.TextMarshaler) ([]byte, error) { return []byte("registered.textmarshaler"), nil }),
				marshalerFunc(func(Marshaler) ([]byte, error) { return []byte("registered.marshaler"), nil }),
			},
			out: [][]string{
				{"CSVTextMarshaler", "PCSVTextMarshaler"},
				{"registered.textmarshaler", "registered.textmarshaler"},
			},
		},
		{
			desc: "registered func - method",
			in: []any{
				&struct {
					PtrRecCSVMarshaler PtrRecCSVMarshaler
				}{},
				struct {
					PtrRecCSVMarshaler PtrRecCSVMarshaler
				}{},
			},
			regFunc: marshalersSlice{
				marshalerFunc((*PtrRecCSVMarshaler).CSV),
			},
			out: [][]string{
				{"PtrRecCSVMarshaler"},
				{"ptrreccsvmarshaler.CSV"},
				{"0"},
			},
		},
		{
			desc: "registered func - fallback error",
			in: []any{
				struct {
					Embedded14
				}{},
			},
			regFunc: marshalersSlice{
				marshalerFunc((*Embedded14).MarshalCSV),
			},
			err: &UnsupportedTypeError{
				Type: reflect.TypeOf(Embedded14{}),
			},
		},
		{
			desc: "registered interface func - returning error",
			in: []any{
				&struct {
					Embedded14 Embedded14
				}{},
			},
			regFunc: marshalersSlice{
				marshalerFunc(func(Marshaler) ([]byte, error) { return nil, Error }),
			},
			err: Error,
		},
		{
			desc: "registered func - returning error",
			in: []any{
				&struct {
					A InvalidType
				}{},
			},
			regFunc: marshalersSlice{
				marshalerFunc(func(*InvalidType) ([]byte, error) { return nil, Error }),
			},
			err: Error,
		},
		{
			desc: "registered func - fallback error on interface",
			in: []any{
				struct {
					Embedded14
				}{},
			},
			regFunc: marshalersSlice{
				marshalerFunc(func(m Marshaler) ([]byte, error) { return nil, nil }),
			},
			err: &UnsupportedTypeError{
				Type: reflect.TypeOf(Embedded14{}),
			},
		},
		{
			desc: "marshaler fallback error",
			in: []any{
				struct {
					Embedded14
				}{},
			},
			err: &UnsupportedTypeError{
				Type: reflect.TypeOf(Embedded14{}),
			},
		},
		{
			desc: "encode different types",
			// This doesnt mean the output csv is valid. Generally this is an invalid
			// use. However, we need to make sure that the encoder is doing what it is
			// asked to... correctly.
			in: []any{
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
			in: []any{
				struct {
					V any
				}{1},
				struct {
					V any
				}{ptr(10)},
				struct {
					V any
				}{pptr(100)},
				struct {
					V any
				}{ppptr(1000)},
				struct {
					V *any
				}{ptr[any](pptr(10000))},
				struct {
					V *any
				}{func() *any {
					var v any = ppptr(100000)
					var vv any = v
					return &vv
				}()},
				struct {
					V any
				}{func() any {
					var v any = &CSVMarshaler{}
					var vv any = v
					return &vv
				}()},
				struct {
					V any
				}{func() any {
					var v any = CSVMarshaler{}
					var vv any = v
					return &vv
				}()},
				struct {
					V any
				}{func() any {
					var v any = &CSVMarshaler{}
					var vv any = v
					return vv
				}()},
				struct {
					V any
				}{
					V: func() any {
						return PtrRecCSVMarshaler(5)
					}(),
				},
				struct {
					V any
				}{
					V: func() any {
						m := PtrRecCSVMarshaler(5)
						return &m
					}(),
				},
				struct {
					V any
				}{func() any {
					var v any
					var vv any = v
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
				{"5"},
				{"ptrreccsvmarshaler"},
				{""},
			},
		},
		{
			desc: "encode NaN",
			in: []any{
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
			in: []any{
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
			in: []any{
				struct{}{},
			},
			out: [][]string{{}, {}},
		},
		{
			desc: "value wrapped in interfaces and pointers",
			in: []any{
				func() (v any) { v = &struct{ A int }{5}; return v }(),
			},
			out: [][]string{{"A"}, {"5"}},
		},
		{
			desc: "csv marshaler error",
			in: []any{
				struct {
					A CSVMarshaler
				}{
					A: CSVMarshaler{Err: Error},
				},
			},
			err: &MarshalerError{Type: reflect.TypeOf(CSVMarshaler{}), MarshalerType: "MarshalCSV", Err: Error},
		},
		{
			desc: "csv marshaler error as registered error",
			in: []any{
				struct {
					A CSVMarshaler
				}{
					A: CSVMarshaler{Err: Error},
				},
			},
			regFunc: marshalersSlice{
				marshalerFunc(CSVMarshaler.MarshalCSV),
			},
			err: Error,
		},
		{
			desc: "text marshaler error",
			in: []any{
				struct {
					A TextMarshaler
				}{
					A: TextMarshaler{Err: Error},
				},
			},
			err: &MarshalerError{Type: reflect.TypeOf(TextMarshaler{}), MarshalerType: "MarshalText", Err: Error},
		},
		{
			desc: "text marshaler fallback error - ptr reciever",
			in: []any{
				struct {
					A Embedded15
				}{},
			},
			err: &UnsupportedTypeError{Type: reflect.TypeOf(Embedded15{})},
		},
		{
			desc: "text marshaler error as registered func",
			in: []any{
				struct {
					A TextMarshaler
				}{
					A: TextMarshaler{Err: Error},
				},
			},
			regFunc: marshalersSlice{
				marshalerFunc(TextMarshaler.MarshalText),
			},
			err: Error,
		},
		{
			desc: "unsupported type",
			in: []any{
				InvalidType{},
			},
			err: &UnsupportedTypeError{
				Type: reflect.TypeOf(struct{}{}),
			},
		},
		{
			desc: "unsupported double pointer type",
			in: []any{
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
			in: []any{
				TypeF{V: TypeA{}},
			},
			err: &UnsupportedTypeError{
				Type: reflect.TypeOf(TypeA{}),
			},
		},
		{
			desc: "encode not a struct",
			in:   []any{int(1)},
			err: &InvalidEncodeError{
				Type: reflect.TypeOf(int(1)),
			},
		},
		{
			desc: "encode nil interface",
			in:   []any{nilIface},
			err: &InvalidEncodeError{
				Type: reflect.TypeOf(nilIface),
			},
		},
		{
			desc: "encode nil ptr",
			in:   []any{nilPtr},
			err:  &InvalidEncodeError{},
		},
		{
			desc: "encode nil interface pointer",
			in:   []any{nilIfacePtr},
			err:  &InvalidEncodeError{},
		},
	}

	for _, f := range fixtures {
		f := f

		do := func(t *testing.T, fn func(*Encoder)) {
			t.Helper()

			var buf bytes.Buffer
			w := csv.NewWriter(&buf)
			enc := NewEncoder(w)
			fn(enc)

			for _, v := range f.in {
				err := enc.Encode(v)
				if f.err != nil {
					if !checkErr(f.err, err) {
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
		}

		if len(f.regFunc) == 0 {
			t.Run(f.desc, func(t *testing.T) {
				do(t, func(e *Encoder) {})
			})
			continue
		}

		t.Run("old register "+f.desc, func(t *testing.T) {
			do(t, func(enc *Encoder) {
				for _, f := range f.regFunc {
					enc.Register(f.RawFunc.Interface())
				}
			})
		})

		t.Run("new register "+f.desc, func(t *testing.T) {
			do(t, func(enc *Encoder) {
				enc.WithMarshalers(NewMarshalers(f.regFunc.Marshalers()...))
			})
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
			v        any
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
			in   any
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
				desc: "struct slice",
				in:   []TypeF{},
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

				if !checkErr(f.err, err) {
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

	t.Run("slice and array", func(t *testing.T) {
		fixtures := []struct {
			desc string
			in   any
			out  [][]string
			err  error
		}{
			{
				desc: "slice",
				in: []TypeI{
					{"1", 1},
					{"2", 2},
				},
				out: [][]string{
					{"String", "int"},
					{"1", "1"},
					{"2", "2"},
				},
			},
			{
				desc: "ptr slice",
				in: &[]TypeI{
					{"1", 1},
					{"2", 2},
				},
				out: [][]string{
					{"String", "int"},
					{"1", "1"},
					{"2", "2"},
				},
			},
			{
				desc: "ptr slice with ptr elements",
				in: &[]*TypeI{
					{"1", 1},
					{"2", 2},
				},
				out: [][]string{
					{"String", "int"},
					{"1", "1"},
					{"2", "2"},
				},
			},
			{
				desc: "array",
				in: [2]TypeI{
					{"1", 1},
					{"2", 2},
				},
				out: [][]string{
					{"String", "int"},
					{"1", "1"},
					{"2", "2"},
				},
			},
			{
				desc: "ptr array",
				in: &[2]TypeI{
					{"1", 1},
					{"2", 2},
				},
				out: [][]string{
					{"String", "int"},
					{"1", "1"},
					{"2", "2"},
				},
			},
			{
				desc: "ptr array with ptr elements",
				in: &[2]*TypeI{
					{"1", 1},
					{"2", 2},
				},
				out: [][]string{
					{"String", "int"},
					{"1", "1"},
					{"2", "2"},
				},
			},
			{
				desc: "array with default val",
				in: [2]TypeI{
					{"1", 1},
				},
				out: [][]string{
					{"String", "int"},
					{"1", "1"},
					{"", ""},
				},
			},
			{
				desc: "no auto header on empty slice",
				in:   []TypeI{},
				out:  [][]string{},
			},
			{
				desc: "no auto header on empty array",
				in:   [0]TypeI{},
				out:  [][]string{},
			},
			{
				desc: "disallow double slice",
				in: [][]TypeI{
					{
						{"1", 1},
					},
				},
				err: &InvalidEncodeError{Type: reflect.TypeOf([][]TypeI{})},
			},
			{
				desc: "disallow double ptr slice",
				in: &[][]TypeI{
					{
						{"1", 1},
					},
				},
				err: &InvalidEncodeError{Type: reflect.TypeOf(&[][]TypeI{})},
			},
			{
				desc: "disallow double ptr slice with ptr slice",
				in: &[]*[]TypeI{
					{
						{"1", 1},
					},
				},
				err: &InvalidEncodeError{Type: reflect.TypeOf(&[]*[]TypeI{})},
			},
			{
				desc: "disallow double array",
				in: [2][2]TypeI{
					{
						{"1", 1},
					},
				},
				err: &InvalidEncodeError{Type: reflect.TypeOf([2][2]TypeI{})},
			},
			{
				desc: "disallow double ptr array",
				in: &[2][2]TypeI{
					{
						{"1", 1},
					},
				},
				err: &InvalidEncodeError{Type: reflect.TypeOf(&[2][2]TypeI{})},
			},
			{
				desc: "disallow interface slice",
				in: []any{
					TypeI{"1", 1},
				},
				err: &InvalidEncodeError{Type: reflect.TypeOf([]any{})},
			},
			{
				desc: "disallow interface array",
				in: [1]any{
					TypeI{"1", 1},
				},
				err: &InvalidEncodeError{Type: reflect.TypeOf([1]any{})},
			},
		}

		for _, f := range fixtures {
			t.Run(f.desc, func(t *testing.T) {
				var buf bytes.Buffer
				w := csv.NewWriter(&buf)
				err := NewEncoder(w).Encode(f.in)

				if f.err != nil {
					if !checkErr(f.err, err) {
						t.Errorf("want err=%v; got %v", f.err, err)
					}
					return
				}

				if err != nil {
					t.Fatalf("want err=nil; got %v", err)
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
	})

	t.Run("with header", func(t *testing.T) {
		t.Run("all present and sorted", func(t *testing.T) {
			fixtures := []struct {
				desc       string
				autoHeader bool
				out        [][]string
			}{
				{
					desc:       "with autoheader",
					autoHeader: true,
					out: [][]string{
						{"C", "B", "D"},
						{"c", "b", "d"},
					},
				},
				{
					desc:       "without autoheader",
					autoHeader: false,
					out: [][]string{
						{"c", "b", "d"},
					},
				},
			}

			for _, f := range fixtures {
				t.Run(f.desc, func(t *testing.T) {
					type Embedded struct {
						D string
					}
					type Foo struct {
						A string
						Embedded
						B string
						C string
					}

					var buf bytes.Buffer
					w := csv.NewWriter(&buf)
					enc := NewEncoder(w)
					enc.SetHeader([]string{"C", "B", "D"})
					enc.AutoHeader = f.autoHeader
					enc.Encode(Foo{
						A: "a",
						Embedded: Embedded{
							D: "d",
						},
						B: "b",
						C: "c",
					})

					w.Flush()

					expected := encodeCSV(t, f.out)
					if expected != buf.String() {
						t.Errorf("want=%s; got %s", expected, buf.String())
					}
				})
			}
		})

		t.Run("missing fields", func(t *testing.T) {
			fixtures := []struct {
				desc       string
				autoHeader bool
				out        [][]string
			}{
				{
					desc:       "with autoheader",
					autoHeader: true,
					out: [][]string{
						{"C", "X", "A", "Z"},
						{"c", "", "a", ""},
					},
				},
				{
					desc:       "without autoheader",
					autoHeader: false,
					out: [][]string{
						{"c", "", "a", ""},
					},
				},
			}

			for _, f := range fixtures {
				t.Run(f.desc, func(t *testing.T) {
					type Foo struct {
						A string
						B string
						C string
					}

					var buf bytes.Buffer
					w := csv.NewWriter(&buf)
					enc := NewEncoder(w)
					enc.SetHeader([]string{"C", "X", "A", "Z"})
					enc.AutoHeader = f.autoHeader
					enc.Encode(Foo{
						A: "a",
						B: "b",
						C: "c",
					})

					w.Flush()

					expected := encodeCSV(t, f.out)
					if expected != buf.String() {
						t.Errorf("want=%q; got %q", expected, buf.String())
					}
				})
			}
		})

		t.Run("duplicates", func(t *testing.T) {
			type Foo struct {
				A string
				B string
				C string
			}

			var buf bytes.Buffer
			w := csv.NewWriter(&buf)
			enc := NewEncoder(w)
			enc.SetHeader([]string{"C", "X", "C", "A", "X", "Z", "A"})
			enc.Encode(Foo{
				A: "a",
				B: "b",
				C: "c",
			})

			w.Flush()

			expected := encodeCSV(t, [][]string{
				{"C", "X", "Z", "A"},
				{"c", "", "", "a"},
			})
			if expected != buf.String() {
				t.Errorf("want=%q; got %q", expected, buf.String())
			}
		})
	})

	t.Run("register panics", func(t *testing.T) {
		var buf bytes.Buffer
		r := csv.NewWriter(&buf)
		enc := NewEncoder(r)

		fixtures := []struct {
			desc string
			arg  any
		}{
			{
				desc: "not a func",
				arg:  1,
			},
			{
				desc: "nil",
				arg:  nil,
			},
			{
				desc: "T == empty interface",
				arg:  func(any) ([]byte, error) { return nil, nil },
			},
			{
				desc: "first out not bytes",
				arg:  func(int) (int, error) { return 0, nil },
			},
			{
				desc: "second out not error",
				arg:  func(int) (int, int) { return 0, 0 },
			},
			{
				desc: "func with one out value",
				arg:  func(int) error { return nil },
			},
			{
				desc: "func with no returns",
				arg:  func(int) {},
			},
		}

		for _, f := range fixtures {
			t.Run(f.desc, func(t *testing.T) {
				var e any
				func() {
					defer func() {
						e = recover()
					}()
					enc.Register(f.arg)
				}()

				if e == nil {
					t.Error("Register was supposed to panic but it didnt")
				}
				t.Log(e)
			})
		}

		t.Run("already registered", func(t *testing.T) {
			f := func(int) ([]byte, error) { return nil, nil }
			enc.Register(f)

			var e any
			func() {
				defer func() {
					e = recover()
				}()
				enc.Register(f)
			}()

			if e == nil {
				t.Error("Register was supposed to panic but it didnt")
			}
			t.Log(e)
		})
	})
}

func BenchmarkEncode(b *testing.B) {
	b.Run("registered type", func(b *testing.B) {
		type Foo struct {
			A int `csv:"a"`
		}

		b.Run("old register", func(b *testing.B) {
			var buf bytes.Buffer
			w := csv.NewWriter(&buf)
			enc := NewEncoder(w)
			enc.AutoHeader = false

			enc.Register(func(v int) ([]byte, error) {
				return []byte(strconv.Itoa(v)), nil
			})

			var a Foo
			for i := 0; i < b.N; i++ {
				if err := enc.Encode(a); err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("new register", func(b *testing.B) {
			var buf bytes.Buffer
			w := csv.NewWriter(&buf)
			enc := NewEncoder(w)
			enc.AutoHeader = false

			enc.WithMarshalers(NewMarshalers(MarshalFunc(
				func(v int) ([]byte, error) {
					return []byte(strconv.Itoa(v)), nil
				},
			)))

			var a Foo
			for i := 0; i < b.N; i++ {
				if err := enc.Encode(a); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

func encode(t *testing.T, buf *bytes.Buffer, v any, tag string) {
	t.Helper()

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
	t.Helper()

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

type marshalers struct {
	*Marshalers
	RawFunc reflect.Value
}

func marshalerFunc[T any](f func(T) ([]byte, error)) marshalers {
	return marshalers{
		Marshalers: MarshalFunc(f),
		RawFunc:    reflect.ValueOf(f),
	}
}

type marshalersSlice []marshalers

func (ms marshalersSlice) Marshalers() (out []*Marshalers) {
	for i := range ms {
		out = append(out, ms[i].Marshalers)
	}
	return out
}
