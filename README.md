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

# A compiler:
```
go build ./cmd/compiler/arm/main.go
```

Compiling a .kl file:
```
./main -o <output file> calculator.kl
# output file's default parameter is "out"
./main calculator.kl
```

### The compiled binary is approximately 30 times faster than the interpreter:
Compiler test:
```
komarulang % time ./out 
./out  0.01s user 0.00s system 2% cpu 0.466 total
```
Interpreter test:
```
komarulang % time ./main calculator.kl
./main calculator.kl  0.31s user 0.01s system 99% cpu 0.320 total
```
On a second thought, increasing a data set for benchmarking has led to a huge performance gap between a compiler and an interpreter.
Compiler's just much much faster. And I have no idea how much, but I've got the urge to say "🚀 blazingly fast".

