package csvutil_test

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/jszwec/csvutil"
)

func ExampleEncoder_encode() {
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
