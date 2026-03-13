package tokenizer

import "slices"

const (
	// operators
	Plus = iota
	Minus

	Star
	Slash
	Percent
	LeftParen
	RightParen
	LeftBrace
	RightBrace
	Semicolon
	
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
	Print // Temporart until gets replaced by stdlib

	// keywords
	Var
	If
	Else
	While

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
