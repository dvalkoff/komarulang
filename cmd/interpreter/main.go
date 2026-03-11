package main

import (
	"fmt"
	"os"

	"github.com/dvalkoff/komarulang/parser"
	"github.com/dvalkoff/komarulang/tokenizer"
)

func evaluate(ast parser.Expression) any {
	switch typed := ast.(type) {
	case parser.BinaryExpression:
		return evaluateBinaryOperation(evaluate(typed.Left), evaluate(typed.Right), typed.Operator)
	case parser.UnaryExpression:
		return evaluateUnaryOperation(evaluate(typed.Right), typed.Operator)
	case parser.BooleanLiteral:
		return typed.Value
	case parser.IntegerLiteral:
		return typed.Value
	}
	return 0
}

func evaluateUnaryOperation(rightOperand any, operator tokenizer.TokenType) any {
	switch operator {
	case tokenizer.Minus:
		right := rightOperand.(int)
		return -right
	case tokenizer.Bang:
		right := rightOperand.(bool)
		return !right
	}
	panic(fmt.Sprintf("can not execute operation %v on: %v", operator, rightOperand))
}

func evaluateBinaryOperation(leftOperand, rightOperand any, operator tokenizer.TokenType) any {
	
	switch operator {
	case tokenizer.Plus:
		left, right := leftOperand.(int), rightOperand.(int)
		return left + right
	case tokenizer.Minus:
		left, right := leftOperand.(int), rightOperand.(int)
		return left - right
	case tokenizer.Star:
		left, right := leftOperand.(int), rightOperand.(int)
		return left * right
	case tokenizer.Slash:
		left, right := leftOperand.(int), rightOperand.(int)
		return left / right
	case tokenizer.Percent:
		left, right := leftOperand.(int), rightOperand.(int)
		return left % right

	case tokenizer.Less:
		left, right := leftOperand.(int), rightOperand.(int)
		return left < right
	case tokenizer.LessEqual:
		left, right := leftOperand.(int), rightOperand.(int)
		return left <= right
	case tokenizer.Greater:
		left, right := leftOperand.(int), rightOperand.(int)
		return left > right
	case tokenizer.GreaterEqual:
		left, right := leftOperand.(int), rightOperand.(int)
		return left >= right

	case tokenizer.EqualEqual:
		return leftOperand == rightOperand
	case tokenizer.BangEqual:
		return leftOperand != rightOperand
	}
	panic(fmt.Sprintf("can not execute operation %v on left: %v and right %v", operator, leftOperand, rightOperand))
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
	fmt.Println(evaluate(tree))
}