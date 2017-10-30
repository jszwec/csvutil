package csvutil_test

import (
	"fmt"

	"github.com/jszwec/csvutil"
)

func ExampleDecoder_unmarshal() {
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
}
