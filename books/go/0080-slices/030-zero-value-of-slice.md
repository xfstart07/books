---
Title: Zero value of slice
Id: 62
SOId: 6807
---
[Zero value](29) of slice is `nil`.

A `nil` slice has [length and capacity](61) of 0.

A `nil` slice has no underlying array.

A non-nil slice can also have length and capacity of 0, like `[]int{}` or `make([]int, 5)[5:]`.

Any type that has `nil` values can be converted to `nil` slice:

```
s = []int(nil)
```

To test whether a slice is empty, use:

```go
if len(s) == 0　{
    fmt.Print("s is empty.\n")
}
```

This works for both `nil` and non-nil slice with no elements.
