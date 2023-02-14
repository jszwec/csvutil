//go:build race
// +build race

package csvutil

import (
	"bytes"
	"encoding/csv"
	"io"
	"sync"
	"testing"
)

func TestCacheDataRaces(t *testing.T) {
	const routines = 50
	const rows = 1000

	v := TypeF{
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
		V:        pptr(100),
		Pv:       ptr[any](pptr(200)),
		Binary:   Binary,
		PBinary:  &Binary,
	}

	t.Run("encoding", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < routines; i++ {
			tag := "csv"
			if i%2 == 0 {
				tag = "custom"
			}

			wg.Add(1)
			go func() {
				defer wg.Done()

				var buf bytes.Buffer
				w := csv.NewWriter(&buf)
				enc := NewEncoder(w)
				enc.Tag = tag
				for i := 0; i < rows; i++ {
					if err := enc.Encode(v); err != nil {
						panic(err)
					}
				}
				w.Flush()
			}()
		}
		wg.Wait()
	})

	t.Run("decoding", func(t *testing.T) {
		vs := make([]*TypeF, 0, rows)
		for i := 0; i < rows; i++ {
			vs = append(vs, &v)
		}

		data, err := Marshal(vs)
		if err != nil {
			t.Fatal(err)
		}

		var wg sync.WaitGroup
		for i := 0; i < routines; i++ {
			tag := "csv"
			if i%2 == 0 {
				tag = "custom"
			}

			wg.Add(1)
			go func() {
				defer wg.Done()

				dec, err := NewDecoder(csv.NewReader(bytes.NewReader(data)))
				if err != nil {
					t.Fatal(err)
				}
				dec.Tag = tag

				for {
					var val TypeF
					if err := dec.Decode(&val); err == io.EOF {
						break
					} else if err != nil {
						panic(err)
					}
				}
			}()
		}
		wg.Wait()
	})
}
