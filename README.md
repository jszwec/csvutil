csvutil [![GoDoc](https://godoc.org/github.com/jszwec/csvutil?status.svg)](http://godoc.org/github.com/jszwec/csvutil) [![Build Status](https://travis-ci.org/jszwec/csvutil.svg?branch=master)](https://travis-ci.org/jszwec/csvutil) [![Build status](https://ci.appveyor.com/api/projects/status/eiyx0htjrieoo821/branch/master?svg=true)](https://ci.appveyor.com/project/jszwec/csvutil/branch/master) [![Go Report Card](https://goreportcard.com/badge/github.com/jszwec/csvutil)](https://goreportcard.com/report/github.com/jszwec/csvutil) [![codecov](https://codecov.io/gh/jszwec/csvutil/branch/master/graph/badge.svg)](https://codecov.io/gh/jszwec/csvutil)
=================

<p align="center">
  <img style="float: right;" src="https://user-images.githubusercontent.com/3941256/33054906-52b4bc08-ce4a-11e7-9651-b70c5a47c921.png"/ width=200>
</p>

Package csvutil provides fast and idiomatic mapping between CSV and Go values.

This package does not provide a CSV parser itself, it is based on the [Reader](https://godoc.org/github.com/jszwec/csvutil#Reader) and [Writer](https://godoc.org/github.com/jszwec/csvutil#Writer)
interfaces which are implemented by eg. std csv package. This gives a possibility
of choosing any other CSV writer or reader which may be more performant.

Installation
------------

    go get github.com/jszwec/csvutil

Example
--------

### Unmarshal ###

Nice and easy Unmarshal is using the std csv.Reader with its default options. Use [Decoder](https://godoc.org/github.com/jszwec/csvutil#Decoder) for streaming and more advanced use cases.

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

Marshal is using the std csv.Writer with its default options. Use [Encoder](https://godoc.org/github.com/jszwec/csvutil#Encoder) for streaming or to use a different Writer.

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

### Unmarshal and metadata ###

It may happen that your CSV input will not always have the same header. In addition
to your base fields you may get extra metadata that you would still like to store.
[Decoder](https://godoc.org/github.com/jszwec/csvutil#Decoder) provides 
[Unused](https://godoc.org/github.com/jszwec/csvutil#Decoder.Unused) method, which after each call to 
[Decode](https://godoc.org/github.com/jszwec/csvutil#Decoder.Decode) can report which header indexes 
were not used during decoding. Based on that, it is possible to handle and store all these extra values.

```go
	type User struct {
		Name      string            `csv:"name"`
		City      string            `csv:"city"`
		Age       int               `csv:"age"`
		OtherData map[string]string `csv:"-"`
	}

	csvReader := csv.NewReader(strings.NewReader(`
name,age,city,zip
alice,25,la,90005
bob,30,ny,10005`))

	dec, err := csvutil.NewDecoder(csvReader)
	if err != nil {
		log.Fatal(err)
	}

	header := dec.Header()
	var users []User
	for {
		u := User{OtherData: make(map[string]string)}

		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		for _, i := range dec.Unused() {
			u.OtherData[header[i]] = dec.Record()[i]
		}
		users = append(users, u)
	}

	fmt.Println(users)

	// Output:
	// [{alice la 25 map[zip:90005]} {bob ny 30 map[zip:10005]}]
```

### But my CSV file has no header... ###

Some CSV files have no header, but if you know how it should look like, it is
possible to define a struct and generate it. All that is left to do, is to pass
it to a decoder.

```go
	type User struct {
		ID   int
		Name string
		Age  int `csv:",omitempty"`
		City string
	}

	csvReader := csv.NewReader(strings.NewReader(`
1,John,27,la
2,Bob,,ny`))

	// in real application this should be done once in init function.
	userHeader, err := csvutil.Header(User{}, "csv")
	if err != nil {
		log.Fatal(err)
	}

	dec, err := csvutil.NewDecoder(csvReader, userHeader...)
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

	fmt.Printf("%+v", users)

	// Output:
	// [{ID:1 Name:John Age:27 City:la} {ID:2 Name:Bob Age:0 City:ny}]
```

Performance
------------

csvutil provides the best encoding and decoding performance with small memory usage.

### Unmarshal ###

benchmark code: https://gist.github.com/jszwec/e8515e741190454fa3494bcd3e1f100f

csvutil:
```
BenchmarkUnmarshal/csvutil.Unmarshal/1_record-8         	  300000	      5852 ns/op	    6900 B/op	      32 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10_records-8       	  100000	     13946 ns/op	    7924 B/op	      41 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100_records-8      	   20000	     95234 ns/op	   18100 B/op	     131 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/1000_records-8     	    2000	    903502 ns/op	  120652 B/op	    1031 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10000_records-8    	     200	   9273741 ns/op	 1134694 B/op	   10031 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100000_records-8   	      20	  94125839 ns/op	11628908 B/op	  100031 allocs/op
```

gocsv:
```
BenchmarkUnmarshal/gocsv.Unmarshal/1_record-8           	  200000	     10363 ns/op	    7651 B/op	      96 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/10_records-8         	   50000	     31308 ns/op	   13747 B/op	     306 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/100_records-8        	   10000	    237417 ns/op	   72499 B/op	    2379 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/1000_records-8       	     500	   2264064 ns/op	  650135 B/op	   23082 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/10000_records-8      	      50	  24189980 ns/op	 7023592 B/op	  230091 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/100000_records-8     	       5	 264797120 ns/op	75483184 B/op	 2300104 allocs/op
```

easycsv:
```
BenchmarkUnmarshal/easycsv.ReadAll/1_record-8           	  100000	     13287 ns/op	    8855 B/op	      81 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/10_records-8         	   20000	     66767 ns/op	   24072 B/op	     391 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/100_records-8        	    3000	    586222 ns/op	  170537 B/op	    3454 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/1000_records-8       	     300	   5630293 ns/op	 1595662 B/op	   34057 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/10000_records-8      	      20	  60513920 ns/op	18870410 B/op	  340068 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/100000_records-8     	       2	 623618489 ns/op	190822456 B/op	 3400084 allocs/op
```

### Marshal ###

benchmark code: https://gist.github.com/jszwec/31980321e1852ebb5615a44ccf374f17

csvutil:
```
BenchmarkMarshal/csvutil.Marshal/1_record-8         	  300000	      5501 ns/op	    6336 B/op	      26 allocs/op
BenchmarkMarshal/csvutil.Marshal/10_records-8       	  100000	     20647 ns/op	    7248 B/op	      36 allocs/op
BenchmarkMarshal/csvutil.Marshal/100_records-8      	   10000	    174656 ns/op	   24656 B/op	     127 allocs/op
BenchmarkMarshal/csvutil.Marshal/1000_records-8     	    1000	   1697202 ns/op	  164961 B/op	    1029 allocs/op
BenchmarkMarshal/csvutil.Marshal/10000_records-8    	     100	  16995940 ns/op	 1522412 B/op	   10032 allocs/op
BenchmarkMarshal/csvutil.Marshal/100000_records-8   	      10	 172411108 ns/op	22363382 B/op	  100036 allocs/op
```

gocsv:
```
BenchmarkMarshal/gocsv.Marshal/1_record-8           	  200000	      7202 ns/op	    5922 B/op	      83 allocs/op
BenchmarkMarshal/gocsv.Marshal/10_records-8         	   50000	     31821 ns/op	    9427 B/op	     390 allocs/op
BenchmarkMarshal/gocsv.Marshal/100_records-8        	    5000	    285885 ns/op	   52773 B/op	    3451 allocs/op
BenchmarkMarshal/gocsv.Marshal/1000_records-8       	     500	   2806405 ns/op	  452517 B/op	   34053 allocs/op
BenchmarkMarshal/gocsv.Marshal/10000_records-8      	      50	  28682052 ns/op	 4412157 B/op	  340065 allocs/op
BenchmarkMarshal/gocsv.Marshal/100000_records-8     	       5	 286836492 ns/op	51969227 B/op	 3400083 allocs/op
```
