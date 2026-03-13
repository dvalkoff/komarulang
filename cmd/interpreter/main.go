package main

import (
	"fmt"
	"os"

	"github.com/dvalkoff/komarulang/parser"
	"github.com/dvalkoff/komarulang/tokenizer"
)

type Environment struct {
	data map[string]any
	parent *Environment
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		data: make(map[string]any),
		parent: parent,
	}
}

func (e *Environment) Exists(key string) bool {
	if e == nil {
		return false
	}
	if _, ok := e.data[key]; ok {
		return true
	}
	return e.parent.Exists(key)
}

func (e *Environment) Get(key string) (any, bool) {
	if e == nil {
		return nil, false
	}
	if val, ok := e.data[key]; ok {
		return val, true
	}
	return e.parent.Get(key)
}

func (e *Environment) New(key string, val any) {
	e.data[key] = val
}

func (e *Environment) Set(key string, val any) {
	if _, ok := e.data[key]; ok {
		e.data[key] = val
		return
	}
	e.parent.Set(key, val)
}


func interpretStmt(env *Environment, stmt parser.Statement) {
	switch typed := stmt.(type) {
	case parser.Block:
		interpretBlock(env, typed)
	case parser.VarDeclaration:
		identifier := typed.Identifier
		if env.Exists(identifier) {
			panic(fmt.Sprintf("variable %v already exist", identifier))
		}
		value := evaluate(env, typed.Expr)
		env.New(identifier, value)
	case parser.VarAssignment:
		identifier := typed.Identifier
		if !env.Exists(identifier) {
			panic(fmt.Sprintf("variable %v does not exist", identifier))
		}
		value := evaluate(env, typed.Expr)
		env.Set(identifier, value)
	case parser.ExprStatement:
		evaluate(env, typed.Expr)
	case parser.PrintStatement:
		result := evaluate(env, typed.Expr)
		fmt.Println(result)
	case parser.IfStatement:
		interpretIf(env, typed)
	case parser.WhileStatement:
		interpretWhile(env, typed)
	}
}

func interpretBlock(parentEnv *Environment, block parser.Block) {
	env := NewEnvironment(parentEnv)
	for _, stmt := range block.Stmts {
		interpretStmt(env, stmt)
	}
}

func interpretWhile(env *Environment, whileStmt parser.WhileStatement) {
	for evaluateCondition(env, whileStmt.Condition) {
		interpretStmt(env, whileStmt.Block)
	}
}

func interpretIf(env *Environment, ifStmt parser.IfStatement) {
	condition := evaluateCondition(env, ifStmt.Condition)
	if condition {
		interpretStmt(env, ifStmt.Block)
	}
	if !condition && ifStmt.ElseBlock != nil {
		interpretStmt(env, ifStmt.ElseBlock)
	}
}

func evaluateCondition(env *Environment, condition parser.Expression) bool {
	conditionValue := evaluate(env, condition)
	cond, ok := conditionValue.(bool)
	if !ok {
		panic(fmt.Sprintf("type mismatch: expected <boolean>, got %t", conditionValue))
	}
	return cond
}

func evaluate(env *Environment, ast parser.Expression) any {
	switch typed := ast.(type) {
	case parser.BinaryExpression:
		return evaluateBinaryOperation(evaluate(env, typed.Left), evaluate(env, typed.Right), typed.Operator)
	case parser.UnaryExpression:
		return evaluateUnaryOperation(evaluate(env, typed.Right), typed.Operator)
	case parser.BooleanLiteral:
		return typed.Value
	case parser.IntegerLiteral:
		return typed.Value
	case parser.IdentifierLiteral:
		if value, ok := env.Get(typed.Value); ok {
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

	case tokenizer.Ampersand:
		left, right := leftOperand.(int), rightOperand.(int)
		return left & right
	case tokenizer.Vbar:
		left, right := leftOperand.(int), rightOperand.(int)
		return left | right
	case tokenizer.Caret:
		left, right := leftOperand.(int), rightOperand.(int)
		return left ^ right
	
	case tokenizer.AmpersandAmpersand:
		left, right := leftOperand.(bool), rightOperand.(bool)
		return left && right
	case tokenizer.VbarVbar:
		left, right := leftOperand.(bool), rightOperand.(bool)
		return left || right
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
	env := NewEnvironment(nil)
	for _, decl := range prog {
		interpretStmt(env, decl)
	}
}