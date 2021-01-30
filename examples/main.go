package main

import (
	"fmt"
	"github.com/modular-id/stramp"
	"reflect"
)

type Person struct {
	Name      Name     `stramp:"name"`
	Nicknames []string `stramp:"nicknames"`
	Age       float64  `stramp:"age"`
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
		Age:       55.5,
	}

	fmt.Printf("A: %v \n", a)

	kv, err := stramp.Stramp(a)

	if err != nil {
		panic(err)
	}

	fmt.Printf("A: %v \n", kv)

	p := reflect.New(reflect.TypeOf(Person{}))
	p.Elem().Set(reflect.Zero(reflect.TypeOf(Person{})))

	if err := stramp.DeStramp(kv, p.Interface()); err != nil {
		panic(err)
	}

	b := p.Elem().Interface()

	fmt.Printf("A: %v \n", b)

	surname, err := stramp.Get("name.surname", b)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Surname: %s\n", surname)
}
