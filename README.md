Building a compiler:
```
go build ./cmd/compiler/arm/main.go
```

Compiling a .kl file:
```
./main -o <output file> calculator.kl
# output file's default parameter is "out"
./main calculator.kl
```

