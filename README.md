csvutil [![GoDoc](https://godoc.org/github.com/jszwec/csvutil?status.svg)](http://godoc.org/github.com/jszwec/csvutil) [![Build Status](https://travis-ci.org/jszwec/csvutil.svg?branch=master)](https://travis-ci.org/jszwec/csvutil) [![Build status](https://ci.appveyor.com/api/projects/status/eiyx0htjrieoo821/branch/master?svg=true)](https://ci.appveyor.com/project/jszwec/csvutil/branch/master)
=================

package csvutil provides a fast and idiomatic way to decode csv inputs.

Installation
------------

    go get github.com/jszwec/csvutil

Performance
------------

csvutil provides the best decoding performance with small memory usage.

benchmark code: https://gist.github.com/jszwec/e8515e741190454fa3494bcd3e1f100f

csvutil:
```
BenchmarkUnmarshal/csvutil.Unmarshal/1_record-8         	  200000	     10073 ns/op	    7568 B/op	      46 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10_records-8       	   50000	     24264 ns/op	    9601 B/op	      73 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100_records-8      	   10000	    163714 ns/op	   29857 B/op	     343 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/1000_records-8     	    1000	   1541715 ns/op	  233232 B/op	    3043 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10000_records-8    	     100	  15955442 ns/op	 2260436 B/op	   30043 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100000_records-8   	      10	 159592311 ns/op	23248254 B/op	  300044 allocs/op
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
