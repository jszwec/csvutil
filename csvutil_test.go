package csvutil

import (
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
		})
	}

	t.Run("test invalid arguments", func(t *testing.T) {
		var fixtures = []struct {
			desc     string
			v        interface{}
			expected string
		}{
			{"nil interface", interface{}(nil), "csvutil: Unmarshal(nil)"},
			{"nil", nil, "csvutil: Unmarshal(nil)"},
			{"non pointer struct", struct{}{}, "csvutil: Unmarshal(non-pointer struct {})"},
			{"non-slice pointer", (*int)(nil), "csvutil: Unmarshal(non-slice pointer)"},
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
