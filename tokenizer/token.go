package tokenizer

import "slices"

const (
	// operators
	Plus TokenType = iota
	Minus

	Star
	Slash
	Percent
	LeftParen
	RightParen
	LeftBrace
	RightBrace
	Semicolon
	Comma

	Ampersand
	AmpersandAmpersand
	Vbar
	VbarVbar
	Caret

	Bang
	BangEqual
	Equal
	EqualEqual
	Greater
	GreaterEqual
	Less
	LessEqual

	// literals
	Integer
	Bool
	Identifier
	Print // Temporary until gets replaced by stdlib

	// keywords
	Var
	If
	Else
	While
	For

	// functions
	Fun
	Return
	Break
	Continue

	// types
	Type

	EOF
	EOL
)

type TokenType int

func (t TokenType) Match(types ...TokenType) bool {
	return slices.Contains(types, t)
}

type Token struct {
	TokenType TokenType
	Value     any
}

func GetEOF() Token {
	return Token{TokenType: EOF, Value: nil}
}
