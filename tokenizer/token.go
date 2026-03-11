package tokenizer

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
	EOL
)

type TokenType int

type Token struct {
	TokenType TokenType
	Value any
}

func GetEOF() Token {
	return Token{TokenType: EOF, Value: nil}
}