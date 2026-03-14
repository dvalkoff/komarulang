package parser

import (
	"fmt"

	token "github.com/dvalkoff/komarulang/tokenizer"
)


type ParserError struct {
	Expected token.TokenType
	Got token.TokenType
}

func (e ParserError) Error() string {
	return fmt.Sprintf("Expected %v  Got: %v", tokenToString(e.Expected), tokenToString(e.Got))
}

type TypeError struct {
	Expected token.VarType
	Got token.VarType
}

func (e TypeError) Error() string {
	return fmt.Sprintf("Expected %v  Got: %v", typeToString(e.Expected), typeToString(e.Got))
}

type NotCompatibleOperationError struct {
	Operation token.TokenType
	Type token.VarType
}

func (e NotCompatibleOperationError) Error() string {
	return fmt.Sprintf("Operation %v  does not support type %v", tokenToString(e.Operation), typeToString(e.Type))
}

func tokenToString(t token.TokenType) string {
	switch t {
	case token.Plus:
		return "+"
	case token.Minus:
		return "-"
	case token.Star:
		return "*"
	case token.Slash:
		return "/"
	case token.Percent:
		return "%"
	case token.LeftParen:
		return "("
	case token.RightParen:
		return ")"
	case token.LeftBrace:
		return "{"
	case token.RightBrace:
		return "}"
	case token.Semicolon:
		return ";"
	case token.Bang:
		return "!"
	case token.BangEqual:
		return "!="
	case token.Equal:
		return "="
	case token.EqualEqual:
		return "=="
	case token.Greater:
		return ">"
	case token.GreaterEqual:
		return ">="
	case token.Less:
		return "<"	
	case token.LessEqual:
		return "<="
	case token.Integer:
		return "<integer>"
	case token.Bool:
		return "<bool>"
	case token.Identifier:
		return "<identifier>"
	case token.Print:
		return "<print>"
	case token.Var:
		return "<var>"
	case token.EOF:
		return "<eof>"
	case token.EOL:
		return "<end of line>"
	case token.If:
		return "<if>"
	case token.Else:
		return "<else>"
	case token.For:
		return "<for>"
	case token.While:
		return "<for>"
	case token.Ampersand:
		return "&"
	case token.AmpersandAmpersand:
		return "&&"
	case token.Vbar:
		return "|"
	case token.VbarVbar:
		return "||"
	case token.Caret:
		return "^"
	case token.Type:
		return "<type>"
	}
	return ""
}


func typeToString(t token.VarType) string {
	switch t {
	case token.VoidType:
		return "<void>"
	case token.IntType:
		return "<int>"
	case token.BoolType:
		return "<bool>"
	case token.IdentifierType:
		return "<identifier>"
	}
	return ""
}