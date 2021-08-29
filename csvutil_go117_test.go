//go:build go1.17
// +build go1.17

package csvutil

import (
	"encoding/csv"
	"reflect"
)

// In Go1.17 csv.ParseError.Column became 1-indexed instead of 0-indexed.
// so we need this file for Go 1.17+.

var testUnmarshalInvalidFirstLineErr = &csv.ParseError{
	StartLine: 1,
	Line:      1,
	Column:    2,
	Err:       csv.ErrQuote,
}

var testUnmarshalInvalidSecondLineErr = &csv.ParseError{
	StartLine: 2,
	Line:      2,
	Column:    2,
	Err:       csv.ErrQuote,
}

var ptrUnexportedEmbeddedDecodeErr = errPtrUnexportedStruct(reflect.TypeOf(new(embedded)))
