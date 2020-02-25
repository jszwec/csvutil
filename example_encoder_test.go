package csvutil_test

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/jszwec/csvutil"
)

func ExampleEncoder_Encode_streaming() {
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

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	enc := csvutil.NewEncoder(w)

	for _, u := range users {
		if err := enc.Encode(u); err != nil {
			fmt.Println("error:", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println(buf.String())

	// Output:
	// Name,City,Country,age
	// John,Boston,USA,26
	// Bob,LA,USA,27
	// Alice,SF,USA,
}

func ExampleEncoder_Encode_all() {
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

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := csvutil.NewEncoder(w).Encode(users); err != nil {
		fmt.Println("error:", err)
	}

	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println(buf.String())

	// Output:
	// Name,City,Country,age
	// John,Boston,USA,26
	// Bob,LA,USA,27
	// Alice,SF,USA,
}

func ExampleEncoder_EncodeHeader() {
	type User struct {
		Name string
		Age  int `csv:"age,omitempty"`
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	enc := csvutil.NewEncoder(w)

	if err := enc.EncodeHeader(User{}); err != nil {
		fmt.Println("error:", err)
	}

	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println(buf.String())

	// Output:
	// Name,age
}

func ExampleEncoder_Encode_inline() {
	type Address struct {
		Street string `csv:"street"`
		City   string `csv:"city"`
	}

	type User struct {
		Name        string  `csv:"name"`
		Address     Address `csv:",inline"`
		HomeAddress Address `csv:"home_address_,inline"`
		WorkAddress Address `csv:"work_address_,inline"`
		Age         int     `csv:"age,omitempty"`
	}

	users := []User{
		{
			Name:        "John",
			Address:     Address{"Washington", "Boston"},
			HomeAddress: Address{"Boylston", "Boston"},
			WorkAddress: Address{"River St", "Cambridge"},
			Age:         26,
		},
	}

	b, err := csvutil.Marshal(users)
	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Printf("%s\n", b)

	// Output:
	// name,street,city,home_address_street,home_address_city,work_address_street,work_address_city,age
	// John,Washington,Boston,Boylston,Boston,River St,Cambridge,26
}
