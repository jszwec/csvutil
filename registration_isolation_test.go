package csvutil

import (
	"bytes"
	"encoding"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestEncoderRegisterDoesNotMutateSharedMarshalers(t *testing.T) {
	marshalers := MarshalFunc(func(value int) ([]byte, error) {
		return []byte(strconv.Itoa(value + 1)), nil
	})
	var first, second bytes.Buffer
	firstWriter, secondWriter := csv.NewWriter(&first), csv.NewWriter(&second)
	a, b := NewEncoder(firstWriter), NewEncoder(secondWriter)
	a.WithMarshalers(marshalers)
	b.WithMarshalers(marshalers)
	a.Register(func(value bool) ([]byte, error) { return []byte("custom"), nil })
	value := struct {
		Number int
		Flag   bool
	}{1, true}
	if err := a.Encode(value); err != nil {
		t.Fatal(err)
	}
	if err := b.Encode(value); err != nil {
		t.Fatal(err)
	}
	firstWriter.Flush()
	secondWriter.Flush()
	if got, want := first.String(), "Number,Flag\n2,custom\n"; got != want {
		t.Fatalf("first: %q, want %q", got, want)
	}
	if got, want := second.String(), "Number,Flag\n2,true\n"; got != want {
		t.Fatalf("second: %q, want %q", got, want)
	}
}

func TestDecoderRegisterDoesNotMutateSharedUnmarshalers(t *testing.T) {
	unmarshalers := UnmarshalFunc(func(data []byte, value *int) error {
		n, err := strconv.Atoi(string(data))
		*value = n + 1
		return err
	})
	input := "Number,Flag\n1,true\n"
	a, err := NewDecoder(csv.NewReader(strings.NewReader(input)))
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewDecoder(csv.NewReader(strings.NewReader(input)))
	if err != nil {
		t.Fatal(err)
	}
	a.WithUnmarshalers(unmarshalers)
	b.WithUnmarshalers(unmarshalers)
	a.Register(func(data []byte, value *bool) error { *value = false; return nil })
	var first, second struct {
		Number int
		Flag   bool
	}
	if err := a.Decode(&first); err != nil {
		t.Fatal(err)
	}
	if err := b.Decode(&second); err != nil {
		t.Fatal(err)
	}
	if first.Number != 2 || first.Flag {
		t.Fatalf("first: %#v", first)
	}
	if second.Number != 2 || !second.Flag {
		t.Fatalf("second: %#v", second)
	}
}

type registrationFlag bool

func (f registrationFlag) Flag() bool      { return bool(f) }
func (f *registrationFlag) SetFlag(v bool) { *f = registrationFlag(v) }

type registrationFlagReader interface{ Flag() bool }
type registrationFlagWriter interface{ SetFlag(bool) }

func TestEncoderRegisterKeepsSharedInterfaceFunctionsIndependent(t *testing.T) {
	defer func() {
		if v := recover(); v != nil {
			t.Fatalf("independent registration panicked: %v", v)
		}
	}()
	marshalers := NewMarshalers(
		MarshalFunc(func(v fmt.Stringer) ([]byte, error) { return []byte(v.String()), nil }),
		MarshalFunc(func(v encoding.TextMarshaler) ([]byte, error) { return v.MarshalText() }),
		MarshalFunc(func(v Marshaler) ([]byte, error) { return v.MarshalCSV() }),
	)
	var first, second bytes.Buffer
	wa, wb := csv.NewWriter(&first), csv.NewWriter(&second)
	a, b := NewEncoder(wa), NewEncoder(wb)
	a.WithMarshalers(marshalers)
	b.WithMarshalers(marshalers)
	a.Register(func(v registrationFlagReader) ([]byte, error) { return []byte("first"), nil })
	b.Register(func(v registrationFlagReader) ([]byte, error) { return []byte("second"), nil })
	row := struct{ Flag registrationFlag }{true}
	if err := a.Encode(row); err != nil {
		t.Fatal(err)
	}
	if err := b.Encode(row); err != nil {
		t.Fatal(err)
	}
	wa.Flush()
	wb.Flush()
	if got, want := first.String(), "Flag\nfirst\n"; got != want {
		t.Fatalf("first: %q, want %q", got, want)
	}
	if got, want := second.String(), "Flag\nsecond\n"; got != want {
		t.Fatalf("second: %q, want %q", got, want)
	}
}

func TestDecoderRegisterKeepsSharedInterfaceFunctionsIndependent(t *testing.T) {
	defer func() {
		if v := recover(); v != nil {
			t.Fatalf("independent registration panicked: %v", v)
		}
	}()
	unmarshalers := NewUnmarshalers(
		UnmarshalFunc(func(data []byte, v fmt.Stringer) error { return nil }),
		UnmarshalFunc(func(data []byte, v encoding.TextUnmarshaler) error { return v.UnmarshalText(data) }),
		UnmarshalFunc(func(data []byte, v Unmarshaler) error { return v.UnmarshalCSV(data) }),
	)
	a, err := NewDecoder(csv.NewReader(strings.NewReader("Flag\nvalue\n")))
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewDecoder(csv.NewReader(strings.NewReader("Flag\nvalue\n")))
	if err != nil {
		t.Fatal(err)
	}
	a.WithUnmarshalers(unmarshalers)
	b.WithUnmarshalers(unmarshalers)
	a.Register(func(data []byte, v registrationFlagWriter) error { v.SetFlag(true); return nil })
	b.Register(func(data []byte, v registrationFlagWriter) error { v.SetFlag(false); return nil })
	var first, second struct{ Flag registrationFlag }
	if err := a.Decode(&first); err != nil {
		t.Fatal(err)
	}
	if err := b.Decode(&second); err != nil {
		t.Fatal(err)
	}
	if !first.Flag {
		t.Fatalf("first: %#v", first)
	}
	if second.Flag {
		t.Fatalf("second: %#v", second)
	}
}
