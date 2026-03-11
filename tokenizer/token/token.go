package token

import (
	"fmt"
	"strconv"
)

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

	EOF
)

type TokenType int

type Token struct {
	TokenType TokenType
	Value any
}

func GetToken(val string) (Token, error) {
	switch val {
	case "+":
		return Token{TokenType: Plus, Value: nil}, nil
	case "-":
		return Token{TokenType: Minus, Value: nil}, nil
	case "/":
		return Token{TokenType: Slash, Value: nil}, nil
	case "*":
		return Token{TokenType: Star, Value: nil}, nil
	case "%":
		return Token{TokenType: Percent, Value: nil}, nil
	case "(":
		return Token{TokenType: LeftParen, Value: nil}, nil
	case ")":
		return Token{TokenType: RightParen, Value: nil}, nil
	case "{":
		return Token{TokenType: LeftBrace, Value: nil}, nil
	case "}":
		return Token{TokenType: RightBrace, Value: nil}, nil
	case "!":
		return Token{TokenType: Bang, Value: nil}, nil
	case "!=":
		return Token{TokenType: BangEqual, Value: nil}, nil
	case "=":
		return Token{TokenType: Equal, Value: nil}, nil
	case "==":
		return Token{TokenType: EqualEqual, Value: nil}, nil
	case ">":
		return Token{TokenType: Greater, Value: nil}, nil
	case ">=":
		return Token{TokenType: GreaterEqual, Value: nil}, nil
	case "<":
		return Token{TokenType: Less, Value: nil}, nil
	case "<=":
		return Token{TokenType: LessEqual, Value: nil}, nil
	}
	if token, ok := tryInteger(val); ok {
		return token, nil
	}
	if token, ok := tryBool(val); ok {
		return token, nil
	}
	return Token{}, fmt.Errorf("Unrecongized token %v", val)
}

func tryInteger(val string) (Token, bool) {
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return Token{}, false
	}
	return Token{TokenType: Integer, Value: intVal}, true
}

func tryBool(val string) (Token, bool) {
	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		return Token{}, false
	}
	return Token{TokenType: Bool, Value: boolVal}, true
}

func GetEOF() Token {
	return Token{TokenType: EOF, Value: nil}
}