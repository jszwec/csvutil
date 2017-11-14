package csvutil

import (
	"reflect"
	"sort"
	"sync"
)

type field struct {
	typ   reflect.Type
	tag   tag
	index []int
}

type fields []field

func (fs fields) byIndex() {
	sort.Slice(fs, func(i, j int) bool {
		if len(fs[i].index) != len(fs[j].index) {
			return len(fs[i].index) < len(fs[j].index)
		}

		for ii, n := range fs[i].index {
			if n != fs[j].index[ii] {
				return n < fs[j].index[ii]
			}
		}
		return false
	})
}

type typeKey struct {
	tag string
	reflect.Type
}

var fieldCache sync.Map // map[typeKey][]field

func cachedFields(t reflect.Type, tagName string) fields {
	k := typeKey{tagName, t}
	if v, ok := fieldCache.Load(k); ok {
		return v.(fields)
	}

	f := buildFields(t, tagName)
	if f == nil {
		f = fields{}
	}

	v, _ := fieldCache.LoadOrStore(k, f)
	return v.(fields)
}

type fieldMap map[string]fields

func (m fieldMap) insert(f field) {
	fs, ok := m[f.tag.name]
	if !ok {
		m[f.tag.name] = append(fs, f)
		return
	}

	// insert only fields with the shortest path.
	if len(fs[0].index) != len(f.index) {
		return
	}

	// fields that are tagged have priority.
	if !f.tag.empty {
		m[f.tag.name] = append([]field{f}, fs...)
		return
	}

	m[f.tag.name] = append(fs, f)
}

func (m fieldMap) fields() fields {
	out := make(fields, 0, len(m))
	for _, v := range m {
		for i, f := range v {
			if f.tag.empty != v[0].tag.empty {
				v = v[:i]
				break
			}
		}
		if len(v) > 1 {
			continue
		}
		out = append(out, v[0])
	}
	out.byIndex()
	return out
}

func buildFields(typ reflect.Type, tagName string) fields {
	q := fields{{typ: typ}}
	visited := make(map[reflect.Type]bool)
	fm := make(fieldMap)

	for len(q) > 0 {
		f := q[0]
		q = q[1:]

		if visited[f.typ] {
			continue
		}
		visited[f.typ] = true

		depth := len(f.index)

		numField := f.typ.NumField()
		for i := 0; i < numField; i++ {
			sf := f.typ.Field(i)

			if sf.PkgPath != "" && !sf.Anonymous {
				// unexported field
				continue
			}

			if sf.Anonymous {
				t := sf.Type
				if t.Kind() == reflect.Ptr {
					t = t.Elem()
				}
				if sf.PkgPath != "" && t.Kind() != reflect.Struct {
					// ignore embedded unexported non-struct fields.
					continue
				}
			}

			tag := parseTag(tagName, sf)
			if tag.ignore {
				continue
			}

			ft := sf.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}

			newf := field{
				typ:   ft,
				tag:   tag,
				index: makeIndex(f.index, i),
			}

			if sf.Anonymous && ft.Kind() == reflect.Struct {
				q = append(q, newf)
				continue
			}

			fm.insert(newf)

			// look for duplicate nodes on the same level. Nodes won't be
			// revisited, so write all fields for the current type now.
			for _, v := range q {
				if len(v.index) != depth {
					break
				}
				if v.typ == f.typ {
					// other nodes can have different path.
					fm.insert(field{
						typ:   ft,
						tag:   tag,
						index: makeIndex(v.index, i),
					})
				}
			}
		}
	}
	return fm.fields()
}

func makeIndex(index []int, v int) []int {
	out := make([]int, len(index), len(index)+1)
	copy(out, index)
	return append(out, v)
}
