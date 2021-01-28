package main

import (
	"fmt"
	"github.com/modular-id/stramp"
)

type Person struct {
	Name      Name     `stramp:"name"`
	Nicknames []string `stramp:"nicknames"`
	Age       int      `stramp:"age"`
}

type Name struct {
	FirstName string `stramp:"first_name"`
	Surname   string `stramp:"surname"`
}

func main() {
	a := Person{
		Name: Name{
			FirstName: "John",
			Surname:   "Smith",
		},
		Nicknames: []string{"Jonny", "James"},
		Age:       55,
	}

	fmt.Printf("A: %v \n", a)

	kv, err := stramp.Stramp(a)

	if err != nil {
		panic(err)
	}

	fmt.Printf("A: %v \n", kv)

	b := Person{}

	if err := stramp.DeStramp(kv, &b); err != nil {
		panic(err)
	}

	fmt.Printf("A: %v \n", b)

	surname, err := stramp.Get("name.surname", b)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Surname: %s\n", surname)
}
