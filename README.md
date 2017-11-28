csvutil [![GoDoc](https://godoc.org/github.com/jszwec/csvutil?status.svg)](http://godoc.org/github.com/jszwec/csvutil) [![Build Status](https://travis-ci.org/jszwec/csvutil.svg?branch=master)](https://travis-ci.org/jszwec/csvutil) [![Build status](https://ci.appveyor.com/api/projects/status/eiyx0htjrieoo821/branch/master?svg=true)](https://ci.appveyor.com/project/jszwec/csvutil/branch/master) [![Go Report Card](https://goreportcard.com/badge/github.com/jszwec/csvutil)](https://goreportcard.com/report/github.com/jszwec/csvutil)
=================

<p align="center">
  <img style="float: right;" src="https://user-images.githubusercontent.com/3941256/33054906-52b4bc08-ce4a-11e7-9651-b70c5a47c921.png"/ width=200>
</p>

package csvutil provides fast and idiomatic mapping between CSV and Go values.

Installation
------------

    go get github.com/jszwec/csvutil

Example
--------

### Unmarshal ###

```go
	var csvInput = []byte(`
name,age
jacek,26
john,27`,
	)

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

### Marshal ###

```go
	type Address struct {
		City    string
		Country string
	}

	type User struct {
		Name string
		Address
		Age int `csv:"age,omitempty"`
	}

	users := []User{
		{Name: "John", Address: Address{"Boston", "USA"}, Age: 26},
		{Name: "Bob", Address: Address{"LA", "USA"}, Age: 27},
		{Name: "Alice", Address: Address{"SF", "USA"}},
	}

	b, err := csvutil.Marshal(users)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(string(b))

	// Output:
	// Name,City,Country,age
	// John,Boston,USA,26
	// Bob,LA,USA,27
	// Alice,SF,USA,
```

Performance
------------

csvutil provides the best encoding and decoding performance with small memory usage.

### Unmarshal ###

benchmark code: https://gist.github.com/jszwec/e8515e741190454fa3494bcd3e1f100f

csvutil:
```
BenchmarkUnmarshal/csvutil.Unmarshal/1_record-8         	  200000	      9407 ns/op	    7408 B/op	      44 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10_records-8       	  100000	     21384 ns/op	    8433 B/op	      53 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100_records-8      	   10000	    140172 ns/op	   18609 B/op	     143 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/1000_records-8     	    1000	   1334816 ns/op	  121183 B/op	    1043 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10000_records-8    	     100	  13203689 ns/op	 1140356 B/op	   10043 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100000_records-8   	      10	 137474932 ns/op	12048059 B/op	  100044 allocs/op
```

gocsv:
```
BenchmarkUnmarshal/gocsv.Unmarshal/1_record-8           	  200000	     10613 ns/op	    7451 B/op	      94 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/10_records-8         	   50000	     36413 ns/op	   13547 B/op	     304 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/100_records-8        	    5000	    287672 ns/op	   72300 B/op	    2377 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/1000_records-8       	     500	   2756252 ns/op	  649932 B/op	   23080 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/10000_records-8      	      50	  29407701 ns/op	 7023391 B/op	  230089 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/100000_records-8     	       5	 311860368 ns/op	75482985 B/op	 2300102 allocs/op
```

easycsv:
```
BenchmarkUnmarshal/easycsv.ReadAll/1_record-8           	  100000	     15636 ns/op	    8863 B/op	      78 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/10_records-8         	   20000	     76797 ns/op	   24080 B/op	     388 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/100_records-8        	    2000	    666465 ns/op	  170548 B/op	    3451 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/1000_records-8       	     200	   6431414 ns/op	 1595751 B/op	   34054 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/10000_records-8      	      20	  70387764 ns/op	18870418 B/op	  340065 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/100000_records-8     	       2	 737079728 ns/op	190822472 B/op	 3400081 allocs/op
```

### Marshal ###

benchmark code: https://gist.github.com/jszwec/31980321e1852ebb5615a44ccf374f17

csvutil:
```
BenchmarkMarshal/csvutil.Marshal/1_record-8         	  200000	      6010 ns/op	    6816 B/op	      28 allocs/op
BenchmarkMarshal/csvutil.Marshal/10_records-8       	  100000	     22391 ns/op	    7728 B/op	      38 allocs/op
BenchmarkMarshal/csvutil.Marshal/100_records-8      	   10000	    189905 ns/op	   25139 B/op	     129 allocs/op
BenchmarkMarshal/csvutil.Marshal/1000_records-8     	    1000	   1812082 ns/op	  165458 B/op	    1031 allocs/op
BenchmarkMarshal/csvutil.Marshal/10000_records-8    	     100	  18112811 ns/op	 1523067 B/op	   10034 allocs/op
BenchmarkMarshal/csvutil.Marshal/100000_records-8   	      10	 183706155 ns/op	22364681 B/op	  100038 allocs/op
```

gocsv:
```
BenchmarkMarshal/gocsv.Marshal/1_record-8           	  200000	      7291 ns/op	    5810 B/op	      82 allocs/op
BenchmarkMarshal/gocsv.Marshal/10_records-8         	   50000	     32093 ns/op	    9316 B/op	     389 allocs/op
BenchmarkMarshal/gocsv.Marshal/100_records-8        	    5000	    284238 ns/op	   52673 B/op	    3450 allocs/op
BenchmarkMarshal/gocsv.Marshal/1000_records-8       	     500	   2777589 ns/op	  452503 B/op	   34052 allocs/op
BenchmarkMarshal/gocsv.Marshal/10000_records-8      	      50	  28477563 ns/op	 4413044 B/op	  340064 allocs/op
BenchmarkMarshal/gocsv.Marshal/100000_records-8     	       5	 286370004 ns/op	51970707 B/op	 3400084 allocs/op
```
