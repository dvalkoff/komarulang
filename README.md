# Language

```
// comment line
var a = 1 // variable declaration, integer
var b = true // bool

a = 2 // reassignment

print(a) // stdout builtin "function". techically it's a keyword for now

{ // code block. variables are local
    var d = 3
    // arithmetic operations
    var c = d + a
    c = c * a
    c = c / a
    c = -c
    c = c % 2
    // comparison
    print(a > d)
    print(a < d)
    print(a >= d)
    print(a <= d)
    print(a != d)
    print(a == d)
}

{
    var e = true
    var f = false
    // logical operations
    print(e && f)
    print(e || f)
    // bool negation
    print(!e)
}

{
    var e = 128
    var f = 64
    // bitwise operations
    print(e & f)
    print(e | f)
    print(e ^ f)
}

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


{
    // for loops
    for var i = 0; i < 10; i = i + 1 {
        print(i)
    }

    var i = 0 // variable doesn't have to be declared inside for statement
    for i = 1; i <= 10; i = i + 1 {
        print(i)
    }
}

{
    var x int = 1 // declaring a variable with a specific type
    var y = 1 + x // type inference
    // y = false - compile time type error
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

