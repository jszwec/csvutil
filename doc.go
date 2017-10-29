// Package csvutil decodes string records to struct types.
//
// A string record, such as CSV, is held in []string type. Reader interface
// defined in this package can read such records. The example implementation
// that satisfies this interface is: csv.Reader.
//
// Decoder uses Reader to read new records and unmarshal them into the given
// struct type.
package csvutil
