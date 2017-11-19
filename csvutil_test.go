package csvutil

import (
	"encoding/csv"
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	fixture := []struct {
		desc string
		src  []byte
		in   interface{}
		out  interface{}
	}{
		{
			desc: "type with two records",
			src:  []byte("String,int\nstring1,1\nstring2,2"),
			in:   new([]TypeI),
			out: &[]TypeI{
				{"string1", 1},
				{"string2", 2},
			},
		},
		{
			desc: "pointer types with two records",
			src:  []byte("String,int\nstring1,1\nstring2,2"),
			in:   &[]*TypeI{},
			out: &[]*TypeI{
				{"string1", 1},
				{"string2", 2},
			},
		},
		{
			desc: "quoted input",
			src:  []byte("\n\n\n\"String\",\"int\"\n\"string1,\n\",\"1\"\n\n\n\n\"string2\",\"2\""),
			in:   &[]TypeI{},
			out: &[]TypeI{
				{"string1,\n", 1},
				{"string2", 2},
			},
		},
		{
			desc: "quoted input - with endline",
			src:  []byte("\n\n\n\"String\",\"int\"\n\"string1,\n\",\"1\"\n\"string2\",\"2\"\n\n\n"),
			in:   &[]TypeI{},
			out: &[]TypeI{
				{"string1,\n", 1},
				{"string2", 2},
			},
		},
		{
			desc: "header only",
			src:  []byte("String,int\n"),
			in:   &[]TypeI{},
			out:  &[]TypeI{},
		},
		{
			desc: "no data",
			src:  []byte(""),
			in:   &[]TypeI{},
			out:  &[]TypeI{},
		},
	}

	for _, f := range fixture {
		t.Run(f.desc, func(t *testing.T) {
			if err := Unmarshal(f.src, f.in); err != nil {
				t.Fatalf("want err=nil; got %v", err)
			}

			if !reflect.DeepEqual(f.in, f.out) {
				t.Errorf("want out=%v; got %v", f.out, f.in)
			}

			out := reflect.ValueOf(f.out).Elem()
			in := reflect.ValueOf(f.in).Elem()
			if cout, cin := out.Cap(), in.Cap(); cout != cin {
				t.Errorf("want cap=%d; got %d", cout, cin)
			}
		})
	}

	t.Run("invalid data", func(t *testing.T) {
		type A struct{}

		fixtures := []struct {
			desc string
			data []byte
			err  error
		}{
			{
				desc: "invalid first line",
				data: []byte(`"`),
				err:  &csv.ParseError{Line: 1, Column: 1, Err: csv.ErrQuote},
			},
			{
				desc: "invalid second line",
				data: []byte("line\n\""),
				err:  &csv.ParseError{Line: 2, Column: 1, Err: csv.ErrQuote},
			},
		}

		for _, f := range fixtures {
			t.Run(f.desc, func(t *testing.T) {
				var a []A
				if err := Unmarshal(f.data, &a); !reflect.DeepEqual(err, f.err) {
					t.Errorf("want err=%v; got %v", f.err, err)
				}
			})
		}
	})

	t.Run("test invalid arguments", func(t *testing.T) {
		n := 1

		var fixtures = []struct {
			desc     string
			v        interface{}
			expected string
		}{
			{"nil interface", interface{}(nil), "csvutil: Unmarshal(nil)"},
			{"nil", nil, "csvutil: Unmarshal(nil)"},
			{"non pointer struct", struct{}{}, "csvutil: Unmarshal(non-pointer struct {})"},
			{"non-slice pointer", (*int)(nil), "csvutil: Unmarshal(non-slice pointer)"},
			{"non-nil non-slice pointer", &n, "csvutil: Unmarshal(non-slice pointer)"},
		}

		for _, f := range fixtures {
			t.Run(f.desc, func(t *testing.T) {
				err := Unmarshal([]byte(""), f.v)
				if err == nil {
					t.Fatalf("want err != nil")
				}
				if got := err.Error(); got != f.expected {
					t.Errorf("want err=%s; got %s", f.expected, got)
				}
			})
		}
	})
}

func TestCountLines(t *testing.T) {
	fixtures := []struct {
		desc string
		data []byte
		out  int
	}{
		{
			desc: "three lines no endline",
			data: []byte(`line1,line1
line2,line2,
line3,line3`),
			out: 3,
		},
		{
			desc: "three lines",
			data: []byte(`line1,line1
line2,line2
line3,line3
`),
			out: 3,
		},
		{
			desc: "no data",
			data: []byte(``),
			out:  0,
		},
		{
			desc: "endline in a quoted string",
			data: []byte(`"line
""1""",line1
line2,"line   
  2"""
`),
			out: 2,
		},
		{
			desc: "empty lines",
			data: []byte("\n\nline1,line1\n\n\n\nline2,line2\n\n"),
			out:  2,
		},
		{
			desc: "1 line ending with quote",
			data: []byte(`"line1","line2"`),
			out:  1,
		},
		{
			desc: "1 line ending with quote - with endline",
			data: []byte(`"line1","line2"
`),
			out: 1,
		},
		{
			desc: "2 lines ending with quote",
			data: []byte(`"line1","line2"
line2,"line2"`),
			out: 2,
		},
	}

	for _, f := range fixtures {
		t.Run(f.desc, func(t *testing.T) {
			if out := countRecords(f.data); out != f.out {
				t.Errorf("want=%d; got %d", f.out, out)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	fixtures := []struct {
		desc string
		v    interface{}
		out  [][]string
		err  error
	}{
		{
			desc: "slice with basic type",
			v: []TypeI{
				{String: "string", Int: 10},
				{String: "", Int: 0},
			},
			out: [][]string{
				{"String", "int"},
				{"string", "10"},
				{"", ""},
			},
		},
		{
			desc: "slice with pointer type",
			v: []*TypeI{
				{String: "string", Int: 10},
				{String: "", Int: 0},
			},
			out: [][]string{
				{"String", "int"},
				{"string", "10"},
				{"", ""},
			},
		},
		{
			desc: "slice pointer",
			v: &[]*TypeI{
				{String: "string", Int: 10},
			},
			out: [][]string{
				{"String", "int"},
				{"string", "10"},
			},
		},
		{
			desc: "not a slice",
			v:    int64(1),
			err:  &InvalidMarshalError{Type: reflect.TypeOf(int64(1))},
		},
		{
			desc: "slice of non pointers",
			v:    []int64{1},
			err:  &InvalidEncodeError{Type: reflect.TypeOf(int64(1))},
		},
	}

	for _, f := range fixtures {
		t.Run(f.desc, func(t *testing.T) {
			b, err := Marshal(f.v)
			if f.err != nil {
				if !reflect.DeepEqual(f.err, err) {
					t.Errorf("want err=%v; got %v", f.err, err)
				}
				return
			} else if err != nil {
				t.Errorf("want err=nil; got %v", err)
			}

			if expected := encodeCSV(t, f.out); string(b) != expected {
				t.Errorf("want %s; got %s", expected, string(b))
			}
		})
	}
}
