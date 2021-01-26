package main

import (
	"fmt"
	"github.com/jontypreston/stramp"
)

type Person struct {
	Name Name `etcd:"name"`
	Nicknames []string `etcd:"nicknames"`
	Age int `etcd:"age"`
}

type Name struct {
	FirstName string `etcd:"first_name"`
	Surname string `etcd:"surname"`
}

func main() {
	a := Person{
		Name: Name{
			FirstName: "John",
			Surname: "Smith",
		},
		Nicknames: []string{"Jonny", "James"},
		Age:  55,
	}

	kv, err := stramp.Stramp(a)

	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n\n", kv)

	b := Person{}

	if err := stramp.DeStramp(&b, kv); err != nil {
		panic(err)
	}

	fmt.Printf("%v\n\n", b)
}
