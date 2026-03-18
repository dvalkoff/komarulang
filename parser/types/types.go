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
	IntPointer
	BoolPointer
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
	case IntPointer:
		return "<*int>"
	case BoolPointer:
		return "<*bool>"
	}
	return ""
}

func FromToken(t token.Token, isPointer bool) (Type, error) {
	if !t.TokenType.Match(token.Type) {
		return NotSpecified, fmt.Errorf("Token %v is not a type",t.TokenType)
	}
	val := t.Value.(string)
	switch {
	case val =="bool" && !isPointer:
		return BoolType, nil
	case val == "int" && !isPointer:
		return IntType, nil
	case val =="bool" && isPointer:
		return BoolPointer, nil
	case val == "int" && isPointer:
		return IntPointer, nil
	}
	return NotSpecified, fmt.Errorf("Not allowed type %v", t.Value)
}

func ToPointer(t Type) (Type, error) {
	switch t {
	case IntType:
		return IntPointer, nil
	case BoolType:
		return BoolPointer, nil
	}
	return NotSpecified, fmt.Errorf("Not allowed type %v", t)
}

func FromPointer(t Type) (Type, error) {
	switch t {
	case IntPointer:
		return IntType, nil
	case BoolPointer:
		return BoolType, nil
	}
	return NotSpecified, fmt.Errorf("Not allowed type %v", t)
}

func Compatible(t1, t2 Type) bool {
	return t1 == t2
}

func CompatibleOperation(t1 Type, op token.TokenType) bool {
	return true // TODO: implement operation compatibility
}
