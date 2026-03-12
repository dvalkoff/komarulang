# Language

```
// comment line
var a = 1 // variable declaration, integer
var b = true // bool

a = 2 // reassignment

{ // code block. variables are local
    var b = 3
    // arithmetic operations
    var c = b + a
    c = c * a
    c = c / a
    c = -c
    c = c % 2
}

print(a) // stdout builtin "function". techically it's a keyword for now

// if else statements full support
if a > 5 {
    print(a)
} else if b {
    print(b)
} else {
    print(false)
}

```

# Interpreter launch

```
go build ./cmd/interpreter/main.go
./main calculator.kl
```

# A compiler (temporarely outdated):
```
go build ./cmd/compiler/arm/main.go
```

Compiling a .kl file:
```
./main -o <output file> calculator.kl
# output file's default parameter is "out"
./main calculator.kl
```

