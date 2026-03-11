package main

import (
	"fmt"
	"os"

	"github.com/dvalkoff/komarulang/parser"
	"github.com/dvalkoff/komarulang/tokenizer"
	"github.com/dvalkoff/komarulang/tokenizer/token"
)

func evaluate(ast parser.Expression) int {
	switch typed := ast.(type) {
	case parser.BinaryExpression:
		return evaluateBinaryOperation(evaluate(typed.Left), evaluate(typed.Right), typed.Operator)
	case parser.IntegerLiteral:
		return typed.Value
	}
	return 0
}

func evaluateBinaryOperation(left, right int, operator token.TokenType) int {
	switch operator {
	case token.Plus:
		return left + right
	case token.Minus:
		return left - right
	}
	return 0
}

func main() {
	fileName := os.Args[1]
	tokenizer := tokenizer.Tokenizer{File: fileName}
	file, err := os.Open(fileName)
	if err != nil {
		panic("Failed to open source file")
	}
	tokens, err := tokenizer.Scan(file)
	if err != nil {
		panic(err)
	}
	p := parser.NewParser(tokens)
	tree, err := p.Expression()
	if err != nil {
		panic(err)
	}
	fmt.Println("evaluated value", evaluate(tree))
}