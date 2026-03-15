package parser

import token "github.com/dvalkoff/komarulang/tokenizer"


func DefaultLiteral(t token.VarType) Expression {
	switch t {
	case token.IntType:
		return &IntegerLiteral{Value: 0}
	case token.BoolType:
		return &BooleanLiteral{Value: false}
	}
	return nil
}