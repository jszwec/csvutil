package csvutil

// Reader provides the interface for reading a single record.
//
// If there is no data left to be read, Read returns nil, io.EOF.
//
// It is implemented by csv.Reader.
type Reader interface {
	Read() ([]string, error)
}

// Unmarshaler is the interface implemented by types that can unmarshal
// a single record's field description of themselves.
type Unmarshaler interface {
	UnmarshalCSV(string) error
}
