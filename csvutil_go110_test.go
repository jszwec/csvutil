// +build go1.10

package csvutil

import (
	"encoding/csv"
	"reflect"
)

var testUnmarshalInvalidFirstLineErr = &csv.ParseError{
	StartLine: 1,
	Line:      1,
	Column:    1,
	Err:       csv.ErrQuote,
}

var testUnmarshalInvalidSecondLineErr = &csv.ParseError{
	StartLine: 2,
	Line:      2,
	Column:    1,
	Err:       csv.ErrQuote,
}

var ptrUnexportedEmbeddedDecodeErr = errPtrUnexportedStruct(reflect.TypeOf(new(embedded)))
