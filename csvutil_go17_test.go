// +build !go1.10

package csvutil

import "encoding/csv"

var testUnmarshalInvalidFirstLineErr = &csv.ParseError{
	Line:   1,
	Column: 1,
	Err:    csv.ErrQuote,
}

var testUnmarshalInvalidSecondLineErr = &csv.ParseError{
	Line:   2,
	Column: 1,
	Err:    csv.ErrQuote,
}

var ptrUnexportedEmbeddedDecodeErr error
