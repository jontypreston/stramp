# Stramp


## What?
Convert a `struct` to a flat `map[string]string`.

## Why?
Golang's `struct` is a natural representation for semantically related key-value pairs.
Lots of key-value stores (like *etcd*) don't have native support for nested structures.
Stramp flattens a possibly nested `struct` into a flat key-value map.

## How?
Define your `struct`.
Any fields which are not exported or are missing the `stramp` tag are ignored.

```golang
type Person struct {
    Name      Name     `stramp:"name"`
    Nicknames []string `stramp:"nicknames"`
    Age       int      `stramp:"age"`
}

type Name struct {
    FirstName string `stramp:"first_name"`
    Surname   string `stramp:"surname"`
}
```

```golang
a := Person{
    Name: Name{
        FirstName: "John",
        Surname:   "Smith",
    },
    Nicknames: []string{"Jonny", "James"},
    Age:       55,
}

kv, _ := stramp.Stramp(a)
```

```golang
b := Person{}

_ = stramp.DeStramp(kv, &b)
```

## The Basics
### Supported Types
 - [x] int, int8, int16, int32, int64
 - [x] uint
 - [ ] float32, float64
 - [x] string
 - [ ] bool
 - [x] slice (int | float64 | string)
 - [x] struct

### Keys
#### Fields
The key for each struct field is given by its tag.

```golang
type A struct {
    Foo string `stramp:"foo"`
}
```

If the tag `etcd` clashes, you can change it:
```golang
stramp.TagKey = "something"
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

If you change the `IndexKey` function, you should also change the inverse `KeyIndex` function to match.
Note that `KeyIndex` is not used within `stramp.Stramp` nor `stramp.DeStramp`, thus you need only change it if you use the getter or setter functions.

## TODO
 - [ ] Unit Testing
 - [ ] Support *all* fundamental types
 - [ ] Support common built-in types (like `time.Duration`)
 - [x] Refactor project to make better use of [CoR](https://refactoring.guru/design-patterns/chain-of-responsibility)