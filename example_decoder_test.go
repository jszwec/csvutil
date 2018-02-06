package csvutil_test

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/jszwec/csvutil"
)

func ExampleDecoder_Decode() {
	type User struct {
		ID   *int   `csv:"id,omitempty"`
		Name string `csv:"name"`
		City string `csv:"city"`
		Age  int    `csv:"age"`
	}

	csvReader := csv.NewReader(strings.NewReader(`
id,name,age,city
,alice,25,la
,bob,30,ny`))

	dec, err := csvutil.NewDecoder(csvReader)
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
}

func ExampleDecoder_Unused() {
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
}

func ExampleDecoder_decodeEmbedded() {
	type Address struct {
		ID    int    `csv:"id"` // same field as in User - this one will be empty
		City  string `csv:"city"`
		State string `csv:"state"`
	}

	type User struct {
		Address
		ID   int    `csv:"id"` // same field as in Address - this one wins
		Name string `csv:"name"`
		Age  int    `csv:"age"`
	}

	csvReader := csv.NewReader(strings.NewReader(
		"id,name,age,city,state\n" +
			"1,alice,25,la,ca\n" +
			"2,bob,30,ny,ny"))

	dec, err := csvutil.NewDecoder(csvReader)
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
	// [{{0 la ca} 1 alice 25} {{0 ny ny} 2 bob 30}]
}
