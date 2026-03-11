package codegen

import (
	"fmt"

	"github.com/dvalkoff/komarulang/parser"
	"github.com/dvalkoff/komarulang/tokenizer/token"
)

type Register int

func (r Register) String() string {
	return fmt.Sprintf("x%d", int(r))
}

type CodeGenerator struct {}

func (c CodeGenerator) Generate(ast parser.Expression) string {
	as, _ := generateFromAST(ast, 0)
	return header() + as + exit()
}

func generateFromAST(ast parser.Expression, register Register) (string, Register) {
	switch typed := ast.(type) {
	case parser.BinaryExpression:
		mov1, register1 := generateFromAST(typed.Left, register)
		mov2, register2 := generateFromAST(typed.Right, register + 1)
		binaryOp, register := generateBinaryOperation(register1, register2, typed.Operator)
		return mov1 + mov2 + binaryOp, register
	case parser.IntegerLiteral:
		return generateIntegerLiteral(typed, register)
	default:
		panic(fmt.Sprintf("Unexpected expression occured %v", ast))
	}
}

func generateBinaryOperation(register1, register2 Register, operator token.TokenType) (string, Register) {
	switch operator {
	case token.Plus:
		return fmt.Sprintf("    add %v, %v, %v\n", register1, register1, register2), register1
	case token.Minus:
		return fmt.Sprintf("    sub %v, %v, %v\n", register1, register1, register2), register1
	default:
		panic(fmt.Sprintf("Unexpected operator occured %v", operator))
	}
}

func generateIntegerLiteral(intLiteral parser.IntegerLiteral, register Register) (string, Register) {
	mov := fmt.Sprintf("    mov %v, #%v\n", register, intLiteral.Value)
	return mov, register
}

func header() string {
	return `.global _main
.align 2

_main:
`
}

func exit() string {
	return `
	mov x16, #1
    svc #0x80
`
}