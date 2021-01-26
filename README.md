# Stramp

## What?
Convert a `struct` to a flat `map[string]string`.

## Why?
Golang's `struct` is a natural representation for semantically related key-value pairs.
Lots of key-value stores (like *etcd*) don't have native support for nested structures.
Stramp flattens a possibly nested `struct` into a flat key-value map.

## How?
Define your `struct`.
Any fields which are not exported or are missing the `etcd` tag are ignored.

```golang
type Person struct {
	Name Name `etcd:"name"`
	Nicknames []string `etcd:"nicknames"`
	Age int `etcd:"age"`
}

type Name struct {
	FirstName string `etcd:"first_name"`
	Surname string `etcd:"surname"`
}
```

```golang
a := Person{
    Name: Name{
        FirstName: "John",
        Surname: "Smith",
    },
    Nicknames: []string{"Jonny", "James"},
    Age:  55,
}

kv, _ := stramp.Stramp(a)
```

```golang
b := Person{}

_ = stramp.DeStramp(&b, kv)
```

## The Basics
### Supported Types
 - [x] int
 - [ ] uint
 - [ ] float32 
 - [x] float64
 - [x] string
 - [ ] bool
 - [x] slice (int | float64 | string)

### Keys
#### Fields
The key for each struct field is given by its tag.

```golang
type A struct {
	Foo string `etcd:"foo"`
}
```

If the tag `etcd` clashes, you can change it:
```golang
stramp.TagKey = "stramp"
```

#### Nesting
Nesting is represented in the keys using a separator.
The default is `.` (a dot).

```golang
// Change separator to a forward slash
stramp.Sep = "/"
```

Slice elements are given a key based on their index.
By default, the index is used unchanged.

```golang
// Change index key representation to use square brackets
stramp.IndexKey = func(i int) string {
	return fmt.Sprintf("[%d]", i)
}
```