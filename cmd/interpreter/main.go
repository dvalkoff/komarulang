package main

import (
	"fmt"
	"os"

	"github.com/dvalkoff/komarulang/parser"
	"github.com/dvalkoff/komarulang/tokenizer"
	"github.com/dvalkoff/komarulang/tokenizer/token"
)

func evaluate(ast parser.Expression) any {
	switch typed := ast.(type) {
	case parser.BinaryExpression:
		return evaluateBinaryOperation(evaluate(typed.Left), evaluate(typed.Right), typed.Operator)
	case parser.IntegerLiteral:
		return typed.Value
	}
	return 0
}

func evaluateBinaryOperation(leftOperand, rightOperand any, operator token.TokenType) any {
	left, right := leftOperand.(int), rightOperand.(int)
	switch operator {
	case token.Plus:
		return left + right
	case token.Minus:
		return left - right
	case token.Star:
		return left * right
	case token.Slash:
		return left / right

	case token.Less:
		return left < right
	case token.LessEqual:
		return left <= right
	case token.Greater:
		return left > right
	case token.GreaterEqual:
		return left >= right

	case token.EqualEqual:
		return leftOperand == rightOperand
	case token.BangEqual:
		return leftOperand != rightOperand
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