package main

import (
	"fmt"
	"os"

	"github.com/dvalkoff/komarulang/env"
	"github.com/dvalkoff/komarulang/parser"
	"github.com/dvalkoff/komarulang/tokenizer"
)

func interpretStmt(env *env.Environment[any], stmt parser.Statement) {
	switch typed := stmt.(type) {
	case *parser.Block:
		interpretBlock(env, typed)
	case *parser.VarDeclaration:
		identifier := typed.Identifier
		if env.Exists(identifier) {
			panic(fmt.Sprintf("variable %v already exist", identifier))
		}
		value := evaluate(env, typed.Expr)
		env.New(identifier, value)
	case *parser.VarAssignment:
		identifier := typed.Identifier
		if !env.Exists(identifier) {
			panic(fmt.Sprintf("variable %v does not exist", identifier))
		}
		value := evaluate(env, typed.Expr)
		env.Set(identifier, value)
	case *parser.ExprStatement:
		evaluate(env, typed.Expr)
	case *parser.PrintStatement:
		result := evaluate(env, typed.Expr)
		fmt.Println(result)
	case *parser.IfStatement:
		interpretIf(env, typed)
	case *parser.WhileStatement:
		interpretWhile(env, typed)
	case *parser.ForStatement:
		interpretFor(env, typed)
	}
}

func interpretBlock(parentEnv *env.Environment[any], block *parser.Block) {
	env := env.NewEnvironment[any](parentEnv)
	for _, stmt := range block.Stmts {
		interpretStmt(env, stmt)
	}
}

func interpretFor(parentEnv *env.Environment[any], forStatement *parser.ForStatement) {
	env := env.NewEnvironment[any](parentEnv)
	interpretStmt(env, forStatement.VarDecl)
	for evaluateCondition(env, forStatement.Condition) {
		interpretStmt(env, forStatement.Block)
		interpretStmt(env, forStatement.Increment)
	}
}

func interpretWhile(env *env.Environment[any], whileStmt *parser.WhileStatement) {
	for evaluateCondition(env, whileStmt.Condition) {
		interpretStmt(env, whileStmt.Block)
	}
}

func interpretIf(env *env.Environment[any], ifStmt *parser.IfStatement) {
	condition := evaluateCondition(env, ifStmt.Condition)
	if condition {
		interpretStmt(env, ifStmt.Block)
	}
	if !condition && ifStmt.ElseBlock != nil {
		interpretStmt(env, ifStmt.ElseBlock)
	}
}

func evaluateCondition(env *env.Environment[any], condition parser.Expression) bool {
	conditionValue := evaluate(env, condition)
	cond, ok := conditionValue.(bool)
	if !ok {
		panic(fmt.Sprintf("type mismatch: expected <boolean>, got %t", conditionValue))
	}
	return cond
}

func evaluate(env *env.Environment[any], ast parser.Expression) any {
	switch typed := ast.(type) {
	case *parser.BinaryExpression:
		return evaluateBinaryOperation(evaluate(env, typed.Left), evaluate(env, typed.Right), typed.Operator)
	case *parser.UnaryExpression:
		return evaluateUnaryOperation(evaluate(env, typed.Right), typed.Operator)
	case *parser.BooleanLiteral:
		return typed.Value
	case *parser.IntegerLiteral:
		return typed.Value
	case *parser.IdentifierLiteral:
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
	resolver := parser.TypeResolver{}
	err = resolver.Resolve(prog)
	if err != nil {
		panic(err)
	}
	
	env := env.NewEnvironment[any](nil)
	for _, decl := range prog {
		interpretStmt(env, decl)
	}
}