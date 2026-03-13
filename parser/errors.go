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
	}
	return ""
}