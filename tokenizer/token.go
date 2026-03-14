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
	For

	// types
	Type

	EOF
	EOL
)

type TokenType int

const (
	NotSpecified VarType = VarType(-1)
	BoolType VarType = iota
	IntType
	VoidType
	IdentifierType
)

type VarType int

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
