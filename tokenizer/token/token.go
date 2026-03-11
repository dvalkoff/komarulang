package token

import (
	"fmt"
	"strconv"
)

const (
	// operators
	Plus = iota
	Minus

	// literals
	Integer

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
	}
	if token, ok := tryInteger(val); ok {
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

func GetEOF() Token {
	return Token{TokenType: EOF, Value: nil}
}