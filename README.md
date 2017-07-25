recenc [![GoDoc](https://godoc.org/github.com/jszwec/recenc?status.svg)](http://godoc.org/github.com/jszwec/recenc) [![Build Status](https://travis-ci.org/jszwec/recenc.svg?branch=master)](https://travis-ci.org/jszwec/recenc) [![Build status](https://ci.appveyor.com/api/projects/status/6t4i7j31he1pdsj9?svg=true)](https://ci.appveyor.com/project/jszwec/recenc)
=================

Package recenc provides fast and idiomatic way to decode string records such as CSV to struct types.

A string record, such as CSV, is held in []string type. Reader interface
defined in this package can read such records. The example implementation
that satisfies this interface is: csv.Reader.

```
type Reader interface {
	Read() ([]string, error)
}
```

Decoder uses Reader to read new records and unmarshal them into the given
struct type.

Installation
------------

    go get github.com/jszwec/recenc

Performance
------------

Decoder uses internal type caching to increase the performance after the first Decode call.

```
BenchmarkDecode/10_field_struct_first_decode-8           200000          7800 ns/op        1588 B/op          23 allocs/op
BenchmarkDecode/10_field_struct_second_decode-8         2000000           810 ns/op           0 B/op           0 allocs/op
```

```
BenchmarkDecode/10_field_struct_1_record-8               200000          7799 ns/op        1588 B/op          23 allocs/op
BenchmarkDecode/10_field_struct_10_records-8             100000         12829 ns/op        1588 B/op          23 allocs/op
BenchmarkDecode/10_field_struct_100_records-8             20000         63463 ns/op        1588 B/op          23 allocs/op
BenchmarkDecode/10_field_struct_1000_records-8             3000        574202 ns/op        1587 B/op          23 allocs/op
BenchmarkDecode/10_field_struct_10000_records-8             300       5582384 ns/op        1589 B/op          23 allocs/op
```

Example
--------

Simple CSV

```
	type User struct {
		ID   *int   `recenc:"id,omitempty"`
		Name string `recenc:"name"`
		City string `recenc:"city"`
		Age  int    `recenc:"age"`
	}

	csvReader := csv.NewReader(strings.NewReader(
		"id,name,age,city\n" +
			",alice,25,la\n" +
			",bob,30,ny\n"))

	dec, err := recenc.NewDecoder(csvReader)
	if err != nil {
		log.Fatal(err)
	}

	var users []User
	for {
		var u User
		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		users = append(users, u)
	}

	fmt.Println(users)

	// Output:
	// [{<nil> alice la 25} {<nil> bob ny 30}]
```
