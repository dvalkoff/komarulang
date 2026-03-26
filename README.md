# komarulang

A statically-typed, compiled programming language that generates ARM64 assembly. Designed with a clean, Go-inspired syntax and first-class support for pointer arithmetic and manual memory management.

---

## Features

- Compiles to ARM64 assembly
- Static typing with type inference
- Functions with recursion support
- `for` and `while` loops
- `if / else if / else` control flow
- Scoped code blocks
- Arithmetic, comparison, logical, and bitwise operators
- Pointer types and pointer arithmetic
- Manual heap memory management (`malloc` / `free`)
- Built-in `print` statement

---

## Building the Compiler

Requires [Go](https://go.dev/) to be installed.

```bash
go build ./cmd/compiler/arm/main.go
```

This produces the `main` binary — the komarulang compiler.

---

## Compiling a `.kl` File

```bash
# Default output file is named "out"
./main calculator.kl

# Specify a custom output file name
./main -o <output_file> calculator.kl
```

The output is a binary.

---

## Language Reference

### Variables

Variables are declared with `var`. Type annotations are optional — the compiler infers types from the assigned value.

```kl
var a = 1          // integer, inferred
var b = true       // bool, inferred
var x int = 1      // explicit type annotation
var y = 1 + x      // type inference from expression
```

Reassignment uses plain assignment (no keyword):

```kl
a = 2
```

### Functions

Functions are declared with the `fun` keyword. Parameters require explicit types; the return type follows the parameter list.

```kl
fun add(a int, b int) int {
    return a + b
}

fun greet() {
    print(42)
}
```

Recursion is fully supported:

```kl
fun FibonacciRecursive(n int) int {
    if n == 0 { return 0 }
    if n == 1 { return 1 }
    var r1 = FibonacciRecursive(n - 1)
    var r2 = FibonacciRecursive(n - 2)
    return r1 + r2
}
```

### Control Flow

**if / else if / else**

```kl
if a > 5 {
    print(a)
} else if b {
    print(b)
} else {
    print(0)
}
```

**while loop**

```kl
var e = 0
while e < 5 {
    print(e)
    e = e + 1
}
```

**for loop**

```kl
for var i = 0; i < 10; i = i + 1 {
    print(i)
}

// The loop variable can be declared outside
var i = 0
for i = 1; i <= 10; i = i + 1 {
    print(i)
}
```

### Operators

**Arithmetic**

| Operator | Description    |
|----------|----------------|
| `+`      | Addition       |
| `-`      | Subtraction / Negation |
| `*`      | Multiplication |
| `/`      | Division       |
| `%`      | Modulo         |

**Comparison**

| Operator | Description           |
|----------|-----------------------|
| `==`     | Equal                 |
| `!=`     | Not equal             |
| `<`      | Less than             |
| `>`      | Greater than          |
| `<=`     | Less than or equal    |
| `>=`     | Greater than or equal |

**Logical**

| Operator | Description |
|----------|-------------|
| `&&`     | Logical AND |
| `\|\|`   | Logical OR  |
| `!`      | Logical NOT |

**Bitwise**

| Operator | Description |
|----------|-------------|
| `&`      | Bitwise AND |
| `\|`     | Bitwise OR  |
| `^`      | Bitwise XOR |

### Scoped Blocks

Variables declared inside a block are local to it:

```kl
{
    var d = 3
    // d is only accessible here
}
```

### Comments

```kl
// This is a single-line comment
```

---

## Pointers and Heap Memory

komarulang supports manual memory management through pointer types and the built-in `malloc` / `free` functions.

### Pointer Types

Use `*T` to declare a pointer to type `T`. Dereference with `*`.

```kl
var arrPointer = malloc(size)
var arrTmp *int = arrPointer

*arrTmp = 42          // write through pointer
var val = *arrTmp     // read through pointer
arrTmp = arrTmp + 8   // advance pointer by one int (8 bytes)
```

### Built-in Memory Functions

| Function       | Description                             |
|----------------|-----------------------------------------|
| `malloc(size)` | Allocates `size` bytes on the heap, returns a pointer |
| `free(ptr)`    | Frees a previously allocated heap block |

### Example: Allocating and Sorting an Array

```kl
fun selectionSort(arrPtr *int, size int) {
    var ptrI = arrPtr
    for var i = 0; i < size - 1; i = i + 1 {
        var minPtr = ptrI
        var minPtrVal = *minPtr

        var ptrJ = ptrI
        for var j = i; j < size; j = j + 1 {
            if *ptrJ < minPtrVal {
                minPtr = ptrJ
                minPtrVal = *ptrJ
            }
            ptrJ = ptrJ + 8
        }

        var temp int = *ptrI
        *ptrI = minPtrVal
        *minPtr = temp

        ptrI = ptrI + 8
    }
}

fun main() {
    var arrSize = 1000 * 100
    var intSize = 8
    var arrPointer = malloc(arrSize * intSize)

    var arrTmp *int = arrPointer
    for var i = 0; i < arrSize; i = i + 1 {
        *arrTmp = i * 12345 + 1  // fill with values
        arrTmp = arrTmp + intSize
    }

    selectionSort(arrPointer, arrSize)
    free(arrPointer)
}
```

> **Note:** Integers are 8 bytes (64-bit) on ARM64. When advancing a pointer through an `int` array, increment by `8`.

---

## Built-ins

| Name         | Type     | Description                              |
|--------------|----------|------------------------------------------|
| `print(val)` | Keyword  | Prints a value to stdout                 |
| `malloc(n)`  | Function | Allocates `n` bytes, returns a pointer   |
| `free(ptr)`  | Function | Frees memory at the given pointer        |

---

## Full Example

```kl
fun Fibonacci(n int) int {
    var a = 0
    var b = 1
    for var i = 0; i < n; i = i + 1 {
        var temp = a + b
        a = b
        b = temp
    }
    return a
}

fun main() {
    var result = Fibonacci(10)
    print(result)   // 55
}
```
