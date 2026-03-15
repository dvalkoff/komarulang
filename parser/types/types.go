package types

import (
	"fmt"

	token "github.com/dvalkoff/komarulang/tokenizer"
)

const (
	NotSpecified Type = iota
	BoolType
	IntType
	IdentifierType
	VoidType
)

type Type int

func (t Type) TypeToString() string {
	switch t {
	case IntType:
		return "<int>"
	case BoolType:
		return "<bool>"
	case IdentifierType:
		return "<identifier>"
	case VoidType:
		return "<void>"
	}
	return ""
}

func FromToken(t token.Token) (Type, error) {
	if !t.TokenType.Match(token.Type) {
		return NotSpecified, fmt.Errorf("Token %v is not a type",t.TokenType)
	}
	val := t.Value.(string)
	switch val {
	case "bool":
		return BoolType, nil
	case "int":
		return IntType, nil
	}
	return NotSpecified, fmt.Errorf("Not allowed type %v", t.Value)
}

func Compatible(t1, t2 Type) bool {
	return t1 == t2
}

func CompatibleOperation(t1 Type, op token.TokenType) bool {
	return true // TODO: implement operation compatibility
}
