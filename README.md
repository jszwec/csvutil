csvutil [![GoDoc](https://godoc.org/github.com/jszwec/csvutil?status.svg)](http://godoc.org/github.com/jszwec/csvutil) [![Build Status](https://travis-ci.org/jszwec/csvutil.svg?branch=master)](https://travis-ci.org/jszwec/csvutil) [![Build status](https://ci.appveyor.com/api/projects/status/eiyx0htjrieoo821/branch/master?svg=true)](https://ci.appveyor.com/project/jszwec/csvutil/branch/master)
=================

package csvutil provides a fast and idiomatic way to decode csv inputs.

Installation
------------

    go get github.com/jszwec/csvutil

Performance
------------

Decoder uses internal type caching to increase the performance after the first Decode call.

csvutil:
```
BenchmarkUnmarshal/csvutil.Unmarshal/1_record-8               200000         11349 ns/op        7753 B/op          64 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10_record-8               50000         30569 ns/op       11097 B/op         181 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100_record-8              10000        222964 ns/op       44554 B/op        1351 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/1000_record-8              1000       2121605 ns/op      379934 B/op       13051 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10000_record-8               50      21779523 ns/op     3737566 B/op      130051 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100000_record-8               5     244131921 ns/op    38752902 B/op     1300053 allocs/op
```

gocsv:
```
BenchmarkUnmarshal/gocsv.Unmarshal/1_record-8                 100000         11683 ns/op        7475 B/op         111 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/10_record-8                 30000         44868 ns/op       13451 B/op         402 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/100_record-8                 5000        359964 ns/op       71004 B/op        3285 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/1000_record-8                 500       3460573 ns/op      636650 B/op       32088 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/10000_record-8                 30      37324845 ns/op     6904008 B/op      320097 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/100000_record-8                 3     410029809 ns/op    75267434 B/op     3200112 allocs/op
```

Example
--------

```
	var csvInput = []byte(`name,age
jacek,26
john,27`)

	type User struct {
		Name string `csv:"name"`
		Age  int    `csv:"age"`
	}

	var users []User
	if err := csvutil.Unmarshal(csvInput, &users); err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v", users)

	// Output:
	// [{Name:jacek Age:26} {Name:john Age:27}]	
```
