# Language

```
// comment line
var a = 1 // variable declaration, integer
var b = true // bool

a = 2 // reassignment

{ // code block. variables are local
    var d = 3
    // arithmetic operations
    var c = d + a
    c = c * a
    c = c / a
    c = -c
    c = c % 2
}

{
    var e = true
    var f = false
    // logical operations
    print(e && f)
    print(e || f)
}

{
    var e = 128
    var f = 64
    // bitwise operations
    print(e & f)
    print(e | f)
    print(e ^ f)
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

{
    // while loops
    var e = 0
    while e < 5 {
        print(e)
        e = e + 1
    }
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

