# A prototype (partly) implementing the proposal golang/go#50554
This is a prototype implementing the proposed new syntax.
It works by converting Go code with the new syntax (from `.gof` files) to the old one and writing it to a new file with the same name but the `.go` extension.

Example:

main.gof:
```go
package main

import "fmt"

func main() {
    name = "Tom"
    fmt.Printf($"Hello %s{name}")
}
```

will be converted to

main.go:
```go
package main

import "fmt"

func main() {
    name = "Tom"
    fmt.Printf("Hello %s", name)
}
```

The rune sequence is ignored in any non Go code. It is not replaced inside comments and string literals.

# Current limitations:
- Does not implement multiline/raw format strings. So this won't be converted: ``fmt.Printf($`Hi %s{name}`)``.
- Does not enforce the rules lined out in the proposal; e.g that the format string has to be the last argument for a variadic parameter.
- Does not implement the variable assignment. You can convert code that assigns format strings to two variables but the result won't compile.