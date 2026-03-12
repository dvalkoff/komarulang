package main

import (
	"fmt"
	"os"

	"github.com/dvalkoff/komarulang/parser"
	"github.com/dvalkoff/komarulang/tokenizer"
)

var globalVariables map[string]any = map[string]any{}

func interpretStmt(stmt parser.Statement) {
	switch typed := stmt.(type) {
	case parser.VarDeclaration:
		identifier := typed.Identifier
		if _, ok := globalVariables[identifier]; ok {
			panic(fmt.Sprintf("variable %v already exist", identifier))
		}
		value := evaluate(typed.Expr)
		globalVariables[identifier] = value
	case parser.VarAssignment:
		identifier := typed.Identifier
		if _, ok := globalVariables[identifier]; !ok {
			panic(fmt.Sprintf("variable %v does not exist", identifier))
		}
		value := evaluate(typed.Expr)
		globalVariables[identifier] = value
	case parser.ExprStatement:
		evaluate(typed.Expr)
	case parser.PrintStatement:
		result := evaluate(typed.Expr)
		fmt.Println(result)
	}
}

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
	case parser.IdentifierLiteral:
		if value, ok := globalVariables[typed.Value]; ok {
			return value
		} else {
			panic(fmt.Sprintf("variable %v does not exist", typed.Value))
		}
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
	prog, err := p.Parse()
	if err != nil {
		panic(err)
	}
	for _, decl := range prog {
		interpretStmt(decl)
	}
}